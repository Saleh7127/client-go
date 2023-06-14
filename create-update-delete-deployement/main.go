package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"k8s.io/client-go/util/retry"
	"log"
	"os"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {

	kubeconfig := flag.String("kubeconfig", "/home/user/.kube/config", "location to your kubeconfig file")
	fmt.Println(kubeconfig)
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		fmt.Printf("Error %s, building config from flags\n", err.Error())
		config, err = rest.InClusterConfig()
		if err != nil {
			fmt.Printf("Error %s, getting inclusterconfig\n", err.Error())
		}
	}

	ClientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Printf("Error %s, creating Clientset\n", err.Error())
	}

	/*
		pods, err := ClientSet.CoreV1().Pods("default").List(context.Background(), metaV1.ListOptions{})
		if err != nil {
			fmt.Printf("Error %s, listing all pods from default ns\n", err.Error())
		}
		fmt.Println("Pod from default ns")
		for _, pod := range pods.Items {
			fmt.Printf("%s\n", pod.Name)
		} */

	deployClient := ClientSet.AppsV1().Deployments(apiv1.NamespaceDefault)

	deployment := &appsv1.Deployment{
		ObjectMeta: metaV1.ObjectMeta{
			Name: "demo-deploy",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(3),
			Selector: &metaV1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "demo",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metaV1.ObjectMeta{
					Labels: map[string]string{
						"app": "demo",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "web",
							Image: "nginx:1.12",
							Ports: []apiv1.ContainerPort{
								{
									Name:          "http",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
		},
	}

	fmt.Println("Creating Deployment")
	status, err := deployClient.Create(context.Background(), deployment, metaV1.CreateOptions{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Deployment Created %q.\n", status.GetObjectMeta().GetName())

	prompt()

	fmt.Println("Updating Deployment")
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		status, err := deployClient.Get(context.Background(), "demo-deploy", metaV1.GetOptions{})
		if err != nil {
			log.Fatal(err)
		}
		status.Spec.Replicas = int32Ptr(2)
		status.Spec.Template.Spec.Containers[0].Image = "nginx.1.13"
		_, err = deployClient.Update(context.Background(), status, metaV1.UpdateOptions{})
		return err
	})
	if retryErr != nil {
		log.Fatal(err)
	}
	fmt.Println("Deployment Updated")

	prompt()

	fmt.Printf("Listing Deployments in namespace %s:\n", apiv1.NamespaceDefault)
	DeploymentList, err := deployClient.List(context.Background(), metaV1.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}
	for _, deploy := range DeploymentList.Items {
		fmt.Printf("Deployment Name: %s and have %v replicas\n", deploy.Name, *deploy.Spec.Replicas)
	}

	prompt()

	fmt.Println("Deleting Deployment")
	deletePolicy := metaV1.DeletePropagationForeground
	err = deployClient.Delete(context.Background(), "demo-deploy", metaV1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Deployment Deleted")
}

func int32Ptr(i int32) *int32 { return &i }

func prompt() {
	fmt.Printf("-> Press Return key to continue.")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		break
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	fmt.Println()
}
