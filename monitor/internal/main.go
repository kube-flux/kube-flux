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

//printPodUsageImp prints the cpu and memory sum of pods in each the importance factor.
func printPodUsageImp(clientSet *kubernetes.Clientset, currNamespace string) {
	namespace := currNamespace
	deployments, err := clientSet.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	// loop through all deployments
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
			fmt.Printf("   Pod name: \t\t\t%s\n", currPodName)
			fmt.Printf("   Current imp: \t\t%s\n", currPodImp)
			// calculate the sum of cpu and memory
			fmt.Printf("Pod Name: " + currPodName)
			err = getPodUsage(currPodName, currPodImp, clientSet, namespace)
			if err != nil {
				panic(err.Error())
			}
		}
		fmt.Println()
	}
	// TODO: CPU_Map["1"] is CPU_map which key is imp "1"  Memo_Map["1"] is Memo_Map which key is imp "1"
	for key := range CPU_Map {
		fmt.Printf("Imp %s:\tSum of CPU is %.2f\n\tSum of Memo is %.2f\n", key, CPU_Map[key], Memo_Map[key])
	}
}

func getPodUsage(podName string, imp string, clientSet *kubernetes.Clientset, currNamespace string) error {
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

	fmt.Printf("   Current CPU usage: \t\t%.2f\n", currentCPUUsage)
	fmt.Printf("   Current Memory usage: \t%.2f\n", currentMemoUsage)
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
	// TODO: hard coded deployment.Items[imp]
	nginx := deployment.Items[imp-1]
	print(*(nginx.Spec.Replicas))
	currReplicaNum := *(nginx.Spec.Replicas)
	fmt.Printf("Previous number of replica-set deployed in imp %d: %d\n", imp, currReplicaNum)
	var newReplicaNum int32
	if currReplicaNum > threshold {
		newReplicaNum = threshold
	} else {
		fmt.Printf("Nothing need to be changed.\n")
		return
	}
	fmt.Printf("Changing the number of replica-set to %d\n", newReplicaNum)
	// update the replica-set number
	*(nginx.Spec.Replicas) = newReplicaNum
	_, _ = clientSet.AppsV1().Deployments(namespace).Update(context.TODO(), &nginx, metav1.UpdateOptions{})
	fmt.Printf("Current number replica-set deployed after change : %d\n", *(nginx.Spec.Replicas))
}

func main() {
	filePath := os.Args[1]                         //Pass .pem file as a command line argument
	clusterIP := os.Args[2]                        //Pass cluster IP address
	currNamespace := os.Args[3]                    //Pass the namespace
	clientSet := authenticate(filePath, clusterIP) //Authenticates with the GCP cluster
	printPodUsageImp(clientSet, currNamespace)     //Print the usage of CPU memory in each imp
	// change the replica-set num accordingly
	if Memo_Map["1"] > 10000 || Memo_Map["2"] > 8000 || Memo_Map["3"] > 6000 {
		//change imp3 3-->0
		//change imp2 5-->3
		//change imp1 6-->4
		changeReplica(clientSet, currNamespace, 3, 0)
		changeReplica(clientSet, currNamespace, 2, 3)
		changeReplica(clientSet, currNamespace, 1, 4)
	} else if Memo_Map["1"] > 80000 || Memo_Map["2"] > 6000 || Memo_Map["3"] > 4000 {
		//change imp3 3-->0
		//change imp2 5-->3
		changeReplica(clientSet, currNamespace, 3, 0)
		changeReplica(clientSet, currNamespace, 2, 3)
	} else if Memo_Map["1"] > 6000 || Memo_Map["2"] > 4000 || Memo_Map["3"] > 2000 {
		//change imp3 3-->0
		changeReplica(clientSet, currNamespace, 3, 0)
	}
}
