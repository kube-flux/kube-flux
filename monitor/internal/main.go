package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	clientcmd "k8s.io/client-go/tools/clientcmd/api"
)

// cpuMap is a map stores importance factor and average CPU usage.
var (
	cpuMap = make(map[string]float64)
)

// memoryMap is a map stores importance factor and average memory usage.
var (
	memoryMap = make(map[string]float64)
)

// PodMetric stores the JSON Array of one single pod information.
type PodMetric struct {
	Kind       string `json:"kind"`
	APIVersion string `json:"apiVersion"`
	Metadata   struct {
		Name              string    `json:"name"`
		Namespace         string    `json:"namespace"`
		SelfLink          string    `json:"selfLink"`
		Imp               string    `json:"imp"`
		CreationTimestamp time.Time `json:"creationTimestamp"`
	} `json:"metadata"`
	Timestamp  time.Time `json:"timestamp"`
	Window     string    `json:"window"`
	Containers []struct {
		Name  string `json:"name"`
		Usage struct {
			CPU    string `json:"cpu"`
			Memory string `json:"memory"`
		} `json:"usage"`
	} `json:"containers"`
}

var singlePodObj PodMetric // single pod structure object

// authenticate is used to authenticate Go-client with GKE cluster.
func authenticate(filePath string, HostIp string) *kubernetes.Clientset {
	MasterUrl := "https://" + HostIp
	ca, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(err)
	}
	config := &rest.Config{
		TLSClientConfig: rest.TLSClientConfig{
			CAData: ca,
		},
		Host:         MasterUrl,
		AuthProvider: &clientcmd.AuthProviderConfig{Name: "gcp"}}
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic("Failed to authenticate IP: " + HostIp)
	}
	return clientSet
}

// printDeploymentInfo prints the cpu and memory average of pods in each the importance factor.
func printDeploymentInfo(clientSet *kubernetes.Clientset, namespace string) {
	deployments, err := clientSet.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	// loop through all deployments
	fmt.Printf("\n---------------------- [List of Deployment] ----------------------\n")
	for i, dep := range deployments.Items {
		currLabel := dep.GetLabels()["app"]
		fmt.Printf("%d) Deployment: \t%s\n", i+1, dep.GetName())
		fmt.Printf("   Label: \t%s\n", dep.GetLabels()["app"])
		MapLabel := "app=" + currLabel
		pods, _ := clientSet.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: MapLabel})
		numOfPods := len(pods.Items)
		fmt.Printf("   Total Number of pods: %d\n", numOfPods)
		currPodImp := ""
		// loop through the pods
		for j := range pods.Items {
			currPod := pods.Items[j]
			currPodImp = currPod.GetAnnotations()["imp"]
			currPodName := currPod.GetName()
			// calculate the sum of cpu and memory and save to map
			err = sumPodUsage(currPodName, currPodImp, clientSet, namespace)
			if err != nil {
				panic(err.Error())
			}
		}
		// get the ave of the pods in each deployment
		if numOfPods != 0 {
			cpuMap[currPodImp] = cpuMap[currPodImp] / float64(numOfPods)
			memoryMap[currPodImp] = memoryMap[currPodImp] / float64(numOfPods)
		}
	}
}

// sumPodUsage print single pod info, then sums the cpu and memory usage of pods in the same importance factor.
func sumPodUsage(podName string, imp string, clientSet *kubernetes.Clientset, namespace string) error {
	absPath := "apis/metrics.k8s.io/v1beta1/namespaces/" + namespace + "/pods/" + podName
	data, err := clientSet.RESTClient().Get().AbsPath(absPath).DoRaw(context.TODO())
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, &singlePodObj)
	if err != nil {
		panic(err.Error())
	}

	tempMemoryString := strings.TrimRight(singlePodObj.Containers[0].Usage.Memory, "Ki")
	tempCPUString := strings.TrimRight(singlePodObj.Containers[0].Usage.CPU, "n")
	currentCPUUsage, _ := strconv.ParseFloat(tempCPUString, 2)
	currentMemoryUsage, _ := strconv.ParseFloat(tempMemoryString, 2)

	fmt.Printf("   Pod name: \t\t\t%s\n", podName)
	fmt.Printf("   Current imp: \t\t%s\n", imp)
	fmt.Printf("   Current CPU usage: \t\t%.2f\n", currentCPUUsage)
	fmt.Printf("   Current Memory usage: \t%.2f\n", currentMemoryUsage)
	//sum the CPU & memory usage in the same imp
	cpuMap[imp] += currentCPUUsage
	memoryMap[imp] += currentMemoryUsage
	return err
}

// changeReplica changes the number of replica-sets of a certain deployment.
func changeReplica(clientSet *kubernetes.Clientset, namespace string, imp int32, num int32, action string) {
	deployment, err := clientSet.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}

	nginx := deployment.Items[imp-1]
	currReplicaNum := *(nginx.Spec.Replicas)
	fmt.Printf("Importance Factor %d) \tPrevious number of replica-set deployed: %d\n", imp, currReplicaNum)
	var newReplicaNum int32

	if currReplicaNum == 0 || currReplicaNum == num {
		fmt.Printf("\t\t\tNothing need to be changed.\n")
		return
	}
	if currReplicaNum > num && action == "subtract" {
		newReplicaNum = num
		fmt.Printf("\t\t\tOff %d replica-set.\n", currReplicaNum-num)
	} else if currReplicaNum < num && action == "add" {
		newReplicaNum = num
		fmt.Printf("\t\t\tOn %d replica-set.\n", num-currReplicaNum)
	} else {
		fmt.Printf("\t\t\tNothing need to be changed.\n")
		return
	}
	// update the replica-set number
	*(nginx.Spec.Replicas) = newReplicaNum
	_, _ = clientSet.AppsV1().Deployments(namespace).Update(context.TODO(), &nginx, metav1.UpdateOptions{})
	fmt.Printf("\t\t\tCurrent number replica-set after change: %d\n", *(nginx.Spec.Replicas))
}

