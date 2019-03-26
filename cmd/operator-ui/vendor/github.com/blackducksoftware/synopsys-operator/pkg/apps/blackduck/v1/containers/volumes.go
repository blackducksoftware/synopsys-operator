package containers

import (
	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
)

func (hc *Creater) getDBSecretVolume() *components.Volume {
	return components.NewSecretVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      "db-passwords",
		MapOrSecretName: "db-creds",
		Items: map[string]horizonapi.KeyAndMode{
			"HUB_POSTGRES_ADMIN_PASSWORD_FILE": {KeyOrPath: "HUB_POSTGRES_ADMIN_PASSWORD_FILE", Mode: util.IntToInt32(420)},
			"HUB_POSTGRES_USER_PASSWORD_FILE":  {KeyOrPath: "HUB_POSTGRES_USER_PASSWORD_FILE", Mode: util.IntToInt32(420)},
		},
		DefaultMode: util.IntToInt32(420),
	})
}

func (hc *Creater) getProxyVolume() *components.Volume {
	return components.NewSecretVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      "blackduck-proxy-certificate",
		MapOrSecretName: "blackduck-proxy-certificate",
		Items: map[string]horizonapi.KeyAndMode{
			"HUB_PROXY_CERT_FILE": {KeyOrPath: "HUB_PROXY_CERT_FILE", Mode: util.IntToInt32(420)},
		},
		DefaultMode: util.IntToInt32(420),
	})
}
