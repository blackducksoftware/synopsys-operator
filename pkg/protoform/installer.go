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
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/blackducksoftware/perceptor-protoform/pkg/api"
	"github.com/blackducksoftware/perceptor-protoform/pkg/apps"
	"github.com/blackducksoftware/perceptor-protoform/pkg/apps/alert"
	"github.com/blackducksoftware/perceptor-protoform/pkg/apps/hubfederator"
	"github.com/blackducksoftware/perceptor-protoform/pkg/apps/perceptor"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/horizon/pkg/deployer"

	"github.com/spf13/viper"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	securityclient "github.com/openshift/client-go/security/clientset/versioned/typed/security/v1"
)

// Installer handles deploying configured components to a cluster
type Installer struct {
	deployer *deployer.Deployer

	config api.ProtoformConfig

	appDefaults map[apps.AppType]interface{}
	apps        []apps.AppInstallerInterface

	osSecurityClient *securityclient.SecurityV1Client
}

// NewInstaller creates an Installer object
func NewInstaller(path string) (*Installer, error) {
	var i Installer
	var config *rest.Config
	var err error

	pc := readConfig(path)

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

	osClient, err := securityclient.NewForConfig(config)
	if err != nil {
		osClient = nil
	}

	d, err := deployer.NewDeployer(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create deployer: %v", err)
	}

	i = Installer{
		deployer:         d,
		config:           *pc,
		appDefaults:      make(map[apps.AppType]interface{}),
		apps:             make([]apps.AppInstallerInterface, 0),
		osSecurityClient: osClient,
	}

	i.prettyPrint(i.config)

	return &i, nil
}

// LoadAppDefault will store the defaults for the provided app
func (i *Installer) LoadAppDefault(app apps.AppType, defaults interface{}) {
	i.appDefaults[app] = defaults
}

// We don't dynamically reload.
// If users want to dynamically reload,
// they can update the individual perceptor containers configmaps.
func readConfig(configPath string) *api.ProtoformConfig {
	config := api.ProtoformConfig{}

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

	viper.SetDefault("ViperSecret", "protoform")

	err := viper.ReadInConfig()
	if err != nil {
		log.Errorf("the input config file path is %s. unable to read the config file due to %+v! Using defaults for everything", configPath, err)
	}

	internalRegistry := []byte(viper.GetString("InternalRegistries"))
	internalRegistries := make([]perceptor.RegistryAuth, 0)
	if !strings.EqualFold(viper.GetString("InternalRegistries"), "") {
		err = json.Unmarshal(internalRegistry, &internalRegistries)
		if err != nil {
			log.Errorf("Internal registry is %s, unable to marshal the internal registries due to %+v", viper.GetString("InternalRegistries"), err)
			os.Exit(1)
		}
		log.Infof("internalRegistries: %+v", internalRegistries)
		viper.Set("InternalRegistries", internalRegistries)
	}

	setViperAppStructs(&config)
	viper.Unmarshal(&config)

	// Set the Log level by reading the loglevel from config
	log.Infof("Log level : %s", config.DefaultLogLevel)
	level, _ := log.ParseLevel(config.DefaultLogLevel)
	log.SetLevel(level)

	log.Debug("*************** [protoform] done reading in config ****************")
	return &config
}

func setViperAppStructs(conf *api.ProtoformConfig) {
	if viper.Get("Apps") != nil {
		conf.Apps = &api.ProtoformApps{}
		if viper.Get("Apps.PerceptorConfig") != nil {
			conf.Apps.PerceptorConfig = &perceptor.AppConfig{}
		}
		if viper.Get("Apps.AlertConfig") != nil {
			conf.Apps.AlertConfig = &alert.AppConfig{}
		}
		if viper.Get("Apps.HubFederatorConfig") != nil {
			conf.Apps.HubFederatorConfig = &hubfederator.AppConfig{}
		}
	}
}

// Run will start the installer
func (i *Installer) Run(stopCh chan struct{}) error {
	if i.config.Apps != nil {
		err := i.createApps()
		if err != nil {
			return err
		}
	}

	err := i.preDeploy()
	if err != nil {
		return err
	}

	if !i.config.DryRun {
		err = i.deployer.Run()
		if err != nil {
			return err
		}

		err = i.postDeploy()
		if err != nil {
			return err
		}

		i.deployer.StartControllers(stopCh)
	} else {
		i.prettyPrint(fmt.Sprintf("%+v", i.deployer))
	}

	return nil
}

