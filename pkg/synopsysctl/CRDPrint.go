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
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"k8s.io/client-go/rest"
	"sigs.k8s.io/yaml"

	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/api"
	alertapi "github.com/blackducksoftware/synopsys-operator/pkg/api/alert/v1"
	blackduckapi "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	opssightapi "github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps"
	alertapp "github.com/blackducksoftware/synopsys-operator/pkg/apps/alert"
	blackduckapp "github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"github.com/blackducksoftware/synopsys-operator/pkg/soperator"
)

// PrintFormat represents the format to print the struct
type PrintFormat string

// Constants for the PrintFormats
const (
	JSON PrintFormat = "JSON"
	YAML PrintFormat = "YAML"
)

func getDefaultApp(cType string) (*apps.App, error) {
	pc := &protoform.Config{}
	pc.SelfSetDefaults()
	pc.DryRun = true
	err := verifyClusterType(cType)
	if err != nil {
		return nil, err
	}
	if strings.EqualFold(strings.ToUpper(cType), clusterTypeOpenshift) {
		pc.IsOpenshift = true
	}
	rc := &rest.Config{}

	protoformDeployer, _ := protoform.NewDeployer(pc, rc, nil)
	return apps.NewApp(protoformDeployer), nil
}

// PrintResource prints a Resource as yaml or json. printKubeComponents allows printing the kuberentes
// resources instead
func PrintResource(crd interface{}, format string, printKubeComponents bool) error {
	// print the CRD
	if !printKubeComponents {
		return PrintComponents([]interface{}{crd}, format)
	}

	app, err := getDefaultApp(nativeClusterType)
	if err != nil {
		return err
	}

	var cList *api.ComponentList

	switch reflect.TypeOf(crd) {
	case reflect.TypeOf(soperator.SpecConfig{}):
		operator := crd.(soperator.SpecConfig)
		cList, err = operator.GetComponents()
		if err != nil {
			return fmt.Errorf("failed to get components: %s", err)
		}
	case reflect.TypeOf(alertapi.Alert{}):
		alert := crd.(alertapi.Alert)
		cList, err = app.Alert().GetComponents(&alert, alertapp.CRDResources)
		if err != nil {
			return fmt.Errorf("failed to get components: %s", err)
		}
	case reflect.TypeOf(blackduckapi.Blackduck{}):
		blackDuck := crd.(blackduckapi.Blackduck)
		cList, err = app.Blackduck().GetComponents(&blackDuck, blackduckapp.CRDResources)
		if err != nil {
			return fmt.Errorf("failed to get components: %s", err)
		}
	case reflect.TypeOf(opssightapi.OpsSight{}):
		opsSight := crd.(opssightapi.OpsSight)
		cList, err = app.OpsSight().GetComponents(&opsSight)
		if err != nil {
			return fmt.Errorf("failed to get components: %s", err)
		}
	default:
		return fmt.Errorf("cannot print a resource with the format: %+v", crd)
	}

	if cList == nil {
		return fmt.Errorf("failed to generate a componentLists for %+v", crd)
	}
	cList.PersistentVolumeClaims = []*components.PersistentVolumeClaim{} // Don't print resources for PVCs
	return PrintComponentListKube(cList, format)
}

// PrintComponentListKube does
func PrintComponentListKube(cList *api.ComponentList, format string) error {
	kubeInterfaces := cList.GetKubeInterfaces()
	return PrintComponents(kubeInterfaces, format)
}

// PrintComponents outputs components for a CRD in the correct format for 'kubectl create -f <file>'
func PrintComponents(objs []interface{}, format string) error {
	for i, obj := range objs {
		_, err := PrintComponent(obj, format)
		if err != nil {
			return fmt.Errorf("failed to print components: %s", err)
		}
		if i != len(objs)-1 && format == "yaml" {
			fmt.Printf("---\n")
		}
	}
	return nil
}

// PrintComponent will print the interface in either json or yaml format
func PrintComponent(v interface{}, format string) (string, error) {
	var b []byte
	var err error
	switch {
	case strings.ToUpper(format) == string(JSON):
		b, err = json.MarshalIndent(v, "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to convert struct to json. err: %+v. struct: %+v", err, v)
		}
		fmt.Println(string(b))
	case strings.ToUpper(format) == string(YAML):
		b, err = yaml.Marshal(v)
		if err != nil {
			return "", fmt.Errorf("failed to convert struct to yaml. err: %+v. struct: %+v", err, v)
		}
		fmt.Println(string(b))
	default:
		return "", fmt.Errorf("'%s' is an invalid format", format)
	}
	return string(b), nil
}
