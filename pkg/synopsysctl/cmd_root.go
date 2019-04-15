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
	"os"

	util "github.com/blackducksoftware/synopsys-operator/pkg/util"
	homedir "github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// Options flags for all commands
var cluster string
var kubeconfig string
var context string
var insecureSkipTLSVerify = false

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "synopsysctl",
	Short: "Command Line Tool for managing Synopsys Resources",
	Args: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 && args[0] == "--help" {
			return fmt.Errorf("Help Called")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Debugf("Running Non-Synopsysctl Command\n")
		out, err := util.RunKubeCmd(restconfig, kube, openshift, args...)
		if err != nil {
			log.Errorf("Error with KubeCmd: %s", out)
			return nil
		}
		fmt.Printf("%+v", out)
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	setResourceClients()              // sets kubeconfig and initializes resource client libraries
	rootCmd.DisableFlagParsing = true // lets rootCmd pass flags to kube/oc
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cluster, "cluster", cluster, "name of the kubeconfig cluster to use")
	rootCmd.PersistentFlags().StringVar(&kubeconfig, "kubeconfig", kubeconfig, "path to the kubeconfig file to use for CLI requests")
	rootCmd.PersistentFlags().StringVar(&context, "context", context, "name of the kubeconfig context to use")
	rootCmd.PersistentFlags().BoolVar(&insecureSkipTLSVerify, "insecure-skip-tls-verify", insecureSkipTLSVerify, "server's certificate won't be validated. HTTPS will be less secure")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".synopsysctl" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".synopsysctl")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
