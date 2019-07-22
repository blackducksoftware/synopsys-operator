package v1

import (
	"fmt"
	"github.com/blackducksoftware/horizon/pkg/components"
	blackduckapi "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/components/rc/utils"
	utils2 "github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/components/utils"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/types"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/database/postgres"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/store"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"

	"k8s.io/client-go/kubernetes"
)

type BdReplicationController struct {
	*types.ReplicationController
	config     *protoform.Config
	kubeClient *kubernetes.Clientset
	blackduck  *blackduckapi.Blackduck
}

func init() {
	store.Register(types.RcPostgresV1, NewBdReplicationController)
}

func (c *BdReplicationController) GetRc() (*components.ReplicationController, error) {

	containerConfig, ok := c.Containers[types.PostgresContainerName]
	if !ok {
		return nil, fmt.Errorf("couldn't find container %s", types.PostgresContainerName)
	}

	name := util.GetResourceName(c.blackduck.Name, util.BlackDuckName, "postgres")

	var pvcName string
	if c.blackduck.Spec.PersistentStorage {
		pvcName = utils.GetPVCName("postgres", c.blackduck)
	}

	p := &postgres.Postgres{
		Name:                   name,
		Namespace:              c.blackduck.Spec.Namespace,
		PVCName:                pvcName,
		Port:                   int32(5432),
		Image:                  containerConfig.Image,
		MinCPU:                 util.Int32ToInt(containerConfig.MinCPU),
		MaxCPU:                 util.Int32ToInt(containerConfig.MaxCPU),
		MinMemory:              util.Int32ToInt(containerConfig.MinMem),
		MaxMemory:              util.Int32ToInt(containerConfig.MaxMem),
		Database:               "blackduck",
		User:                   "blackduck",
		PasswordSecretName:     util.GetResourceName(c.blackduck.Name, util.BlackDuckName, "db-creds"),
		UserPasswordSecretKey:  "HUB_POSTGRES_ADMIN_PASSWORD_FILE",
		AdminPasswordSecretKey: "HUB_POSTGRES_POSTGRES_PASSWORD_FILE",
		MaxConnections:         300,
		SharedBufferInMB:       1024,
		EnvConfigMapRefs:       []string{util.GetResourceName(c.blackduck.Name, util.BlackDuckName, "db-config")},
		Labels:                 utils2.GetLabel("postgres", c.blackduck.Name),
		IsOpenshift:            c.config.IsOpenshift,
	}

	return p.GetPostgresReplicationController()
}
func NewBdReplicationController(replicationController *types.ReplicationController, config *protoform.Config, kubeClient *kubernetes.Clientset, blackduck *blackduckapi.Blackduck) types.ReplicationControllerInterface {
	return &BdReplicationController{ReplicationController: replicationController, config: config, kubeClient: kubeClient, blackduck: blackduck}
}
