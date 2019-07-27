package v1

import (
	"fmt"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	blackduckapi "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/store"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/types"
	apputils "github.com/blackducksoftware/synopsys-operator/pkg/apps/utils"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"k8s.io/client-go/kubernetes"
)

// BdService holds the Black Duck service configuration
type BdService struct {
	config     *protoform.Config
	kubeClient *kubernetes.Clientset
	blackDuck  *blackduckapi.Blackduck
}

// GetService returns the service
func (b BdService) GetService() *components.Service {
	// TODO: remove GetResourceName method until the HUB-20482 is fixed. once it if fixed, add them back
	return util.CreateService("authentication", apputils.GetLabel("authentication", b.blackDuck.Name), b.blackDuck.Spec.Namespace, int32(8443), int32(8443), horizonapi.ServiceTypeServiceIP, apputils.GetVersionLabel("authentication", b.blackDuck.Name, b.blackDuck.Spec.Version))
}

// NewBdService returns the Black Duck service configuration
func NewBdService(config *protoform.Config, kubeClient *kubernetes.Clientset, cr interface{}) (types.ServiceInterface, error) {
	blackDuck, ok := cr.(*blackduckapi.Blackduck)
	if !ok {
		return nil, fmt.Errorf("unable to cast the interface to Black Duck object")
	}
	return &BdService{config: config, kubeClient: kubeClient, blackDuck: blackDuck}, nil
}

func init() {
	store.Register(blackduck.BlackDuckAuthentivationServiceV1, NewBdService)
}
