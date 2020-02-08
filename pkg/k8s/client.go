package k8s

import (
	"context"
	"github.com/keikoproj/iam-manager/pkg/log"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
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

//NewK8sClient gets the new k8s go client
func NewK8sClientDoOrDie() *Client {
	config, err := rest.InClusterConfig()
	if err != nil {
		// Do i need to panic here?
		//How do i test this from local?
		//Lets get it from local config file
		config, err = clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	dClient, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	cl := &Client{
		client,
		dClient,
	}
	return cl
}

//Iface defines required functions to be implemented by receivers
type Iface interface {
	IamrolesCount(ctx context.Context, ns string)
	GetConfigMap(ctx context.Context, ns string, name string) *v1.ConfigMap
	SetUpEventHandler(ctx context.Context) record.EventRecorder
}

//IamrolesCount function lists the "Iamrole" for a provided namespace
func (c *Client) IamrolesCount(ctx context.Context, ns string) (int, error) {
	log := log.Logger(ctx, "k8s", "client", "IamrolesCount")
	log.WithValues("namespace", ns)
	log.V(1).Info("list api call")
	iamCR := schema.GroupVersionResource{
		Group:    "iammanager.keikoproj.io",
		Version:  "v1alpha1",
		Resource: "iamroles",
	}

	roleList, err := c.dCl.Resource(iamCR).Namespace(ns).List(metav1.ListOptions{})
	if err != nil {
		log.Error(err, "unable to list iamroles resources")
		return 0, err
	}
	log.Info("Total number of roles", "count", len(roleList.Items))
	return len(roleList.Items), nil
}

func (c *Client) GetConfigMap(ctx context.Context, ns string, name string) *v1.ConfigMap {
	log := log.Logger(ctx, "k8s", "client", "GetConfigMap")
	log.WithValues("namespace", ns)
	log.Info("Retrieving config map")
	res, err := c.cl.CoreV1().ConfigMaps(ns).Get(name, metav1.GetOptions{})
	if err != nil {
		log.Error(err, "unable to get config map")
		panic(err)
	}

	return res
}

//SetUpEventHandler sets up event handler with client-go recorder instead of creating events directly
func (c *Client) SetUpEventHandler(ctx context.Context) record.EventRecorder {
	log := log.Logger(ctx, "k8s", "client", "SetUpEventHandler")
	//This was re-written based on job-controller in kuberentest repo
	//For more info refer: https://github.com/kubernetes/kubernetes/blob/master/pkg/controller/job/job_controller.go
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: c.cl.CoreV1().Events("")})
	log.V(1).Info("Successfully add event broadcaster")
	return eventBroadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: "iam-manager"})
}
