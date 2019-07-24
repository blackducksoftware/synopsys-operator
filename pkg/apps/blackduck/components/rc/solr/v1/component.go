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
	store.Register(types.RcSolrV1, NewBdReplicationController)
}

func (c *BdReplicationController) GetRc() (*components.ReplicationController, error) {

	containerConfig, ok := c.Containers[types.SolrContainerName]
	if !ok {
		return nil, fmt.Errorf("couldn't find container %s", types.SolrContainerName)
	}
	solrVolumeMount := c.getSolrVolumeMounts()
	solrContainerConfig := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "solr", Image: containerConfig.Image, PullPolicy: horizonapi.PullAlways},
		EnvConfigs:      []*horizonapi.EnvConfig{utils.GetHubConfigEnv(c.blackduck.Name)},
		VolumeMounts:    solrVolumeMount,
		PortConfig:      []*horizonapi.PortConfig{{ContainerPort: int32(8983), Protocol: horizonapi.ProtocolTCP}},
	}

	utils2.SetLimits(solrContainerConfig.ContainerConfig, containerConfig)

	if c.blackduck.Spec.LivenessProbes {
		solrContainerConfig.LivenessProbeConfigs = []*horizonapi.ProbeConfig{{
			ActionConfig: horizonapi.ActionConfig{
				Type:    horizonapi.ActionTypeCommand,
				Command: []string{"/usr/local/bin/docker-healthcheck.sh", "http://localhost:8983/solr/project/admin/ping?wt=json"},
			},
			Delay:           240,
			Interval:        30,
			Timeout:         10,
			MinCountFailure: 10,
		}}
	}

	podConfig := &util.PodConfig{
		Volumes:             c.getSolrVolumes(),
		Containers:          []*util.Container{solrContainerConfig},
		ImagePullSecrets:    c.blackduck.Spec.RegistryConfiguration.PullSecrets,
		Labels:              apputils.GetVersionLabel("solr", c.blackduck.Name, c.blackduck.Spec.Version),
		NodeAffinityConfigs: utils.GetNodeAffinityConfigs("solr", &c.blackduck.Spec),
	}

	if !c.config.IsOpenshift {
		podConfig.FSGID = util.IntToInt64(0)
	}

	return util.CreateReplicationControllerFromContainer(
		&horizonapi.ReplicationControllerConfig{Namespace: c.blackduck.Spec.Namespace, Name: apputils.GetResourceName(c.blackduck.Name, util.BlackDuckName, "solr"), Replicas: util.IntToInt32(1)},
		podConfig, apputils.GetLabel("solr", c.blackduck.Name))
}

// getSolrVolumes will return the solr volumes
func (c *BdReplicationController) getSolrVolumes() []*components.Volume {
	var solrVolume *components.Volume
	if c.blackduck.Spec.PersistentStorage {
		solrVolume, _ = util.CreatePersistentVolumeClaimVolume("dir-solr", utils.GetPVCName("solr", c.blackduck))
	} else {
		solrVolume, _ = util.CreateEmptyDirVolumeWithoutSizeLimit("dir-solr")
	}

	volumes := []*components.Volume{solrVolume}
	return volumes
}

// getSolrVolumeMounts will return the solr volume mounts
func (c *BdReplicationController) getSolrVolumeMounts() []*horizonapi.VolumeMountConfig {
	volumesMounts := []*horizonapi.VolumeMountConfig{
		{Name: "dir-solr", MountPath: "/opt/blackduck/hub/solr/cores.data"},
	}
	return volumesMounts
}

func NewBdReplicationController(replicationController *types.ReplicationController, config *protoform.Config, kubeClient *kubernetes.Clientset, blackduck *blackduckapi.Blackduck) types.ReplicationControllerInterface {
	return &BdReplicationController{ReplicationController: replicationController, config: config, kubeClient: kubeClient, blackduck: blackduck}
}
