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

package v2

import (
	"fmt"

	"github.com/blackducksoftware/horizon/pkg/components"
	"k8s.io/client-go/kubernetes"

	blackduckapi "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/store"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/types"
	commoncomponents "github.com/blackducksoftware/synopsys-operator/pkg/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
)

// BlackDuckPVCV2 describes the V2 deployments of PVCs for Black Duck
type BlackDuckPVCV2 struct {
	commoncomponents.PVC
	blackDuck *blackduckapi.Blackduck
}

// NewPvc returns the Black Duck PVCV2 configuration
func NewPvc(config *protoform.Config, kubeClient *kubernetes.Clientset, cr interface{}) (types.PVCInterface, error) {
	blackDuck, ok := cr.(*blackduckapi.Blackduck)
	if !ok {
		return nil, fmt.Errorf("unable to cast the interface to Black Duck object")
	}
	pvc := commoncomponents.NewPVC(config, kubeClient)
	return &BlackDuckPVCV2{pvc, blackDuck}, nil
}

// GetPVCs returns the PVC
func (b BlackDuckPVCV2) GetPVCs() ([]*components.PersistentVolumeClaim, error) {
	spec := b.blackDuck.Spec

	defaultPVC := map[string]string{
		"blackduck-authentication":   "2Gi",
		"blackduck-cfssl":            "2Gi",
		"blackduck-registration":     "2Gi",
		"blackduck-webapp":           "2Gi",
		"blackduck-logstash":         "20Gi",
		"blackduck-zookeeper":        "4Gi",
		"blackduck-uploadcache-data": "100Gi",
	}

	if b.blackDuck.Spec.ExternalPostgres == nil {

		defaultPVC["blackduck-postgres"] = "150Gi"
	}

	return b.Generate(defaultPVC, spec.PVC, spec.PVCStorageClass, b.blackDuck.Name, spec.Namespace, b.blackDuck.Annotations["synopsys.com/created.by"])
}

func init() {
	store.Register(types.BlackDuckPVCV2, NewPvc)
}
