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
	store.Register(types.RcJobrunnerV1, NewBdReplicationController)
}

func (c *BdReplicationController) GetRc() (*components.ReplicationController, error) {

	containerConfig, ok := c.Containers[types.JobrunnerContainerName]
	if !ok {
		return nil, fmt.Errorf("couldn't find container %s", types.JobrunnerContainerName)
	}

	if containerConfig.MaxMem == nil {
		return nil, fmt.Errorf("Maxmem must be set for %s", types.JobrunnerContainerName)
	}

	jobRunnerEmptyDir, _ := util.CreateEmptyDirVolumeWithoutSizeLimit("dir-jobrunner")
	jobRunnerEnvs := []*horizonapi.EnvConfig{utils.GetHubConfigEnv(c.blackduck.Name), utils.GetHubDBConfigEnv(c.blackduck.Name)}
	jobRunnerEnvs = append(jobRunnerEnvs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "HUB_MAX_MEMORY", KeyOrVal: fmt.Sprintf("%dM", *containerConfig.MaxMem-512)})
	jobRunnerContainerConfig := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "jobrunner", Image: containerConfig.Image, PullPolicy: horizonapi.PullAlways},
		EnvConfigs:      jobRunnerEnvs,
		VolumeMounts: []*horizonapi.VolumeMountConfig{
			{Name: "db-passwords", MountPath: "/tmp/secrets/HUB_POSTGRES_ADMIN_PASSWORD_FILE", SubPath: "HUB_POSTGRES_ADMIN_PASSWORD_FILE"},
			{Name: "db-passwords", MountPath: "/tmp/secrets/HUB_POSTGRES_USER_PASSWORD_FILE", SubPath: "HUB_POSTGRES_USER_PASSWORD_FILE"},
			{Name: "dir-jobrunner", MountPath: "/opt/blackduck/hub/jobrunner/security"},
		},
		PortConfig: []*horizonapi.PortConfig{{ContainerPort: int32(3001), Protocol: horizonapi.ProtocolTCP}},
	}

	utils2.SetLimits(jobRunnerContainerConfig.ContainerConfig, containerConfig)

	if c.blackduck.Spec.LivenessProbes {
		jobRunnerContainerConfig.LivenessProbeConfigs = []*horizonapi.ProbeConfig{{
			ActionConfig: horizonapi.ActionConfig{
				Type:    horizonapi.ActionTypeCommand,
				Command: []string{"/usr/local/bin/docker-healthcheck.sh"},
			},
			Delay:           240,
			Interval:        30,
			Timeout:         10,
			MinCountFailure: 10,
		}}
	}

	jobRunnerVolumes := []*components.Volume{utils.GetDBSecretVolume(c.blackduck.Name), jobRunnerEmptyDir}

	// Mount the HTTPS proxy certificate if provided
	if len(c.blackduck.Spec.ProxyCertificate) > 0 {
		jobRunnerContainerConfig.VolumeMounts = append(jobRunnerContainerConfig.VolumeMounts, &horizonapi.VolumeMountConfig{
			Name:      "proxy-certificate",
			MountPath: "/tmp/secrets/HUB_PROXY_CERT_FILE",
			SubPath:   "HUB_PROXY_CERT_FILE",
		})
		jobRunnerVolumes = append(jobRunnerVolumes, utils.GetProxyVolume(c.blackduck.Name))
	}

	podConfig := &util.PodConfig{
		Volumes:             jobRunnerVolumes,
		Containers:          []*util.Container{jobRunnerContainerConfig},
		ImagePullSecrets:    c.blackduck.Spec.RegistryConfiguration.PullSecrets,
		Labels:              utils2.GetVersionLabel("jobrunner", c.blackduck.Name, c.blackduck.Spec.Version),
		NodeAffinityConfigs: utils.GetNodeAffinityConfigs("jobrunner", &c.blackduck.Spec),
	}
	if !c.config.IsOpenshift {
		podConfig.FSGID = util.IntToInt64(0)
	}
	return util.CreateReplicationControllerFromContainer(
		&horizonapi.ReplicationControllerConfig{Namespace: c.blackduck.Spec.Namespace, Name: util.GetResourceName(c.blackduck.Name, util.BlackDuckName, "jobrunner"), Replicas: util.IntToInt32(c.Replicas)},
		podConfig, utils2.GetLabel("jobrunner", c.blackduck.Name))
}
func NewBdReplicationController(replicationController *types.ReplicationController, config *protoform.Config, kubeClient *kubernetes.Clientset, blackduck *blackduckapi.Blackduck) types.ReplicationControllerInterface {
	return &BdReplicationController{ReplicationController: replicationController, config: config, kubeClient: kubeClient, blackduck: blackduck}
}
