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
	store.Register(blackduck.BlackDuckScanRCV1, NewBdReplicationController)
}

// GetRc returns the RC
func (c *BdReplicationController) GetRc() (*components.ReplicationController, error) {
	containerConfig, ok := c.Containers[blackduck.ScanContainerName]
	if !ok {
		return nil, fmt.Errorf("couldn't find container %s", blackduck.ScanContainerName)
	}

	if containerConfig.MaxMem == nil {
		return nil, fmt.Errorf("Maxmem must be set for %s", blackduck.ScanContainerName)
	}

	scannerEnvs := []*horizonapi.EnvConfig{utils.GetBlackDuckConfigEnv(c.blackDuck.Name), utils.GetBlackDuckDBConfigEnv(c.blackDuck.Name)}
	scannerEnvs = append(scannerEnvs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "HUB_MAX_MEMORY", KeyOrVal: fmt.Sprintf("%dM", *containerConfig.MaxMem-512)})
	hubScanEmptyDir, _ := util.CreateEmptyDirVolumeWithoutSizeLimit("dir-scan")
	hubScanContainerConfig := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "scan", Image: containerConfig.Image, PullPolicy: horizonapi.PullAlways},
		EnvConfigs:      scannerEnvs,
		VolumeMounts: []*horizonapi.VolumeMountConfig{
			{
				Name:      "db-passwords",
				MountPath: "/tmp/secrets/HUB_POSTGRES_ADMIN_PASSWORD_FILE",
				SubPath:   "HUB_POSTGRES_ADMIN_PASSWORD_FILE",
			},
			{
				Name:      "db-passwords",
				MountPath: "/tmp/secrets/HUB_POSTGRES_USER_PASSWORD_FILE",
				SubPath:   "HUB_POSTGRES_USER_PASSWORD_FILE",
			},
			{
				Name:      "dir-scan",
				MountPath: "/opt/blackduck/hub/hub-scan/security",
			},
		},
		PortConfig: []*horizonapi.PortConfig{{ContainerPort: int32(8443), Protocol: horizonapi.ProtocolTCP}},
	}

	apputils.SetLimits(hubScanContainerConfig.ContainerConfig, containerConfig)

	if c.blackDuck.Spec.LivenessProbes {
		hubScanContainerConfig.LivenessProbeConfigs = []*horizonapi.ProbeConfig{{
			ActionConfig: horizonapi.ActionConfig{
				Type: horizonapi.ActionTypeCommand,
				Command: []string{
					"/usr/local/bin/docker-healthcheck.sh",
					"https://127.0.0.1:8443/api/health-checks/liveness",
					"/opt/blackduck/hub/hub-scan/security/root.crt",
					"/opt/blackduck/hub/hub-scan/security/blackduck_system.crt",
					"/opt/blackduck/hub/hub-scan/security/blackduck_system.key",
				},
			},
			Delay:           240,
			Interval:        30,
			Timeout:         10,
			MinCountFailure: 10,
		}}
	}

	hubScanVolumes := []*components.Volume{hubScanEmptyDir, utils.GetDBSecretVolume(c.blackDuck.Name)}

	// Mount the HTTPS proxy certificate if provided
	if len(c.blackDuck.Spec.ProxyCertificate) > 0 {
		hubScanContainerConfig.VolumeMounts = append(hubScanContainerConfig.VolumeMounts, &horizonapi.VolumeMountConfig{
			Name:      "proxy-certificate",
			MountPath: "/tmp/secrets/HUB_PROXY_CERT_FILE",
			SubPath:   "HUB_PROXY_CERT_FILE",
		})
		hubScanVolumes = append(hubScanVolumes, utils.GetProxyVolume(c.blackDuck.Name))
	}

	podConfig := &util.PodConfig{
		Volumes:             hubScanVolumes,
		Containers:          []*util.Container{hubScanContainerConfig},
		ImagePullSecrets:    c.blackDuck.Spec.RegistryConfiguration.PullSecrets,
		Labels:              apputils.GetVersionLabel("scan", c.blackDuck.Name, c.blackDuck.Spec.Version),
		NodeAffinityConfigs: utils.GetNodeAffinityConfigs("scan", &c.blackDuck.Spec),
	}
	if !c.config.IsOpenshift {
		podConfig.FSGID = util.IntToInt64(0)
	}

	return util.CreateReplicationControllerFromContainer(
		&horizonapi.ReplicationControllerConfig{Namespace: c.blackDuck.Spec.Namespace, Name: apputils.GetResourceName(c.blackDuck.Name, util.BlackDuckName, "scan"), Replicas: util.IntToInt32(c.Replicas)},
		podConfig, apputils.GetLabel("scan", c.blackDuck.Name))
}

// NewBdReplicationController returns the Black Duck RC configuration
func NewBdReplicationController(replicationController *types.ReplicationController, config *protoform.Config, kubeClient *kubernetes.Clientset, cr interface{}) (types.ReplicationControllerInterface, error) {
	blackDuck, ok := cr.(*blackduckapi.Blackduck)
	if !ok {
		return nil, fmt.Errorf("unable to cast the interface to Black Duck object")
	}
	return &BdReplicationController{ReplicationController: replicationController, config: config, kubeClient: kubeClient, blackDuck: blackDuck}, nil
}
