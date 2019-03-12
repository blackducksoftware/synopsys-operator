package containers

import (
	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
)

// GetFrontendDeployment returns the front end deployment
func (g *RgpDeployer) GetFrontendDeployment() *components.Deployment {
	deployConfig := &horizonapi.DeploymentConfig{
		Name:      "frontend-service",
		Namespace: g.Grspec.Namespace,
	}
	return util.CreateDeployment(deployConfig, g.GetFrontendPod())
}

// GetFrontendPod returns the front end pod
func (g *RgpDeployer) GetFrontendPod() *components.Pod {
	var containers []*util.Container
	container := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{
			Name:       "frontend-service",
			Image:      "gcr.io/snps-swip-staging/reporting-frontend-service:latest",
			PullPolicy: horizonapi.PullIfNotPresent,
			//MinMem:     "500Mi",
			//MaxMem:     "",
			//MinCPU:     "250m",
			//MaxCPU:     "",
		},
		EnvConfigs: g.getFrontendEnvConfigs(),
		PortConfig: []*horizonapi.PortConfig{
			{ContainerPort: "8080"},
		},
	}

	containers = append(containers, container)
	pod := util.CreatePod("frontend-service", "", nil, containers, nil, nil)
	//pod.AddHostAliases([]string{"frontend-service"})
	return pod
}

// GetFrontendService returns the front end service
func (g *RgpDeployer) GetFrontendService() *components.Service {
	service := components.NewService(horizonapi.ServiceConfig{
		Name:          "frontend-service",
		Namespace:     g.Grspec.Namespace,
		IPServiceType: horizonapi.ClusterIPServiceTypeDefault,
	})
	service.AddSelectors(map[string]string{
		"app": "frontend-service",
	})
	service.AddPort(horizonapi.ServicePortConfig{Name: "80", Port: 80, Protocol: horizonapi.ProtocolTCP, TargetPort: "8080"})
	return service
}

func (g *RgpDeployer) getFrontendEnvConfigs() []*horizonapi.EnvConfig {
	var envs []*horizonapi.EnvConfig
	envs = append(envs, g.getSwipEnvConfigs()...)
	return envs
}
