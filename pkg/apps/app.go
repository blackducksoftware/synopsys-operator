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

package apps

import (
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"k8s.io/client-go/rest"
)

// App struct
type App struct {
	config     *protoform.Config
	kubeConfig *rest.Config
}

// NewApp will return an App
func NewApp(config *protoform.Config, kubeConfig *rest.Config) *App {
	return &App{config: config, kubeConfig: kubeConfig}
}

// Blackduck will return a Blackduck
func (a *App) Blackduck() *blackduck.Blackduck {
	return blackduck.NewBlackduck(a.config, a.kubeConfig)
}
