package main

import (
	"context"
	"encoding/json"
	"fmt"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// PodMetricsList stores the JSON Array of pod information.
type PodMetricsList struct {
	Kind       string `json:"kind"`
	APIVersion string `json:"apiVersion"`
	Metadata   struct {
		SelfLink string `json:"selfLink"`
	} `json:"metadata"`
	Items []struct {
		Metadata struct {
			Name              string    `json:"name"`
			Namespace         string    `json:"namespace"`
			SelfLink          string    `json:"selfLink"`
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
	} `json:"items"`
}

var pods PodMetricsList // pods structure object

// getMetrics fetches the pod information by making a rest API call to the metrics server.
// It receives a JSON array in bytes.
// The JSON array is then converted to a structure which is pointed to by an object.
func getMetrics(clientSet *kubernetes.Clientset, pods *PodMetricsList) error {
	data, err := clientSet.RESTClient().Get().AbsPath("apis/metrics.k8s.io/v1beta1/pods").DoRaw(context.TODO())
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, &pods)
	return err
}

//changeImportanceFactor changes the importance factor of a replica-set.
//replica-sets are first deployed with importance factor under annotations meta-data.
//for every deployment, it finds it's replica-set via label and updates the importance factor stored under annotations.
func changeImportanceFactor(clientSet *kubernetes.Clientset) {
	namespace := apiv1.NamespaceDefault
	deployments, err := clientSet.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	for i, dep := range deployments.Items {
		currLabel := dep.GetLabels()["app"]
		//TODO: remove the random number generator & label check
		if currLabel == "ngnix" {
			fmt.Printf("%d) Deployment: %s\n", i+1, dep.GetName())
			fmt.Printf("Label is:%s\n", dep.GetLabels()["app"])
			MapLabel := "app=" + currLabel
			pods, _ := clientSet.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: MapLabel})
			rand.Seed(time.Now().UnixNano())
			importanceNumber := strconv.FormatInt(int64(rand.Intn(5)+1), 10)
			impMap := make(map[string]string)
			impMap["imp"] = importanceNumber
			for j := range pods.Items {
				currPod := pods.Items[j]
				fmt.Printf("Current importance of a pod %s is:  %s\n", currPod.GetName(), currPod.GetAnnotations()["imp"])
				currPod.SetAnnotations(impMap)
				fmt.Printf("Setting importance of pod %s to %s\n", currPod.GetName(), importanceNumber)
				fmt.Println()
				_, _ = clientSet.CoreV1().Pods(namespace).Update(context.TODO(), &currPod, metav1.UpdateOptions{})
			}
		}
	}
}

func main() {
	cfg := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", cfg)
	if err != nil {
		log.Printf("Error reading Kubernetes config file")
		os.Exit(1)
	}
	//create the client
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Printf("Error creating Go client")
		os.Exit(1)
	}
	changeImportanceFactor(clientSet)
	//fetch the metrics
	err = getMetrics(clientSet, &pods)
	if err != nil {
		panic(err.Error())
	}
	//print the information
	for i, pod := range pods.Items {
		fmt.Printf("\nPod Name: %s, Pod Namespace: %s\n", pod.Metadata.Name, pod.Metadata.Namespace)
		for j, container := range pods.Items[i].Containers {
			fmt.Printf("%d) Container Name: %s, CPU Usage: %s, Memory Usage: %s\n", j+1, container.Name, container.Usage.CPU, container.Usage.Memory)
		}
	}
}
