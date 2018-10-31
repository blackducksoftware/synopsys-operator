package hub

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
}
