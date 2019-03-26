package apps

import (
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"k8s.io/client-go/rest"
)

// App struct
type App struct {
	config     *protoform.Config
	kubeConfig *rest.Config
}

// NewApp will return an App
func NewApp(config *protoform.Config, kubeConfig *rest.Config) *App {
	return &App{config: config, kubeConfig: kubeConfig}
}

// Blackduck will return a Blackduck
func (a *App) Blackduck() *blackduck.Blackduck {
	return blackduck.NewBlackduck(a.config, a.kubeConfig)
}
