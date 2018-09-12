package controller_plugins

// This is a controller that updates the configmap
// in perceptor periodically.
// It is assumed that the configmap in perceptor will
// roll over any time this is updated, and if not, that
// there is a problem in the orchestration environment.

import (
	"encoding/json"
	"fmt"

	"github.com/blackducksoftware/horizon/pkg/api"
	hubclient "github.com/blackducksoftware/perceptor-protoform/pkg/hub/client/clientset/versioned"
	opssiteclient "github.com/blackducksoftware/perceptor-protoform/pkg/opssight/client/clientset/versioned"

	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
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

// Run ...
func (p *PerceptorConfigMap) Run(c api.ControllerResources, ch chan struct{}) error {

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
