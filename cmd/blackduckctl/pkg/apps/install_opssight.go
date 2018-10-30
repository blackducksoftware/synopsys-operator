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

import (
	"fmt"

	horizondep "github.com/blackducksoftware/horizon/pkg/deployer"

	bdutil "github.com/blackducksoftware/perceptor-protoform/cmd/protoform-installer/blackduckctl/pkg/util"
	versioned "github.com/blackducksoftware/perceptor-protoform/pkg/opssight/client/clientset/versioned"

	opssightv1 "github.com/blackducksoftware/perceptor-protoform/pkg/api/opssight/v1"
	"github.com/blackducksoftware/perceptor-protoform/pkg/opssight"
	"github.com/blackducksoftware/perceptor-protoform/pkg/util"
	"github.com/spf13/cobra"
)

var InstallOpsSight = &cobra.Command{
	Use:   "install-opssight",
	Short: "Install a opssight instance (or export the YAML file for doing so).",
	Args: func(cmd *cobra.Command, args []string) error {
		_, err1 := cmd.PersistentFlags().GetBool("dry-run")
		_, err2 := cmd.PersistentFlags().GetString("namespace")
		if err1 != nil || err2 != nil {
			return fmt.Errorf("Args incorrect: %v %v %v", err1, err2)
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		dryRun, _ := cmd.PersistentFlags().GetBool("dry-run")
		namespace, _ := cmd.PersistentFlags().GetString("namespace")
		spec := opssightv1.OpsSightSpec{
			Perceptor: &opssightv1.Perceptor{
				Name:  "perceptor",
				Port:  3001,
				Image: "gcr.io/saas-hub-stg/blackducksoftware/perceptor:master",
				CheckForStalledScansPauseHours: 999999,
				StalledScanClientTimeoutHours:  999999,
				ModelMetricsPauseSeconds:       15,
				UnknownImagePauseMilliseconds:  15000,
				ClientTimeoutMilliseconds:      100000,
			},
			Perceiver: &opssightv1.Perceiver{
				EnableImagePerceiver: false,
				EnablePodPerceiver:   true,
				Port:                 3002,
				ImagePerceiver: &opssightv1.ImagePerceiver{
					Name:  "image-perceiver",
					Image: "gcr.io/saas-hub-stg/blackducksoftware/image-perceiver:master",
				},
				PodPerceiver: &opssightv1.PodPerceiver{
					Name:  "pod-perceiver",
					Image: "gcr.io/saas-hub-stg/blackducksoftware/pod-perceiver:master",
				},
				ServiceAccount:            "perceiver",
				AnnotationIntervalSeconds: 30,
				DumpIntervalMinutes:       30,
			},
			ScannerPod: &opssightv1.ScannerPod{
				Name: "perceptor-scanner",
				ImageFacade: &opssightv1.ImageFacade{
					Port:               3004,
					InternalRegistries: []opssightv1.RegistryAuth{},
					Image:              "gcr.io/saas-hub-stg/blackducksoftware/perceptor-imagefacade:master",
					ServiceAccount:     "perceptor-scanner",
					Name:               "perceptor-imagefacade",
				},
				Scanner: &opssightv1.Scanner{
					Name:                 "perceptor-scanner",
					Port:                 3003,
					Image:                "gcr.io/saas-hub-stg/blackducksoftware/perceptor-scanner:master",
					ClientTimeoutSeconds: 600,
				},
				ReplicaCount: 1,
			},
			Skyfire: &opssightv1.Skyfire{
				Image:          "gcr.io/saas-hub-stg/blackducksoftware/skyfire:master",
				Name:           "skyfire",
				Port:           3005,
				ServiceAccount: "skyfire",
			},
			Hub: &opssightv1.Hub{
				User:                         "sysadmin",
				Port:                         443,
				ConcurrentScanLimit:          2,
				TotalScanLimit:               1000,
				PasswordEnvVar:               "PCP_HUBUSERPASSWORD",
				Password:                     "blackduck",
				InitialCount:                 1,
				MaxCount:                     1,
				DeleteHubThresholdPercentage: 50,
				HubSpec: bdutil.GetHubDefaultValue(),
			},
			EnableMetrics: true,
			EnableSkyfire: false,
			DefaultCPU:    "300m",
			DefaultMem:    "1300Mi",
			LogLevel:      "debug",
			SecretName:    "perceptor",
		}
		sc := opssight.NewSpecConfig(&spec)
		compList, _ := sc.GetComponents()
		deployer, _ := horizondep.NewDeployer(nil)
		if dryRun == true {
			for _, crb := range compList.ClusterRoleBindings {
				deployer.AddClusterRoleBinding(crb)
			}
			for _, cr := range compList.ClusterRoles {
				deployer.AddClusterRole(cr)
			}
			for _, cm := range compList.ConfigMaps {
				deployer.AddConfigMap(cm)
			}
			for _, dep := range compList.Deployments {
				deployer.AddDeployment(dep)
			}
			for _, rc := range compList.ReplicationControllers {
				deployer.AddReplicationController(rc)
			}
			for _, sec := range compList.Secrets {
				deployer.AddSecret(sec)
			}
			for _, sa := range compList.ServiceAccounts {
				deployer.AddServiceAccount(sa)
			}
			for _, svc := range compList.Services {
				deployer.AddService(svc)
			}
			// print it all out
			for _, v := range deployer.Export() {
				fmt.Println(v)
			}
			return
		} else {
			// TODO catch errors
			restconf, _ := util.GetKubeConfig()
			cs, _ := versioned.NewForConfig(restconf)
			opssight := &opssightv1.OpsSight{
				Spec: spec,
			}
			cs.Synopsys().OpsSights(namespace).Create(opssight)
		}

	},
}

// implementing init is important ! thats how cobra knows to bind your 'app' to top level command.
func init() {
	RootCmd.AddCommand(InstallOperatorCommand)
	InstallOpsSight.PersistentFlags().Bool("dry-run", false, "Print the yaml and exit.")
	InstallOpsSight.PersistentFlags().Int32("postgres-restart-minutes", 3, "Time before postgres is restarted.")
	InstallOpsSight.PersistentFlags().String("nfs-path", "", "Path to an NFS mount that operator will use to make PV's against.")
	InstallOpsSight.PersistentFlags().String("namespace", "blackduck", "The namespace you want to install blackduck into.")
}
