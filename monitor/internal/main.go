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

// CPU_Map is a map stores importance factor and sumOfCpu.
var (
	CPU_Map = make(map[string]float64)
)

// Memo_Map is a map stores importance factor and sumOfMemo.
var (
	Memo_Map = make(map[string]float64)
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

// printDeploymentInfo prints the cpu and memory sum of pods in each the importance factor.
func printDeploymentInfo(clientSet *kubernetes.Clientset, currNamespace string) {
	namespace := currNamespace
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
		fmt.Printf("   Total Number of pods: %d\n", len(pods.Items))

		// loop through the pods
		for j := range pods.Items {
			currPod := pods.Items[j]
			currPodImp := currPod.GetAnnotations()["imp"]
			currPodName := currPod.GetName()
			// calculate the sum of cpu and memory and save to map
			err = printSumPodUsage(currPodName, currPodImp, clientSet, namespace)
			if err != nil {
				panic(err.Error())
			}
		}
	}
}

// printSumPodUsage print single pod info, then sums the cpu and memory usage of pods in the same importance factor.
func printSumPodUsage(podName string, imp string, clientSet *kubernetes.Clientset, currNamespace string) error {
	absPath := "apis/metrics.k8s.io/v1beta1/namespaces/" + currNamespace + "/pods/" + podName
	data, err := clientSet.RESTClient().Get().AbsPath(absPath).DoRaw(context.TODO())
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, &singlePodObj)
	if err != nil {
		panic(err.Error())
	}

	tempMemoString := strings.TrimRight(singlePodObj.Containers[0].Usage.Memory, "Ki")
	tempCPUString := strings.TrimRight(singlePodObj.Containers[0].Usage.CPU, "n")
	currentCPUUsage, _ := strconv.ParseFloat(tempCPUString, 2)
	currentMemoUsage, _ := strconv.ParseFloat(tempMemoString, 2)

	fmt.Printf("   Pod name: \t\t\t%s\n", podName)
	fmt.Printf("   Current imp: \t\t%s\n", imp)
	fmt.Printf("   Current CPU usage: \t\t%.2f\n", currentCPUUsage)
	fmt.Printf("   Current Memory usage: \t%.2f\n", currentMemoUsage)
	//sum the CPU & memory usage in the same imp
	CPU_Map[imp] += currentCPUUsage
	Memo_Map[imp] += currentMemoUsage
	return err
}

// changeReplica changes the number of replica-sets of a certain deployment.
func changeReplica(clientSet *kubernetes.Clientset, currNamespace string, imp int32, threshold int32) {
	namespace := currNamespace
	deployment, err := clientSet.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}

	nginx := deployment.Items[imp-1]
	currReplicaNum := *(nginx.Spec.Replicas)
	fmt.Printf("Importance Factor %d) \tPrevious number of replica-set deployed: %d\n", imp, currReplicaNum)
	var newReplicaNum int32
	// if current replica-set num is not less threshold, subtract the num from threshold
	if threshold == 0 || currReplicaNum == 0 {
		fmt.Printf("\t\t\tNothing need to be changed.\n")
		return
	} else if currReplicaNum >= threshold {
		newReplicaNum = currReplicaNum - threshold
		fmt.Printf("\t\t\tOff %d replica-set.\n", threshold)
	} else {
		fmt.Printf("\t\t\tOff %d replica-set.\n", currReplicaNum)
		newReplicaNum = 0
	}

	// if currReplicaNum > threshold {
	// 	newReplicaNum = threshold
	// } else {
	// 	fmt.Printf("\t\t\tNothing need to be changed.\n")
	// 	return
	// }
	// fmt.Printf("\t\t\tChanging the number of replica-set to: \t\t\t%d\n", newReplicaNum)

	// update the replica-set number
	*(nginx.Spec.Replicas) = newReplicaNum
	_, _ = clientSet.AppsV1().Deployments(namespace).Update(context.TODO(), &nginx, metav1.UpdateOptions{})
	fmt.Printf("\t\t\tCurrent number replica-set after change: %d\n", *(nginx.Spec.Replicas))
}

