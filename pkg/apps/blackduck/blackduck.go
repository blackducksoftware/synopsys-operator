package blackduck

import (
	"fmt"
	"strings"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	routeclient "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	securityclient "github.com/openshift/client-go/security/clientset/versioned/typed/security/v1"

	"github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"

	blackduckclientset "github.com/blackducksoftware/synopsys-operator/pkg/blackduck/client/clientset/versioned"

	latestblackduck "github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/latest"
	v1blackduck "github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/v1"
)

// Blackduck is used for the Blackduck deployment
type Blackduck struct {
	config           *protoform.Config
	kubeConfig       *rest.Config
	kubeClient       *kubernetes.Clientset
	blackduckClient  *blackduckclientset.Clientset
	osSecurityClient *securityclient.SecurityV1Client
	routeClient      *routeclient.RouteV1Client
	creaters         []Creater
}

// NewBlackduck will return a Blackduck
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
		latestblackduck.NewCreater(config, kubeConfig, kubeclient, blackduckClient, osClient, routeClient, false),
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

// Delete will be used to delete a blackduck instance
func (b Blackduck) Delete(name string) {
	// TODO  - We should not delete the namespace but instead search for the current version (annotation??) then delete the Blackduck instance.
}

// Versions returns the versions that the operator supports
func (b Blackduck) Versions() []string {
	var versions []string
	for _, c := range b.creaters {
		for _, v := range c.Versions() {
			versions = append(versions, v)
		}
	}
	return versions
}

// Ensure will make sure the instance is correctly deployed or deploy it if needed
func (b Blackduck) Ensure(bd *v1.Blackduck) error {
	// TODO - Collect PV name, IP,
	creater, err := b.getCreater(bd.Spec.Version)
	if err != nil {
		return err
	}
	return creater.Ensure(bd)
}
