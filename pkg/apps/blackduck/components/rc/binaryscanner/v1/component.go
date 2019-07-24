package v1

import (
	"fmt"
	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	blackduckapi "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/components/rc/utils"
	utils2 "github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/components/utils"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/types"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/store"
	apputils "github.com/blackducksoftware/synopsys-operator/pkg/apps/utils"
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
	store.Register(types.RcBinaryScannerV1, NewBdReplicationController)
}

func (c *BdReplicationController) GetRc() (*components.ReplicationController, error) {
	containerConfig, ok := c.Containers[types.BinaryScannerContainerName]
	if !ok {
		return nil, fmt.Errorf("couldn't find container %s", types.BinaryScannerContainerName)
	}

	binaryScannerContainerConfig := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "binaryscanner", Image: containerConfig.Image,
			PullPolicy: horizonapi.PullAlways, Command: []string{"/docker-entrypoint.sh"}},
		EnvConfigs: []*horizonapi.EnvConfig{utils.GetHubConfigEnv(c.blackduck.Name)},
		PortConfig: []*horizonapi.PortConfig{{ContainerPort: int32(3001), Protocol: horizonapi.ProtocolTCP}},
	}
	utils2.SetLimits(binaryScannerContainerConfig.ContainerConfig, containerConfig)

	podConfig := &util.PodConfig{
		Containers:          []*util.Container{binaryScannerContainerConfig},
		ImagePullSecrets:    c.blackduck.Spec.RegistryConfiguration.PullSecrets,
		Labels:              apputils.GetVersionLabel("binaryscanner", c.blackduck.Name, c.blackduck.Spec.Version),
		NodeAffinityConfigs: utils.GetNodeAffinityConfigs("binaryscanner", &c.blackduck.Spec),
	}

	if !c.config.IsOpenshift {
		podConfig.FSGID = util.IntToInt64(0)
	}

	return util.CreateReplicationControllerFromContainer(
		&horizonapi.ReplicationControllerConfig{Namespace: c.blackduck.Spec.Namespace, Name: apputils.GetResourceName(c.blackduck.Name, util.BlackDuckName, "binaryscanner"), Replicas: util.IntToInt32(1)},
		podConfig, apputils.GetLabel("binaryscanner", c.blackduck.Name))
}

func NewBdReplicationController(replicationController *types.ReplicationController, config *protoform.Config, kubeClient *kubernetes.Clientset, blackduck *blackduckapi.Blackduck) types.ReplicationControllerInterface {
	return &BdReplicationController{ReplicationController: replicationController, config: config, kubeClient: kubeClient, blackduck: blackduck}
}
