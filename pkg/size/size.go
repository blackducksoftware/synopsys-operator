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

package size

import (
	"strings"

	sizev1 "github.com/blackducksoftware/synopsys-operator/pkg/api/size/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/types"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetDefaultSize returns the default size. This will be used ny synopsysctl to create the Size custom resources during the deployment
func GetDefaultSize(name string) *sizev1.Size {
	switch strings.ToUpper(name) {
	case "SMALL":
		return &sizev1.Size{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Spec: sizev1.SizeSpec{
				PodResources: map[string]sizev1.PodResource{
					"authentication": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types.AuthenticationContainerName): {
								MinMem: util.IntToInt32(1024),
								MaxMem: util.IntToInt32(1024),
							},
						},
					},
					"binaryscanner": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types.BinaryScannerContainerName): {
								MinCPU: util.IntToInt32(1),
								MaxCPU: util.IntToInt32(1),
								MinMem: util.IntToInt32(2048),
								MaxMem: util.IntToInt32(2048),
							},
						},
					},
					"cfssl": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types.CfsslContainerName): {
								MinMem: util.IntToInt32(640),
								MaxMem: util.IntToInt32(640),
							},
						},
					},
					"documentation": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types.DocumentationContainerName): {
								MinMem: util.IntToInt32(512),
								MaxMem: util.IntToInt32(512),
							},
						},
					},
					"jobrunner": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types.JobrunnerContainerName): {
								MinCPU: util.IntToInt32(1),
								MaxCPU: util.IntToInt32(1),
								MinMem: util.IntToInt32(4608),
								MaxMem: util.IntToInt32(4608),
							},
						},
					},
					"rabbitmq": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types.RabbitMQContainerName): {
								MinMem: util.IntToInt32(1024),
								MaxMem: util.IntToInt32(1024),
							},
						},
					},
					"registration": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types.RegistrationContainerName): {
								MinMem: util.IntToInt32(1024),
								MaxMem: util.IntToInt32(1024),
							},
						},
					},
					"scan": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types.ScanContainerName): {
								MinMem: util.IntToInt32(2560),
								MaxMem: util.IntToInt32(2560),
							},
						},
					},
					"solr": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types.SolrContainerName): {
								MinMem: util.IntToInt32(640),
								MaxMem: util.IntToInt32(640),
							},
						},
					},
					"uploadcache": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types.UploadCacheContainerName): {
								MinMem: util.IntToInt32(512),
								MaxMem: util.IntToInt32(512),
							},
						},
					},
					"webapp-logstash": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types.WebappContainerName): {
								MinCPU: util.IntToInt32(1),
								MaxCPU: util.IntToInt32(1),
								MinMem: util.IntToInt32(2560),
								MaxMem: util.IntToInt32(2560),
							},
							string(types.LogstashContainerName): {
								MinMem: util.IntToInt32(1024),
								MaxMem: util.IntToInt32(1024),
							},
						},
					},
					"webserver": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types.WebserverContainerName): {
								MinMem: util.IntToInt32(512),
								MaxMem: util.IntToInt32(512),
							},
						},
					},
					"zookeeper": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types.ZookeeperContainerName): {
								MinMem: util.IntToInt32(640),
								MaxMem: util.IntToInt32(640),
							},
						},
					},
					"postgres": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types.PostgresContainerName): {
								MinCPU: util.IntToInt32(1),
								MaxCPU: util.IntToInt32(1),
								MinMem: util.IntToInt32(3072),
								MaxMem: util.IntToInt32(3072),
							},
						},
					},
				}},
		}
	case "MEDIUM":
		return &sizev1.Size{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Spec: sizev1.SizeSpec{
				PodResources: map[string]sizev1.PodResource{
					"authentication": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types.AuthenticationContainerName): {
								MinMem: util.IntToInt32(1024),
								MaxMem: util.IntToInt32(1024),
							},
						},
					},
					"binaryscanner": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types.BinaryScannerContainerName): {
								MinCPU: util.IntToInt32(1),
								MaxCPU: util.IntToInt32(1),
								MinMem: util.IntToInt32(2048),
								MaxMem: util.IntToInt32(2048),
							},
						},
					},
					"cfssl": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types.CfsslContainerName): {
								MinMem: util.IntToInt32(640),
								MaxMem: util.IntToInt32(640),
							},
						},
					},
					"documentation": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types.DocumentationContainerName): {
								MinMem: util.IntToInt32(512),
								MaxMem: util.IntToInt32(512),
							},
						},
					},
					"jobrunner": {
						Replica: 4,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types.JobrunnerContainerName): {
								MinCPU: util.IntToInt32(4),
								MaxCPU: util.IntToInt32(4),
								MinMem: util.IntToInt32(7168),
								MaxMem: util.IntToInt32(7168),
							},
						},
					},
					"rabbitmq": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types.RabbitMQContainerName): {
								MinMem: util.IntToInt32(1024),
								MaxMem: util.IntToInt32(1024),
							},
						},
					},
					"registration": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types.RegistrationContainerName): {
								MinMem: util.IntToInt32(1024),
								MaxMem: util.IntToInt32(1024),
							},
						},
					},
					"scan": {
						Replica: 2,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types.ScanContainerName): {
								MinMem: util.IntToInt32(5120),
								MaxMem: util.IntToInt32(5120),
							},
						},
					},
					"solr": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types.SolrContainerName): {
								MinMem: util.IntToInt32(640),
								MaxMem: util.IntToInt32(640),
							},
						},
					},
					"uploadcache": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types.UploadCacheContainerName): {
								MinMem: util.IntToInt32(512),
								MaxMem: util.IntToInt32(512),
							},
						},
					},
					"webapp-logstash": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types.WebappContainerName): {
								MinCPU: util.IntToInt32(2),
								MaxCPU: util.IntToInt32(2),
								MinMem: util.IntToInt32(5120),
								MaxMem: util.IntToInt32(5120),
							},
							string(types.LogstashContainerName): {
								MinMem: util.IntToInt32(1024),
								MaxMem: util.IntToInt32(1024),
							},
						},
					},
					"webserver": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types.WebserverContainerName): {
								MinMem: util.IntToInt32(2048),
								MaxMem: util.IntToInt32(2048),
							},
						},
					},
					"zookeeper": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types.ZookeeperContainerName): {
								MinMem: util.IntToInt32(640),
								MaxMem: util.IntToInt32(640),
							},
						},
					},
					"postgres": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types.PostgresContainerName): {
								MinCPU: util.IntToInt32(2),
								MaxCPU: util.IntToInt32(2),
								MinMem: util.IntToInt32(8192),
								MaxMem: util.IntToInt32(8192),
							},
						},
					},
				},
			},
		}
	case "LARGE":
		return &sizev1.Size{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Spec: sizev1.SizeSpec{
				PodResources: map[string]sizev1.PodResource{
					"authentication": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types.AuthenticationContainerName): {
								MinMem: util.IntToInt32(1024),
								MaxMem: util.IntToInt32(1024),
							},
						},
					},
					"binaryscanner": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types.BinaryScannerContainerName): {
								MinCPU: util.IntToInt32(1),
								MaxCPU: util.IntToInt32(1),
								MinMem: util.IntToInt32(2048),
								MaxMem: util.IntToInt32(2048),
							},
						},
					},
					"cfssl": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types.CfsslContainerName): {
								MinMem: util.IntToInt32(640),
								MaxMem: util.IntToInt32(640),
							},
						},
					},
					"documentation": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types.DocumentationContainerName): {
								MinMem: util.IntToInt32(512),
								MaxMem: util.IntToInt32(512),
							},
						},
					},
					"jobrunner": {
						Replica: 6,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types.JobrunnerContainerName): {
								MinCPU: util.IntToInt32(1),
								MaxCPU: util.IntToInt32(1),
								MinMem: util.IntToInt32(13824),
								MaxMem: util.IntToInt32(13824),
							},
						},
					},
					"rabbitmq": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types.RabbitMQContainerName): {
								MinMem: util.IntToInt32(1024),
								MaxMem: util.IntToInt32(1024),
							},
						},
					},
					"registration": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types.RegistrationContainerName): {
								MinMem: util.IntToInt32(1024),
								MaxMem: util.IntToInt32(1024),
							},
						},
					},
					"scan": {
						Replica: 3,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types.ScanContainerName): {
								MinMem: util.IntToInt32(9728),
								MaxMem: util.IntToInt32(9728),
							},
						},
					},
					"solr": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types.SolrContainerName): {
								MinMem: util.IntToInt32(640),
								MaxMem: util.IntToInt32(640),
							},
						},
					},
					"uploadcache": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types.UploadCacheContainerName): {
								MinMem: util.IntToInt32(512),
								MaxMem: util.IntToInt32(512),
							},
						},
					},
					"webapp-logstash": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types.WebappContainerName): {
								MinCPU: util.IntToInt32(2),
								MaxCPU: util.IntToInt32(2),
								MinMem: util.IntToInt32(9728),
								MaxMem: util.IntToInt32(9728),
							},
							string(types.LogstashContainerName): {
								MinMem: util.IntToInt32(1024),
								MaxMem: util.IntToInt32(1024),
							},
						},
					},
					"webserver": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types.WebserverContainerName): {
								MinMem: util.IntToInt32(2048),
								MaxMem: util.IntToInt32(2048),
							},
						},
					},
					"zookeeper": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types.ZookeeperContainerName): {
								MinMem: util.IntToInt32(640),
								MaxMem: util.IntToInt32(640),
							},
						},
					},
					"postgres": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types.PostgresContainerName): {
								MinCPU: util.IntToInt32(2),
								MaxCPU: util.IntToInt32(2),
								MinMem: util.IntToInt32(12288),
								MaxMem: util.IntToInt32(12288),
							},
						},
					},
				},
			},
		}
	}
	return nil
}