func (i *Installer) createApps() error {
	if i.config.Apps.PerceptorConfig != nil {
		if len(i.config.Apps.PerceptorConfig.LogLevel) == 0 {
			i.config.Apps.PerceptorConfig.LogLevel = i.config.DefaultLogLevel
		}

		// Remove this override once secrets are created by app
		i.config.Apps.PerceptorConfig.HubUserPassword = i.config.HubUserPassword

		p, err := perceptor.NewApp(i.appDefaults[apps.Perceptor])
		if err != nil {
			return fmt.Errorf("failed to load perceptor: %v", err)
		}
		err = i.addApp(p, i.config.Apps.PerceptorConfig)
		if err != nil {
			return fmt.Errorf("failed to create perceptor app: %v", err)
		}
	}

	if i.config.Apps.AlertConfig != nil {
		a, err := alert.NewApp(i.appDefaults[apps.Alert])
		if err != nil {
			return fmt.Errorf("failed to load alert: %v", err)
		}
		err = i.addApp(a, i.config.Apps.AlertConfig)
		if err != nil {
			return fmt.Errorf("failed to create alert app: %v", err)
		}
	}

	if i.config.Apps.HubFederatorConfig != nil {
		h, err := hubfederator.NewApp(i.appDefaults[apps.HubFederator])
		if err != nil {
			return fmt.Errorf("failed to load hub federator: %v", err)
		}
		err = i.addApp(h, i.config.Apps.HubFederatorConfig)
		if err != nil {
			return fmt.Errorf("failed to create hub federator app: %v", err)
		}
	}

	if !i.config.DryRun {
		namespaces := []string{}
		for _, n := range i.apps {
			namespaces = i.appendIfMissing(n.GetNamespace(), namespaces)
		}
		i.deployer.AddController("Pod List Controller", NewPodListController(namespaces))
	}

	return nil
}

func (i *Installer) addApp(app apps.AppInstallerInterface, config interface{}) error {
	err := app.Configure(config)
	if err != nil {
		return fmt.Errorf("failed to configure app: %+v", err)
	}
	i.apps = append(i.apps, app)

	return nil
}

func (i *Installer) appendIfMissing(new string, list []string) []string {
	for _, o := range list {
		if strings.Compare(new, o) == 0 {
			return list
		}
	}
	return append(list, new)
}

func (i *Installer) preDeploy() error {
	for _, app := range i.apps {
		appComponents, err := app.GetComponents()
		if err != nil {
			return err
		}

		if appComponents != nil {
			i.addNS(app.GetNamespace())
			i.addRCs(appComponents.ReplicationControllers)
			i.addSvcs(appComponents.Services)
			i.addCMs(appComponents.ConfigMaps)
			i.addSAs(appComponents.ServiceAccounts)
			i.addCRs(appComponents.ClusterRoles)
			i.addCRBs(appComponents.ClusterRoleBindings)
			i.addDeploys(appComponents.Deployments)
			i.addSecrets(appComponents.Secrets)
		}
	}
	return nil
}

func (i *Installer) addNS(ns string) {
	comp := components.NewNamespace(horizonapi.NamespaceConfig{
		Name: ns,
	})

	i.deployer.AddNamespace(comp)
}

func (i *Installer) addRCs(list []*components.ReplicationController) {
	if len(list) > 0 {
		for _, rc := range list {
			i.deployer.AddReplicationController(rc)
		}
	}
}

func (i *Installer) addSvcs(list []*components.Service) {
	if len(list) > 0 {
		for _, svc := range list {
			i.deployer.AddService(svc)
		}
	}
}

func (i *Installer) addCMs(list []*components.ConfigMap) {
	if len(list) > 0 {
		for _, cm := range list {
			i.deployer.AddConfigMap(cm)
		}
	}
}

func (i *Installer) addSAs(list []*components.ServiceAccount) {
	if len(list) > 0 {
		for _, sa := range list {
			i.deployer.AddServiceAccount(sa)
		}
	}
}

func (i *Installer) addCRs(list []*components.ClusterRole) {
	if len(list) > 0 {
		for _, cr := range list {
			i.deployer.AddClusterRole(cr)
		}
	}
}

func (i *Installer) addCRBs(list []*components.ClusterRoleBinding) {
	if len(list) > 0 {
		for _, crb := range list {
			i.deployer.AddClusterRoleBinding(crb)
		}
	}
}

func (i *Installer) addDeploys(list []*components.Deployment) {
	if len(list) > 0 {
		for _, d := range list {
			i.deployer.AddDeployment(d)
		}
	}
}

func (i *Installer) addSecrets(list []*components.Secret) {
	if len(list) > 0 {
		for _, s := range list {
			i.deployer.AddSecret(s)
		}
	}
}

func (i *Installer) postDeploy() error {
	if i.osSecurityClient != nil {
		// Since there is a security client it means the cluster target is openshift
		// TODO this should be moved inside the perceptor app structure
		if i.config.Apps.PerceptorConfig != nil {
			// Need to add the perceptor-scanner service account to the privelged scc
			scc, err := i.osSecurityClient.SecurityContextConstraints().Get("privileged", meta_v1.GetOptions{})
			if err != nil {
				return fmt.Errorf("failed to get scc privileged: %v", err)
			}

			var scannerAccount string
			for _, o := range i.apps {
				if p, ok := o.(*perceptor.App); ok {
					s := p.ScannerServiceAccount()
					scannerAccount = fmt.Sprintf("system:serviceaccount:%s:%s", p.GetNamespace(), s.GetName())
					break
				}
			}

			// Only add the service account if it isn't already in the list of users for the privileged scc
			exists := false
			for _, u := range scc.Users {
				if strings.Compare(u, scannerAccount) == 0 {
					exists = true
					break
				}
			}

			if !exists {
				scc.Users = append(scc.Users, scannerAccount)

				_, err = i.osSecurityClient.SecurityContextConstraints().Update(scc)
				if err != nil {
					return fmt.Errorf("failed to update scc privileged: %v", err)
				}
			}
		}
	}

	return nil
}

func (i *Installer) prettyPrint(v interface{}) {
	b, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(b))
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
