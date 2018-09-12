package plugins

// This is a controller that updates the configmap
// in perceptor periodically.
// It is assumed that the configmap in perceptor will
// roll over any time this is updated, and if not, that
// there is a problem in the orchestration environment.

import (
	"k8s.io/apimachinery/pkg/util/wait"
	"encoding/json"
	"fmt"

	"github.com/blackducksoftware/horizon/pkg/api"
	hubclient "github.com/blackducksoftware/perceptor-protoform/pkg/hub/client/clientset/versioned"
	opssiteclient "github.com/blackducksoftware/perceptor-protoform/pkg/opssight/client/clientset/versioned"
 	"github.com/kubernetes/kubernetes/pkg/kubelet/kubeletconfig/util/log"

	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type hubConfig struct {
	Hosts                     []string
	User                      string
	PasswordEnvVar            string
	ClientTimeoutMilliseconds int
	Port                      int
	ConcurrentScanLimit       int
	TotalScanLimit            int
}

type timings struct {
	CheckForStalledScansPauseHours int
	StalledScanClientTimeoutHours  int
	ModelMetricsPauseSeconds       int
	UnknownImagePauseMilliseconds  int
}

type perceptorConfig struct {
	Hub         *hubConfig
	Timings     *timings
	UseMockMode bool
	Port        int
	LogLevel    string
}

// PerceptorConfigMap ...
type PerceptorConfigMap struct{}

// sendHubs is one possible way to configure the perceptor hub family.
// TODO replace w/ configmap mutation if we want to.
func sendHubs(kubeClient *kubernetes.Clientset, namespace string, hubs []string) error {
	configmapList, err := kubeClient.Core().ConfigMaps(namespace).List(meta_v1.ListOptions{})
	if err != nil {
		return err
	}

	var configMap *v1.ConfigMap
	for _, cm := range configmapList.Items {
		if cm.Name == "perceptor-config" {
			configMap = &cm
			break
		}
	}

	if configMap == nil {
		return fmt.Errorf("unable to find configmap perceptor-config")
	}

	var value perceptorConfig
	err = json.Unmarshal([]byte(configMap.Data["perceptor_conf.yaml"]), &value)
	if err != nil {
		return err
	}

	value.Hub.Hosts = hubs

	jsonBytes, err := json.Marshal(value)
	if err != nil {
		return err
	}

	configMap.Data["perceptor_conf.yaml"] = string(jsonBytes)
	kubeClient.Core().ConfigMaps(namespace).Update(configMap)

	return nil
}

// Run is a BLOCKING function which should be run by the framework .
func (p *PerceptorConfigMap) Run(c api.ControllerResources, ch chan struct{}) {
	syncFunc := func() {
		p.updateAllHubs(c, ch)
	}
	lw := &cache.ListWatch{
		ListFunc: func(opts metav1.ListOptions) (runtime.Object, error) {
			return hubclient.New(c.KubeClient.RESTClient()).SynopsysV1().Hubs(v1.NamespaceAll).List()
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return hubclient.New(c.KubeClient.RESTClient()).SynopsysV1().Hubs(v1.NamespaceAll).Watch()
		},
	}
	st, ctrl := cache.NewInformer(lw,
		&extensions.Deployment{},
		c.SyncPeriod,
		cache.ResourceEventHandlerFuncs{
			// TODO kinda dumb, we just do a complete re-list of all hubs, 
			// every time an event happens... But thats all we need to do, so its good enough.
			DeleteFunc: func(obj interface{}) {
				logrus.Infof("Hub deleted ! %v ",obj)
				syncFunc()
			},
			OnAdd: func(obj interface{}){
				logrus.Infof("Hub added ! %v ",obj)
				syncFunc()
			}
		},
	)
	logrus.Infof("Starting controller for hub<->perceptor updates... this blocks, so running in a go func.")
	
	// make sure this is called from a go func.
	// This blocks!  
	ctrl.Run(ch)
}

// Run ...
func (p *PerceptorConfigMap) updateAllHubs(c api.ControllerResources, ch chan struct{}) error {
	allHubNamespaces := func() []string {
		allHubNamespaces := []string{}
		hubsList, _ := hubclient.New(c.KubeClient.RESTClient()).SynopsysV1().Hubs(v1.NamespaceAll).List(meta_v1.ListOptions{})
		hubs := hubsList.Items
		for _, hub := range hubs {
			ns := hub.Namespace
			allHubNamespaces = append(allHubNamespaces, ns)
			logrus.Infof("Hub config map controller, namespace is %v", ns)
		}
		return allHubNamespaces
	}()

	// for opssight 3.0, only support one opssight
	opssiteList, err := opssiteclient.New(c.KubeClient.RESTClient()).SynopsysV1().OpsSights(v1.NamespaceAll).List(meta_v1.ListOptions{})
	if err != nil {
		return err
	}

	// curl perceptor w/ the latest hub list
	for _, opssight := range opssiteList.Items {
		sendHubs(c.KubeClient, opssight.Namespace, allHubNamespaces)
	}
	return nil
}
