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
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Root Command Options and Defaults
var cfgFile string
var kubeConfigPath = ""
var insecureSkipTLSVerify = false
var logLevelCtl = "info"

// synopsysctlVersion is the current version of the synopsysctl utility
var synopsysctlVersion string

// rootCmd is the base command of synopsyctl that all other commands are added to
var rootCmd = &cobra.Command{
	Use:   "synopsysctl",
	Short: "synopsysctl is a command line tool for managing Synopsys resources",
	Args: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	// This function is run before every subcommand
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if err := setSynopsysctlLogLevel(); err != nil {
			return err
		}

		// Determine if synopsysctl is running in mock mode
		mockMode := false
		mockModeFlagExists := cmd.Flags().Lookup("mock")
		if mockModeFlagExists != nil && mockModeFlagExists.Changed {
			mockMode = true
		}

		// Determine if synopsysctl is running in native command
		nativeMode := strings.Contains(cmd.CommandPath(), "native")

		// Determine if synopsysctl is 'updating' a resource
		updatingResource := strings.Contains(cmd.CommandPath(), "update")

		// Don't set cluster resources if we are in mock mode or native mode (aka the command doesn't need access the cluster)
		// This allows users to use native/mock when not connected to a cluster
		// Note: If you are updating a resource you can run mock mode and still need access to the cluster
		if updatingResource || (!mockMode && !nativeMode) {
			if err := setGlobalKubeConfigPath(cmd); err != nil {
				log.Error(err)
				os.Exit(1)
			}
			if err := setGlobalRestConfig(); err != nil {
				log.Error(err)
				os.Exit(1)
			}
			if err := setGlobalKubeClient(); err != nil {
				log.Error(err)
				os.Exit(1)
			}
			if err := setGlobalResourceClients(); err != nil {
				log.Error(err)
				os.Exit(1)
			}
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
		log.Errorf("synopsyctl failed: %+v", err)
		os.Exit(1)
	}
}

func init() {
	//(PassCmd) rootCmd.DisableFlagParsing = true // lets rootCmd pass flags to kube/oc

	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&kubeConfigPath, "kubeconfig", kubeConfigPath, "Path to a kubeconfig file with the context set to a cluster for synopsysctl to access")
	rootCmd.PersistentFlags().BoolVar(&insecureSkipTLSVerify, "insecure-skip-tls-verify", insecureSkipTLSVerify, "Server's certificate won't be validated. HTTPS will be irrelevant")
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
