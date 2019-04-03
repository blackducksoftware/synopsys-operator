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

package containers

import (
	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	"k8s.io/apimachinery/pkg/api/resource"
)

// GetPVCs will return the PVCs
func (c *Creater) GetPVCs() []*components.PersistentVolumeClaim {
	var pvcs []*components.PersistentVolumeClaim

	defaultPVC := map[string]string{
		"blackduck-postgres":          "150Gi",
		"blackduck-authentication":    "2Gi",
		"blackduck-cfssl":             "2Gi",
		"blackduck-registration":      "2Gi",
		"blackduck-solr":              "2Gi",
		"blackduck-webapp":            "2Gi",
		"blackduck-logstash":          "20Gi",
		"blackduck-zookeeper-data":    "2Gi",
		"blackduck-zookeeper-datalog": "2Gi",
		"blackduck-rabbitmq":          "5Gi",
		"blackduck-uploadcache-data":  "100Gi",
		"blackduck-uploadcache-key":   "2Gi",
	}

	if c.hubSpec.PersistentStorage {
		for _, claim := range c.hubSpec.PVC {
			var defaultsize string
			// Set default value if size isn't specified
			if v, ok := defaultPVC[claim.Name]; ok {
				defaultsize = v
			}
			pvcs = append(pvcs, c.createPVC(claim.Name, claim.Size, defaultsize, claim.StorageClass, horizonapi.ReadWriteOnce))
		}
	}

	return pvcs
}

func (c *Creater) createPVC(name string, requestedSize string, defaultSize string, storageclass string, accessMode horizonapi.PVCAccessModeType) *components.PersistentVolumeClaim {
	// Workaround so that storageClass does not get set to "", which prevent Kube from using the default storageClass
	var class *string
	if len(storageclass) > 0 {
		class = &storageclass
	} else if len(c.hubSpec.PVCStorageClass) > 0 {
		class = &c.hubSpec.PVCStorageClass
	} else {
		class = nil
	}

	var size string
	_, err := resource.ParseQuantity(requestedSize)
	if err != nil {
		size = defaultSize
	} else {
		size = requestedSize
	}

	pvc, _ := components.NewPersistentVolumeClaim(horizonapi.PVCConfig{
		Name:      name,
		Namespace: c.hubSpec.Namespace,
		Size:      size,
		Class:     class,
	})

	pvc.AddAccessMode(accessMode)

	return pvc
}
