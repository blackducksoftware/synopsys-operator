package v1

import (
	"github.com/blackducksoftware/horizon/pkg/components"
	blackduckapi "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/types"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/database/postgres"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/store"
	apputils "github.com/blackducksoftware/synopsys-operator/pkg/apps/utils"
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
	name := apputils.GetResourceName(b.blackduck.Name, util.BlackDuckName, "postgres")

	p := &postgres.Postgres{
		Name:        name,
		Namespace:   b.blackduck.Spec.Namespace,
		Port:        int32(5432),
		Database:    "blackduck",
		User:        "blackduck",
		Labels:      apputils.GetLabel("postgres", b.blackduck.Name),
		IsOpenshift: b.config.IsOpenshift,
	}

	return p.GetPostgresService()
}

func NewBdService(config *protoform.Config, kubeClient *kubernetes.Clientset, blackduck *blackduckapi.Blackduck) types.ServiceInterface {
	return &BdService{config: config, kubeClient: kubeClient, blackduck: blackduck}
}

func init() {
	store.Register(types.ServicePostgresV1, NewBdService)
}