// printCurrPodUsage loops through CPU map and memory map to print.
func printCurrPodUsage() {
	// CPU_Map["1"] is CPU_map which key is imp "1"  Memo_Map["1"] is Memo_Map which key is imp "1"
	for key := range CPU_Map {
		fmt.Printf("Importance Factor %s) \tSum of CPU: \t%.2f\n\t\t\tSum of Memo: \t%.2f\n", key, CPU_Map[key], Memo_Map[key])
	}
}

// audoAdjustReplica changes the replica-set num based on CPU and memory usage.
func audoAdjustReplica(clientSet *kubernetes.Clientset, currNamespace string) {
	fmt.Printf("\n--------------------- [Change of Relica-set] ---------------------\n")
	if Memo_Map["1"] > 20000 || Memo_Map["2"] > 22000 || Memo_Map["3"] > 24000 {
		//deduct 8*2500=20000 20000-->0
		changeReplica(clientSet, currNamespace, 3, 5)
		changeReplica(clientSet, currNamespace, 2, 2)
		changeReplica(clientSet, currNamespace, 1, 1)
	} else if Memo_Map["1"] > 18000 || Memo_Map["2"] > 20000 || Memo_Map["3"] > 22000 {
		//deduct 7*2500=17500 18000-->0
		changeReplica(clientSet, currNamespace, 3, 5)
		changeReplica(clientSet, currNamespace, 2, 2)
		changeReplica(clientSet, currNamespace, 1, 0)
	} else if Memo_Map["1"] > 15000 || Memo_Map["2"] > 17000 || Memo_Map["3"] > 19000 {
		//deduct 6*2500=1500 15000-->0
		changeReplica(clientSet, currNamespace, 3, 5)
		changeReplica(clientSet, currNamespace, 2, 1)
		changeReplica(clientSet, currNamespace, 1, 0)
	} else if Memo_Map["1"] > 10000 || Memo_Map["2"] > 12000 || Memo_Map["3"] > 14000 {
		//deduct 4*2500=10000 10000-->0
		changeReplica(clientSet, currNamespace, 3, 4)
		changeReplica(clientSet, currNamespace, 2, 0)
		changeReplica(clientSet, currNamespace, 1, 0)
	}
}

// sumCurrPodUsage sums the cpu and memory usage of pods in the same importance factor.
func sumCurrPodUsage(clientSet *kubernetes.Clientset, currNamespace string) error {
	namespace := currNamespace
	deployments, err := clientSet.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}

	// clear map before recalculate
	for key, _ := range CPU_Map {
		CPU_Map[key] = 0
		Memo_Map[key] = 0
	}
	// loop through all deployments
	for _, dep := range deployments.Items {
		currLabel := dep.GetLabels()["app"]
		MapLabel := "app=" + currLabel
		pods, _ := clientSet.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: MapLabel})
		// loop through the pods
		for j := range pods.Items {
			currPod := pods.Items[j]
			currPodImp := currPod.GetAnnotations()["imp"]
			currPodName := currPod.GetName()
			// calculate the sum of cpu and memory and save to map
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
			tempMemoString := strings.TrimRight(singlePodObj.Containers[0].Usage.Memory, "Ki")
			tempCPUString := strings.TrimRight(singlePodObj.Containers[0].Usage.CPU, "n")
			currentCPUUsage, _ := strconv.ParseFloat(tempCPUString, 2)
			currentMemoUsage, _ := strconv.ParseFloat(tempMemoString, 2)
			//sum the CPU & memory usage in the same imp
			CPU_Map[currPodImp] += currentCPUUsage
			Memo_Map[currPodImp] += currentMemoUsage
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

	fmt.Printf("\n------------------------- [Sum of Usage] -------------------------\n")
	printCurrPodUsage()

	// change the replica-set num accordingly
	audoAdjustReplica(clientSet, currNamespace)

	// calling Sleep method
	fmt.Println("\nWaiting for changing the relica-set number.....")
	time.Sleep(20 * time.Second)
	fmt.Println("Done!")

	sumCurrPodUsage(clientSet, currNamespace)
	fmt.Printf("\n--------------------- [Current Sum of Usage] ---------------------\n")
	printCurrPodUsage()

}
