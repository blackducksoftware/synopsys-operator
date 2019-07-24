package v1

import (
	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	blackduckapi "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/types"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/store"
	apputils "github.com/blackducksoftware/synopsys-operator/pkg/apps/utils"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"k8s.io/client-go/kubernetes"
	"strings"
)

type BdService struct {
	config     *protoform.Config
	kubeClient *kubernetes.Clientset
	blackduck  *blackduckapi.Blackduck
}

func (b BdService) GetService() *components.Service {
	var svc *components.Service

	switch strings.ToUpper(b.blackduck.Spec.ExposeService) {
	case util.NODEPORT:
		svc = util.CreateService(apputils.GetResourceName(b.blackduck.Name, util.BlackDuckName, "webserver-exposed"), apputils.GetLabel("webserver", b.blackduck.Name), b.blackduck.Spec.Namespace, int32(443), int32(8443), horizonapi.ServiceTypeNodePort, apputils.GetLabel("webserver-exposed", b.blackduck.Name))

		break
	case util.LOADBALANCER:
		svc = util.CreateService(apputils.GetResourceName(b.blackduck.Name, util.BlackDuckName, "webserver-exposed"), apputils.GetLabel("webserver", b.blackduck.Name), b.blackduck.Spec.Namespace, int32(443), int32(8443), horizonapi.ServiceTypeLoadBalancer, apputils.GetLabel("webserver-exposed", b.blackduck.Name))
		break
	default:
	}

	return svc
}

func NewBdService(config *protoform.Config, kubeClient *kubernetes.Clientset, blackduck *blackduckapi.Blackduck) types.ServiceInterface {
	return &BdService{config: config, kubeClient: kubeClient, blackduck: blackduck}
}

func init() {
	store.Register(types.ServiceExposeV1, NewBdService)
}
