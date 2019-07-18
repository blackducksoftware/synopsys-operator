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
	store.Register(types.RcDocumentationV1, NewBdReplicationController)
}

func (c *BdReplicationController) GetRc() (*components.ReplicationController, error) {

	containerConfig, ok := c.Containers[types.DocumentationContainerName]
	if !ok {
		return nil, fmt.Errorf("couldn't find container %s", types.DocumentationContainerName)
	}

	documentationEmptyDir, _ := util.CreateEmptyDirVolumeWithoutSizeLimit("dir-documentation")
	documentationContainerConfig := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "documentation", Image: containerConfig.Image,
			PullPolicy: horizonapi.PullAlways, MinMem: fmt.Sprintf("%dM", containerConfig.MinMem), MaxMem: fmt.Sprintf("%dM", containerConfig.MaxMem), MinCPU: fmt.Sprintf("%d", containerConfig.MinCPU), MaxCPU: fmt.Sprintf("%d", containerConfig.MaxCPU)},
		EnvConfigs: []*horizonapi.EnvConfig{utils.GetHubConfigEnv(c.blackduck.Name)},
		VolumeMounts: []*horizonapi.VolumeMountConfig{
			{Name: "dir-documentation", MountPath: "/opt/blackduck/hub/hub-documentation/security"},
		},
		PortConfig: []*horizonapi.PortConfig{{ContainerPort: int32(8443), Protocol: horizonapi.ProtocolTCP}},
	}

	if c.blackduck.Spec.LivenessProbes {
		documentationContainerConfig.LivenessProbeConfigs = []*horizonapi.ProbeConfig{{
			ActionConfig: horizonapi.ActionConfig{
				Type:    horizonapi.ActionTypeCommand,
				Command: []string{"/usr/local/bin/docker-healthcheck.sh", "https://127.0.0.1:8443/hubdoc/health-checks/liveness", "/opt/blackduck/hub/hub-documentation/security/root.crt"},
			},
			Delay:           240,
			Interval:        30,
			Timeout:         10,
			MinCountFailure: 10,
		}}
	}

	podConfig := &util.PodConfig{
		Volumes:             []*components.Volume{documentationEmptyDir},
		Containers:          []*util.Container{documentationContainerConfig},
		ImagePullSecrets:    c.blackduck.Spec.RegistryConfiguration.PullSecrets,
		Labels:              utils2.GetVersionLabel("documentation", c.blackduck.Name, c.blackduck.Spec.Version),
		NodeAffinityConfigs: utils.GetNodeAffinityConfigs("documentation", &c.blackduck.Spec),
	}

	if !c.config.IsOpenshift {
		podConfig.FSGID = util.IntToInt64(0)
	}

	return util.CreateReplicationControllerFromContainer(
		&horizonapi.ReplicationControllerConfig{Namespace: c.blackduck.Spec.Namespace, Name: util.GetResourceName(c.blackduck.Name, util.BlackDuckName, "documentation"), Replicas: util.IntToInt32(1)},
		podConfig, utils2.GetLabel("documentation", c.blackduck.Name))
}

func NewBdReplicationController(replicationController *types.ReplicationController, config *protoform.Config, kubeClient *kubernetes.Clientset, blackduck *blackduckapi.Blackduck) types.ReplicationControllerInterface {
	return &BdReplicationController{ReplicationController: replicationController, config: config, kubeClient: kubeClient, blackduck: blackduck}
}