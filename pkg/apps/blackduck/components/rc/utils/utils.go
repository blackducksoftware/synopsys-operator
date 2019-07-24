package utils

import (
	"fmt"
	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	v1 "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/utils"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
)

func GetDBSecretVolume(name string) *components.Volume {
	return components.NewSecretVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      "db-passwords",
		MapOrSecretName: utils.GetResourceName(name, util.BlackDuckName, "db-creds"),
		Items: []horizonapi.KeyPath{
			{Key: "HUB_POSTGRES_ADMIN_PASSWORD_FILE", Path: "HUB_POSTGRES_ADMIN_PASSWORD_FILE", Mode: util.IntToInt32(420)},
			{Key: "HUB_POSTGRES_USER_PASSWORD_FILE", Path: "HUB_POSTGRES_USER_PASSWORD_FILE", Mode: util.IntToInt32(420)},
		},
		DefaultMode: util.IntToInt32(420),
	})
}

func GetProxyVolume(name string) *components.Volume {
	return components.NewSecretVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      "blackduck-proxy-certificate",
		MapOrSecretName: utils.GetResourceName(name, util.BlackDuckName, "proxy-certificate"),
		Items: []horizonapi.KeyPath{
			{Key: "HUB_PROXY_CERT_FILE", Path: "HUB_PROXY_CERT_FILE", Mode: util.IntToInt32(420)},
		},
		DefaultMode: util.IntToInt32(420),
	})
}

func GetHubConfigEnv(name string) *horizonapi.EnvConfig {
	return &horizonapi.EnvConfig{Type: horizonapi.EnvFromConfigMap, FromName: utils.GetResourceName(name, util.BlackDuckName, "config")}
}

func GetHubDBConfigEnv(name string) *horizonapi.EnvConfig {
	return &horizonapi.EnvConfig{Type: horizonapi.EnvFromConfigMap, FromName: utils.GetResourceName(name, util.BlackDuckName, "db-config")}
}

var affTypeMap = map[string]horizonapi.AffinityType{
	"Hard": horizonapi.AffinityHard,
	"Soft": horizonapi.AffinitySoft,
}

var nodeOperatorMap = map[string]horizonapi.NodeOperator{
	"In":           horizonapi.NodeOperatorIn,
	"NotIn":        horizonapi.NodeOperatorNotIn,
	"Exists":       horizonapi.NodeOperatorExists,
	"DoesNotExist": horizonapi.NodeOperatorDoesNotExist,
	"Gt":           horizonapi.NodeOperatorGt,
	"Lt":           horizonapi.NodeOperatorLt,
}

func GetNodeAffinityConfigs(podName string, bdspec *v1.BlackduckSpec) map[horizonapi.AffinityType][]*horizonapi.NodeAffinityConfig {

	// make an empty NodeAffinityMap
	nodeAffinityMap := make(map[horizonapi.AffinityType][]*horizonapi.NodeAffinityConfig)

	for _, affinity := range bdspec.NodeAffinities[podName] {
		nodeAffinityMap[affTypeMap[affinity.AffinityType]] = append(nodeAffinityMap[affTypeMap[affinity.AffinityType]],
			&horizonapi.NodeAffinityConfig{
				Expressions: []horizonapi.NodeExpression{
					{
						Key:    affinity.Key,
						Op:     nodeOperatorMap[affinity.Op],
						Values: affinity.Values,
					},
				},
			},
		)
	}

	return nodeAffinityMap
}

func GetPVCName(name string, blackduck *v1.Blackduck) string {
	if blackduck.Annotations["synopsys.com/created.by"] == "pre-2019.6.0" {
		return fmt.Sprintf("blackduck-%s", name)
	}
	return utils.GetResourceName(blackduck.Name, util.BlackDuckName, name)
}
