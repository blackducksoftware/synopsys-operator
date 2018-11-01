package containers

package containers
 
import (
	"testing"

	horizonapi "github.com/jayunit100/horizon/pkg/api"
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
 

func TestImageTag(t *testing.T) {
	hubVersion := "5.0.0"
	externalVersion := "1.0.0"
	hubSpec := &v1.HubSpec{ImageTagMap: map[string]string{"authentication": hubVersion, "documentation": hubVersion, "jobrunner": hubVersion,
		"registration": hubVersion, "scan": hubVersion, "webapp": hubVersion, "cfssl": externalVersion, "logstash": externalVersion,
		"nginx": externalVersion, "solr": externalVersion, "zookeeper": externalVersion}, HubVersion: "4.5.0"}
	creater := NewCreater(nil, hubSpec, nil, []*horizonapi.EnvConfig{}, []*horizonapi.EnvConfig{}, nil, nil)

	external100 := []string{"zookeeper", "nginx", "solr", "logstash", "cfssl"}
	internal50 := []string{"registration", "webapp", "jobrunner", "documentation", "scan", "authentication"}
	for _, v := range external100 {
		if creater.getTag(v) == externalVersion {
			fmt.Printf("%s: %s\n", v, creater.getTag(v))
		} else {
			t.Fail()
		}
	}
	for _, v := range internal50 {
		if creater.getTag(v) == hubVersion {
			fmt.Printf("%s: %s\n", v, creater.getTag(v))
		} else {
			t.Fail()
		}
	}

	hubSpec1 := &v1.HubSpec{HubVersion: "4.5.0"}
	creater = NewCreater(nil, hubSpec1, nil, []*horizonapi.EnvConfig{}, []*horizonapi.EnvConfig{}, nil, nil)
	all50 := []string{"zookeeper", "nginx", "solr", "logstash", "cfssl", "registration", "webapp", "jobrunner", "documentation", "scan", "authentication"}
	for _, v := range all50 {
		if creater.getTag(v) == "4.5.0" {
			fmt.Printf("%s: %s\n", v, creater.getTag(v))
		} else {
			t.Fail()
		}
	}
}
