package k8s

import (
	"context"
	"fmt"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/tools/clientcmd"
	"os"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Client struct {
	cl  kubernetes.Interface
	dCl dynamic.Interface
}

//NewK8sClient gets the new k8s go client
func NewK8sClient() (*Client, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		// Do i need to panic here?
		//How do i test this from local?
		//Lets get it from local config file
		config, err = clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	dClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	cl := &Client{
		client,
		dClient,
	}
	return cl, nil
}

//Iface defines required functions to be implemented by receivers
type Iface interface {
	IamrolesCount(ctx context.Context, ns string)
	GetConfigMap(ctx context.Context, ns string, name string)
}

//IamrolesCount function lists the "Iamrole" for a provided namespace
func (c *Client) IamrolesCount(ctx context.Context, ns string) (int, error) {
	fmt.Println("Iam at least herereeeeeeeeeeee")
	iamCR := schema.GroupVersionResource{
		Group:    "iammanager.keikoproj.io",
		Version:  "v1alpha1",
		Resource: "iamroles",
	}

	roleList, err := c.dCl.Resource(iamCR).Namespace(ns).List(metav1.ListOptions{})
	if err != nil {
		fmt.Printf("error = %s\n", err.Error())
		return 0, err
	}
	fmt.Printf("Total number of roles = %d\n", len(roleList.Items))
	return len(roleList.Items), nil
}

func (c *Client) GetConfigMap(ctx context.Context, ns string, name string) *v1.ConfigMap {
	res, err := c.cl.CoreV1().ConfigMaps(ns).Get(name, metav1.GetOptions{})
	if err != nil {
		fmt.Printf("error = %s\n", err.Error())
		panic(err)
	}

	return res
}