// printCurrPodUsage loops through CPU map and memory map to print.
func printCurrPodUsage() {
	// cpuMap["1"] is cpuMap which key is imp "1"  memoryMap["1"] is memoryMap which key is imp "1"
	for key := range cpuMap {
		fmt.Printf("Importance Factor %s) \tAverage of CPU: \t%.2f\n\t\t\tAverage of Memory: \t%.2f\n", key, cpuMap[key], memoryMap[key])
	}
}

// autoAdjustReplica changes the replica-set num based on CPU and memory usage.
func autoAdjustReplica(clientSet *kubernetes.Clientset, namespace string) {
	fmt.Printf("\n--------------------- [Change of Relica-set] ---------------------\n")

	if cpuMap["1"] > 1000 || cpuMap["2"] > 1000 || cpuMap["3"] > 1000 {
		// if too high, subtract down
		fmt.Printf("Subtracting...\n")
		changeReplica(clientSet, namespace, 3, 3, "subtract")
		changeReplica(clientSet, namespace, 2, 8, "subtract")
		changeReplica(clientSet, namespace, 1, 10, "subtract")
	} else if cpuMap["1"] < 800 || cpuMap["2"] < 800 || cpuMap["3"] < 800 {
		// if too low, scale up
		fmt.Printf("Adding...\n")
		changeReplica(clientSet, namespace, 3, 10, "add")
		changeReplica(clientSet, namespace, 2, 10, "add")
		changeReplica(clientSet, namespace, 1, 10, "add")
	} else {
		fmt.Printf("Nothing need to be changed.\n")
	}

	return
}

// aveCurrPodUsage calculate the average cpu and memory usage of pods in the same importance factor.
func aveCurrPodUsage(clientSet *kubernetes.Clientset, namespace string) error {
	deployments, err := clientSet.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}

	// clear map before recalculate
	for key, _ := range cpuMap {
		cpuMap[key] = 0
		memoryMap[key] = 0
	}
	// loop through all deployments
	for _, dep := range deployments.Items {
		currLabel := dep.GetLabels()["app"]
		MapLabel := "app=" + currLabel
		pods, _ := clientSet.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: MapLabel})
		numOfPods := len(pods.Items)
		currPodImp := ""
		// loop through the pods
		for j := range pods.Items {
			currPod := pods.Items[j]
			currPodImp = currPod.GetAnnotations()["imp"]
			currPodName := currPod.GetName()
			// calculate the average of cpu and memory and save to map
			absPath := "apis/metrics.k8s.io/v1beta1/namespaces/" + namespace + "/pods/" + currPodName
			data, err := clientSet.RESTClient().Get().AbsPath(absPath).DoRaw(context.TODO())
			if err != nil {
				return err
			}
			err = json.Unmarshal(data, &singlePodObj)
			if err != nil {
				panic(err.Error())
			}
			if err != nil {
				panic(err.Error())
			}
			tempMemoryString := strings.TrimRight(singlePodObj.Containers[0].Usage.Memory, "Ki")
			tempCPUString := strings.TrimRight(singlePodObj.Containers[0].Usage.CPU, "n")
			currentCPUUsage, _ := strconv.ParseFloat(tempCPUString, 2)
			currentMemoryUsage, _ := strconv.ParseFloat(tempMemoryString, 2)
			//sum the CPU & memory usage in the same imp
			cpuMap[currPodImp] += currentCPUUsage
			memoryMap[currPodImp] += currentMemoryUsage
		}
		if numOfPods != 0 {
			cpuMap[currPodImp] = cpuMap[currPodImp] / float64(numOfPods)
			memoryMap[currPodImp] = memoryMap[currPodImp] / float64(numOfPods)
		}
	}
	return err
}

func main() {
	filePath := os.Args[1]                         //Pass .pem file as a command line argument
	clusterIP := os.Args[2]                        //Pass cluster IP address
	currNamespace := os.Args[3]                    //Pass the namespace
	clientSet := authenticate(filePath, clusterIP) //Authenticates with the GCP cluster
	printDeploymentInfo(clientSet, currNamespace)  //Print the usage of CPU memory in each imp
	// calculate the usage and print
	fmt.Printf("\n----------------------------- [Usage] ----------------------------\n")
	printCurrPodUsage()
	// change the replica-set num accordingly
	autoAdjustReplica(clientSet, currNamespace)
	// calling Sleep method
	fmt.Println("\nWaiting for changing the relica-set number.....")
	time.Sleep(15 * time.Second)
	fmt.Println("Done!")
	// calculate average usage and print again
	aveCurrPodUsage(clientSet, currNamespace)
	fmt.Printf("\n---------------------- [Usage after change] ----------------------\n")
	printCurrPodUsage()
}
