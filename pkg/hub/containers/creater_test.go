package containers

import (
	"testing"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/perceptor-protoform/pkg/util"
	"github.com/sirupsen/logrus"
)

type TG struct {
}

func (t *TG) getUID(s string) *int64 {
	x := int64(100)
	return &x
}
func (t *TG) getTag(s string) string {
	return "CORRECT"
}
func TestC(t *testing.T) {
	c := &TG{}
	myCont := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "documentation", Image: "a/b/c:tag"},
	}
	if myCont.ContainerConfig.Image == "a/b/c:CORRECT" {
		logrus.Infof("test setup isnt right")
		t.Fail()
	}
	PostEdit(myCont, c)
	if myCont.ContainerConfig.Image != "a/b/c:CORRECT" {
		logrus.Infof("Got wrong tag %v", myCont.ContainerConfig.Image)
		t.Fail()
	}
	if *myCont.ContainerConfig.UID != 100 {
		t.Fail()
	}
}
