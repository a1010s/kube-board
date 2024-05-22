package main

import (
	"context"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"time"

	v1 "k8s.io/api/core/v1"
//	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type PodInfo struct {
	Name            string
	Namespace       string
	Status          string
	Reason          string
	Message         string
	ContainerStatus []ContainerInfo
}

type ContainerInfo struct {
	Name  string
	Ready bool
}

type DeploymentInfo struct {
	Name      string
	Namespace string
	Replicas  int32
	Ready     int32
}

var clientset *kubernetes.Clientset

func main() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating Kubernetes clientset: %s", err.Error())
	}

	http.HandleFunc("/pod", podHandler)
	http.HandleFunc("/deploy", deployHandler)

	go startScanning(10 * time.Second)

	fmt.Println("Starting server at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func podHandler(w http.ResponseWriter, r *http.Request) {
	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Error listing pods: %s", err.Error())
	}

	var podInfos []PodInfo
	for _, pod := range pods.Items {
		podInfo := checkPodHealth(&pod)
		podInfos = append(podInfos, podInfo)
	}

	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		log.Fatalf("Error parsing template: %s", err.Error())
	}

	tmpl.Execute(w, podInfos)
}

func deployHandler(w http.ResponseWriter, r *http.Request) {
	deployments, err := clientset.AppsV1().Deployments("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Error listing deployments: %s", err.Error())
	}

	var deployInfos []DeploymentInfo
	for _, deploy := range deployments.Items {
		deployInfo := DeploymentInfo{
			Name:      deploy.Name,
			Namespace: deploy.Namespace,
			Replicas:  *deploy.Spec.Replicas,
			Ready:     deploy.Status.ReadyReplicas,
		}
		deployInfos = append(deployInfos, deployInfo)
	}

	tmpl, err := template.ParseFiles("templates/deploy.html")
	if err != nil {
		log.Fatalf("Error parsing template: %s", err.Error())
	}

	tmpl.Execute(w, deployInfos)
}

func checkPodHealth(pod *v1.Pod) PodInfo {
	podInfo := PodInfo{
		Name:      pod.Name,
		Namespace: pod.Namespace,
		Status:    getPodStatus(pod),
	}

	if pod.Status.Phase == v1.PodFailed {
		podInfo.Reason = pod.Status.Reason
		podInfo.Message = pod.Status.Message
	} else if pod.Status.Phase == v1.PodRunning {
		for _, containerStatus := range pod.Status.ContainerStatuses {
			containerInfo := ContainerInfo{
				Name:  containerStatus.Name,
				Ready: containerStatus.Ready,
			}
			podInfo.ContainerStatus = append(podInfo.ContainerStatus, containerInfo)
		}
	} else if pod.Status.Phase == v1.PodPending {
		if len(pod.Status.ContainerStatuses) > 0 {
			containerStatus := pod.Status.ContainerStatuses[0]
			if !containerStatus.Ready {
				podInfo.Reason = containerStatus.State.Waiting.Reason
				podInfo.Message = containerStatus.State.Waiting.Message
			}
		}
	}

	return podInfo
}

func getPodStatus(pod *v1.Pod) string {
	for _, condition := range pod.Status.Conditions {
		if condition.Type == v1.PodReady && condition.Status == v1.ConditionTrue {
			return "Ready"
		}
	}
	for _, containerStatus := range pod.Status.ContainerStatuses {
		if !containerStatus.Ready {
			return "Not Ready"
		}
	}
	return string(pod.Status.Phase)
}

func startScanning(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			scanAndLog()
		}
	}
}

func scanAndLog() {
	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Error listing pods: %s", err.Error())
	}

	fmt.Printf("Scanning at %s...\n", time.Now().Format("2006-01-02 15:04:05"))

	for _, pod := range pods.Items {
		podInfo := checkPodHealth(&pod)
		fmt.Printf("Pod Name: %s, Namespace: %s, Status: %s\n", podInfo.Name, podInfo.Namespace, podInfo.Status)
	}
}
