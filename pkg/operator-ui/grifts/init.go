package grifts

import (
	"github.com/blackducksoftware/perceptor-protoform/pkg/operator-ui/actions"
	"github.com/gobuffalo/buffalo"
)

func init() {
	buffalo.Grifts(actions.App())
}
