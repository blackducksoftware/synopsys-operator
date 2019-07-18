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
	"k8s.io/client-go/kubernetes"
)

type BdRSecret struct {
	config     *protoform.Config
	kubeClient *kubernetes.Clientset
	blackduck  *blackduckapi.Blackduck
}

func (b BdRSecret) GetSecrets() []*components.Secret {
	var secrets []*components.Secret

	var adminPassword, userPassword, postgresPassword string
	var err error
	if b.blackduck.Spec.ExternalPostgres != nil {

		adminPassword, err = util.Base64Decode(b.blackduck.Spec.ExternalPostgres.PostgresAdminPassword)
		if err != nil {
			return nil
		}

		userPassword, err = util.Base64Decode(b.blackduck.Spec.ExternalPostgres.PostgresUserPassword)
		if err != nil {
			return nil
		}

	} else {
		adminPassword, err = util.Base64Decode(b.blackduck.Spec.AdminPassword)
		if err != nil {
			return nil
		}

		userPassword, err = util.Base64Decode(b.blackduck.Spec.UserPassword)
		if err != nil {
			return nil
		}

		postgresPassword, err = util.Base64Decode(b.blackduck.Spec.PostgresPassword)
		if err != nil {
			return nil
		}

	}

	postgresSecret := components.NewSecret(horizonapi.SecretConfig{Namespace: b.blackduck.Spec.Namespace, Name: util.GetResourceName(b.blackduck.Name, util.BlackDuckName, "db-creds"), Type: horizonapi.SecretTypeOpaque})

	if b.blackduck.Spec.ExternalPostgres != nil {
		postgresSecret.AddData(map[string][]byte{"HUB_POSTGRES_ADMIN_PASSWORD_FILE": []byte(adminPassword), "HUB_POSTGRES_USER_PASSWORD_FILE": []byte(userPassword)})
	} else {
		postgresSecret.AddData(map[string][]byte{"HUB_POSTGRES_ADMIN_PASSWORD_FILE": []byte(adminPassword), "HUB_POSTGRES_USER_PASSWORD_FILE": []byte(userPassword), "HUB_POSTGRES_POSTGRES_PASSWORD_FILE": []byte(postgresPassword)})
	}
	postgresSecret.AddLabels(utils2.GetVersionLabel("postgres", b.blackduck.Name, b.blackduck.Spec.Version))

	secrets = append(secrets, postgresSecret)
	return secrets
}

func NewBdRSecret(config *protoform.Config, kubeClient *kubernetes.Clientset, blackduck *blackduckapi.Blackduck) types.SecretInterface {
	return &BdRSecret{config: config, kubeClient: kubeClient, blackduck: blackduck}
}

func init() {
	store.Register(types.SecretPostgresV1, NewBdRSecret)
}
