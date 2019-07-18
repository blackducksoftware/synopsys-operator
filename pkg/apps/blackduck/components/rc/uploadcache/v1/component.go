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
	store.Register(types.RcUploadCacheV1, NewBdReplicationController)
}

func (c *BdReplicationController) GetRc() (*components.ReplicationController, error) {

	containerConfig, ok := c.Containers[types.UploadCacheContainerName]
	if !ok {
		return nil, fmt.Errorf("couldn't find container %s", types.UploadCacheContainerName)
	}
	volumeMounts := c.getUploadCacheVolumeMounts()

	uploadCacheContainerConfig := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "uploadcache", Image: containerConfig.Image,
			PullPolicy: horizonapi.PullAlways, MinMem: fmt.Sprintf("%dM", containerConfig.MinMem), MaxMem: fmt.Sprintf("%dM", containerConfig.MaxMem), MinCPU: fmt.Sprintf("%d", containerConfig.MinCPU), MaxCPU: fmt.Sprintf("%d", containerConfig.MaxCPU)},
		EnvConfigs:   []*horizonapi.EnvConfig{utils.GetHubConfigEnv(c.blackduck.Name)},
		VolumeMounts: volumeMounts,
		PortConfig: []*horizonapi.PortConfig{{ContainerPort: int32(9443), Protocol: horizonapi.ProtocolTCP},
			{ContainerPort: int32(9444), Protocol: horizonapi.ProtocolTCP}},
	}

	if c.blackduck.Spec.LivenessProbes {
		uploadCacheContainerConfig.LivenessProbeConfigs = []*horizonapi.ProbeConfig{{
			ActionConfig: horizonapi.ActionConfig{
				Type:    horizonapi.ActionTypeCommand,
				Command: []string{"curl", "--insecure", "-X", "GET", "--verbose", "http://localhost:8086/live?full=1"},
			},
			Delay:           240,
			Interval:        30,
			Timeout:         10,
			MinCountFailure: 5,
		}}
	}
	podConfig := &util.PodConfig{
		Volumes:             c.getUploadCacheVolumes(),
		Containers:          []*util.Container{uploadCacheContainerConfig},
		ImagePullSecrets:    c.blackduck.Spec.RegistryConfiguration.PullSecrets,
		Labels:              utils2.GetVersionLabel("uploadcache", c.blackduck.Name, c.blackduck.Spec.Version),
		NodeAffinityConfigs: utils.GetNodeAffinityConfigs("uploadcache", &c.blackduck.Spec),
	}

	if !c.config.IsOpenshift {
		podConfig.FSGID = util.IntToInt64(0)
	}

	return util.CreateReplicationControllerFromContainer(
		&horizonapi.ReplicationControllerConfig{Namespace: c.blackduck.Spec.Namespace, Name: util.GetResourceName(c.blackduck.Name, util.BlackDuckName, "uploadcache"), Replicas: util.IntToInt32(1)},
		podConfig, utils2.GetLabel("uploadcache", c.blackduck.Name))
}

// getUploadCacheVolumes will return the uploadCache volumes
func (c *BdReplicationController) getUploadCacheVolumes() []*components.Volume {
	uploadCacheSecurityEmptyDir, _ := util.CreateEmptyDirVolumeWithoutSizeLimit("dir-uploadcache-security")
	sealKeySecretVol, _ := util.CreateSecretVolume("dir-seal-key", util.GetResourceName(c.blackduck.Name, util.BlackDuckName, "upload-cache"), 0444)
	var uploadCacheDataDir *components.Volume
	if c.blackduck.Spec.PersistentStorage {
		uploadCacheDataDir, _ = util.CreatePersistentVolumeClaimVolume("dir-uploadcache-data", utils.GetPVCName("uploadcache-data", c.blackduck))
	} else {
		uploadCacheDataDir, _ = util.CreateEmptyDirVolumeWithoutSizeLimit("dir-uploadcache-data")
	}
	volumes := []*components.Volume{uploadCacheSecurityEmptyDir, uploadCacheDataDir, sealKeySecretVol}
	return volumes
}

// getUploadCacheVolumeMounts will return the uploadCache volume mounts
func (c *BdReplicationController) getUploadCacheVolumeMounts() []*horizonapi.VolumeMountConfig {
	volumesMounts := []*horizonapi.VolumeMountConfig{
		{Name: "dir-uploadcache-security", MountPath: "/opt/blackduck/hub/blackduck-upload-cache/security"},
		{Name: "dir-uploadcache-data", MountPath: "/opt/blackduck/hub/blackduck-upload-cache/uploads", SubPath: "uploads"},
		{Name: "dir-uploadcache-data", MountPath: "/opt/blackduck/hub/blackduck-upload-cache/keys", SubPath: "keys"},
		{Name: "dir-seal-key", MountPath: "/tmp/secrets"},
	}
	return volumesMounts
}

func NewBdReplicationController(replicationController *types.ReplicationController, config *protoform.Config, kubeClient *kubernetes.Clientset, blackduck *blackduckapi.Blackduck) types.ReplicationControllerInterface {
	return &BdReplicationController{ReplicationController: replicationController, config: config, kubeClient: kubeClient, blackduck: blackduck}
}
