package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
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

var (
	// cpuMap is a map stores importance factor and average CPU usage.
	cpuMap = make(map[string]float64)
	// memoryMap is a map stores importance factor and average memory usage.
	memoryMap     = make(map[string]float64)
	clientSet     *kubernetes.Clientset
	policy        *Policy
	currNamespace string
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

	if currReplicaNum == num {
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

// ChangeReplicaPolicy changes the number of replica-sets of a certain deployment.
func ChangeReplicaPolicy(clientSet *kubernetes.Clientset, policy *Policy) {
	deployment, err := clientSet.AppsV1().Deployments(currNamespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	d1 := deployment.Items[0]
	d2 := deployment.Items[1]
	d3 := deployment.Items[2]
	fmt.Printf("Number of replica-set deployed currently : %d %d %d\n ", *(d1.Spec.Replicas), *(d2.Spec.Replicas), *(d3.Spec.Replicas))

	high := policy.Factor[policy.Status]["High"]
	medium := policy.Factor[policy.Status]["Medium"]
	low := policy.Factor[policy.Status]["Low"]
	fmt.Printf("Changing the # of replica-set to top/medium/low %d %d %d\n", high, medium, low)
	*(d1.Spec.Replicas) = high
	*(d2.Spec.Replicas) = medium
	*(d3.Spec.Replicas) = low
	_, _ = clientSet.AppsV1().Deployments(currNamespace).Update(context.TODO(), &d1, metav1.UpdateOptions{})
	_, _ = clientSet.AppsV1().Deployments(currNamespace).Update(context.TODO(), &d2, metav1.UpdateOptions{})
	_, _ = clientSet.AppsV1().Deployments(currNamespace).Update(context.TODO(), &d3, metav1.UpdateOptions{})
	fmt.Printf("Number of replica-set deployed after change : %d %d %d\n", high, medium, low)

}

// printCurrPodUsage loops through CPU map and memory map to print.
func printCurrPodUsage() {
	// cpuMap["1"] is cpuMap which key is imp "1"  memoryMap["1"] is memoryMap which key is imp "1"
	for key := range cpuMap {
		fmt.Printf("Importance Factor %s) \tAverage of CPU: \t%.2f\n\t\t\tAverage of Memory: \t%.2f\n", key, cpuMap[key], memoryMap[key])
	}
}

// autoAdjustReplica changes the replica-set num based on CPU and memory usage.
func autoAdjustReplica(clientSet *kubernetes.Clientset, namespace string) (factor map[string]map[string]int32) {
	fmt.Printf("\n--------------------- [Change of Relica-set] ---------------------\n")
	fmt.Println("Tong", cpuMap["1"], cpuMap["2"], cpuMap["3"])
	if cpuMap["1"] > 1000000 {
		// if too high, subtract down
		fmt.Printf("Subtracting...\n")
		if policy.Status == "Green" {
			changeReplica(clientSet, namespace, 3, 3, "subtract")
			changeReplica(clientSet, namespace, 2, 3, "subtract")
			changeReplica(clientSet, namespace, 1, 10, "subtract")
		} else if policy.Status == "Yellow" {
			changeReplica(clientSet, namespace, 3, 2, "subtract")
			changeReplica(clientSet, namespace, 2, 2, "subtract")
			changeReplica(clientSet, namespace, 1, 8, "subtract")
		} else {
			changeReplica(clientSet, namespace, 3, 1, "subtract")
			changeReplica(clientSet, namespace, 2, 1, "subtract")
			changeReplica(clientSet, namespace, 1, 3, "subtract")
		}
		green := map[string]int32{"High": 10, "Medium": 3, "Low": 3}
		yellow := map[string]int32{"High": 8, "Medium": 2, "Low": 2}
		red := map[string]int32{"High": 3, "Medium": 1, "Low": 1}
		factor = map[string]map[string]int32{"Green": green, "Yellow": yellow, "Red": red}
	} else if cpuMap["1"] < 100 {
		// if too low, scale up
		fmt.Printf("Adding...\n")
		if policy.Status == "Green" {
			changeReplica(clientSet, namespace, 3, 10, "add")
			changeReplica(clientSet, namespace, 2, 10, "add")
			changeReplica(clientSet, namespace, 1, 10, "add")
		} else if policy.Status == "Yellow" {
			changeReplica(clientSet, namespace, 3, 8, "add")
			changeReplica(clientSet, namespace, 2, 8, "add")
			changeReplica(clientSet, namespace, 1, 8, "add")
		} else {
			changeReplica(clientSet, namespace, 3, 3, "add")
			changeReplica(clientSet, namespace, 2, 3, "add")
			changeReplica(clientSet, namespace, 1, 3, "add")
		}
		green := map[string]int32{"High": 10, "Medium": 10, "Low": 10}
		yellow := map[string]int32{"High": 8, "Medium": 8, "Low": 8}
		red := map[string]int32{"High": 3, "Medium": 3, "Low": 3}
		factor = map[string]map[string]int32{"Green": green, "Yellow": yellow, "Red": red}
	} else {
		if policy.Status == "Green" {
			changeReplica(clientSet, namespace, 3, 6, "add")
			changeReplica(clientSet, namespace, 2, 6, "add")
			changeReplica(clientSet, namespace, 1, 10, "add")
		} else if policy.Status == "Yellow" {
			changeReplica(clientSet, namespace, 3, 4, "add")
			changeReplica(clientSet, namespace, 2, 4, "add")
			changeReplica(clientSet, namespace, 1, 8, "add")
		} else {
			changeReplica(clientSet, namespace, 3, 2, "add")
			changeReplica(clientSet, namespace, 2, 2, "add")
			changeReplica(clientSet, namespace, 1, 3, "add")
		}
		green := map[string]int32{"High": 10, "Medium": 6, "Low": 6}
		yellow := map[string]int32{"High": 8, "Medium": 4, "Low": 4}
		red := map[string]int32{"High": 3, "Medium": 2, "Low": 2}
		factor = map[string]map[string]int32{"Green": green, "Yellow": yellow, "Red": red}
	}
	return factor
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
		} else {
			cpuMap[currPodImp] = 0
			memoryMap[currPodImp] = 0
		}
	}
	return err
}

// ----------------- policy -----------------

type Policy struct {
	Status string
	Factor map[string]map[string]int32
}

// backend handles policy & importance factor
func Backend(w http.ResponseWriter, req *http.Request) {
	// Setup response
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS, PUT")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

	if req.Method == "OPTIONS" {
		return
	}

	if req.Method == "GET" {
		log.Println("func", "ServeHTTP", "Handling GET request /policy")

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(policy)

		log.Println("func", "ServeHTTP", "Handled GET request for /policy")
		return
	}

	if req.Method == "PUT" {
		log.Println("func", "ServeHTTP", "Handling PUT request")
		w.WriteHeader(http.StatusNoContent)

		log.Println("func", "NewPolicyHandler", "Decoding request body")
		decoder := json.NewDecoder(req.Body)
		var request Policy
		if err := decoder.Decode(&request); err != nil {
			err := errors.New("failed to decode to Policy")
			log.Println("func", "ServeHTTP", "decode policy from request err:", err)
		}

		if policy.Status == request.Status {
			log.Println("func", "ServeHTTP", "Same status")
			return
		}

		policy.Status = request.Status
		ChangeReplicaPolicy(clientSet, policy)
		w.WriteHeader(http.StatusNoContent)
		return
	}
}

func monitor() {
	for {
		//printDeploymentInfo(clientSet, currNamespace) //Print the usage of CPU memory in each imp
		// calculate the usage and print
		//fmt.Printf("\n----------------------------- [Usage] ----------------------------\n")
		//printCurrPodUsage()
		// calculate average usage and print again
		aveCurrPodUsage(clientSet, currNamespace)
		// change the replica-set num accordingly
		policy.Factor = autoAdjustReplica(clientSet, currNamespace)
		// calling Sleep method
		fmt.Println("\nWaiting for changing the relica-set number.....")
		time.Sleep(10 * time.Second)
		fmt.Println("Done!")

		//fmt.Printf("\n---------------------- [Usage after change] ----------------------\n")
		//printCurrPodUsage()
	}
}

// curl -X PUT -H "Content-Type: application/json" -d '{"Red": {"TOP": 1, "Medium": 1, "LOW": 0}, "Yellow": {"TOP": 2, "Medium": 2, "LOW": 0}, "Green": {"TOP": 4, "Medium": 4, "LOW": 0}}' http://localhost:8888/factor
func main() {
	filePath := os.Args[1]                        //Pass .pem file as a command line argument
	clusterIP := os.Args[2]                       //Pass cluster IP address
	currNamespace = os.Args[3]                    //Pass the namespace
	clientSet = authenticate(filePath, clusterIP) //Authenticates with the GCP cluster

	green := map[string]int32{"High": 10, "Medium": 10, "Low": 10}
	yellow := map[string]int32{"High": 8, "Medium": 8, "Low": 8}
	red := map[string]int32{"High": 3, "Medium": 3, "Low": 3}
	initFactor := map[string]map[string]int32{"Green": green, "Yellow": yellow, "Red": red}
	policy = &Policy{
		Status: "Green",
		Factor: initFactor,
	}

	go monitor()
	http.HandleFunc("/policy", Backend)
	http.ListenAndServe(":8888", nil)

}
