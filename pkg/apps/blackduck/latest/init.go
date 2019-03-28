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

package blackduck

import (
	"fmt"
	"strings"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	horizon "github.com/blackducksoftware/horizon/pkg/deployer"
	"github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	containers "github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/latest/containers"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	log "github.com/sirupsen/logrus"
)

func (hc *Creater) init(deployer *horizon.Deployer, createHub *v1.BlackduckSpec, hubContainerFlavor *containers.ContainerFlavor) error {
	// Create a namespaces
	_, err := util.GetNamespace(hc.KubeClient, createHub.Namespace)
	if err != nil {
		log.Debugf("unable to find the namespace %s", createHub.Namespace)
		deployer.AddNamespace(components.NewNamespace(horizonapi.NamespaceConfig{Name: createHub.Namespace}))
	}

	// Create a service account
	deployer.AddServiceAccount(util.CreateServiceAccount(createHub.Namespace, createHub.Namespace))

	// Create a cluster role binding and associated it to a service account
	deployer.AddClusterRoleBinding(util.CreateClusterRoleBinding(createHub.Namespace, createHub.Namespace, createHub.Namespace, "", "ClusterRole", "synopsys-operator-admin"))

	// We only start the postgres container if the external DB configuration struct is empty
	if createHub.PersistentStorage {
		for _, claim := range createHub.PVC {
			storageClass := createHub.PVCStorageClass
			if len(claim.StorageClass) > 0 {
				storageClass = claim.StorageClass
			}

			if !hc.isBinaryAnalysisEnabled && (strings.Contains(claim.Name, "blackduck-rabbitmq") || strings.Contains(claim.Name, "blackduck-uploadcache")) {
				continue
			}

			var size string

			// Set default value if size isn't specified
			// TODO JD - check the if the size is using a support format Gi, etc
			switch claim.Name {
			case "blackduck-postgres":
				size = "150Gi"
				if len(claim.Size) > 0 {
					size = claim.Size
				}
			case "blackduck-authentication":
				size = "2Gi"
				if len(claim.Size) > 0 {
					size = claim.Size
				}
			case "blackduck-cfssl":
				size = "2Gi"
				if len(claim.Size) > 0 {
					size = claim.Size
				}
			case "blackduck-registration":
				size = "2Gi"
				if len(claim.Size) > 0 {
					size = claim.Size
				}
			case "blackduck-solr":
				size = "2Gi"
				if len(claim.Size) > 0 {
					size = claim.Size
				}
			case "blackduck-webapp":
				size = "2Gi"
				if len(claim.Size) > 0 {
					size = claim.Size
				}
			case "blackduck-logstash":
				size = "20Gi"
				if len(claim.Size) > 0 {
					size = claim.Size
				}
			case "blackduck-zookeeper-data":
				size = "2Gi"
				if len(claim.Size) > 0 {
					size = claim.Size
				}
			case "blackduck-zookeeper-datalog":
				size = "2Gi"
				if len(claim.Size) > 0 {
					size = claim.Size
				}
			case "blackduck-rabbitmq":
				size = "5Gi"
				if len(claim.Size) > 0 {
					size = claim.Size
				}
			case "blackduck-uploadcache-data":
				size = "100Gi"
				if len(claim.Size) > 0 {
					size = claim.Size
				}
			case "blackduck-uploadcache-key":
				size = "2Gi"
				if len(claim.Size) > 0 {
					size = claim.Size
				}
			default:
				size = claim.Size
			}

			pvc, err := util.CreatePersistentVolumeClaim(claim.Name, createHub.Namespace, size, storageClass, horizonapi.ReadWriteOnce)
			if err != nil {
				return fmt.Errorf("failed to create the postgres PVC %s in namespace %s because %+v", claim.Name, createHub.Namespace, err)
			}
			deployer.AddPVC(pvc)

		}
	}

	return nil
}
