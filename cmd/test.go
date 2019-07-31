package main

import (
	"fmt"
	_ "github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"github.com/prometheus/common/log"
	"os"
)

var version string

func main() {
	log.Infof("version: %s", version)
	if len(os.Args) > 1 {
		configPath := os.Args[1]
		runProtoform(configPath, version)
		fmt.Printf("Config path: %s", configPath)
		return
	}
	log.Warn("no config file sent. running operator with environment variable and default settings")
	runProtoform("", version)
}

func runProtoform(configPath string, version string) {

	config, err := protoform.GetConfig(configPath, version)
	if err != nil {
		panic("Failed to load configuration")
	}
	if config == nil {
		panic("expected non-nil config, but got nil")
	}

	config.DryRun = true

	kubeConfig, err := protoform.GetKubeConfig("", false)
	if err != nil {
		panic("unable to create config for both in-cluster and external to cluster")
	}

	kubeClientSet, err := protoform.GetKubeClientSet(kubeConfig)
	if err != nil {
		panic("unable to create Kubernetes clientset")
	}

	_, err = protoform.NewDeployer(config, kubeConfig, kubeClientSet)

	if err != nil {
		panic(err.Error())
	}

}
