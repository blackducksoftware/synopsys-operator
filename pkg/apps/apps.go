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
	"reflect"

	"github.com/blackducksoftware/horizon/pkg/components"
)

// AppInstallerInterface defines the interface for apps
type AppInstallerInterface interface {
	Configure(interface{}) error
	GetComponents() (*ComponentList, error)
	GetNamespace() string
}

// ComponentList defines the list of components for an app
type ComponentList struct {
	ReplicationControllers []*components.ReplicationController
	Services               []*components.Service
	ConfigMaps             []*components.ConfigMap
	ServiceAccounts        []*components.ServiceAccount
	ClusterRoleBindings    []*components.ClusterRoleBinding
	ClusterRoles           []*components.ClusterRole
	Deployments            []*components.Deployment
}

// AppType defines the type of application
type AppType int

// Types of apps
const (
	PerceptorApp AppType = iota
)

// MergeConfig will merge 2 configuation structs of the same type.
// Any entries in new will replace existing entries in config
func MergeConfig(new interface{}, config interface{}) error {
	if config == nil {
		config = new
		return nil
	}

	if new == nil {
		return nil
	}

	newConfig := reflect.ValueOf(new).Elem()
	existingConfig := reflect.ValueOf(config).Elem()
	for cnt := 0; cnt < newConfig.NumField(); cnt++ {
		fieldName := existingConfig.Type().Field(cnt).Name
		field := existingConfig.Field(cnt)
		newField := newConfig.Field(cnt)
		newValue := newConfig.FieldByName(fieldName)
		if newValue.IsValid() {
			kind := newConfig.Type().Field(cnt).Type.Kind()
			switch kind {
			case reflect.String,
				reflect.Slice,
				reflect.Array,
				reflect.Map:
				if newField.Len() != 0 {
					field.Set(newValue)
				}
			case reflect.Ptr:
				if !newField.IsNil() {
					field.Set(newValue)
				}
			default:
				return fmt.Errorf("unknown field type: %v", kind)
			}
		}
	}

	return nil
}
