package util

import (
	"bytes"
	"fmt"
	"os/exec"
	"time"

	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var openshift bool
var kube bool

func GetKubernetesClient() *kubernetes.Clientset {
	config, err := clientcmd.BuildConfigFromFlags("", "/$HOME/.kube/config")
	if err != nil {
		fmt.Println(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)

	if err != nil {
		fmt.Println(err.Error())
	}
	return clientset
}

func init() {
	_, exists := exec.LookPath("kubectl")
	if exists == nil {
		kube = true
	}
	_, ocexists := exec.LookPath("oc")
	if ocexists == nil {
		openshift = true
	}
	logrus.Infof("Clients: Openshift: %v, Kube: %v", openshift, kube)
}

// RunCmd is a simple wrapper to oc/kubectl exec that captures output.
// TODO consider replacing w/ go api but not crucial for now.
func RunKubeCmd(args ...string) error {
	var cmd2 *exec.Cmd

	// cluster-info in kube doesnt seem to be in
	// some versions of oc, but status is.
	// double check this.
	if args[0] == "cluster-info" && openshift {
		args[0] = "status"
	}
	if openshift {
		cmd2 = exec.Command("oc", args...)
	} else if kube {
		cmd2 = exec.Command("kubectl", args...)
	}
	stdoutErr, err := cmd2.CombinedOutput()
	fmt.Printf("%s\n", stdoutErr)
	if err != nil {
		fmt.Printf("Error running command !!!")
		return err
	}
	fmt.Printf("%s\n", stdoutErr)
	time.Sleep(1 * time.Second)
	return nil
}

// runWithTimeout runs a command and times it out at the specified duration
func RunWithTimeout(cmd *exec.Cmd, d time.Duration) (error, string) {
	timeout := time.After(d)

	// Use a bytes.Buffer to get the output
	var buf bytes.Buffer
	cmd.Stdout = &buf

	cmd.Start()

	// Use a channel to signal completion so we can use a select statement
	done := make(chan error)
	go func() { done <- cmd.Wait() }()

	// The select statement allows us to execute based on which channel
	// we get a message from first.
	select {
	case <-timeout:
		// Timeout happened first, kill the process and print a message.
		cmd.Process.Kill()
		return fmt.Errorf("Killed due to timeout"), buf.String()
	case err := <-done:
		if err != nil {
			return nil, buf.String()
		} else {
			return err, buf.String()
		}
	}
}
