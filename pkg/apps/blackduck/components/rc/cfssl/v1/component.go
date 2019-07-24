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
	store.Register(types.RcCfsslV1, NewBdReplicationController)
}

func (c *BdReplicationController) GetRc() (*components.ReplicationController, error) {

	containerConfig, ok := c.Containers[types.CfsslContainerName]
	if !ok {
		return nil, fmt.Errorf("couldn't find container %s", types.CfsslContainerName)
	}

	cfsslVolumeMounts := c.getCfsslolumeMounts()
	cfsslContainerConfig := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "cfssl", Image: containerConfig.Image,
			PullPolicy: horizonapi.PullAlways},
		EnvConfigs:   []*horizonapi.EnvConfig{utils.GetHubConfigEnv(c.blackduck.Name)},
		VolumeMounts: cfsslVolumeMounts,
		PortConfig:   []*horizonapi.PortConfig{{ContainerPort: int32(8888), Protocol: horizonapi.ProtocolTCP}},
	}

	utils2.SetLimits(cfsslContainerConfig.ContainerConfig, containerConfig)

	if c.blackduck.Spec.LivenessProbes {
		cfsslContainerConfig.LivenessProbeConfigs = []*horizonapi.ProbeConfig{{
			ActionConfig: horizonapi.ActionConfig{
				Type:    horizonapi.ActionTypeCommand,
				Command: []string{"/usr/local/bin/docker-healthcheck.sh", "http://localhost:8888/api/v1/cfssl/scaninfo"},
			},
			Delay:           240,
			Interval:        30,
			Timeout:         10,
			MinCountFailure: 10,
		}}
	}

	podConfig := &util.PodConfig{
		Volumes:             c.getCfsslVolumes(),
		Containers:          []*util.Container{cfsslContainerConfig},
		ImagePullSecrets:    c.blackduck.Spec.RegistryConfiguration.PullSecrets,
		Labels:              apputils.GetVersionLabel("cfssl", c.blackduck.Name, c.blackduck.Spec.Version),
		NodeAffinityConfigs: utils.GetNodeAffinityConfigs("cfssl", &c.blackduck.Spec),
	}

	if !c.config.IsOpenshift {
		podConfig.FSGID = util.IntToInt64(0)
	}

	return util.CreateReplicationControllerFromContainer(
		&horizonapi.ReplicationControllerConfig{Namespace: c.blackduck.Spec.Namespace, Name: apputils.GetResourceName(c.blackduck.Name, util.BlackDuckName, "cfssl"), Replicas: util.IntToInt32(1)},
		podConfig, apputils.GetLabel("cfssl", c.blackduck.Name))
}

// getCfsslVolumes will return the cfssl volumes
func (c *BdReplicationController) getCfsslVolumes() []*components.Volume {
	var cfsslVolume *components.Volume
	if c.blackduck.Spec.PersistentStorage {
		cfsslVolume, _ = util.CreatePersistentVolumeClaimVolume("dir-cfssl", utils.GetPVCName("cfssl", c.blackduck))
	} else {
		cfsslVolume, _ = util.CreateEmptyDirVolumeWithoutSizeLimit("dir-cfssl")
	}

	volumes := []*components.Volume{cfsslVolume}
	return volumes
}

// getCfsslolumeMounts will return the cfssl volume mounts
func (c *BdReplicationController) getCfsslolumeMounts() []*horizonapi.VolumeMountConfig {
	volumesMounts := []*horizonapi.VolumeMountConfig{
		{Name: "dir-cfssl", MountPath: "/etc/cfssl"},
	}
	return volumesMounts
}

func NewBdReplicationController(replicationController *types.ReplicationController, config *protoform.Config, kubeClient *kubernetes.Clientset, blackduck *blackduckapi.Blackduck) types.ReplicationControllerInterface {
	return &BdReplicationController{ReplicationController: replicationController, config: config, kubeClient: kubeClient, blackduck: blackduck}
}
