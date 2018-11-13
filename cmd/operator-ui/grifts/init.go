package grifts

import (
	"github.com/blackducksoftware/synopsys-operator/cmd/operator-ui/actions"
	"github.com/gobuffalo/buffalo"
)

func init() {
	buffalo.Grifts(actions.App())
}
