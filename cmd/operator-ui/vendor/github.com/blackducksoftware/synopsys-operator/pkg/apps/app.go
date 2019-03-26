package apps

import (
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"k8s.io/client-go/rest"
)

type App struct {
	config           *protoform.Config
	kubeConfig       *rest.Config
}

func NewApp(config *protoform.Config, kubeConfig *rest.Config) *App {
	return &App{config: config, kubeConfig: kubeConfig}
}

func (a *App) Blackduck() *blackduck.Blackduck {
	return blackduck.NewBlackduck(a.config, a.kubeConfig)
}