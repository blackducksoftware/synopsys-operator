package blackduck

import (
	"testing"
)

func TestEnv(t *testing.T) {
	i, e := GetHubKnobs()
	if len(i) < 10 {
		t.Fail()
	}
	if len(e) < 10 {
		t.Fail()
	}
	for i, v := range e {
		t.Log(i, v)
	}

	for ii, v := range i {
		t.Log(ii, v)
	}

}
