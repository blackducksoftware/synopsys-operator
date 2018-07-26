/*
Copyright (C) 2018 Synopsys, Inc.

Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements. See the NOTICE file
distributed with this work for additional information
regarding copyright ownership. The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License. You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied. See the License for the
specific language governing permissions and limitations
under the License.
*/

package protoform

import (
	"encoding/json"
	"flag"
	"fmt"
	"path/filepath"
	"reflect"

	log "github.com/sirupsen/logrus"

	"github.com/blackducksoftware/perceptor-protoform/pkg/api"

	"github.com/blackducksoftware/horizon/pkg/deployer"

	"github.com/spf13/viper"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// Installer handles deploying configured components to a cluster
type Installer struct {
	*deployer.Deployer

	// Config will have all viper inputs and default values
	Config protoformConfig
}

// NewInstaller creates an Installer object
func NewInstaller(defaults *api.ProtoformDefaults, path string) (*Installer, error) {
	var i Installer
	var config *rest.Config
	var err error

	pc := readConfig(path)
	setDefaults(defaults, pc)

	if !pc.DryRun {
		// creates the in-cluster config
		config, err = rest.InClusterConfig()
		if err != nil {
			log.Infof("unable to get in cluster config: %v", err)
			log.Infof("trying to use local config")
			config, err = newKubeConfigFromOutsideCluster()
			if err != nil {
				log.Errorf("unable to retrive the local config: %v", err)
				return nil, fmt.Errorf("failed to find a valid cluster config")
			}
		}
	} else {
		config = &rest.Config{}
	}

	d, err := deployer.NewDeployer(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create deployer: %v", err)
	}

	i = Installer{d, *pc}

	i.prettyPrint(i.Config)

	return &i, nil
}

func setDefaults(defaults *api.ProtoformDefaults, config *protoformConfig) {
	configFields := reflect.ValueOf(config).Elem()
	defaultFields := reflect.ValueOf(defaults).Elem()
	for cnt := 0; cnt < configFields.NumField(); cnt++ {
		fieldName := configFields.Type().Field(cnt).Name
		field := configFields.Field(cnt)
		defaultValue := defaultFields.FieldByName(fieldName)
		if defaultValue.IsValid() {
			switch configFields.Type().Field(cnt).Type.Kind().String() {
			case "string":
				if field.Len() == 0 {
					field.Set(defaultValue)
				}
			case "slice":
				if field.Len() == 0 {
					field.Set(defaultValue)
				}
			case "int":
				if field.Int() == 0 {
					field.Set(defaultValue)
				}
			}
		}
	}
}

// We don't dynamically reload.
// If users want to dynamically reload,
// they can update the individual perceptor containers configmaps.
func readConfig(configPath string) *protoformConfig {
	config := protoformConfig{}

	log.Debug("*************** [protoform] initializing  ****************")
	log.Infof("Config Path: %s", configPath)
	viper.SetConfigFile(configPath)

	// these need to be set before we read in the config!
	viper.SetEnvPrefix("PCP")
	viper.BindEnv("HubUserPassword")
	if viper.GetString("hubuserpassword") == "" {
		viper.Debug()
		log.Panic("no hub database password secret supplied.  Please inject PCP_HUBUSERPASSWORD as a secret and restart")
	}

	config.HubUserPasswordEnvVar = "PCP_HUBUSERPASSWORD"
	config.ViperSecret = "protoform"

	err := viper.ReadInConfig()
	if err != nil {
		log.Errorf("unable to read the config file! The input config file path is %s. Using defaults for everything", configPath)
	}

	internalRegistry := viper.GetStringSlice("InternalDockerRegistries")
	viper.Set("InternalDockerRegistries", internalRegistry)

	viper.Unmarshal(&config)

	// Set the Log level by reading the loglevel from config
	log.Infof("Log level : %s", config.LogLevel)
	level, _ := log.ParseLevel(config.LogLevel)
	log.SetLevel(level)

	log.Debug("*************** [protoform] done reading in config ****************")
	return &config
}

// AddPerceptorResources method is to support perceptor projects TODO: Remove perceptor specific code
func (i *Installer) AddPerceptorResources() {
	i.configServiceAccounts()
	isValid := i.sanityCheckServices()
	if isValid == false {
		log.Panic("Please set the service accounts correctly!")
	}

	i.substituteDefaultImageVersion()
	i.addPerceptorResources()
}

func (i *Installer) substituteDefaultImageVersion() {
	if len(i.Config.PerceptorImageVersion) == 0 {
		i.Config.PerceptorImageVersion = i.Config.DefaultVersion
	}
	if len(i.Config.ScannerImageVersion) == 0 {
		i.Config.ScannerImageVersion = i.Config.DefaultVersion
	}
	if len(i.Config.PerceiverImageVersion) == 0 {
		i.Config.PerceiverImageVersion = i.Config.DefaultVersion
	}
	if len(i.Config.ImageFacadeImageVersion) == 0 {
		i.Config.ImageFacadeImageVersion = i.Config.DefaultVersion
	}
	if len(i.Config.SkyfireImageVersion) == 0 {
		i.Config.SkyfireImageVersion = i.Config.DefaultVersion
	}
}

func (i *Installer) configServiceAccounts() {
	// TODO Viperize these env vars.
	if len(i.Config.ServiceAccounts) == 0 {
		log.Info("No service accounts exist.  Using defaults")

		svcAccounts := map[string]string{
			// WARNING: These service accounts need to exist !
			"pod-perceiver":          "perceiver",
			"image-perceiver":        "perceiver",
			"perceptor-image-facade": "perceptor-scanner",
		}
		// TODO programatically validate rather then sanity check.
		i.prettyPrint(svcAccounts)
		i.Config.ServiceAccounts = svcAccounts
	}
}

func (i *Installer) addPerceptorResources() {
	// Add Perceptor
	i.AddReplicationController(i.PerceptorReplicationController())
	i.AddService(i.PerceptorService())
	i.AddConfigMap(i.PerceptorConfigMap())

	// Add Perceptor Scanner
	rc, err := i.ScannerReplicationController()
	if err != nil {
		panic(fmt.Errorf("failed to create scanner replication controller: %v", err))
	}
	i.AddReplicationController(rc)
	i.AddService(i.ScannerService())
	i.AddService(i.ImageFacadeService())
	i.AddConfigMap(i.ScannerConfigMap())
	i.AddConfigMap(i.ImageFacadeConfigMap())
	i.AddServiceAccount(i.ScannerServiceAccount())
	i.AddClusterRoleBinding(i.ScannerClusterRoleBinding())

	if i.Config.PodPerceiver {
		rc, err := i.PodPerceiverReplicationController()
		if err != nil {
			panic(fmt.Errorf("failed to create pod perceiver: %v", err))
		}
		i.AddReplicationController(rc)
		i.AddService(i.PodPerceiverService())
		i.AddConfigMap(i.PerceiverConfigMap())
		i.AddServiceAccount(i.PodPerceiverServiceAccount())
		cr := i.PodPerceiverClusterRole()
		i.AddClusterRole(cr)
		i.AddClusterRoleBinding(i.PodPerceiverClusterRoleBinding(cr))
	}

	if i.Config.ImagePerceiver {
		rc, err := i.ImagePerceiverReplicationController()
		if err != nil {
			panic(fmt.Errorf("failed to create image perceiver: %v", err))
		}
		i.AddReplicationController(rc)
		i.AddService(i.ImagePerceiverService())
		i.AddConfigMap(i.PerceiverConfigMap())
		i.AddServiceAccount(i.ImagePerceiverServiceAccount())
		cr := i.ImagePerceiverClusterRole()
		i.AddClusterRole(cr)
		i.AddClusterRoleBinding(i.ImagePerceiverClusterRoleBinding(cr))
	}

	if i.Config.PerceptorSkyfire {
		rc, err := i.PerceptorSkyfireReplicationController()
		if err != nil {
			panic(fmt.Errorf("failed to create skyfire: %v", err))
		}
		i.AddReplicationController(rc)
		i.AddService(i.PerceptorSkyfireService())
		i.AddConfigMap(i.PerceptorSkyfireConfigMap())
		i.AddServiceAccount(i.PerceptorSkyfireServiceAccount())
		cr := i.PerceptorSkyfireClusterRole()
		i.AddClusterRole(cr)
		i.AddClusterRoleBinding(i.PerceptorSkyfireClusterRoleBinding(cr))
	}

	if i.Config.Metrics {
		dep, err := i.PerceptorMetricsDeployment()
		if err != nil {
			panic(fmt.Errorf("failed to create metrics: %v", err))
		}
		i.AddDeployment(dep)
		i.AddService(i.PerceptorMetricsService())
		i.AddConfigMap(i.PerceptorMetricsConfigMap())
	}

	if !i.Config.DryRun {
		i.AddController("Pod List Controller", NewPodListController(i.Config.Namespace))
	}

}

func (i *Installer) sanityCheckServices() bool {
	isValid := func(cn string) bool {
		for _, valid := range []string{"perceptor", "pod-perceiver", "image-perceiver", "perceptor-scanner", "perceptor-image-facade"} {
			if cn == valid {
				return true
			}
		}
		return false
	}
	for cn := range i.Config.ServiceAccounts {
		if !isValid(cn) {
			log.Panic("[protoform] failed at verifiying that the container name for a svc account was valid!")
		}
	}
	return true
}

func (i *Installer) prettyPrint(v interface{}) {
	b, _ := json.MarshalIndent(v, "", "  ")
	println(string(b))
}

func newKubeConfigFromOutsideCluster() (*rest.Config, error) {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Errorf("error creating default client config: %s", err)
		return nil, err
	}
	return config, err
}
