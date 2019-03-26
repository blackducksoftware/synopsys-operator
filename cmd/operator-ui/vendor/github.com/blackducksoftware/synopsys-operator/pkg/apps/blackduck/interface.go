package blackduck

import (
"github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
)

type Creater interface {
	Create(blackduck *v1.Blackduck) error
	Start(blackduck *v1.Blackduck) error
	Stop(blackduck *v1.Blackduck) error
	Delete(namespace string) error
	Versions() []string
}