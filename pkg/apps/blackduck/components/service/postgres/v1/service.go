package v1

import (
	"fmt"
	"github.com/blackducksoftware/horizon/pkg/components"
	blackduckapi "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/database/postgres"
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
	name := apputils.GetResourceName(b.blackDuck.Name, util.BlackDuckName, "postgres")

	p := &postgres.Postgres{
		Name:        name,
		Namespace:   b.blackDuck.Spec.Namespace,
		Port:        int32(5432),
		Database:    "blackduck",
		User:        "blackduck",
		Labels:      apputils.GetLabel("postgres", b.blackDuck.Name),
		IsOpenshift: b.config.IsOpenshift,
	}

	return p.GetPostgresService()
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
	store.Register(types.BlackDuckPostgresServiceV1, NewBdService)
}
