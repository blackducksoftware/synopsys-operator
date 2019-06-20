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

	homedir "github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Root Command Options and Defaults
var cfgFile string
var kubeconfig = ""
var insecureSkipTLSVerify = false
var logLevelCtl = "info"

// synopsysctlVersion is the current version of the synopsysctl utility
var synopsysctlVersion string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "synopsysctl",
	Short: "synopsysctl is a command line tool for managing Synopsys resources",
	Args: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Set the Log Level
		lvl, err := log.ParseLevel(logLevelCtl)
		if err != nil {
			log.Errorf("ctl-log-Level '%s' is not a valid level: %s", logLevelCtl, err)
		}
		log.SetLevel(lvl)
		if !cmd.Flags().Lookup("kubeconfig").Changed { // if kubeconfig wasn't set, check the environ
			if kubeconfigEnvVal, exists := os.LookupEnv("KUBECONFIG"); exists { // set kubeconfig if environ is set
				kubeconfig = kubeconfigEnvVal
			}
		}
		// Sets kubeconfig and initializes resource client libraries
		if err := setResourceClients(); err != nil {
			log.Error(err)
			os.Exit(1)
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("must specify sub-command")
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(version string) {
	rootCmd.Version = version
	synopsysOperatorImage = fmt.Sprintf("docker.io/blackducksoftware/synopsys-operator:%s", version)
	initFlags()
	if err := rootCmd.Execute(); err != nil {
		log.Errorf("unable to execute command due to %+v", err)
		os.Exit(1)
	}
}

func init() {
	//(PassCmd) rootCmd.DisableFlagParsing = true // lets rootCmd pass flags to kube/oc
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&kubeconfig, "kubeconfig", kubeconfig, "Path to the kubeconfig file to use for CLI requests")
	rootCmd.PersistentFlags().BoolVar(&insecureSkipTLSVerify, "insecure-skip-tls-verify", insecureSkipTLSVerify, "Server's certificate won't be validated. HTTPS will be less secure")
	rootCmd.PersistentFlags().StringVarP(&logLevelCtl, "verbose-level", "v", logLevelCtl, "Log level for synopsysctl [trace|debug|info|warn|error|fatal|panic]")
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
			log.Errorf("unable to find the home directory due to %+v", err)
			os.Exit(1)
		}

		// Search config in home directory with name ".synopsysctl" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".synopsysctl")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Infof("using config file '%s'", viper.ConfigFileUsed())
	}
}
