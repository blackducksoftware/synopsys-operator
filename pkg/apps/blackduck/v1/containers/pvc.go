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

package containers

import (
	"github.com/blackducksoftware/horizon/pkg/components"
	blackduckutil "github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/util"
)

// GetPVCs will return the PVCs
func (c *Creater) GetPVCs() ([]*components.PersistentVolumeClaim, error) {
	defaultPVC := map[string]string{
		"blackduck-authentication": "2Gi",
		"blackduck-cfssl":          "2Gi",
		"blackduck-registration":   "2Gi",
		"blackduck-webapp":         "2Gi",
		"blackduck-logstash":       "20Gi",
		"blackduck-zookeeper":      "4Gi",
	}

	if c.blackDuck.Spec.ExternalPostgres == nil {
		defaultPVC["blackduck-postgres"] = "150Gi"
	}

	if c.isBinaryAnalysisEnabled {
		defaultPVC["blackduck-uploadcache-data"] = "100Gi"
	}

	return blackduckutil.GenPVC(defaultPVC, c.blackDuck)
}
