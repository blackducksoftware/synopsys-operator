package containers

import (
	"fmt"
	"strings"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
)

func (g *RgpDeployer) getCommonEnvConfigs() []*horizonapi.EnvConfig {
	var envs []*horizonapi.EnvConfig
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "CONNECTION_POOL_SIZE", KeyOrVal: "10"})
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "LOG_LEVEL", KeyOrVal: "INFO"})
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "SPRING_PROFILE", KeyOrVal: "production"})
	return envs
}

func (g *RgpDeployer) getSwipEnvConfigs() []*horizonapi.EnvConfig {
	var envs []*horizonapi.EnvConfig
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "SWIP_ROOT_DOMAIN", KeyOrVal: g.Grspec.IngressHost})
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "SWIP_ENVIRONMENT_NAME", KeyOrVal: g.Grspec.Namespace})
	return envs
}

func (g *RgpDeployer) getPostgresEnvConfigs() []*horizonapi.EnvConfig {
	var envs []*horizonapi.EnvConfig
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "POSTGRES_HOST", KeyOrVal: "postgres"})
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "POSTGRES_PORT", KeyOrVal: "5432"})
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "POSTGRES_USERNAME", KeyOrVal: "postgres"})
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvFromSecret, NameOrPrefix: "POSTGRES_PASSWORD", KeyOrVal: "POSTGRES_PASSWORD", FromName: "db-creds"})
	return envs
}

func (g *RgpDeployer) getMongoEnvConfigs() []*horizonapi.EnvConfig {
	var envs []*horizonapi.EnvConfig
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "MONGODB_HOST", KeyOrVal: "mongodb"})
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "MONGODB_PORT", KeyOrVal: "27017"})
	return envs
}

func (g *RgpDeployer) getEventStoreLegacyEnvConfigs() []*horizonapi.EnvConfig {
	var envs []*horizonapi.EnvConfig
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "EVENT_STORE_ADDR", KeyOrVal: "eventstore"})
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "EVENT_STORE_USERNAME", KeyOrVal: "admin"})
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvFromSecret, NameOrPrefix: "EVENT_STORE_PASSWORD", KeyOrVal: "password", FromName: "swip-eventstore-creds"})
	return envs
}

func (g *RgpDeployer) getEventStoreEnvConfigs(role string) []*horizonapi.EnvConfig {
	var envs []*horizonapi.EnvConfig
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "EVENT_STORE_ADDR", KeyOrVal: "eventstore"})
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: fmt.Sprintf("EVENT_STORE_%s_USERNAME", strings.ToUpper(role)), KeyOrVal: role})
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvFromSecret, NameOrPrefix: fmt.Sprintf("EVENT_STORE_%s_PASSWORD", strings.ToUpper(role)), KeyOrVal: "password", FromName: "swip-eventstore-creds"})
	return envs
}
