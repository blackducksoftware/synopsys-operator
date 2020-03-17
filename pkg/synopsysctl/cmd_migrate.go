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

	"github.com/spf13/cobra"
)

// migrateCmd migrates a resource before upgrading Synopsys Operator
var migrateCmd = &cobra.Command{
	Use:           "migrate",
	Example:       "synopsysctl migrate -n <namespace>",
	Short:         "Migrate a Synopsys resource before upgrading the operator",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		// Check number of arguments
		if len(args) != 0 {
			cmd.Help()
			return fmt.Errorf("this command takes 0 argument")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

//func addIfNotEmpty( value interface{} ,path []string, helmConfig map[string]interface{} ) {
//
//	util.SetHelmValueInMap(helmConfig, path, value)
//}

func init() {
	// Add Migrate Commands
	rootCmd.AddCommand(migrateCmd)
	migrateCmd.Flags().StringVarP(&namespace, "namespace", "n", namespace, "Namespace of the instance(s)")

}
