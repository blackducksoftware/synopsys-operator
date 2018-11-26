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

package apps

import (
	"fmt"

	horizondep "github.com/blackducksoftware/horizon/pkg/deployer"
	"github.com/blackducksoftware/synopsys-operator/pkg/api/hub/v2"
	"github.com/blackducksoftware/synopsys-operator/pkg/hub"
	"github.com/blackducksoftware/synopsys-operator/pkg/hub/containers"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/spf13/cobra"
)

/**
type Config struct {
	DryRun                bool
	LogLevel              string
	Namespace             string
	Threadiness           int
	PostgresRestartInMins int
	NFSPath               string
	HubFederatorConfig    *HubFederatorConfig
}
**/

// InstallBlackduck ...
var InstallBlackduck = &cobra.Command{
	Use:   "install-blackduck",
	Short: "Install a blackduck instance (or export the YAML file for doing so).",
	Args: func(cmd *cobra.Command, args []string) error {
		_, err1 := cmd.PersistentFlags().GetBool("dry-run")
		_, err2 := cmd.PersistentFlags().GetString("namespace")
		if err1 != nil || err2 != nil {
			return fmt.Errorf("Args incorrect: %v %v", err1, err2)
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		dryRun, _ := cmd.PersistentFlags().GetBool("dry-run")
		namespace, _ := cmd.PersistentFlags().GetString("namespace")
		nfsPath, _ := cmd.PersistentFlags().GetString("nfs-path")
		postgresRestartMinutes, _ := cmd.PersistentFlags().GetInt32("postgres-restart-minutes")
		export, _ := cmd.PersistentFlags().GetBool("export")
		creator := hub.Creater{
			Config: &protoform.Config{
				DryRun:                dryRun,
				Namespace:             namespace,
				NFSPath:               nfsPath,
				PostgresRestartInMins: int(postgresRestartMinutes),
			},
		}

		if dryRun {
			deployer, _ := util.NewDeployer(nil)
			if export == true {
				for _, v := range deployer.Deployer.Export() {
					fmt.Println(v)
				}
				return
			}
		} else {
			restconf, _ := util.GetKubeConfig()
			deployer, _ := horizondep.NewDeployer(restconf)
			hubSpec := &v2.HubSpec{
				Namespace:  namespace,
				Flavor:     "small",
				HubVersion: "5.0.0.",
				ScanType:   "master",
				HubType:    "small",
			}

			creator.AddToDeployer(deployer, hubSpec, containers.GetContainersFlavor("SMALL"), nil)
		}
	},
}

// implementing init is important ! thats how cobra knows to bind your 'app' to top level command.
func init() {
	RootCmd.AddCommand(InstallOperatorCommand)
	InstallBlackduck.PersistentFlags().Bool("dry-run", false, "Print the yaml and exit.")
	InstallBlackduck.PersistentFlags().Int32("postgres-restart-minutes", 3, "Time before postgres is restarted.")
	InstallBlackduck.PersistentFlags().String("nfs-path", "", "Path to an NFS mount that operator will use to make PV's against.")
	InstallBlackduck.PersistentFlags().String("namespace", "blackduck", "The namespace you want to install blackduck into.")
}
