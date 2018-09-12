package controller_plugins

// This is a controller that updates the configmap
// in perceptor periodically.
// It is assumed that the configmap in perceptor will
// roll over any time this is updated, and if not, that
// there is a problem in the orchestration environment.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/blackducksoftware/horizon/pkg/api"
	hubclient "github.com/blackducksoftware/perceptor-protoform/pkg/hub/client/clientset/versioned"
	opssiteclient "github.com/blackducksoftware/perceptor-protoform/pkg/opssight/client/clientset/versioned"

	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PerceptorConfigMap struct {
}

type PerceptorPayload struct {
	HubURLs []string `json:"HubURLs"`
}

// sendHubs is one possible way to configure the perceptor hub family.
// TODO replace w/ configmap mutation if we want to.
func sendHubs(url string, hubs []string) {
	data := PerceptorPayload{
		HubURLs: hubs,
	}
	payloadBytes, err := json.Marshal(data)
	if err != nil {
		// handle err
	}
	body := bytes.NewReader(payloadBytes)

	req, err := http.NewRequest("PUT", url, body)
	if err != nil {
		// handle err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// handle err
	}
	defer resp.Body.Close()
}

func (p *PerceptorConfigMap) Run(c api.ControllerResources, ch chan struct{}) error {

	// for matt .
	allHubNamespaces := func() []string {
		allHubNamespaces := []string{}
		hubsList, _ := hubclient.New(c.KubeClient.RESTClient()).SynopsysV1().Hubs("").List(v1.ListOptions{})
		hubs := hubsList.Items
		for _, hub := range hubs {
			ns := hub.Namespace
			allHubNamespaces = append(allHubNamespaces, ns)
			logrus.Infof("Hub config map controller, namespace is %v", ns)
		}
		return allHubNamespaces
	}()

	// for opssight 3.0, only support one opssight
	opssiteList, _ := opssiteclient.New(c.KubeClient.RESTClient()).SynopsysV1().OpsSights("").List(v1.ListOptions{})

	// curl perceptor w/ the latest hub list
	for _, opssight := range opssiteList.Items {
		sendHubs(fmt.Sprintf("%v.svc.cluster.local", opssight.Namespace), allHubNamespaces)
	}
	return nil
}
