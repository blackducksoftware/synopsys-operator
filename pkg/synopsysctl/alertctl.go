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
	"fmt"

	alertv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/alert/v1"
	crddefaults "github.com/blackducksoftware/synopsys-operator/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type AlertCtl struct {
	Spec                   *alertv1.AlertSpec
	AlertRegistry          string
	AlertImagePath         string
	AlertAlertImageName    string
	AlertAlertImageVersion string
	AlertCfsslImageName    string
	AlertCfsslImageVersion string
	AlertBlackduckHost     string
	AlertBlackduckUser     string
	AlertBlackduckPort     int
	AlertPort              int
	AlertStandAlone        bool
	AlertAlertMemory       string
	AlertCfsslMemory       string
	AlertState             string
}

func NewAlertCtl() *AlertCtl {
	return &AlertCtl{
		Spec:                   &alertv1.AlertSpec{},
		AlertRegistry:          "",
		AlertImagePath:         "",
		AlertAlertImageName:    "",
		AlertAlertImageVersion: "",
		AlertCfsslImageName:    "",
		AlertCfsslImageVersion: "",
		AlertBlackduckHost:     "",
		AlertBlackduckUser:     "",
		AlertBlackduckPort:     0,
		AlertPort:              0,
		AlertStandAlone:        false,
		AlertAlertMemory:       "",
		AlertCfsslMemory:       "",
		AlertState:             "",
	}
}

func (ctl *AlertCtl) SetDefault(createAlertSpecType string) error {
	switch createAlertSpecType {
	case "empty":
		ctl.Spec = &alertv1.AlertSpec{}
	case "spec1":
		ctl.Spec = crddefaults.GetAlertDefaultValue()
	case "spec2":
		ctl.Spec = crddefaults.GetAlertDefaultValue2()
	default:
		return fmt.Errorf("Alert Spec Type %s does not match: empty, spec1, spec2", createAlertSpecType)
	}
	return nil
}

// Create Alert Spec Flags
func (ctl *AlertCtl) AddAlertSpecFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&ctl.AlertRegistry, "alert-registry", ctl.AlertRegistry, "TODO")
	cmd.Flags().StringVar(&ctl.AlertImagePath, "image-path", ctl.AlertImagePath, "TODO")
	cmd.Flags().StringVar(&ctl.AlertAlertImageName, "alert-image-name", ctl.AlertAlertImageName, "TODO")
	cmd.Flags().StringVar(&ctl.AlertAlertImageVersion, "alert-image-version", ctl.AlertAlertImageVersion, "TODO")
	cmd.Flags().StringVar(&ctl.AlertCfsslImageName, "cfssl-image-name", ctl.AlertCfsslImageName, "TODO")
	cmd.Flags().StringVar(&ctl.AlertCfsslImageVersion, "cfssl-image-version", ctl.AlertCfsslImageVersion, "TODO")
	cmd.Flags().StringVar(&ctl.AlertBlackduckHost, "blackduck-host", ctl.AlertBlackduckHost, "TODO")
	cmd.Flags().StringVar(&ctl.AlertBlackduckUser, "blackduck-user", ctl.AlertBlackduckUser, "TODO")
	cmd.Flags().IntVar(&ctl.AlertBlackduckPort, "blackduck-port", ctl.AlertBlackduckPort, "TODO")
	cmd.Flags().IntVar(&ctl.AlertPort, "port", ctl.AlertPort, "TODO")
	cmd.Flags().BoolVar(&ctl.AlertStandAlone, "stand-alone", ctl.AlertStandAlone, "TODO")
	cmd.Flags().StringVar(&ctl.AlertAlertMemory, "alert-memory", ctl.AlertAlertMemory, "TODO")
	cmd.Flags().StringVar(&ctl.AlertCfsslMemory, "cfssl-memory", ctl.AlertCfsslMemory, "TODO")
	cmd.Flags().StringVar(&ctl.AlertState, "alert-state", ctl.AlertState, "TODO")
}

func (ctl *AlertCtl) SetAlertFlags(f *pflag.Flag) {
	if f.Changed {
		log.Debugf("Flag %s: CHANGED\n", f.Name)
		switch f.Name {
		case "alert-registry":
			ctl.Spec.Registry = ctl.AlertRegistry
		case "image-path":
			ctl.Spec.ImagePath = ctl.AlertImagePath
		case "alert-image-name":
			ctl.Spec.AlertImageName = ctl.AlertAlertImageName
		case "alert-image-version":
			ctl.Spec.AlertImageVersion = ctl.AlertAlertImageVersion
		case "cfssl-image-name":
			ctl.Spec.CfsslImageName = ctl.AlertCfsslImageName
		case "cfssl-image-version":
			ctl.Spec.CfsslImageVersion = ctl.AlertCfsslImageVersion
		case "blackduck-host":
			ctl.Spec.BlackduckHost = ctl.AlertBlackduckHost
		case "blackduck-user":
			ctl.Spec.BlackduckUser = ctl.AlertBlackduckUser
		case "blackduck-port":
			fmt.Printf("Shouldn't be here\n")
			ctl.Spec.BlackduckPort = &ctl.AlertBlackduckPort
		case "port":
			fmt.Printf("Flag Value: %s\n", f.Value)
			fmt.Printf("AddFlg AlertPort: %+v\n", ctl.AlertPort)
			fmt.Printf("AddFlg &AlertPort: %+v\n", &ctl.AlertPort)
			ctl.Spec.Port = &ctl.AlertPort
			fmt.Printf("AddFlg AlertPort: %+v\n", ctl.AlertPort)
			fmt.Printf("AddFlg &AlertPort: %+v\n", &ctl.AlertPort)
		case "stand-alone":
			ctl.Spec.StandAlone = &ctl.AlertStandAlone
		case "alert-memory":
			ctl.Spec.AlertMemory = ctl.AlertAlertMemory
		case "cfssl-memory":
			ctl.Spec.CfsslMemory = ctl.AlertCfsslMemory
		default:
			log.Debugf("Flag %s: Not Found\n", f.Name)
		}
	} else {
		log.Debugf("Flag %s: UNCHANGED\n", f.Name)
	}
}
