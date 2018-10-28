package util

import (
	"os/exec"
	"testing"
	"time"
)

func TestA(t *testing.T) {
	cmd := &exec.Cmd{
		Args: []string{"sleep", "100"},
	}
	timer := time.NewTimer(5 * time.Second)
	go func() {
		<-timer.C
		t.Fail()
		return
	}()
	RunWithTimeout(cmd, 2*time.Second)
	t.Log("passed! finished before channel tripped")

}
