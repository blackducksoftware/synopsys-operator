package v1

import (
	"fmt"
	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	blackduckapi "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/types"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/store"
	apputils "github.com/blackducksoftware/synopsys-operator/pkg/apps/utils"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"k8s.io/client-go/kubernetes"
	"strings"
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

	hubConfig := components.NewConfigMap(horizonapi.ConfigMapConfig{Namespace: b.blackduck.Spec.Namespace, Name: apputils.GetResourceName(b.blackduck.Name, util.BlackDuckName, "config")})

	hubData := map[string]string{
		"RUN_SECRETS_DIR": "/tmp/secrets",
		"HUB_VERSION":     b.blackduck.Spec.Version,
	}

	blackduckServiceData := map[string]string{
		// TODO: commented the below 2 environs until the HUB-20482 is fixed. once it if fixed, uncomment them
		//"HUB_AUTHENTICATION_HOST": util.GetResourceName(b.blackduck.Name, util.BlackDuckName, "authentication"),
		//"AUTHENTICATION_HOST":     fmt.Sprintf("%s:%d", util.GetResourceName(b.blackduck.Name, util.BlackDuckName, "authentication"), int32(8443)),
		"CLIENT_CERT_CN":        apputils.GetResourceName(b.blackduck.Name, util.BlackDuckName, "binaryscanner"),
		"CFSSL":                 fmt.Sprintf("%s:8888", apputils.GetResourceName(b.blackduck.Name, util.BlackDuckName, "cfssl")),
		"HUB_CFSSL_HOST":        apputils.GetResourceName(b.blackduck.Name, util.BlackDuckName, "cfssl"),
		"BLACKDUCK_CFSSL_HOST":  apputils.GetResourceName(b.blackduck.Name, util.BlackDuckName, "cfssl"),
		"BLACKDUCK_CFSSL_PORT":  "8888",
		"HUB_DOC_HOST":          apputils.GetResourceName(b.blackduck.Name, util.BlackDuckName, "documentation"),
		"HUB_JOBRUNNER_HOST":    apputils.GetResourceName(b.blackduck.Name, util.BlackDuckName, "jobrunner"),
		"HUB_LOGSTASH_HOST":     apputils.GetResourceName(b.blackduck.Name, util.BlackDuckName, "logstash"),
		"RABBIT_MQ_HOST":        apputils.GetResourceName(b.blackduck.Name, util.BlackDuckName, "rabbitmq"),
		"BROKER_URL":            fmt.Sprintf("amqps://%s/protecodesc", apputils.GetResourceName(b.blackduck.Name, util.BlackDuckName, "rabbitmq")),
		"HUB_REGISTRATION_HOST": apputils.GetResourceName(b.blackduck.Name, util.BlackDuckName, "registration"),
		"HUB_SCAN_HOST":         apputils.GetResourceName(b.blackduck.Name, util.BlackDuckName, "scan"),
		"HUB_SOLR_HOST":         apputils.GetResourceName(b.blackduck.Name, util.BlackDuckName, "solr"),
		// TODO: commented the below 2 environs until the HUB-20412 is fixed. once it if fixed, uncomment them
		// "BLACKDUCK_UPLOAD_CACHE_HOST": util.GetResourceName(c.blackDuck.Name, util.BlackDuckName, "uploadcache"),
		// "HUB_UPLOAD_CACHE_HOST":       util.GetResourceName(c.blackDuck.Name, util.BlackDuckName, "uploadcache"),
		// TODO: commented the below environs until the HUB-20462 is fixed. once it if fixed, uncomment them
		// "HUB_WEBAPP_HOST":    util.GetResourceName(c.blackDuck.Name, util.BlackDuckName, "webapp"),
		"HUB_WEBSERVER_HOST": apputils.GetResourceName(b.blackduck.Name, util.BlackDuckName, "webserver"),
		"HUB_ZOOKEEPER_HOST": apputils.GetResourceName(b.blackduck.Name, util.BlackDuckName, "zookeeper"),
	}
	hubData = util.MergeEnvMaps(blackduckServiceData, hubData)

	for _, value := range b.blackduck.Spec.Environs {
		values := strings.SplitN(value, ":", 2)
		if len(values) == 2 {
			mapKey := strings.TrimSpace(values[0])
			mapValue := strings.TrimSpace(values[1])
			if len(mapKey) > 0 && len(mapValue) > 0 {
				hubData[mapKey] = mapValue
			}
		}
	}

	// merge default and input environs
	environs := GetHubKnobs()
	hubData = util.MergeEnvMaps(hubData, environs)

	hubConfig.AddData(hubData)
	hubConfig.AddLabels(apputils.GetVersionLabel("configmap", b.blackduck.Name, b.blackduck.Spec.Version))
	configMaps = append(configMaps, hubConfig)

	return configMaps
}

func GetHubKnobs() map[string]string {
	return map[string]string{
		"IPV4_ONLY":                         "0",
		"USE_ALERT":                         "0",
		"USE_BINARY_UPLOADS":                "0",
		"RABBIT_MQ_PORT":                    "5671",
		"BROKER_USE_SSL":                    "yes",
		"SCANNER_CONCURRENCY":               "1",
		"HTTPS_VERIFY_CERTS":                "yes",
		"RABBITMQ_DEFAULT_VHOST":            "protecodesc",
		"RABBITMQ_SSL_FAIL_IF_NO_PEER_CERT": "false",
		"ENABLE_SOURCE_UPLOADS":             "false",
		"DATA_RETENTION_IN_DAYS":            "180",
		"MAX_TOTAL_SOURCE_SIZE_MB":          "4000",
	}
}

func init() {
	store.Register(types.GlobalConfigmapV1, NewBdConfigmap)
}
