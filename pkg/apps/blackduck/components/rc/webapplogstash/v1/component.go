package v1

import (
	"fmt"
	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	blackduckapi "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
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
	store.Register(types.BlackDuckWebappLogstashRCV1, NewBdReplicationController)
}

// GetRc returns the RC
func (c *BdReplicationController) GetRc() (*components.ReplicationController, error) {
	webappConfig, ok := c.Containers[types.WebappContainerName]
	if !ok {
		return nil, fmt.Errorf("couldn't find container %s", types.WebappContainerName)
	}

	logstashConfig, ok := c.Containers[types.LogstashContainerName]
	if !ok {
		return nil, fmt.Errorf("couldn't find container %s", types.LogstashContainerName)
	}

	// hubMaxMemory is the amount of memory allocated to the JVM. We keep 512mb for alpine
	hubMaxMemory := 2048
	if webappConfig.MaxMem != nil && *webappConfig.MaxMem > 512 {
		hubMaxMemory = int(*webappConfig.MaxMem - 512)
	}

	webappEnvs := []*horizonapi.EnvConfig{utils.GetBlackDuckConfigEnv(c.blackDuck.Name), utils.GetBlackDuckDBConfigEnv(c.blackDuck.Name)}
	webappEnvs = append(webappEnvs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "HUB_MAX_MEMORY", KeyOrVal: fmt.Sprintf("%dm", hubMaxMemory)})

	webappVolumeMounts := c.getWebappVolumeMounts()

	webappContainerConfig := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "webapp", Image: webappConfig.Image, PullPolicy: horizonapi.PullAlways},
		EnvConfigs:      webappEnvs,
		VolumeMounts:    webappVolumeMounts,
		PortConfig:      []*horizonapi.PortConfig{{ContainerPort: int32(8443), Protocol: horizonapi.ProtocolTCP}},
	}

	apputils.SetLimits(webappContainerConfig.ContainerConfig, webappConfig)
	if c.blackDuck.Spec.LivenessProbes {
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
		ContainerConfig: &horizonapi.ContainerConfig{Name: "logstash", Image: logstashConfig.Image, PullPolicy: horizonapi.PullAlways},
		EnvConfigs:      []*horizonapi.EnvConfig{utils.GetBlackDuckConfigEnv(c.blackDuck.Name)},
		VolumeMounts:    logstashVolumeMounts,
		PortConfig:      []*horizonapi.PortConfig{{ContainerPort: int32(5044), Protocol: horizonapi.ProtocolTCP}},
	}

	apputils.SetLimits(logstashContainerConfig.ContainerConfig, logstashConfig)

	if c.blackDuck.Spec.LivenessProbes {
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
		ImagePullSecrets:    c.blackDuck.Spec.RegistryConfiguration.PullSecrets,
		Labels:              apputils.GetVersionLabel("webapp-logstash", c.blackDuck.Name, c.blackDuck.Spec.Version),
		NodeAffinityConfigs: utils.GetNodeAffinityConfigs("webapp-logstash", &c.blackDuck.Spec),
	}

	if !c.config.IsOpenshift {
		podConfig.FSGID = util.IntToInt64(0)
	}

	return util.CreateReplicationControllerFromContainer(
		&horizonapi.ReplicationControllerConfig{Namespace: c.blackDuck.Spec.Namespace, Name: apputils.GetResourceName(c.blackDuck.Name, util.BlackDuckName, "webapp-logstash"), Replicas: util.IntToInt32(1)},
		podConfig, apputils.GetLabel("webapp-logstash", c.blackDuck.Name))
}

// getWebappLogtashVolumes will return the webapp and logstash volumes
func (c *BdReplicationController) getWebappLogtashVolumes() []*components.Volume {
	webappSecurityEmptyDir, _ := util.CreateEmptyDirVolumeWithoutSizeLimit("dir-webapp-security")
	var webappVolume *components.Volume
	if c.blackDuck.Spec.PersistentStorage {
		webappVolume, _ = util.CreatePersistentVolumeClaimVolume("dir-webapp", utils.GetPVCName("webapp", c.blackDuck))
	} else {
		webappVolume, _ = util.CreateEmptyDirVolumeWithoutSizeLimit("dir-webapp")
	}

	var logstashVolume *components.Volume
	if c.blackDuck.Spec.PersistentStorage {
		logstashVolume, _ = util.CreatePersistentVolumeClaimVolume("dir-logstash", utils.GetPVCName("logstash", c.blackDuck))
	} else {
		logstashVolume, _ = util.CreateEmptyDirVolumeWithoutSizeLimit("dir-logstash")
	}

	volumes := []*components.Volume{webappSecurityEmptyDir, webappVolume, logstashVolume, utils.GetDBSecretVolume(c.blackDuck.Name)}
	// Mount the HTTPS proxy certificate if provided
	if len(c.blackDuck.Spec.ProxyCertificate) > 0 {
		volumes = append(volumes, utils.GetProxyVolume(c.blackDuck.Name))
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
	if len(c.blackDuck.Spec.ProxyCertificate) > 0 {
		volumesMounts = append(volumesMounts, &horizonapi.VolumeMountConfig{
			Name:      "proxy-certificate",
			MountPath: "/tmp/secrets/HUB_PROXY_CERT_FILE",
			SubPath:   "HUB_PROXY_CERT_FILE",
		})
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
