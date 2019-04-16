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

var cfgFile string

// Options flags for all commands
var cluster string
var kubeconfig string
var context string
var insecureSkipTLSVerify = false
var logLevelCtl = "warn"

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "synopsysctl",
	Short: "Command Line Tool for managing Synopsys Resources",
	Args: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		flagset := cmd.Flags()

		if flagset.Changed("cluster") { // changes the cluster that is being pointed to (delete this comment)
			log.Warnf("Flag %s is Not Implemented", "cluster")
		}
		if flagset.Changed("context") { // sets the context (delete this comment)
			log.Warnf("Flag %s is Not Implemented", "context")
		}
		// Set the Log Level
		lvl, err := log.ParseLevel(logLevelCtl)
		if err != nil {
			log.Errorf("ctl-log-Level %s is not a valid level: %s", logLevelCtl, err)
		}
		log.SetLevel(lvl)
		// Sets kubeconfig and initializes resource client libraries
		setResourceClients()
		return nil
	},
	//(PassCmd) PreRunE: func(cmd *cobra.Command, args []string) error {
	//(PassCmd) 	if len(args) == 1 && args[0] == "--help" {
	//(PassCmd) 		return fmt.Errorf("Help Called")
	//(PassCmd) 	}
	//(PassCmd) 	return nil
	//(PassCmd) },
	RunE: func(cmd *cobra.Command, args []string) error {
		//(PassCmd) log.Debugf("Running Non-Synopsysctl Command\n")
		//(PassCmd) out, err := util.RunKubeCmd(restconfig, kube, openshift, args...)
		//(PassCmd) if err != nil {
		//(PassCmd) 	log.Errorf("Error with KubeCmd: %s", out)
		//(PassCmd) 	return nil
		//(PassCmd) }
		//(PassCmd) fmt.Printf("%+v", out)
		//(PassCmd) return nil
		return fmt.Errorf("Not a Valid Command")
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
	//(PassCmd) rootCmd.DisableFlagParsing = true // lets rootCmd pass flags to kube/oc
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cluster, "cluster", cluster, "name of the kubeconfig cluster to use")
	rootCmd.PersistentFlags().StringVar(&kubeconfig, "kubeconfig", kubeconfig, "path to the kubeconfig file to use for CLI requests")
	rootCmd.PersistentFlags().StringVar(&context, "context", context, "name of the kubeconfig context to use")
	rootCmd.PersistentFlags().BoolVar(&insecureSkipTLSVerify, "insecure-skip-tls-verify", insecureSkipTLSVerify, "server's certificate won't be validated. HTTPS will be less secure")
	rootCmd.PersistentFlags().StringVar(&logLevelCtl, "ctl-log-level", logLevelCtl, "Log Level for the Synopsysctl")
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
