package v1

import (
	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	blackduckapi "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/components/utils"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/types"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/store"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"k8s.io/client-go/kubernetes"
)

type BdService struct {
	config     *protoform.Config
	kubeClient *kubernetes.Clientset
	blackduck  *blackduckapi.Blackduck
}

func (b BdService) GetService() *components.Service {
	return util.CreateService(util.GetResourceName(b.blackduck.Name, util.BlackDuckName, "webserver"), utils.GetLabel("webserver", b.blackduck.Name), b.blackduck.Spec.Namespace, int32(443), int32(8443), horizonapi.ServiceTypeServiceIP, utils.GetVersionLabel("webserver", b.blackduck.Name, b.blackduck.Spec.Version))
}

func NewBdService(config *protoform.Config, kubeClient *kubernetes.Clientset, blackduck *blackduckapi.Blackduck) types.ServiceInterface {
	return &BdService{config: config, kubeClient: kubeClient, blackduck: blackduck}
}

func init() {
	store.Register(types.ServiceWebserverV1, NewBdService)
}