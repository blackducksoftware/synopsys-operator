package v1

import (
	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	blackduckapi "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	utils2 "github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/components/utils"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/types"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/store"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
)

type BdRSecret struct {
	config     *protoform.Config
	kubeClient *kubernetes.Clientset
	blackduck  *blackduckapi.Blackduck
}

func (b BdRSecret) GetSecrets() []*components.Secret {
	var secrets []*components.Secret

	uploadCacheSecret := components.NewSecret(horizonapi.SecretConfig{Namespace: b.blackduck.Spec.Namespace, Name: util.GetResourceName(b.blackduck.Name, util.BlackDuckName, "upload-cache"), Type: horizonapi.SecretTypeOpaque})

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

	uploadCacheSecret.AddLabels(utils2.GetVersionLabel("uploadcache", b.blackduck.Name, b.blackduck.Spec.Version))
	secrets = append(secrets, uploadCacheSecret)

	return secrets
}

func NewBdRSecret(config *protoform.Config, kubeClient *kubernetes.Clientset, blackduck *blackduckapi.Blackduck) types.SecretInterface {
	return &BdRSecret{config: config, kubeClient: kubeClient, blackduck: blackduck}
}

func init() {
	store.Register(types.SecretUploadCacheV1, NewBdRSecret)
}
