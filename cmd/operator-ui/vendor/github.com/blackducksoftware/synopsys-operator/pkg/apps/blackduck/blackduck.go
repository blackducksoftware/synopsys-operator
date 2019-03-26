package blackduck

import (
	"fmt"
	"github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"

	v1blackduck "github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/v1"
	//v2blackduck "github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/v2"

	blackduckclientset "github.com/blackducksoftware/synopsys-operator/pkg/blackduck/client/clientset/versioned"

	routeclient "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	securityclient "github.com/openshift/client-go/security/clientset/versioned/typed/security/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"strings"
)

type Blackduck struct {
	config           *protoform.Config
	kubeConfig       *rest.Config
	kubeClient       *kubernetes.Clientset
	blackduckClient  *blackduckclientset.Clientset
	osSecurityClient *securityclient.SecurityV1Client
	routeClient      *routeclient.RouteV1Client
	creaters         []Creater
}

func NewBlackduck(config *protoform.Config, kubeConfig *rest.Config) *Blackduck {
	// Initialiase the clienset using kubeConfig
	kubeclient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil
	}

	blackduckClient, err := blackduckclientset.NewForConfig(kubeConfig)
	if err != nil {
		return nil
	}

	osClient, err := securityclient.NewForConfig(kubeConfig)
	if err != nil {
		osClient = nil
	} else {
		_, err := util.GetOpenShiftSecurityConstraint(osClient, "anyuid")
		if err != nil && strings.Contains(err.Error(), "could not find the requested resource") && strings.Contains(err.Error(), "openshift.io") {
			osClient = nil
		}
	}

	routeClient, err := routeclient.NewForConfig(kubeConfig)
	if err != nil {
		routeClient = nil
	} else {
		_, err := util.GetOpenShiftRoutes(routeClient, "default", "docker-registry")
		if err != nil && strings.Contains(err.Error(), "could not find the requested resource") && strings.Contains(err.Error(), "openshift.io") {
			routeClient = nil
		}
	}

	creaters := []Creater{
		v1blackduck.NewCreater(config, kubeConfig, kubeclient, blackduckClient, osClient, routeClient, false),
		//v2blackduck.NewCreater(config, kubeConfig, kubeclient, blackduckClient, osClient, routeClient, false),
	}

	return &Blackduck{
		kubeConfig:       kubeConfig,
		kubeClient:       kubeclient,
		blackduckClient:  blackduckClient,
		osSecurityClient: osClient,
		routeClient:      routeClient,
		creaters:         creaters,
	}
}

func (b Blackduck) getCreater(version string) (Creater, error) {
	for _, c := range b.creaters {
		for _, v := range c.Versions() {
			if strings.Compare(v, version) == 0 {
				return c, nil
			}
		}
	}
	return nil, fmt.Errorf("version %s is not supported", version)
}

func (b Blackduck) Delete(name string) {
	// TODO  - We should not delete the namespace but instead search for the current version (annotation??) then delete the Blackduck instance.
}

func (b Blackduck) Create(bd *v1.Blackduck) error {
	creater, err := b.getCreater(bd.Spec.Version)
	if err != nil {
		return err
	}

	return creater.Create(bd)
}

func (b Blackduck) Start(bd *v1.Blackduck) error {
	creater, err := b.getCreater(bd.Spec.Version)
	if err != nil {
		return err
	}

	return creater.Start(nil)
}

func (b Blackduck) Stop(bd *v1.Blackduck) error {
	creater, err := b.getCreater(bd.Spec.Version)
	if err != nil {
		return err
	}

	return creater.Stop(nil)
}

func (b Blackduck) Versions() []string {
	var versions []string
	for _, c := range b.creaters {
		for _, v := range c.Versions() {
			versions = append(versions, v)
		}
	}
	return versions
}

func (b Blackduck) Ensure() error {
	// TODO - Collect PV name, IP,
	return nil
}
