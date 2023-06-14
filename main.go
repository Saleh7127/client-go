package main

import (
	"context"
	"flag"
	"fmt"
	apiv1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {

	kubeconfig := flag.String("kubeconfig", "/home/user/.kube/config", "location to your kubeconfig file")
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)

	if err != nil {
		fmt.Printf("Error %s, building config from flags\n", err.Error())

		config, err = rest.InClusterConfig()

		if err != nil {
			fmt.Printf("Error %s, getting inclusterconfig\n", err.Error())
		}

	}

	clientset, err := kubernetes.NewForConfig(config)

	if err != nil {
		fmt.Printf("Error %s, creating clientset\n", err.Error())
	}

	pods, err := clientset.CoreV1().Pods("default").List(context.Background(), metaV1.ListOptions{})

	if err != nil {
		fmt.Printf("Error %s, listing all pods from default ns\n", err.Error())
	}

	fmt.Println("Pod from default ns")
	for _, pod := range pods.Items {
		fmt.Printf("%s\n", pod.Name)
	}

	deploy, err := clientset.AppsV1().Deployments(apiv1.NamespaceDefault).List(context.Background(), metaV1.ListOptions{})

	if err != nil {
		fmt.Printf("Error %s, listing all deployments from default ns\n", err.Error())
	}

	fmt.Printf("Deployments name\n")
	for _, deploy := range deploy.Items {
		fmt.Printf("%s\n", deploy.Name)
	}

}
