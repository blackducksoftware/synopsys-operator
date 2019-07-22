package utils

import (
	"fmt"
	"github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/types"
)

func SetLimits(containerConfig *api.ContainerConfig, config types.Container) {
	if config.MinCPU != nil {
		containerConfig.MinCPU = fmt.Sprintf("%d", *config.MinCPU)
	}

	if config.MaxCPU != nil {
		containerConfig.MaxCPU = fmt.Sprintf("%d", *config.MaxCPU)
	}

	if config.MinMem != nil {
		containerConfig.MinMem = fmt.Sprintf("%dM", *config.MinMem)
	}

	if config.MaxMem != nil {
		containerConfig.MaxMem = fmt.Sprintf("%dM", *config.MaxMem)
	}
}
