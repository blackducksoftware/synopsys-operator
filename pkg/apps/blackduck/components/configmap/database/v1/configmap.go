package v1

import (
	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	blackduckapi "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/components/utils"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/types"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/store"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"k8s.io/client-go/kubernetes"
	"strconv"
)

type BdConfigmap struct {
	config     *protoform.Config
	kubeClient *kubernetes.Clientset
	blackduck  *blackduckapi.Blackduck
}

func NewBdConfigmap(config *protoform.Config, kubeClient *kubernetes.Clientset, blackduck *blackduckapi.Blackduck) types.ConfigMapInterface {
	return &BdConfigmap{config: config, kubeClient: kubeClient, blackduck: blackduck}
}

func (b BdConfigmap) GetCM() []*components.ConfigMap {

	var configMaps []*components.ConfigMap
	// DB
	hubDbConfig := components.NewConfigMap(horizonapi.ConfigMapConfig{Namespace: b.blackduck.Spec.Namespace, Name: util.GetResourceName(b.blackduck.Name, util.BlackDuckName, "db-config")})
	if b.blackduck.Spec.ExternalPostgres != nil {
		hubDbConfig.AddData(map[string]string{
			"HUB_POSTGRES_ADMIN": b.blackduck.Spec.ExternalPostgres.PostgresAdmin,
			"HUB_POSTGRES_USER":  b.blackduck.Spec.ExternalPostgres.PostgresUser,
			"HUB_POSTGRES_PORT":  strconv.Itoa(b.blackduck.Spec.ExternalPostgres.PostgresPort),
			"HUB_POSTGRES_HOST":  b.blackduck.Spec.ExternalPostgres.PostgresHost,
		})
	} else {
		hubDbConfig.AddData(map[string]string{
			"HUB_POSTGRES_ADMIN": "blackduck",
			"HUB_POSTGRES_USER":  "blackduck_user",
			"HUB_POSTGRES_PORT":  "5432",
			"HUB_POSTGRES_HOST":  util.GetResourceName(b.blackduck.Name, util.BlackDuckName, "postgres"),
		})
	}

	if b.blackduck.Spec.ExternalPostgres != nil {
		hubDbConfig.AddData(map[string]string{"HUB_POSTGRES_ENABLE_SSL": strconv.FormatBool(b.blackduck.Spec.ExternalPostgres.PostgresSsl)})
		if b.blackduck.Spec.ExternalPostgres.PostgresSsl {
			hubDbConfig.AddData(map[string]string{"HUB_POSTGRES_ENABLE_SSL_CERT_AUTH": "false"})
		}
	} else {
		hubDbConfig.AddData(map[string]string{"HUB_POSTGRES_ENABLE_SSL": "false"})
	}
	hubDbConfig.AddLabels(utils.GetVersionLabel("postgres", b.blackduck.Name, b.blackduck.Spec.Version))
	configMaps = append(configMaps, hubDbConfig)
	return configMaps
}
func init() {
	store.Register(types.DatabaseConfigmapV1, NewBdConfigmap)
}
