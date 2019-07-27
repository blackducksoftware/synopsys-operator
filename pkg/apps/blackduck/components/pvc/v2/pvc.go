package v2

import (
	"fmt"

	"github.com/blackducksoftware/horizon/pkg/components"
	blackduckapi "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/store"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/types"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"k8s.io/client-go/kubernetes"
)

// BdPVC holds the Black Duck PVC configuration
type BdPVC struct {
	config     *protoform.Config
	kubeClient *kubernetes.Clientset
	blackDuck  *blackduckapi.Blackduck
}

// NewPvc returns the Black Duck PVC configuration
func NewPvc(config *protoform.Config, kubeClient *kubernetes.Clientset, cr interface{}) (types.PVCInterface, error) {
	blackDuck, ok := cr.(*blackduckapi.Blackduck)
	if !ok {
		return nil, fmt.Errorf("unable to cast the interface to Black Duck object")
	}
	return &BdPVC{config: config, kubeClient: kubeClient, blackDuck: blackDuck}, nil
}

// GetPVCs returns the PVC
func (b BdPVC) GetPVCs() ([]*components.PersistentVolumeClaim, error) {
	defaultPVC := map[string]string{
		"blackduck-authentication":   "2Gi",
		"blackduck-cfssl":            "2Gi",
		"blackduck-registration":     "2Gi",
		"blackduck-webapp":           "2Gi",
		"blackduck-logstash":         "20Gi",
		"blackduck-zookeeper":        "4Gi",
		"blackduck-uploadcache-data": "100Gi",
	}

	if b.blackDuck.Spec.ExternalPostgres == nil {
		defaultPVC["blackduck-postgres"] = "150Gi"
	}

	return b.blackDuck.GenPVC(defaultPVC)
}

func init() {
	store.Register(blackduck.BlackDuckPVCV2, NewPvc)
}
