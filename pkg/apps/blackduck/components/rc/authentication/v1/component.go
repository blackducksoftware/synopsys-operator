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
	store.Register(types.RcAuthenticationV1, NewBdReplicationController)
}

func (c *BdReplicationController) GetRc() (*components.ReplicationController, error) {

	containerConfig, ok := c.Containers[types.AuthenticationContainerName]
	if !ok {
		return nil, fmt.Errorf("couldn't find container %s", types.AuthenticationContainerName)
	}

	if containerConfig.MaxMem == nil {
		return nil, fmt.Errorf("Maxmem must be set for %s", types.AuthenticationContainerName)
	}

	volumeMounts := c.getAuthenticationVolumeMounts()
	var authEnvs []*horizonapi.EnvConfig
	authEnvs = append(authEnvs, utils.GetHubDBConfigEnv(c.blackduck.Name))
	authEnvs = append(authEnvs, utils.GetHubConfigEnv(c.blackduck.Name))
	authEnvs = append(authEnvs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "HUB_MAX_MEMORY", KeyOrVal: fmt.Sprintf("%dM", *containerConfig.MaxMem-512)})
	hubAuthContainerConfig := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "authentication", Image: containerConfig.Image, PullPolicy: horizonapi.PullAlways},
		EnvConfigs:      authEnvs,
		VolumeMounts:    volumeMounts,
		PortConfig:      []*horizonapi.PortConfig{{ContainerPort: int32(8443), Protocol: horizonapi.ProtocolTCP}},
	}

	utils2.SetLimits(hubAuthContainerConfig.ContainerConfig, containerConfig)

	if c.blackduck.Spec.LivenessProbes {
		hubAuthContainerConfig.LivenessProbeConfigs = []*horizonapi.ProbeConfig{{
			ActionConfig: horizonapi.ActionConfig{
				Type: horizonapi.ActionTypeCommand,
				Command: []string{
					"/usr/local/bin/docker-healthcheck.sh",
					"https://127.0.0.1:8443/api/health-checks/liveness",
					"/opt/blackduck/hub/hub-authentication/security/root.crt",
					"/opt/blackduck/hub/hub-authentication/security/blackduck_system.crt",
					"/opt/blackduck/hub/hub-authentication/security/blackduck_system.key",
				},
			},
			Delay:           240,
			Interval:        30,
			Timeout:         10,
			MinCountFailure: 10,
		}}
	}

	podConfig := &util.PodConfig{
		Volumes:             c.getAuthenticationVolumes(),
		Containers:          []*util.Container{hubAuthContainerConfig},
		ImagePullSecrets:    c.blackduck.Spec.RegistryConfiguration.PullSecrets,
		Labels:              utils2.GetVersionLabel("authentication", c.blackduck.Name, c.blackduck.Spec.Version),
		NodeAffinityConfigs: utils.GetNodeAffinityConfigs("authentication", &c.blackduck.Spec),
	}

	if !c.config.IsOpenshift {
		podConfig.FSGID = util.IntToInt64(0)
	}

	return util.CreateReplicationControllerFromContainer(
		&horizonapi.ReplicationControllerConfig{Namespace: c.blackduck.Spec.Namespace, Name: util.GetResourceName(c.blackduck.Name, util.BlackDuckName, "authentication"), Replicas: util.IntToInt32(1)},
		podConfig, utils2.GetLabel("authentication", c.blackduck.Name))
}

// getAuthenticationVolumes will return the authentication volumes
func (c *BdReplicationController) getAuthenticationVolumes() []*components.Volume {
	hubAuthSecurityEmptyDir, _ := util.CreateEmptyDirVolumeWithoutSizeLimit("dir-authentication-security")

	var hubAuthVolume *components.Volume
	if c.blackduck.Spec.PersistentStorage {
		hubAuthVolume, _ = util.CreatePersistentVolumeClaimVolume("dir-authentication", utils.GetPVCName("authentication", c.blackduck))
	} else {
		hubAuthVolume, _ = util.CreateEmptyDirVolumeWithoutSizeLimit("dir-authentication")
	}

	volumes := []*components.Volume{hubAuthVolume, utils.GetDBSecretVolume(c.blackduck.Name), hubAuthSecurityEmptyDir}

	// Mount the HTTPS proxy certificate if provided
	if len(c.blackduck.Spec.ProxyCertificate) > 0 {
		volumes = append(volumes, utils.GetProxyVolume(c.blackduck.Name))
	}

	// Custom CA auth
	if len(c.blackduck.Spec.AuthCustomCA) > 1 {
		authCustomCaVolume, _ := util.CreateSecretVolume("auth-custom-ca", "auth-custom-ca", 0444)
		volumes = append(volumes, authCustomCaVolume)
	}
	return volumes
}

// getAuthenticationVolumeMounts will return the authentication volume mounts
func (c *BdReplicationController) getAuthenticationVolumeMounts() []*horizonapi.VolumeMountConfig {
	volumesMounts := []*horizonapi.VolumeMountConfig{
		{Name: "db-passwords", MountPath: "/tmp/secrets/HUB_POSTGRES_ADMIN_PASSWORD_FILE", SubPath: "HUB_POSTGRES_ADMIN_PASSWORD_FILE"},
		{Name: "db-passwords", MountPath: "/tmp/secrets/HUB_POSTGRES_USER_PASSWORD_FILE", SubPath: "HUB_POSTGRES_USER_PASSWORD_FILE"},
		{Name: "dir-authentication", MountPath: "/opt/blackduck/hub/hub-authentication/ldap"},
		{Name: "dir-authentication-security", MountPath: "/opt/blackduck/hub/hub-authentication/security"},
	}

	// Mount the HTTPS proxy certificate if provided
	if len(c.blackduck.Spec.ProxyCertificate) > 0 {
		volumesMounts = append(volumesMounts, &horizonapi.VolumeMountConfig{
			Name:      "blackduck-proxy-certificate",
			MountPath: "/tmp/secrets/HUB_PROXY_CERT_FILE",
			SubPath:   "HUB_PROXY_CERT_FILE",
		})
	}

	if len(c.blackduck.Spec.AuthCustomCA) > 1 {
		volumesMounts = append(volumesMounts, &horizonapi.VolumeMountConfig{
			Name:      "auth-custom-ca",
			MountPath: "/tmp/secrets/AUTH_CUSTOM_CA",
			SubPath:   "AUTH_CUSTOM_CA",
		})
	}

	return volumesMounts
}

func NewBdReplicationController(replicationController *types.ReplicationController, config *protoform.Config, kubeClient *kubernetes.Clientset, blackduck *blackduckapi.Blackduck) types.ReplicationControllerInterface {
	return &BdReplicationController{ReplicationController: replicationController, config: config, kubeClient: kubeClient, blackduck: blackduck}
}
