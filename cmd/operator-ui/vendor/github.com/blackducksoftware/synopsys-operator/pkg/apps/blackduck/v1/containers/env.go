package containers

import (


	horizonapi "github.com/blackducksoftware/horizon/pkg/api"

)
func (c *Creater) getHubConfigEnv() *horizonapi.EnvConfig {
	return &horizonapi.EnvConfig{Type: horizonapi.EnvFromConfigMap, FromName: "hub-config"}
}

func (c *Creater) getHubDBConfigEnv() *horizonapi.EnvConfig {
	return &horizonapi.EnvConfig{Type: horizonapi.EnvFromConfigMap, FromName: "hub-db-config"}
}
