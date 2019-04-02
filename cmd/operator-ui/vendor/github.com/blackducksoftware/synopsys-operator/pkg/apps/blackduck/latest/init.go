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

package blackduck

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	horizon "github.com/blackducksoftware/horizon/pkg/deployer"

	"github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"

	containers "github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/latest/containers"
)

func (hc *Creater) init(deployer *horizon.Deployer, bdspec *v1.BlackduckSpec, hubContainerFlavor *containers.ContainerFlavor) error {
	// Create a namespaces
	_, err := util.GetNamespace(hc.KubeClient, bdspec.Namespace)
	if err != nil {
		log.Debugf("unable to find the namespace %s", bdspec.Namespace)
		deployer.AddNamespace(components.NewNamespace(horizonapi.NamespaceConfig{Name: bdspec.Namespace}))
	}

	// Create a service account
	deployer.AddServiceAccount(util.CreateServiceAccount(bdspec.Namespace, bdspec.Namespace))

	// Create a cluster role binding and associated it to a service account
	deployer.AddClusterRoleBinding(util.CreateClusterRoleBinding(bdspec.Namespace, bdspec.Namespace, bdspec.Namespace, "", "ClusterRole", "synopsys-operator-admin"))

	// We only start the postgres container if the external DB configuration struct is empty
	if bdspec.PersistentStorage {
		for _, claim := range bdspec.PVC {
			storageClass := bdspec.PVCStorageClass
			if len(claim.StorageClass) > 0 {
				storageClass = claim.StorageClass
			}

			if !hc.isBinaryAnalysisEnabled(bdspec) && (strings.Contains(claim.Name, "blackduck-rabbitmq") || strings.Contains(claim.Name, "blackduck-uploadcache")) {
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

			pvc, err := util.CreatePersistentVolumeClaim(claim.Name, bdspec.Namespace, size, storageClass, horizonapi.ReadWriteOnce)
			if err != nil {
				return fmt.Errorf("failed to create the postgres PVC %s in namespace %s because %+v", claim.Name, bdspec.Namespace, err)
			}
			deployer.AddPVC(pvc)

		}
	}

	return nil
}
