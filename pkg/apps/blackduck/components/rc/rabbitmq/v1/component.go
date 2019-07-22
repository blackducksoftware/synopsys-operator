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
	store.Register(types.RcRabbitmqV1, NewBdReplicationController)
}

func (c *BdReplicationController) GetRc() (*components.ReplicationController, error) {

	containerConfig, ok := c.Containers[types.RabbitMQContainerName]
	if !ok {
		return nil, fmt.Errorf("couldn't find container %s", types.RabbitMQContainerName)
	}

	volumeMounts := c.getRabbitmqVolumeMounts()

	rabbitmqContainerConfig := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "rabbitmq", Image: containerConfig.Image, PullPolicy: horizonapi.PullAlways},
		EnvConfigs:      []*horizonapi.EnvConfig{utils.GetHubConfigEnv(c.blackduck.Name)},
		VolumeMounts:    volumeMounts,
		PortConfig:      []*horizonapi.PortConfig{{ContainerPort: int32(5671), Protocol: horizonapi.ProtocolTCP}},
	}
	utils2.SetLimits(rabbitmqContainerConfig.ContainerConfig, containerConfig)

	podConfig := &util.PodConfig{
		Volumes:             c.getRabbitmqVolumes(),
		Containers:          []*util.Container{rabbitmqContainerConfig},
		ImagePullSecrets:    c.blackduck.Spec.RegistryConfiguration.PullSecrets,
		Labels:              utils2.GetVersionLabel("rabbitmq", c.blackduck.Name, c.blackduck.Spec.Version),
		NodeAffinityConfigs: utils.GetNodeAffinityConfigs("rabbitmq", &c.blackduck.Spec),
	}

	return util.CreateReplicationControllerFromContainer(
		&horizonapi.ReplicationControllerConfig{Namespace: c.blackduck.Spec.Namespace, Name: util.GetResourceName(c.blackduck.Name, util.BlackDuckName, "rabbitmq"), Replicas: util.IntToInt32(1)},
		podConfig, utils2.GetLabel("rabbitmq", c.blackduck.Name))
}

// getRabbitmqVolumes will return the rabbitmq volumes
func (c *BdReplicationController) getRabbitmqVolumes() []*components.Volume {
	rabbitmqSecurityEmptyDir, _ := util.CreateEmptyDirVolumeWithoutSizeLimit("dir-rabbitmq-security")
	volumes := []*components.Volume{rabbitmqSecurityEmptyDir}
	return volumes
}

// getRabbitmqVolumeMounts will return the rabbitmq volume mounts
func (c *BdReplicationController) getRabbitmqVolumeMounts() []*horizonapi.VolumeMountConfig {
	volumesMounts := []*horizonapi.VolumeMountConfig{
		{Name: "dir-rabbitmq-security", MountPath: "/opt/blackduck/rabbitmq/security"},
	}
	return volumesMounts
}

func NewBdReplicationController(replicationController *types.ReplicationController, config *protoform.Config, kubeClient *kubernetes.Clientset, blackduck *blackduckapi.Blackduck) types.ReplicationControllerInterface {
	return &BdReplicationController{ReplicationController: replicationController, config: config, kubeClient: kubeClient, blackduck: blackduck}
}
