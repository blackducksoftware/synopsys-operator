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

package alert

import (
	"github.com/blackducksoftware/synopsys-operator/pkg/api"
	alertapi "github.com/blackducksoftware/synopsys-operator/pkg/api/alert/v1"
)

// Creater interface for Alert
// An Alert can have multiple Creaters where each Creater supports
// different versions of Alert
type Creater interface {
	Versions() []string
	Ensure(alert *alertapi.Alert) error
	GetComponents(alert *alertapi.Alert) (*api.ComponentList, error)
}
