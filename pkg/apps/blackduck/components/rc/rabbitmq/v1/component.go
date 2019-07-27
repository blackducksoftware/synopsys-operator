package v1

import (
	"fmt"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	blackduckapi "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/components/rc/utils"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/store"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/types"
	apputils "github.com/blackducksoftware/synopsys-operator/pkg/apps/utils"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"k8s.io/client-go/kubernetes"
)

// BdReplicationController holds the Black Duck RC configuration
type BdReplicationController struct {
	*types.ReplicationController
	config     *protoform.Config
	kubeClient *kubernetes.Clientset
	blackDuck  *blackduckapi.Blackduck
}

func init() {
	store.Register(blackduck.BlackDuckRabbitMQRCV1, NewBdReplicationController)
}

// GetRc returns the RC
func (c *BdReplicationController) GetRc() (*components.ReplicationController, error) {
	containerConfig, ok := c.Containers[blackduck.RabbitMQContainerName]
	if !ok {
		return nil, fmt.Errorf("couldn't find container %s", blackduck.RabbitMQContainerName)
	}

	volumeMounts := c.getRabbitmqVolumeMounts()

	rabbitmqContainerConfig := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "rabbitmq", Image: containerConfig.Image, PullPolicy: horizonapi.PullAlways},
		EnvConfigs:      []*horizonapi.EnvConfig{utils.GetBlackDuckConfigEnv(c.blackDuck.Name)},
		VolumeMounts:    volumeMounts,
		PortConfig:      []*horizonapi.PortConfig{{ContainerPort: int32(5671), Protocol: horizonapi.ProtocolTCP}},
	}
	apputils.SetLimits(rabbitmqContainerConfig.ContainerConfig, containerConfig)

	podConfig := &util.PodConfig{
		Volumes:             c.getRabbitmqVolumes(),
		Containers:          []*util.Container{rabbitmqContainerConfig},
		ImagePullSecrets:    c.blackDuck.Spec.RegistryConfiguration.PullSecrets,
		Labels:              apputils.GetVersionLabel("rabbitmq", c.blackDuck.Name, c.blackDuck.Spec.Version),
		NodeAffinityConfigs: utils.GetNodeAffinityConfigs("rabbitmq", &c.blackDuck.Spec),
	}

	return util.CreateReplicationControllerFromContainer(
		&horizonapi.ReplicationControllerConfig{Namespace: c.blackDuck.Spec.Namespace, Name: apputils.GetResourceName(c.blackDuck.Name, util.BlackDuckName, "rabbitmq"), Replicas: util.IntToInt32(1)},
		podConfig, apputils.GetLabel("rabbitmq", c.blackDuck.Name))
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

// NewBdReplicationController returns the Black Duck RC configuration
func NewBdReplicationController(replicationController *types.ReplicationController, config *protoform.Config, kubeClient *kubernetes.Clientset, cr interface{}) (types.ReplicationControllerInterface, error) {
	blackDuck, ok := cr.(*blackduckapi.Blackduck)
	if !ok {
		return nil, fmt.Errorf("unable to cast the interface to Black Duck object")
	}
	return &BdReplicationController{ReplicationController: replicationController, config: config, kubeClient: kubeClient, blackDuck: blackDuck}, nil
}
