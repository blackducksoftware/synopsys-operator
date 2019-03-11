package containers

import (
	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
)

func (g *GrDeployer) GetToolsPortfolioDeployment() *components.Deployment {
	deployConfig := &horizonapi.DeploymentConfig{
		Name:      "tools-portfolio-service",
		Namespace: g.Grspec.Namespace,
	}
	return util.CreateDeployment(deployConfig, g.GetToolsPortfolioPod())
}

func (g *GrDeployer) GetToolsPortfolioPod() *components.Pod {
	envs := g.getToolsPortfolioEnvConfigs()

	var containers []*util.Container

	container := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{
			Name:       "tools-portfolio-service",
			Image:      "gcr.io/snps-swip-staging/reporting-tools-portfolio-service:latest",
			PullPolicy: horizonapi.PullIfNotPresent,
			//MinMem:     "1Gi",
			//MaxMem:     "",
			//MinCPU:     "250m",
			//MaxCPU:     "",
		},
		VolumeMounts: g.getToolsPortfolioVolumeMounts(),
		EnvConfigs:   envs,
		PortConfig: []*horizonapi.PortConfig{
			{ContainerPort: "60281"},
		},
	}

	containers = append(containers, container)
	return util.CreatePod("tools-portfolio-service", "", g.getToolsPortfolioVolumes(), containers, nil, nil)
}


func (g *GrDeployer) GetToolsPortfolioService() *components.Service {
	service := components.NewService(horizonapi.ServiceConfig{
		Name:          "tools-portfolio-service",
		Namespace:     g.Grspec.Namespace,
		IPServiceType: horizonapi.ClusterIPServiceTypeDefault,
	})
	service.AddSelectors(map[string]string{
		"app": "tools-portfolio-service",
	})
	service.AddPort(horizonapi.ServicePortConfig{Name: "60289", Port: 60281, Protocol: horizonapi.ProtocolTCP, TargetPort: "60281"})
	service.AddPort(horizonapi.ServicePortConfig{Name: "admin", Port: 8081, Protocol: horizonapi.ProtocolTCP, TargetPort: "8081"})
	return service
}

func (g *GrDeployer) getToolsPortfolioVolumes() []*components.Volume {
	var volumes []*components.Volume

	volumes = append(volumes, components.NewSecretVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      "vault-cacert",
		MapOrSecretName: "vault-ca-certificate",
		Items: map[string]horizonapi.KeyAndMode{
			"vault_cacrt": {KeyOrPath:"tls.crt", Mode: util.IntToInt32(420)},
		},
	}))

	volumes = append(volumes, components.NewSecretVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      "vault-server-key",
		MapOrSecretName: "auth-server-tls-certificate",
		Items: map[string]horizonapi.KeyAndMode{
			"vault_server_key": {KeyOrPath:"tls.key", Mode: util.IntToInt32(420)},
		},
	}))

	volumes = append(volumes, components.NewSecretVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      "vault-server-cert",
		MapOrSecretName: "auth-server-tls-certificate",
		Items: map[string]horizonapi.KeyAndMode{
			"vault_server_cert": {KeyOrPath:"tls.crt", Mode: util.IntToInt32(420)},
		},
	}))

	return volumes
}


func (g *GrDeployer) getToolsPortfolioVolumeMounts() []*horizonapi.VolumeMountConfig {
	var volumeMounts []*horizonapi.VolumeMountConfig
	volumeMounts = append(volumeMounts, &horizonapi.VolumeMountConfig{Name: "vault-cacert", MountPath: "/mnt/vault/ca"})
	volumeMounts = append(volumeMounts, &horizonapi.VolumeMountConfig{Name: "vault-server-key", MountPath: "/mnt/vault/key"})
	volumeMounts = append(volumeMounts, &horizonapi.VolumeMountConfig{Name: "vault-server-cert", MountPath: "/mnt/vault/cert"})

	return volumeMounts
}


func (g *GrDeployer) getToolsPortfolioEnvConfigs() []*horizonapi.EnvConfig {
	var envs []*horizonapi.EnvConfig
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "SWIP_VAULT_ADDRESS", KeyOrVal: "https://vault:8200"})
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "VAULT_CACERT", KeyOrVal: "/mnt/vault/ca/vault_cacrt"})
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "VAULT_CLIENT_KEY", KeyOrVal: "/mnt/vault/key/vault_server_key"})
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "VAULT_CLIENT_CERT", KeyOrVal: "/mnt/vault/cert/vault_server_cert"})

	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "CONNECTION_POOL_SIZE", KeyOrVal: "10"})
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "LOG_LEVEL", KeyOrVal: "INFO"})
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "SPRING_PROFILE", KeyOrVal: "production"})

	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "POSTGRES_HOST", KeyOrVal: "postgres"})
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "POSTGRES_PORT", KeyOrVal: "5432"})
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "POSTGRES_PASSWORD", KeyOrVal: "test"})
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "POSTGRES_USERNAME", KeyOrVal: "admin"})

	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "SWIP_ROOT_DOMAIN", KeyOrVal: g.Grspec.IngressHost})
	return envs
}