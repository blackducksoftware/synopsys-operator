package rgp

import (
	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Vault stores the value configuration
type Vault struct {
	Namespace string
}

// GetConfigmap returns the vault configmap
func (v *Vault) GetConfigmap() *components.ConfigMap {
	cm := components.NewConfigMap(horizonapi.ConfigMapConfig{Namespace: v.Namespace, Name: "vault-policies"})
	cm.AddData(map[string]string{
		"auth-server.hcl": `path "secret/data/auth/*" {
      capabilities = ["create", "read", "update", "delete", "list"]
    }`,
		"auth-client.hcl": `path "secret/data/auth/public/*" {
      capabilities = ["list", "read"]
    }`,
	})

	return cm
}

// GetJob returns the vault job
func (v *Vault) GetJob() *v1.Job {

	job := &v1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: "vault-init",
		},
		Spec: v1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					ServiceAccountName: "vault-init",
					Containers: []corev1.Container{
						{
							Name:            "vault-init",
							ImagePullPolicy: corev1.PullIfNotPresent,
							Image:           "gcr.io/snps-swip-staging/vault-util:latest",
							Command:         []string{"vault-tls-init"},
							Env: []corev1.EnvVar{
								{
									Name:  "VAULT_SERVICE_NAME",
									Value: "vault",
								},
								{
									Name:  "VAULT_KUBERNETES_NAMESPACE",
									Value: v.Namespace,
								},
								{
									Name:  "VAULT_CLIENT_CERTIFICATES",
									Value: "auth-server,auth-client",
								},
							},
						},
					},
					RestartPolicy: corev1.RestartPolicyOnFailure,
				},
			},
		},
	}

	return job
}

// GetDeployment returns the vault deployment
func (v *Vault) GetDeployment() *components.Deployment {

	var containers []*util.Container

	container := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{
			Name:       "vault-init",
			Image:      "gcr.io/snps-swip-staging/vault-util:latest",
			PullPolicy: horizonapi.PullIfNotPresent,
			MinMem:     "",
			MaxMem:     "",
			MinCPU:     "",
			MaxCPU:     "",
			Command: []string{
				"vault-init",
			},
		},
		EnvConfigs:   v.getVaultEnvConfigs(),
		VolumeMounts: v.getVaultVolumeMounts(),
	}

	containers = append(containers, container)

	deployConfig := &horizonapi.DeploymentConfig{
		Name:      "vault-init",
		Namespace: v.Namespace,
		Replicas:  util.IntToInt32(1),
	}

	return util.CreateDeploymentFromContainer(deployConfig, "vault-init", containers, v.getVaultVolumes(), nil, nil)
}

func (v *Vault) getVaultVolumes() []*components.Volume {
	var volumes []*components.Volume

	volumes = append(volumes, components.NewSecretVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      "vault-tls-certificate",
		MapOrSecretName: "vault-tls-certificate",
	}))

	volumes = append(volumes, components.NewConfigMapVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      "vault-policy-configs",
		MapOrSecretName: "vault-policies",
	}))

	volumes = append(volumes, components.NewSecretVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      "auth-server-tls-certificate",
		MapOrSecretName: "auth-server-tls-certificate",
	}))

	volumes = append(volumes, components.NewSecretVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      "auth-client-tls-certificate",
		MapOrSecretName: "auth-client-tls-certificate",
	}))

	return volumes
}

// getConsulVolumeMounts will return the postgres volume mounts
func (v *Vault) getVaultVolumeMounts() []*horizonapi.VolumeMountConfig {
	var volumeMounts []*horizonapi.VolumeMountConfig
	volumeMounts = append(volumeMounts, &horizonapi.VolumeMountConfig{Name: "vault-tls-certificate", MountPath: "/vault/tls"})
	volumeMounts = append(volumeMounts, &horizonapi.VolumeMountConfig{Name: "auth-server-tls-certificate", MountPath: "/auth-server-tls-certificate"})
	volumeMounts = append(volumeMounts, &horizonapi.VolumeMountConfig{Name: "auth-client-tls-certificate", MountPath: "/auth-client-tls-certificate"})
	volumeMounts = append(volumeMounts, &horizonapi.VolumeMountConfig{Name: "vault-policy-configs", MountPath: "/vault/policies"})

	return volumeMounts
}

func (v *Vault) getVaultEnvConfigs() []*horizonapi.EnvConfig {
	var envs []*horizonapi.EnvConfig
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "VAULT_ADDR", KeyOrVal: "https://vault:8200"})
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "VAULT_CACERT", KeyOrVal: "/vault/tls/ca.crt"})
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "VAULT_INIT_SECRET", KeyOrVal: "vault-init-secret"})
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "VAULT_SECRET_ENGINE_VERSION", KeyOrVal: "v2"})
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "VAULT_KUBERNETES_NAMESPACE", KeyOrVal: v.Namespace})
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "VAULT_POLICY_CONFIGS", KeyOrVal: "/vault/policies"})
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "AUTH_SERVER_VAULT_CLIENT_CERTIFICATE", KeyOrVal: "/auth-server-tls-certificate/tls.crt"})
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "AUTH_CLIENT_VAULT_CLIENT_CERTIFICATE", KeyOrVal: "/auth-client-tls-certificate/tls.crt"})

	return envs
}

// GetSidecarUnsealContainer returns the side car container
func (v *Vault) GetSidecarUnsealContainer() *components.Container {
	container := components.NewContainer(horizonapi.ContainerConfig{
		Name:       "vault-sidecar",
		Image:      "gcr.io/snps-swip-staging/vault-util:latest",
		PullPolicy: horizonapi.PullIfNotPresent,
		Command:    []string{"vault-sidecar", "/vault/init"},
	})

	container.AddEnv(horizonapi.EnvConfig{
		Type:         horizonapi.EnvVal,
		NameOrPrefix: "VAULT_ADDR",
		KeyOrVal:     "https://localhost:8200",
	})

	container.AddEnv(horizonapi.EnvConfig{
		Type:         horizonapi.EnvVal,
		NameOrPrefix: "VAULT_CACERT",
		KeyOrVal:     "/vault/tls/ca.crt",
	})

	container.AddVolumeMount(horizonapi.VolumeMountConfig{
		Name:      "vault-tls-certificate",
		MountPath: "/vault/tls",
	})

	container.AddVolumeMount(horizonapi.VolumeMountConfig{
		Name:      "vault-init-secret",
		MountPath: "/vault/init",
	})

	return container
}
