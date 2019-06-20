/*
Copyright (C) 2019 Synopsys, Inc.

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

package synopsysctl

import (
	"fmt"

	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/types"
)

// migrateCmd migrates a resource before upgrading Synopsys Operator
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate a Synopsys resource before upgrading the operator",
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("not a valid command")
	},
}

// migrateBlackduckCmd migrates one or many Black Duck instances
var migrateBlackduckCmd = &cobra.Command{
	Use:     "blackduck NAME...",
	Example: "synopsysctl migrate blackduck <name>\nsynopsysctl migrate blackduck <name1> <name2> <name3>\nsynopsysctl migrate blackduck <name> -n <namespace>\nsynopsysctl migrate blackduck <name1> <name2> <name3> -n <namespace>",
	Aliases: []string{"blackducks"},
	Short:   "Migrate one or many Black Duck instances",
	Args: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		blackDuckNamespace := namespace
		for _, blackDuckName := range args {
			if len(namespace) == 0 {
				blackDuckNamespace = blackDuckName
			}
			log.Infof("migrating Black Duck '%s' in namespace '%s'...", blackDuckName, blackDuckNamespace)

			// ASSUMING ALL PASSWORDS HAVE REMAINED THE SAME, no need to pull from secret
			defaultPassword := util.Base64Encode([]byte("blackduck"))

			patch := fmt.Sprintf("{\"spec\":{\"adminPassword\":\"%s\",\"userPassword\":\"%s\", \"postgresPassword\":\"%s\"}}", defaultPassword, defaultPassword, defaultPassword)
			_, err := blackDuckClient.SynopsysV1().Blackducks(blackDuckNamespace).Patch(blackDuckNamespace, types.MergePatchType, []byte(patch))
			if err != nil {
				log.Errorf("error migrating Black Duck '%s' in namespace '%s' due to %+v", blackDuckName, blackDuckNamespace, err)
			}
			log.Infof("successfully submitted migrate Black Duck '%s' in namespace '%s'", blackDuckName, blackDuckNamespace)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)

	// Add Migrate Commands
	migrateBlackduckCmd.Flags().StringVarP(&namespace, "namespace", "n", namespace, "Namespace of the instance(s)")
	migrateCmd.AddCommand(migrateBlackduckCmd)
}
