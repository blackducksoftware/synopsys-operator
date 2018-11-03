package grifts

import (
	"github.com/blackducksoftware/perceptor-protoform/cmd/operator_ui/actions"
	"github.com/gobuffalo/buffalo"
)

func init() {
	buffalo.Grifts(actions.App())
}
