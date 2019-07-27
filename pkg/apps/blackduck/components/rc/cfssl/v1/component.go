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
	store.Register(blackduck.BlackDuckCfsslRCV1, NewBdReplicationController)
}

// GetRc returns the RC
func (c *BdReplicationController) GetRc() (*components.ReplicationController, error) {
	containerConfig, ok := c.Containers[blackduck.CfsslContainerName]
	if !ok {
		return nil, fmt.Errorf("couldn't find container %s", blackduck.CfsslContainerName)
	}

	cfsslVolumeMounts := c.getCfsslolumeMounts()
	cfsslContainerConfig := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "cfssl", Image: containerConfig.Image,
			PullPolicy: horizonapi.PullAlways},
		EnvConfigs:   []*horizonapi.EnvConfig{utils.GetBlackDuckConfigEnv(c.blackDuck.Name)},
		VolumeMounts: cfsslVolumeMounts,
		PortConfig:   []*horizonapi.PortConfig{{ContainerPort: int32(8888), Protocol: horizonapi.ProtocolTCP}},
	}

	apputils.SetLimits(cfsslContainerConfig.ContainerConfig, containerConfig)

	if c.blackDuck.Spec.LivenessProbes {
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
		ImagePullSecrets:    c.blackDuck.Spec.RegistryConfiguration.PullSecrets,
		Labels:              apputils.GetVersionLabel("cfssl", c.blackDuck.Name, c.blackDuck.Spec.Version),
		NodeAffinityConfigs: utils.GetNodeAffinityConfigs("cfssl", &c.blackDuck.Spec),
	}

	if !c.config.IsOpenshift {
		podConfig.FSGID = util.IntToInt64(0)
	}

	return util.CreateReplicationControllerFromContainer(
		&horizonapi.ReplicationControllerConfig{Namespace: c.blackDuck.Spec.Namespace, Name: apputils.GetResourceName(c.blackDuck.Name, util.BlackDuckName, "cfssl"), Replicas: util.IntToInt32(1)},
		podConfig, apputils.GetLabel("cfssl", c.blackDuck.Name))
}

// getCfsslVolumes will return the cfssl volumes
func (c *BdReplicationController) getCfsslVolumes() []*components.Volume {
	var cfsslVolume *components.Volume
	if c.blackDuck.Spec.PersistentStorage {
		cfsslVolume, _ = util.CreatePersistentVolumeClaimVolume("dir-cfssl", utils.GetPVCName("cfssl", c.blackDuck))
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

// NewBdReplicationController returns the Black Duck RC configuration
func NewBdReplicationController(replicationController *types.ReplicationController, config *protoform.Config, kubeClient *kubernetes.Clientset, cr interface{}) (types.ReplicationControllerInterface, error) {
	blackDuck, ok := cr.(*blackduckapi.Blackduck)
	if !ok {
		return nil, fmt.Errorf("unable to cast the interface to Black Duck object")
	}
	return &BdReplicationController{ReplicationController: replicationController, config: config, kubeClient: kubeClient, blackDuck: blackDuck}, nil
}
