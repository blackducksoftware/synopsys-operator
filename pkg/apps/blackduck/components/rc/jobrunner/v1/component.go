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
	store.Register(blackduck.BlackDuckJobRunnerRCV1, NewBdReplicationController)
}

// GetRc returns the RC
func (c *BdReplicationController) GetRc() (*components.ReplicationController, error) {
	containerConfig, ok := c.Containers[blackduck.JobrunnerContainerName]
	if !ok {
		return nil, fmt.Errorf("couldn't find container %s", blackduck.JobrunnerContainerName)
	}

	if containerConfig.MaxMem == nil {
		return nil, fmt.Errorf("Maxmem must be set for %s", blackduck.JobrunnerContainerName)
	}

	jobRunnerEmptyDir, _ := util.CreateEmptyDirVolumeWithoutSizeLimit("dir-jobrunner")
	jobRunnerEnvs := []*horizonapi.EnvConfig{utils.GetBlackDuckConfigEnv(c.blackDuck.Name), utils.GetBlackDuckDBConfigEnv(c.blackDuck.Name)}
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

	apputils.SetLimits(jobRunnerContainerConfig.ContainerConfig, containerConfig)

	if c.blackDuck.Spec.LivenessProbes {
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

	jobRunnerVolumes := []*components.Volume{utils.GetDBSecretVolume(c.blackDuck.Name), jobRunnerEmptyDir}

	// Mount the HTTPS proxy certificate if provided
	if len(c.blackDuck.Spec.ProxyCertificate) > 0 {
		jobRunnerContainerConfig.VolumeMounts = append(jobRunnerContainerConfig.VolumeMounts, &horizonapi.VolumeMountConfig{
			Name:      "proxy-certificate",
			MountPath: "/tmp/secrets/HUB_PROXY_CERT_FILE",
			SubPath:   "HUB_PROXY_CERT_FILE",
		})
		jobRunnerVolumes = append(jobRunnerVolumes, utils.GetProxyVolume(c.blackDuck.Name))
	}

	podConfig := &util.PodConfig{
		Volumes:             jobRunnerVolumes,
		Containers:          []*util.Container{jobRunnerContainerConfig},
		ImagePullSecrets:    c.blackDuck.Spec.RegistryConfiguration.PullSecrets,
		Labels:              apputils.GetVersionLabel("jobrunner", c.blackDuck.Name, c.blackDuck.Spec.Version),
		NodeAffinityConfigs: utils.GetNodeAffinityConfigs("jobrunner", &c.blackDuck.Spec),
	}
	if !c.config.IsOpenshift {
		podConfig.FSGID = util.IntToInt64(0)
	}
	return util.CreateReplicationControllerFromContainer(
		&horizonapi.ReplicationControllerConfig{Namespace: c.blackDuck.Spec.Namespace, Name: apputils.GetResourceName(c.blackDuck.Name, util.BlackDuckName, "jobrunner"), Replicas: util.IntToInt32(c.Replicas)},
		podConfig, apputils.GetLabel("jobrunner", c.blackDuck.Name))
}

// NewBdReplicationController returns the Black Duck RC configuration
func NewBdReplicationController(replicationController *types.ReplicationController, config *protoform.Config, kubeClient *kubernetes.Clientset, cr interface{}) (types.ReplicationControllerInterface, error) {
	blackDuck, ok := cr.(*blackduckapi.Blackduck)
	if !ok {
		return nil, fmt.Errorf("unable to cast the interface to Black Duck object")
	}
	return &BdReplicationController{ReplicationController: replicationController, config: config, kubeClient: kubeClient, blackDuck: blackDuck}, nil
}
