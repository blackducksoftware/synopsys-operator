// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
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

package synopsysctl

import (
	alertv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/alert/v1"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Gloabal Specs
var globalAlertSpec = &alertv1.AlertSpec{}

// Create Alert Spec Flags
var alertRegistry = ""
var alertImagePath = ""
var alertAlertImageName = ""
var alertAlertImageVersion = ""
var alertCfsslImageName = ""
var alertCfsslImageVersion = ""
var alertBlackduckHost = ""
var alertBlackduckUser = ""
var alertBlackduckPort = 0
var alertPort = 0
var alertStandAlone = false
var alertAlertMemory = ""
var alertCfsslMemory = ""
var alertState = ""

func addAlertSpecFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&alertRegistry, "alert-registry", alertRegistry, "TODO")
	cmd.Flags().StringVar(&alertImagePath, "image-path", alertImagePath, "TODO")
	cmd.Flags().StringVar(&alertAlertImageName, "alert-image-name", alertAlertImageName, "TODO")
	cmd.Flags().StringVar(&alertAlertImageVersion, "alert-image-version", alertAlertImageVersion, "TODO")
	cmd.Flags().StringVar(&alertCfsslImageName, "cfssl-image-name", alertCfsslImageName, "TODO")
	cmd.Flags().StringVar(&alertCfsslImageVersion, "cfssl-image-version", alertCfsslImageVersion, "TODO")
	cmd.Flags().StringVar(&alertBlackduckHost, "blackduck-host", alertBlackduckHost, "TODO")
	cmd.Flags().StringVar(&alertBlackduckUser, "blackduck-user", alertBlackduckUser, "TODO")
	cmd.Flags().IntVar(&alertBlackduckPort, "blackduck-port", alertBlackduckPort, "TODO")
	cmd.Flags().IntVar(&alertPort, "port", alertPort, "TODO")
	cmd.Flags().BoolVar(&alertStandAlone, "stand-alone", alertStandAlone, "TODO")
	cmd.Flags().StringVar(&alertAlertMemory, "alert-memory", alertAlertMemory, "TODO")
	cmd.Flags().StringVar(&alertCfsslMemory, "cfssl-memory", alertCfsslMemory, "TODO")
	cmd.Flags().StringVar(&alertState, "alert-state", alertState, "TODO")
}

func setAlertFlags(f *pflag.Flag) {
	if f.Changed {
		log.Debugf("Flag %s: CHANGED\n", f.Name)
		switch f.Name {
		case "alert-registry":
			globalAlertSpec.Registry = alertRegistry
		case "image-path":
			globalAlertSpec.ImagePath = alertImagePath
		case "alert-image-name":
			globalAlertSpec.AlertImageName = alertAlertImageName
		case "alert-image-version":
			globalAlertSpec.AlertImageVersion = alertAlertImageVersion
		case "cfssl-image-name":
			globalAlertSpec.CfsslImageName = alertCfsslImageName
		case "cfssl-image-version":
			globalAlertSpec.CfsslImageVersion = alertCfsslImageVersion
		case "blackduck-host":
			globalAlertSpec.BlackduckHost = alertBlackduckHost
		case "blackduck-user":
			globalAlertSpec.BlackduckUser = alertBlackduckUser
		case "blackduck-port":
			globalAlertSpec.BlackduckPort = &alertBlackduckPort
		case "port":
			globalAlertSpec.Port = &alertPort
		case "stand-alone":
			globalAlertSpec.StandAlone = &alertStandAlone
		case "alert-memory":
			globalAlertSpec.AlertMemory = alertAlertMemory
		case "cfssl-memory":
			globalAlertSpec.CfsslMemory = alertCfsslMemory
		default:
			log.Debugf("Flag %s: Not Found\n", f.Name)
		}
	}
	log.Debugf("Flag %s: UNCHANGED\n", f.Name)
}
