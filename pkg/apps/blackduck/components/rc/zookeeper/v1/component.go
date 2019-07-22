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
	store.Register(types.RcZookeeperV1, NewBdReplicationController)
}

func (c *BdReplicationController) GetRc() (*components.ReplicationController, error) {

	containerConfig, ok := c.Containers[types.ZookeeperContainerName]
	if !ok {
		return nil, fmt.Errorf("couldn't find container %s", types.ZookeeperContainerName)
	}

	volumeMounts := c.getZookeeperVolumeMounts()

	zookeeperContainerConfig := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "zookeeper", Image: containerConfig.Image, PullPolicy: horizonapi.PullAlways},
		EnvConfigs:      []*horizonapi.EnvConfig{utils.GetHubConfigEnv(c.blackduck.Name)},
		VolumeMounts:    volumeMounts,
		PortConfig:      []*horizonapi.PortConfig{{ContainerPort: int32(2181), Protocol: horizonapi.ProtocolTCP}},
	}

	utils2.SetLimits(zookeeperContainerConfig.ContainerConfig, containerConfig)
	if c.blackduck.Spec.LivenessProbes {
		zookeeperContainerConfig.LivenessProbeConfigs = []*horizonapi.ProbeConfig{{
			ActionConfig: horizonapi.ActionConfig{
				Type:    horizonapi.ActionTypeCommand,
				Command: []string{"zkServer.sh", "status", "/opt/blackduck/zookeeper/conf/zoo.cfg"},
			},
			Delay:           240,
			Interval:        30,
			Timeout:         10,
			MinCountFailure: 10,
		}}
	}

	podConfig := &util.PodConfig{
		Volumes:             c.getZookeeperVolumes(),
		Containers:          []*util.Container{zookeeperContainerConfig},
		ImagePullSecrets:    c.blackduck.Spec.RegistryConfiguration.PullSecrets,
		Labels:              utils2.GetVersionLabel("zookeeper", c.blackduck.Name, c.blackduck.Spec.Version),
		NodeAffinityConfigs: utils.GetNodeAffinityConfigs("zookeeper", &c.blackduck.Spec),
	}

	if !c.config.IsOpenshift {
		podConfig.FSGID = util.IntToInt64(0)
	}

	return util.CreateReplicationControllerFromContainer(
		&horizonapi.ReplicationControllerConfig{Namespace: c.blackduck.Spec.Namespace, Name: util.GetResourceName(c.blackduck.Name, util.BlackDuckName, "zookeeper"), Replicas: util.IntToInt32(1)},
		podConfig, utils2.GetLabel("zookeeper", c.blackduck.Name))
}

// getZookeeperVolumes will return the zookeeper volumes
func (c *BdReplicationController) getZookeeperVolumes() []*components.Volume {
	var zookeeperVolume *components.Volume

	if c.blackduck.Spec.PersistentStorage {
		zookeeperVolume, _ = util.CreatePersistentVolumeClaimVolume("dir-zookeeper", utils.GetPVCName("zookeeper", c.blackduck))
	} else {
		zookeeperVolume, _ = util.CreateEmptyDirVolumeWithoutSizeLimit("dir-zookeeper")
	}

	volumes := []*components.Volume{zookeeperVolume}
	return volumes
}

// getZookeeperVolumeMounts will return the zookeeper volume mounts
func (c *BdReplicationController) getZookeeperVolumeMounts() []*horizonapi.VolumeMountConfig {
	volumesMounts := []*horizonapi.VolumeMountConfig{
		{Name: "dir-zookeeper", MountPath: "/opt/blackduck/zookeeper/data", SubPath: "data"},
		{Name: "dir-zookeeper", MountPath: "/opt/blackduck/zookeeper/datalog", SubPath: "datalog"},
	}
	return volumesMounts
}

func NewBdReplicationController(replicationController *types.ReplicationController, config *protoform.Config, kubeClient *kubernetes.Clientset, blackduck *blackduckapi.Blackduck) types.ReplicationControllerInterface {
	return &BdReplicationController{ReplicationController: replicationController, config: config, kubeClient: kubeClient, blackduck: blackduck}
}
