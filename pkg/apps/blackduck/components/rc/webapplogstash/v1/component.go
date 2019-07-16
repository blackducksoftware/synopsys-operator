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
	store.Register(types.RcWebappLogstashV1, NewBdReplicationController)
}

func (c *BdReplicationController) GetRc() (*components.ReplicationController, error) {

	containerConfig, ok := c.Containers[types.WebappContainerName]
	if !ok {
		return nil, fmt.Errorf("couldn't find container %s", types.WebappContainerName)
	}

	lontainerConfig, ok := c.Containers[types.LogstashContainerName]
	if !ok {
		return nil, fmt.Errorf("couldn't find container %s", types.LogstashContainerName)
	}

	webappEnvs := []*horizonapi.EnvConfig{utils.GetHubConfigEnv(c.blackduck.Name), utils.GetHubDBConfigEnv(c.blackduck.Name)}
	webappEnvs = append(webappEnvs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "HUB_MAX_MEMORY", KeyOrVal: fmt.Sprintf("%dM", containerConfig.MaxMem-512)})

	webappVolumeMounts := c.getWebappVolumeMounts()

	webappContainerConfig := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "webapp", Image: containerConfig.Image,
			PullPolicy: horizonapi.PullAlways, MinMem: fmt.Sprintf("%dM", containerConfig.MinMem), MaxMem: fmt.Sprintf("%dM", containerConfig.MaxMem), MinCPU: fmt.Sprintf("%d", containerConfig.MinCPU), MaxCPU: fmt.Sprintf("%d", containerConfig.MaxCPU)},
		EnvConfigs:   webappEnvs,
		VolumeMounts: webappVolumeMounts,
		PortConfig:   []*horizonapi.PortConfig{{ContainerPort: int32(8443), Protocol: horizonapi.ProtocolTCP}},
	}

	if c.blackduck.Spec.LivenessProbes {
		webappContainerConfig.LivenessProbeConfigs = []*horizonapi.ProbeConfig{{
			ActionConfig: horizonapi.ActionConfig{
				Type: horizonapi.ActionTypeCommand,
				Command: []string{
					"/usr/local/bin/docker-healthcheck.sh",
					"https://127.0.0.1:8443/api/health-checks/liveness",
					"/opt/blackduck/hub/hub-webapp/security/root.crt",
					"/opt/blackduck/hub/hub-webapp/security/blackduck_system.crt",
					"/opt/blackduck/hub/hub-webapp/security/blackduck_system.key",
				},
			},
			Delay:           360,
			Interval:        30,
			Timeout:         10,
			MinCountFailure: 1000,
		}}
	}

	logstashVolumeMounts := c.getLogstashVolumeMounts()

	logstashContainerConfig := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "logstash", Image: lontainerConfig.Image,
			PullPolicy: horizonapi.PullAlways, MinMem: fmt.Sprintf("%dM", lontainerConfig.MinMem), MaxMem: fmt.Sprintf("%dM", lontainerConfig.MaxMem), MinCPU: fmt.Sprintf("%d", lontainerConfig.MinCPU), MaxCPU: fmt.Sprintf("%d", lontainerConfig.MaxCPU)},
		EnvConfigs:   []*horizonapi.EnvConfig{utils.GetHubConfigEnv(c.blackduck.Name)},
		VolumeMounts: logstashVolumeMounts,
		PortConfig:   []*horizonapi.PortConfig{{ContainerPort: int32(5044), Protocol: horizonapi.ProtocolTCP}},
	}

	if c.blackduck.Spec.LivenessProbes {
		logstashContainerConfig.LivenessProbeConfigs = []*horizonapi.ProbeConfig{{
			ActionConfig: horizonapi.ActionConfig{
				Type:    horizonapi.ActionTypeCommand,
				Command: []string{"/usr/local/bin/docker-healthcheck.sh", "http://localhost:9600/"},
			},
			Delay:           240,
			Interval:        30,
			Timeout:         10,
			MinCountFailure: 1000,
		}}
	}
	podConfig := &util.PodConfig{
		Volumes:             c.getWebappLogtashVolumes(),
		Containers:          []*util.Container{webappContainerConfig, logstashContainerConfig},
		ImagePullSecrets:    c.blackduck.Spec.RegistryConfiguration.PullSecrets,
		Labels:              utils2.GetVersionLabel("webapp-logstash", c.blackduck.Name, c.blackduck.Spec.Version),
		NodeAffinityConfigs: utils.GetNodeAffinityConfigs("webapp-logstash", &c.blackduck.Spec),
	}

	if !c.config.IsOpenshift {
		podConfig.FSGID = util.IntToInt64(0)
	}

	return util.CreateReplicationControllerFromContainer(
		&horizonapi.ReplicationControllerConfig{Namespace: c.blackduck.Spec.Namespace, Name: util.GetResourceName(c.blackduck.Name, util.BlackDuckName, "webapp-logstash"), Replicas: util.IntToInt32(1)},
		podConfig, utils2.GetLabel("webapp-logstash", c.blackduck.Name))
}

