package blackduck

import (
	"github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
)

// Creater interface
type Creater interface {
	Ensure(blackduck *v1.Blackduck) error
	Versions() []string
}
