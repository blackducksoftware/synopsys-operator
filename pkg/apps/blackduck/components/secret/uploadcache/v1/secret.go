package v1

import (
	"fmt"
	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	blackduckapi "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/store"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/types"
	apputils "github.com/blackducksoftware/synopsys-operator/pkg/apps/utils"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
)

// BdRSecret holds the Black Duck secret configuration
type BdRSecret struct {
	config     *protoform.Config
	kubeClient *kubernetes.Clientset
	blackDuck  *blackduckapi.Blackduck
}

// GetSecrets returns the secret
func (b BdRSecret) GetSecrets() []*components.Secret {
	var secrets []*components.Secret

	uploadCacheSecret := components.NewSecret(horizonapi.SecretConfig{Namespace: b.blackDuck.Spec.Namespace, Name: apputils.GetResourceName(b.blackDuck.Name, util.BlackDuckName, "upload-cache"), Type: horizonapi.SecretTypeOpaque})

	if !b.config.DryRun {
		secret, err := util.GetSecret(b.kubeClient, b.config.Namespace, "blackduck-secret")
		if err != nil {
			logrus.Errorf("unable to find Synopsys Operator blackduck-secret in %s namespace due to %+v", b.config.Namespace, err)
			return nil
		}
		uploadCacheSecret.AddData(map[string][]byte{"SEAL_KEY": secret.Data["SEAL_KEY"]})
	} else {
		uploadCacheSecret.AddData(map[string][]byte{"SEAL_KEY": {}})
	}

	uploadCacheSecret.AddLabels(apputils.GetVersionLabel("uploadcache", b.blackDuck.Name, b.blackDuck.Spec.Version))
	secrets = append(secrets, uploadCacheSecret)

	return secrets
}

// NewBdRSecret returns the Black Duck secret configuration
func NewBdRSecret(config *protoform.Config, kubeClient *kubernetes.Clientset, cr interface{}) (types.SecretInterface, error) {
	blackDuck, ok := cr.(*blackduckapi.Blackduck)
	if !ok {
		return nil, fmt.Errorf("unable to cast the interface to Black Duck object")
	}
	return &BdRSecret{config: config, kubeClient: kubeClient, blackDuck: blackDuck}, nil
}

func init() {
	store.Register(types.BlackDuckUploadCacheSecretV1, NewBdRSecret)
}
