package final

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	clientcmd "k8s.io/client-go/tools/clientcmd/api"
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

var podsObj PodMetricsList // pods structure object

// authenticate is used to authenticate Go-client with GKE cluster.
func Authenticate(filePath string, HostIp string) *kubernetes.Clientset {
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
		panic(err.Error())
	}
	return clientSet
}

// listPodsByNamespace lists the number of pods in the cluster.
func listPodsByNamespace(clientSet *kubernetes.Clientset) {
	namespace := apiv1.NamespaceDefault
	pods, err := clientSet.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	for _, pod := range pods.Items {
		fmt.Printf("The pods are : %s\n", pod.GetName())
	}
}

// changeReplica changes the number of replica-sets of a certain deployment.
func ChangeReplica(clientSet *kubernetes.Clientset, policy *Policy) {
	namespace := apiv1.NamespaceDefault
	deployment, err := clientSet.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}

	d1 := deployment.Items[0]
	d2 := deployment.Items[1]
	d3 := deployment.Items[2]
	fmt.Printf("Number of replica-set deployed currently : %d %d %d\n ", *(d1.Spec.Replicas), *(d2.Spec.Replicas), *(d3.Spec.Replicas))

	top := policy.Factor[policy.Status]["Top"]
	medium := policy.Factor[policy.Status]["Medium"]
	low := policy.Factor[policy.Status]["Low"]
	fmt.Printf("Changing the # of replica-set to top/medium/low %d %d %d\n", top, medium, low)
	*(d1.Spec.Replicas) = top
	*(d2.Spec.Replicas) = medium
	*(d3.Spec.Replicas) = low
	_, _ = clientSet.AppsV1().Deployments(namespace).Update(context.TODO(), &d1, metav1.UpdateOptions{})
	_, _ = clientSet.AppsV1().Deployments(namespace).Update(context.TODO(), &d2, metav1.UpdateOptions{})
	_, _ = clientSet.AppsV1().Deployments(namespace).Update(context.TODO(), &d3, metav1.UpdateOptions{})
	fmt.Printf("Number of replica-set deployed after change : %d %d %d\n", top, medium, low)

}

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

//getPodMetrics prints out the CPU & Memory usages of the pods.
func getPodMetrics(clientSet *kubernetes.Clientset) {
	err := getMetrics(clientSet, &podsObj)
	if err != nil {
		panic(err.Error())
	}
	//print the information
	for i, pod := range podsObj.Items {
		fmt.Printf("\nPod Name: %s, Pod Namespace: %s\n", pod.Metadata.Name, pod.Metadata.Namespace)
		for j, container := range podsObj.Items[i].Containers {
			fmt.Printf("%d) Container Name: %s, CPU Usage: %s, Memory Usage: %s\n", j+1, container.Name, container.Usage.CPU, container.Usage.Memory)
		}
	}

}

//func main() {
//	filePath := os.Args[1]                         //Pass .pem file as a command line argument
//	clusterIP := os.Args[2]                        //Pass cluster IP address
//	clientSet := authenticate(filePath, clusterIP) //Authenticates with the GCP cluster
//	listPodsByNamespace(clientSet)                 //Lists the pods in the cluster
//	changeReplica(clientSet)                       //change the number of replica-sets
//	getPodMetrics(clientSet)                       //Gets the CPU & Memory usages of the pods
//}
