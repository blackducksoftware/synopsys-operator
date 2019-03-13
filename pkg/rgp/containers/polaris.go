package containers

import (
	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
)

// GetPolarisDeployment returns the polaris deployment
func (g *RgpDeployer) GetPolarisDeployment() *components.Deployment {
	deployConfig := &horizonapi.DeploymentConfig{
		Name:      "polaris-service",
		Namespace: g.Grspec.Namespace,
	}
	return util.CreateDeployment(deployConfig, g.GetPolarisPod())
}

// GetPolarisPod returns the polaris pod
func (g *RgpDeployer) GetPolarisPod() *components.Pod {
	envs := g.getPolarisEnvConfigs()

	var containers []*util.Container

	container := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{
			Name:       "polaris-service",
			Image:      "gcr.io/snps-swip-staging/reporting-polaris-service:latest",
			PullPolicy: horizonapi.PullIfNotPresent,
			//MinMem:     "1Gi",
			//MaxMem:     "",
			//MinCPU:     "500m",
			//MaxCPU:     "",
		},
		VolumeMounts: g.getPolarisVolumeMounts(),
		EnvConfigs:   envs,
		PortConfig: []*horizonapi.PortConfig{
			{ContainerPort: "7788"},
		},
	}

	containers = append(containers, container)
	return util.CreatePod("polaris-service", "", g.getReportVolumes(), containers, nil, nil)
}

// GetPolarisService returns the polaris service
func (g *RgpDeployer) GetPolarisService() *components.Service {
	service := components.NewService(horizonapi.ServiceConfig{
		Name:          "polaris-service",
		Namespace:     g.Grspec.Namespace,
		IPServiceType: horizonapi.ClusterIPServiceTypeDefault,
	})
	service.AddSelectors(map[string]string{
		"app": "polaris-service",
	})
	service.AddPort(horizonapi.ServicePortConfig{Name: "7788", Port: 7788, Protocol: horizonapi.ProtocolTCP, TargetPort: "7788"})
	service.AddPort(horizonapi.ServicePortConfig{Name: "admin", Port: 8081, Protocol: horizonapi.ProtocolTCP, TargetPort: "8081"})
	return service
}

func (g *RgpDeployer) getPolarisVolumes() []*components.Volume {
	var volumes []*components.Volume

	volumes = append(volumes, components.NewSecretVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      "vault-cacert",
		MapOrSecretName: "vault-ca-certificate",
		Items: map[string]horizonapi.KeyAndMode{
			"vault_cacrt": {KeyOrPath: "tls.crt", Mode: util.IntToInt32(420)},
		},
	}))

	volumes = append(volumes, components.NewSecretVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      "vault-client-key",
		MapOrSecretName: "auth-client-tls-certificate",
		Items: map[string]horizonapi.KeyAndMode{
			"vault_client_key": {KeyOrPath: "tls.key", Mode: util.IntToInt32(420)},
		},
	}))

	volumes = append(volumes, components.NewSecretVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      "vault-client-cert",
		MapOrSecretName: "auth-client-tls-certificate",
		Items: map[string]horizonapi.KeyAndMode{
			"vault_client_cert": {KeyOrPath: "tls.crt", Mode: util.IntToInt32(420)},
		},
	}))

	return volumes
}

func (g *RgpDeployer) getPolarisVolumeMounts() []*horizonapi.VolumeMountConfig {
	var volumeMounts []*horizonapi.VolumeMountConfig
	volumeMounts = append(volumeMounts, &horizonapi.VolumeMountConfig{Name: "vault-cacert", MountPath: "/mnt/vault/ca"})
	volumeMounts = append(volumeMounts, &horizonapi.VolumeMountConfig{Name: "vault-client-key", MountPath: "/mnt/vault/key"})
	volumeMounts = append(volumeMounts, &horizonapi.VolumeMountConfig{Name: "vault-client-cert", MountPath: "/mnt/vault/cert"})

	return volumeMounts
}

func (g *RgpDeployer) getPolarisEnvConfigs() []*horizonapi.EnvConfig {
	var envs []*horizonapi.EnvConfig
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "SWIP_VAULT_ADDRESS", KeyOrVal: "https://vault:8200"})
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "VAULT_CACERT", KeyOrVal: "/mnt/vault/ca/vault_cacrt"})
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "VAULT_CLIENT_KEY", KeyOrVal: "/mnt/vault/key/vault_client_key"})
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "VAULT_CLIENT_CERT", KeyOrVal: "/mnt/vault/cert/vault_client_cert"})

	envs = append(envs, g.getCommonEnvConfigs()...)
	envs = append(envs, g.getSwipEnvConfigs()...)
	envs = append(envs, g.getPostgresEnvConfigs()...)

	return envs
}
