package k8s

//go:generate mockgen -destination=mocks/mock_clientiface.go -package=mock_k8s sigs.k8s.io/controller-runtime/pkg/client Client

import (
	"context"
	"fmt"
	"github.com/keikoproj/iam-manager/pkg/log"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	clientv1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog"
	"os"
	"time"
)

type Client struct {
	Cl  kubernetes.Interface
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
		Cl:  client,
		dCl: dClient,
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
		Cl:  client,
		dCl: dClient,
	}
	return cl
}

// Iface defines required functions to be implemented by receivers
type Iface interface {
	IamrolesCount(ctx context.Context, ns string)
	GetConfigMap(ctx context.Context, ns string, name string) *v1.ConfigMap
	SetUpEventHandler(ctx context.Context) record.EventRecorder
	GetNamespace(ctx context.Context, ns string) *v1.Namespace
	CreateOrUpdateServiceAccount(ctx context.Context, saName string, ns string) error
	EnsureServiceAccount(ctx context.Context, req ServiceAccountRequest) error
	PatchServiceAccountAnnotation(ctx context.Context, saName string, ns string, annotation string, value string) error
	GetServiceAccount(ctx context.Context, saName string, ns string) (*v1.ServiceAccount, error)
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
	res, err := c.Cl.CoreV1().ConfigMaps(ns).Get(name, metav1.GetOptions{})
	if err != nil {
		log.Error(err, "unable to get config map")
		panic(err)
	}

	return res
}

//GetNamespace gets the namespace metadata. This will be used to validate if the namespace is annotated for privileged namespace.
func (c *Client) GetNamespace(ctx context.Context, ns string) (*v1.Namespace, error) {
	log := log.Logger(ctx, "k8s", "client", "GetNamespace")
	log.WithValues("namespace", ns)
	log.Info("Retrieving Namespace")
	resp := &v1.Namespace{}
	resp, err := c.Cl.CoreV1().Namespaces().Get(ns, metav1.GetOptions{})
	if err != nil {
		log.Error(err, "unable to get the namespace details")
		return nil, err
	}

	if err != nil {
		log.Error(err, "unable to get namespace metadata")
		return nil, err
	}
	return resp, nil
}

func (c *Client) ClientInterface() kubernetes.Interface {
	return c.Cl
}

// GetConfigMapInformer returns shared informer for given config map
func GetConfigMapInformer(ctx context.Context, nsName string, cmName string) cache.SharedIndexInformer {
	log := log.Logger(context.Background(), "pkg.k8s.client", "GetConfigMapInformer")
	clientset, err := NewK8sClient()
	if err != nil {
		log.Error(err, "failed to get clientset")
		return nil
	}

	listOptions := func(options *metav1.ListOptions) {
		options.FieldSelector = fmt.Sprintf("metadata.name=%s", cmName)
	}

	// default resync period 24 hours
	cmInformer := clientv1.NewFilteredConfigMapInformer(clientset.ClientInterface(), nsName, 24*time.Hour, cache.Indexers{}, listOptions)
	return cmInformer
}

//SetUpEventHandler sets up event handler with client-go recorder instead of creating events directly
func (c *Client) SetUpEventHandler(ctx context.Context) record.EventRecorder {
	log := log.Logger(ctx, "k8s", "client", "SetUpEventHandler")
	//This was re-written based on job-controller in kuberentest repo
	//For more info refer: https://github.com/kubernetes/kubernetes/blob/master/pkg/controller/job/job_controller.go
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: c.Cl.CoreV1().Events("")})
	log.V(1).Info("Successfully added event broadcaster")
	return eventBroadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: "iam-manager"})
}
