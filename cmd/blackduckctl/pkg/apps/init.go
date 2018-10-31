package apps

// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// This command encodes all the bootstrap components into the cluster.
// It should be run once : When the user first installs the blackduck operator.

import (
	"fmt"

	bdutil "github.com/blackducksoftware/perceptor-protoform/cmd/blackduckctl/pkg/util"

	"github.com/blackducksoftware/perceptor-protoform/pkg/alert"
	"github.com/blackducksoftware/perceptor-protoform/pkg/hub"
	"github.com/sirupsen/logrus"

	horizondep "github.com/blackducksoftware/horizon/pkg/deployer"
	"github.com/blackducksoftware/perceptor-protoform/pkg/opssight"
	"github.com/blackducksoftware/perceptor-protoform/pkg/protoform"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
)

var InstallOperatorCommand = &cobra.Command{
	Use:   "init",
	Short: "Initialize the blackduck operator in your cluster.",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			fmt.Printf("No config provided, will use default secret and default configs!")
		}
		return nil
	},

	Run: func(cmd *cobra.Command, args []string) {
		fs := cmd.Flags()
		regKey, _ := fs.GetString("defaultPassword")
		ns, _ := fs.GetString("namespace")
		containerVersion, _ := fs.GetString("containerVersion")
		configPath, err := fs.GetString("configPath")
		bootstrap, _ := fs.GetBool("boostrap")
		export, _ := fs.GetBool("dry-run")

		if err != nil {
			panic("Cannot proceed, no configPath given for the input to protoform.")
		}

		// --bootstrap installs the product for the first time.
		if bootstrap {
			rc, svc, cm, sa, crb, svc2, rc2, cm2 := protoform.GetBootstrapComponents(ns, containerVersion, regKey)

			//
			// BOOTSTRAP: ITS ALIVE !!!
			// TODO, test if this works from openshift, and build a YAML export option.
			//
			config, err := clientcmd.BuildConfigFromFlags("", "/$HOME/.kube/config")
			if err != nil {
				panic(fmt.Sprintf("Can't proceed : no client available: %v", err))
			}
			deployer, _ := horizondep.NewDeployer(config)
			deployer.AddReplicationController(rc)
			deployer.AddService(svc)
			deployer.AddConfigMap(cm)
			deployer.AddServiceAccount(sa)
			deployer.AddClusterRoleBinding(crb)
			deployer.AddService(svc2)
			deployer.AddReplicationController(rc2)
			deployer.AddConfigMap(cm2)
			if export == true {
				for _, v := range deployer.Export() {
					fmt.Println(v)
				}
				return
			}
			err1 := deployer.Run()
			if err != nil {
				logrus.Infof("Errors during deployment bootstrap: %v, are you sure it went ok / are you sure you wanted to bootstrap?", err1)
			}
		}
		// Now run the operator.
		if configPath != "" {
			runProtoform(configPath)
		}
	},
}

// implementing init is important ! thats how cobra knows to bind your 'app' to top level command.
func init() {
	RootCmd.AddCommand(InstallOperatorCommand)
	InstallOperatorCommand.PersistentFlags().String("defaultPassword", "blackduck", "Default password to use for 'blackduck' instances.")
	InstallOperatorCommand.PersistentFlags().String("namespace", "blackduck-operator", "Namespace to run the operator in.")
	InstallOperatorCommand.PersistentFlags().String("containerVersion", "master", "Code branch to run blackduck-operator off of.")
	InstallOperatorCommand.PersistentFlags().String("configPath", "", "Path to YAML for custom config options.")
	InstallOperatorCommand.PersistentFlags().String("bootstrap", "false", "Wether or not to bootstrap all operator components.")
	InstallOperatorCommand.PersistentFlags().String("export", "false", "Wether or not to export bootstrap components as plain text (i.e. to create manually)")

}

func runProtoform(configPath string) {
	deployer, err := protoform.NewController(configPath)
	if err != nil {
		panic(err)
	}

	stopCh := make(chan struct{})

	alertController := alert.NewCRDInstaller(deployer.Config, deployer.KubeConfig, deployer.KubeClientSet, bdutil.GetAlertDefaultValue(), stopCh)
	deployer.AddController(alertController)

	hubController := hub.NewCRDInstaller(deployer.Config, deployer.KubeConfig, deployer.KubeClientSet, bdutil.GetHubDefaultValue(), stopCh)
	deployer.AddController(hubController)

	opssSightController, err := opssight.NewCRDInstaller(&opssight.Config{
		Config:        deployer.Config,
		KubeConfig:    deployer.KubeConfig,
		KubeClientSet: deployer.KubeClientSet,
		Defaults:      bdutil.GetOpsSightDefaultValue(),
		Threadiness:   deployer.Config.Threadiness,
		StopCh:        stopCh,
	})
	if err != nil {
		panic(err)
	}
	deployer.AddController(opssSightController)

	logrus.Info("Starting deployer.  All controllers have been added to horizon.")
	if err = deployer.Deploy(); err != nil {
		logrus.Errorf("ran into errors during deployment, but continuing anyway: %s", err.Error())
	}

	<-stopCh
}
