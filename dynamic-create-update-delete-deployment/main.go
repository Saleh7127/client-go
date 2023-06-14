package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"log"
	"os"

	apiv1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/retry"
)

func main() {

	kubeconfig := flag.String("Kubeconfig", "/home/user/.kube/config", "location to your kubeconfig file")
	fmt.Println(kubeconfig)
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		fmt.Printf("Error %s, building config from flags\n", err.Error())
		config, err = rest.InClusterConfig()
		if err != nil {
			fmt.Printf("Error %s, getting inclusterconfig\n", err.Error())
		}
	}

	ClientSet, err := dynamic.NewForConfig(config)
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

	deploymentRes := schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}

	deployment := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name": "demo-dynamic-deploy",
			},
			"spec": map[string]interface{}{
				"replicas": 2,
				"selector": map[string]interface{}{
					"matchLabels": map[string]interface{}{
						"app": "demo-dynamic-deploy",
					},
				},
				"template": map[string]interface{}{
					"metadata": map[string]interface{}{
						"labels": map[string]interface{}{
							"app": "demo-dynamic-deploy",
						},
					},
					"spec": map[string]interface{}{
						"containers": []map[string]interface{}{
							{
								"name":  "web",
								"image": "nginx:1.12",
								"ports": []map[string]interface{}{
									{
										"name":          "http",
										"protocol":      "TCP",
										"containerPort": 80,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	fmt.Println("Creating Deployment")
	status, err := ClientSet.Resource(deploymentRes).Namespace(apiv1.NamespaceDefault).Create(context.Background(), deployment, metaV1.CreateOptions{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Deployment Created %q.\n", status.GetName())

	prompt()

	fmt.Println("Updating Deployment")
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		status, err := ClientSet.Resource(deploymentRes).Namespace(apiv1.NamespaceDefault).Get(context.Background(), "demo-dynamic-deploy", metaV1.GetOptions{})
		if err != nil {
			log.Fatal(err)
		}
		err = unstructured.SetNestedField(status.Object, int64(1), "spec", "replicas")
		if err != nil {
			log.Fatal(err)
		}
		containers, found, err := unstructured.NestedSlice(status.Object, "spec", "template", "spec", "containers")
		if err != nil || found == false || containers == nil {
			log.Fatal(err)
		}
		err = unstructured.SetNestedField(containers[0].(map[string]interface{}), "nginx:1.13", "image")
		if err != nil {
			panic(err)
		}
		if err := unstructured.SetNestedField(status.Object, containers, "spec", "template", "spec", "containers"); err != nil {
			panic(err)
		}
		_, err = ClientSet.Resource(deploymentRes).Namespace(apiv1.NamespaceDefault).Update(context.Background(), status, metaV1.UpdateOptions{})

		return err
	})

	if retryErr != nil {
		log.Fatal(err)
	}
	fmt.Println("Deployment Updated")

	prompt()

	fmt.Printf("Listing Deployments in namespace %s:\n", apiv1.NamespaceDefault)
	DeploymentList, err := ClientSet.Resource(deploymentRes).Namespace(apiv1.NamespaceDefault).List(context.Background(), metaV1.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}
	for _, deploy := range DeploymentList.Items {
		replicas, found, err := unstructured.NestedInt64(deploy.Object, "spec", "replicas")
		if err != nil || found == false {
			fmt.Printf("Replicas not found for deployment %s: error=%s", deploy.GetName(), err)
			continue
		}
		fmt.Printf("Deployment Name: %s and have %v replicas\n", deploy.GetName(), replicas)
	}

	prompt()

	fmt.Println("Deleting Deployment")
	deletePolicy := metaV1.DeletePropagationForeground
	err = ClientSet.Resource(deploymentRes).Namespace(apiv1.NamespaceDefault).Delete(context.Background(), "demo-dynamic-deploy", metaV1.DeleteOptions{
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
