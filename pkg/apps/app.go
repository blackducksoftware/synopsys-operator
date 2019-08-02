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
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/alert"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/opssight"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
)

// App struct
type App struct {
	protoformDeployer *protoform.Deployer
}

// NewApp will return an App
func NewApp(protoformDeployer *protoform.Deployer) *App {
	return &App{protoformDeployer: protoformDeployer}
}

// Alert will return an Alert
func (a *App) Alert() *alert.Alert {
	return alert.NewAlert(a.protoformDeployer)
}

// Blackduck will return a Blackduck
func (a *App) Blackduck() *blackduck.Blackduck {
	return blackduck.NewBlackduck(a.protoformDeployer)
}

// OpsSight will return a OpsSight
func (a *App) OpsSight() *opssight.OpsSight {
	return opssight.NewOpsSight(a.protoformDeployer)
}
