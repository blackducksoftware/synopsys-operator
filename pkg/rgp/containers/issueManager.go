package containers

import (
	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
)

// GetIssueManagerDeployment returns the issue manager deployment
func (g *RgpDeployer) GetIssueManagerDeployment() *components.Deployment {
	deployConfig := &horizonapi.DeploymentConfig{
		Name:      "rp-issue-manager",
		Namespace: g.Grspec.Namespace,
	}
	return util.CreateDeployment(deployConfig, g.GetIssueManagerPod())
}

// GetIssueManagerPod returns the issue manager pod
func (g *RgpDeployer) GetIssueManagerPod() *components.Pod {
	envs := g.getIssueManagerEnvConfigs()

	var containers []*util.Container

	container := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{
			Name:       "rp-issue-manager",
			Image:      "gcr.io/snps-swip-staging/reporting-rp-issue-manager:latest",
			PullPolicy: horizonapi.PullIfNotPresent,
			//MinMem:     "1Gi",
			//MaxMem:     "",
			//MinCPU:     "500m",
			//MaxCPU:     "",
		},
		VolumeMounts: g.getIssueManagerVolumeMounts(),
		EnvConfigs:   envs,
		PortConfig: []*horizonapi.PortConfig{
			{ContainerPort: "6888"},
		},
	}

	containers = append(containers, container)
	return util.CreatePod("rp-issue-manager", "", g.getReportVolumes(), containers, nil, nil)
}

// GetIssueManagerService returns the issue manager service
func (g *RgpDeployer) GetIssueManagerService() *components.Service {
	service := components.NewService(horizonapi.ServiceConfig{
		Name:          "rp-issue-manager",
		Namespace:     g.Grspec.Namespace,
		IPServiceType: horizonapi.ClusterIPServiceTypeDefault,
	})
	service.AddSelectors(map[string]string{
		"app": "rp-issue-manager",
	})
	service.AddPort(horizonapi.ServicePortConfig{Name: "6888", Port: 6888, Protocol: horizonapi.ProtocolTCP, TargetPort: "6888"})
	service.AddPort(horizonapi.ServicePortConfig{Name: "admin", Port: 8081, Protocol: horizonapi.ProtocolTCP, TargetPort: "8081"})
	return service
}

func (g *RgpDeployer) getIssueManagerVolumes() []*components.Volume {
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

func (g *RgpDeployer) getIssueManagerVolumeMounts() []*horizonapi.VolumeMountConfig {
	var volumeMounts []*horizonapi.VolumeMountConfig
	volumeMounts = append(volumeMounts, &horizonapi.VolumeMountConfig{Name: "vault-cacert", MountPath: "/mnt/vault/ca"})
	volumeMounts = append(volumeMounts, &horizonapi.VolumeMountConfig{Name: "vault-client-key", MountPath: "/mnt/vault/key"})
	volumeMounts = append(volumeMounts, &horizonapi.VolumeMountConfig{Name: "vault-client-cert", MountPath: "/mnt/vault/cert"})

	return volumeMounts
}

func (g *RgpDeployer) getIssueManagerEnvConfigs() []*horizonapi.EnvConfig {
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
