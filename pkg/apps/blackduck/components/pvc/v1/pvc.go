package v1

import (
	"github.com/blackducksoftware/horizon/pkg/components"
	blackduckapi "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/types"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/store"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"k8s.io/client-go/kubernetes"
)

type BdPVC struct {
	config     *protoform.Config
	kubeClient *kubernetes.Clientset
	blackduck  *blackduckapi.Blackduck
}

func NewPvc(config *protoform.Config, kubeClient *kubernetes.Clientset, blackduck *blackduckapi.Blackduck) types.PVCInterface {
	return &BdPVC{config: config, kubeClient: kubeClient, blackduck: blackduck}
}

func (b BdPVC) GetPVCs() ([]*components.PersistentVolumeClaim, error) {

	defaultPVC := map[string]string{
		"blackduck-authentication":   "2Gi",
		"blackduck-cfssl":            "2Gi",
		"blackduck-registration":     "2Gi",
		"blackduck-solr":             "2Gi",
		"blackduck-webapp":           "2Gi",
		"blackduck-logstash":         "20Gi",
		"blackduck-zookeeper":        "4Gi",
		"blackduck-uploadcache-data": "100Gi",
	}

	if b.blackduck.Spec.ExternalPostgres == nil {
		defaultPVC["blackduck-postgres"] = "150Gi"
	}

	return b.blackduck.GenPVC(defaultPVC)
}

func init() {
	store.Register(types.PVCV1, NewPvc)
}
