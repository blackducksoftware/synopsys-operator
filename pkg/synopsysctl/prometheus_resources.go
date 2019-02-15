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

package synopsysctl

import (
	"fmt"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	horizoncomponents "github.com/blackducksoftware/horizon/pkg/components"
)

func (specConfig *PrometheusSpecConfig) GetPrometheusService() *horizoncomponents.Service {

	// Add Service for Prometheus
	prometheusService := horizoncomponents.NewService(horizonapi.ServiceConfig{
		APIVersion: "v1",
		//ClusterName:              "string",
		Name:      "prometheus",
		Namespace: specConfig.Namespace,
		//ExternalName:             "string",
		IPServiceType: horizonapi.ClusterIPServiceTypeNodePort,
		//ClusterIP:                "string",
		//PublishNotReadyAddresses: "bool",
		//TrafficPolicy:            "TrafficPolicyType",
		//Affinity:                 "string",
	})
	prometheusService.AddAnnotations(map[string]string{"prometheus.io/scrape": "true"})
	prometheusService.AddLabels(map[string]string{"name": "prometheus"})
	prometheusService.AddSelectors(map[string]string{"app": "prometheus"})
	prometheusService.AddPort(horizonapi.ServicePortConfig{
		Name:       "prometheus",
		Port:       9090,
		TargetPort: "9090",
		//NodePort:   "int32",
		Protocol: horizonapi.ProtocolTCP,
	})

	return prometheusService
}

func (specConfig *PrometheusSpecConfig) GetPrometheusDeployment() *horizoncomponents.Deployment {
	// Deployment
	var prometheusDeploymentReplicas int32 = 1
	prometheusDeployment := horizoncomponents.NewDeployment(horizonapi.DeploymentConfig{
		APIVersion: "extensions/v1beta1",
		//ClusterName:             "string",
		Name:      "prometheus",
		Namespace: specConfig.Namespace,
		Replicas:  &prometheusDeploymentReplicas,
		//Recreate:                "bool",
		//MaxUnavailable:          "string",
		//MaxExtra:                "string",
		//MinReadySeconds:         "int32",
		//RevisionHistoryLimit:    "*int32",
		//Paused:                  "bool",
		//ProgressDeadlineSeconds: "*int32",
	})
	prometheusDeployment.AddMatchLabelsSelectors(map[string]string{"app": "prometheus"})

	prometheusPod := horizoncomponents.NewPod(horizonapi.PodConfig{
		APIVersion: "v1",
		//ClusterName          :  "string",
		Name:      "prometheus",
		Namespace: specConfig.Namespace,
		//ServiceAccount       :  "string",
		//RestartPolicy        :  "RestartPolicyType",
		//TerminationGracePeriod : "*int64",
		//ActiveDeadline       :  "*int64",
		//Node                 :  "string",
		//FSGID                :  "*int64",
		//Hostname             :  "string",
		//SchedulerName        :  "string",
		//DNSPolicy           :   "DNSPolicyType",
		//PriorityValue       :   "*int32",
		//PriorityClass        :  "string",
		//SELinux              :  "*SELinuxType",
		//RunAsUser            :  "*int64",
		//RunAsGroup           :  "*int64",
		//ForceNonRoot         :  "*bool",
	})

	prometheusContainer := horizoncomponents.NewContainer(horizonapi.ContainerConfig{
		Name: "prometheus",
		Args: []string{"--log.level=debug", "--config.file=/etc/prometheus/prometheus.yml", "--storage.tsdb.path=/tmp/data/"},
		//Command:                  "[]string",
		Image: specConfig.PrometheusImage,
		//PullPolicy:               "PullPolicyType",
		//MinCPU:                   "string",
		//MaxCPU:                   "string",
		//MinMem:                   "string",
		//MaxMem:                   "string",
		//Privileged:               "*bool",
		//AllowPrivilegeEscalation: "*bool",
		//ReadOnlyFS:               "*bool",
		//ForceNonRoot:             "*bool",
		//SELinux:                  "*SELinuxType",
		//UID:                      "*int64",
		//AllocateStdin:            "bool",
		//StdinOnce:                "bool",
		//AllocateTTY:              "bool",
		//WorkingDirectory:         "string",
		//TerminationMsgPath:       "string",
		//TerminationMsgPolicy:     "TerminationMessagePolicyType",
	})

	prometheusContainer.AddPort(horizonapi.PortConfig{
		Name: "web",
		//Protocol:      "ProtocolType",
		//IP:            "string",
		//HostPort:      "string",
		ContainerPort: "9090",
	})

	prometheusContainer.AddVolumeMount(horizonapi.VolumeMountConfig{
		MountPath: "/data",
		//Propagation: "*MountPropagationType",
		Name: "data",
		//SubPath:     "string",
		//ReadOnly:    "*bool",
	})
	prometheusContainer.AddVolumeMount(horizonapi.VolumeMountConfig{
		MountPath: "/etc/prometheus",
		//Propagation: "*MountPropagationType",
		Name: "config-volume",
		//SubPath:     "string",
		//ReadOnly:    "*bool",
	})

	prometheusEmptyDirVolume, err := horizoncomponents.NewEmptyDirVolume(horizonapi.EmptyDirVolumeConfig{
		VolumeName: "data",
		//Medium:     "StorageMediumType",
		//SizeLimit:  "string",
	})
	if err != nil {
		fmt.Printf("Error creating EmptyDirVolume for Prometheus: %s", err)
		return nil
	}
	prometheusConfigMapVolume := horizoncomponents.NewConfigMapVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      "config-volume",
		MapOrSecretName: "prometheus",
		//Items:           "map[string]KeyAndMode",
		//DefaultMode:     "*int32",
		//Required:        "*bool",
	})

	prometheusPod.AddContainer(prometheusContainer)
	prometheusPod.AddVolume(prometheusEmptyDirVolume)
	prometheusPod.AddVolume(prometheusConfigMapVolume)
	prometheusDeployment.AddPod(prometheusPod)

	return prometheusDeployment
}

func (specConfig *PrometheusSpecConfig) GetPrometheusConfigMap() *horizoncomponents.ConfigMap {
	// Add prometheus config map
	prometheusConfigMap := horizoncomponents.NewConfigMap(horizonapi.ConfigMapConfig{
		APIVersion: "v1",
		//ClusterName: "string",
		Name:      "prometheus",
		Namespace: specConfig.Namespace,
	})
	prometheusConfigMap.AddData(map[string]string{"prometheus.yml": "{'global':{'scrape_interval':'5s'},'scrape_configs':[{'job_name':'synopsys-operator-scrape','scrape_interval':'5s','static_configs':[{'targets':['synopsys-operator:8080', 'synopsys-operator-ui:3000']}]}]}"})

	return prometheusConfigMap
}
