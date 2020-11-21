package main

import (
	"context"
	//"encoding/json"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
	"net/http"
	"sync"
)

// CPU_Map is a map stores importance factor and sumOfCpu.
var (
	cpuMap = make(map[string]int64)
)

// Memo_Map is a map stores importance factor and sumOfMemory.
var (
	memoryMap = make(map[string]int64)
)

func handler(w http.ResponseWriter, _ *http.Request) {
	cfg, err := rest.InClusterConfig()

	if err != nil {
		panic(err)
	}

	clientSet, err := kubernetes.NewForConfig(cfg)

	if err != nil {
		panic("Error creating Go client" + err.Error())
	}
	mc, err := metrics.NewForConfig(cfg)
	if err != nil {
		panic("Error creating Metrics client: " + err.Error())
	}

	namespace := "test"

	deployments, err := clientSet.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic("Error fetching Deployments: " + err.Error())
	}
	//Mutex locks prevents concurrent writes to the hashmap.
	var mutex = &sync.Mutex{}
	mutex.Lock()
	for i, dep := range deployments.Items {
		currLabel := dep.GetLabels()["app"]
		_, _ = fmt.Fprintf(w, "%d) Deployment: \t%s\n", i+1, dep.GetName())
		_, _ = fmt.Fprintf(w, "   Label: \t%s\n", dep.GetLabels()["app"])
		MapLabel := "app=" + currLabel
		pods, _ := clientSet.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: MapLabel})
		_, _ = fmt.Fprintf(w, "   Total Number of pods: %d\n", len(pods.Items))

		for j := range pods.Items {
			currPod := pods.Items[j]
			currPodImp := currPod.GetAnnotations()["imp"]
			currPodName := currPod.GetName()
			_, _ = fmt.Fprintf(w, "   Pod name: \t\t\t%s\n", currPodName)
			_, _ = fmt.Fprintf(w, "   Current imp: \t\t%s\n", currPodImp)
			metricsPod, _ := mc.MetricsV1beta1().PodMetricses(namespace).Get(context.TODO(), currPodName, metav1.GetOptions{})
			cpu, _ := metricsPod.Containers[0].Usage.Cpu().AsInt64()
			mem, _ := metricsPod.Containers[0].Usage.Memory().AsInt64()
			cpuMap[currPodImp] += cpu
			memoryMap[currPodImp] += mem
		}
	}
	for key := range cpuMap {
		_, _ = fmt.Fprintf(w, "\nImp %s:\tSum of CPU is %d\n\tSum of Memo is %d\n", key, cpuMap[key], memoryMap[key])
	}
	_, _ = fmt.Fprintf(w, "Add change replica-set logic here")
	//Clearing contents inside the maps
	cpuMap = make(map[string]int64)
	memoryMap = make(map[string]int64)
	mutex.Unlock()
}

func main() {
	http.HandleFunc("/", handler)
	fmt.Println("Web server is running on port 8080")
	http.ListenAndServe(":8080", nil)
}