// getWebappLogtashVolumes will return the webapp and logstash volumes
func (c *BdReplicationController) getWebappLogtashVolumes() []*components.Volume {
	webappSecurityEmptyDir, _ := util.CreateEmptyDirVolumeWithoutSizeLimit("dir-webapp-security")
	var webappVolume *components.Volume
	if c.blackduck.Spec.PersistentStorage {
		webappVolume, _ = util.CreatePersistentVolumeClaimVolume("dir-webapp", utils.GetPVCName("webapp", c.blackduck))
	} else {
		webappVolume, _ = util.CreateEmptyDirVolumeWithoutSizeLimit("dir-webapp")
	}

	var logstashVolume *components.Volume
	if c.blackduck.Spec.PersistentStorage {
		logstashVolume, _ = util.CreatePersistentVolumeClaimVolume("dir-logstash", utils.GetPVCName("logstash", c.blackduck))
	} else {
		logstashVolume, _ = util.CreateEmptyDirVolumeWithoutSizeLimit("dir-logstash")
	}

	volumes := []*components.Volume{webappSecurityEmptyDir, webappVolume, logstashVolume, utils.GetDBSecretVolume(c.blackduck.Name)}
	// Mount the HTTPS proxy certificate if provided
	if len(c.blackduck.Spec.ProxyCertificate) > 0 {
		volumes = append(volumes, utils.GetProxyVolume(c.blackduck.Name))
	}

	return volumes
}

// getLogstashVolumeMounts will return the Logstash volume mounts
func (c *BdReplicationController) getLogstashVolumeMounts() []*horizonapi.VolumeMountConfig {
	volumesMounts := []*horizonapi.VolumeMountConfig{
		{Name: "dir-logstash", MountPath: "/var/lib/logstash/data"},
	}
	return volumesMounts
}

// getWebappVolumeMounts will return the Webapp volume mounts
func (c *BdReplicationController) getWebappVolumeMounts() []*horizonapi.VolumeMountConfig {
	volumesMounts := []*horizonapi.VolumeMountConfig{
		{Name: "db-passwords", MountPath: "/tmp/secrets/HUB_POSTGRES_ADMIN_PASSWORD_FILE", SubPath: "HUB_POSTGRES_ADMIN_PASSWORD_FILE"},
		{Name: "db-passwords", MountPath: "/tmp/secrets/HUB_POSTGRES_USER_PASSWORD_FILE", SubPath: "HUB_POSTGRES_USER_PASSWORD_FILE"},
		{Name: "dir-webapp", MountPath: "/opt/blackduck/hub/hub-webapp/ldap"},
		{Name: "dir-webapp-security", MountPath: "/opt/blackduck/hub/hub-webapp/security"},
		{Name: "dir-logstash", MountPath: "/opt/blackduck/hub/logs"},
	}

	// Mount the HTTPS proxy certificate if provided
	if len(c.blackduck.Spec.ProxyCertificate) > 0 {
		volumesMounts = append(volumesMounts, &horizonapi.VolumeMountConfig{
			Name:      "proxy-certificate",
			MountPath: "/tmp/secrets/HUB_PROXY_CERT_FILE",
			SubPath:   "HUB_PROXY_CERT_FILE",
		})
	}

	return volumesMounts
}

func NewBdReplicationController(replicationController *types.ReplicationController, config *protoform.Config, kubeClient *kubernetes.Clientset, blackduck *blackduckapi.Blackduck) types.ReplicationControllerInterface {
	return &BdReplicationController{ReplicationController: replicationController, config: config, kubeClient: kubeClient, blackduck: blackduck}
}
