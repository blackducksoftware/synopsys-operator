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

package soperator

import (
	"encoding/json"
	"github.com/blackducksoftware/synopsys-operator/meta-builder/utils"
	"github.com/juju/errors"
	appv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"strings"
	//routev1 "github.com/openshift/api/route/v1"
)

// GetPrometheusService creates a Horizon Service component for Prometheus
func (specConfig *PrometheusSpecConfig) GetPrometheusService() []*v1.Service {
	services := []*v1.Service{}
	prometheusService := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "prometheus",
			Namespace: specConfig.Namespace,
			Labels: map[string]string{
				"app":       "synopsys-operator",
				"component": "prometheus",
			},
			Annotations: map[string]string{
				"prometheus.io/scrape": "true",
			},
		},
		Spec: v1.ServiceSpec{
			Selector: map[string]string{
				"app":       "synopsys-operator",
				"component": "prometheus",
			},
			Ports: []v1.ServicePort{
				{
					Name:       "prometheus",
					Protocol:   v1.ProtocolTCP,
					Port:       9090,
					TargetPort: intstr.FromInt(9090),
				},
			},
		},
	}

	services = append(services, prometheusService)

	if strings.EqualFold(specConfig.Expose, utils.NODEPORT) || strings.EqualFold(specConfig.Expose, utils.LOADBALANCER) {

		var exposedServiceType v1.ServiceType
		if strings.EqualFold(specConfig.Expose, utils.NODEPORT) {
			exposedServiceType = v1.ServiceTypeNodePort
		} else {
			exposedServiceType = v1.ServiceTypeLoadBalancer
		}

		prometheusExposedService := &v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "prometheus-exposed",
				Namespace: specConfig.Namespace,
				Labels: map[string]string{
					"app":       "synopsys-operator",
					"component": "prometheus",
				},
				Annotations: map[string]string{
					"prometheus.io/scrape": "true",
				},
			},
			Spec: v1.ServiceSpec{
				Type:     exposedServiceType,
				Selector: map[string]string{},
				Ports: []v1.ServicePort{
					{
						Name:       "prometheus",
						Protocol:   v1.ProtocolTCP,
						Port:       9090,
						TargetPort: intstr.FromInt(9090),
					},
				},
			},
		}
		services = append(services, prometheusExposedService)
	}
	return services
}

// GetPrometheusDeployment creates a Horizon Deployment component for Prometheus
func (specConfig *PrometheusSpecConfig) GetPrometheusDeployment() (*appv1.Deployment, error) {
	// Deployment
	var prometheusDeploymentReplicas int32 = 1

	prometheusDeployment := &appv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "prometheus",
			Namespace: specConfig.Namespace,
			Labels: map[string]string{
				"app":       "synopsys-operator",
				"component": "prometheus",
			},
		},
		Spec: appv1.DeploymentSpec{
			Replicas: &prometheusDeploymentReplicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":       "synopsys-operator",
					"component": "prometheus",
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "prometheus",
					Namespace: specConfig.Namespace,
					Labels: map[string]string{
						"app":       "synopsys-operator",
						"component": "prometheus",
					},
				},
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{
						{
							Name: "data",
							VolumeSource: v1.VolumeSource{
								EmptyDir: &v1.EmptyDirVolumeSource{
									Medium: v1.StorageMediumDefault,
								},
							},
						},
						{
							Name: "config-volume",
							VolumeSource: v1.VolumeSource{
								ConfigMap: &v1.ConfigMapVolumeSource{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "prometheus",
									},
									DefaultMode: utils.IntToInt32(420),
								},
							},
						},
					},
					InitContainers: nil,
					Containers: []v1.Container{
						{
							Name:  "prometheus",
							Image: specConfig.Image,
							Args:  []string{"--log.level=debug", "--config.file=/etc/prometheus/prometheus.yml", "--storage.tsdb.path=/tmp/data/"},
							Ports: []v1.ContainerPort{
								{
									Name:          "web",
									ContainerPort: 9090,
									Protocol:      v1.ProtocolTCP,
								},
							},
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "data",
									MountPath: "/data",
								},
								{
									Name:      "config-volume",
									MountPath: "/etc/prometheus",
								},
							},
						},
					},
				},
			},
		},
	}

	return prometheusDeployment, nil
}

// GetPrometheusConfigMap creates a Horizon ConfigMap component for Prometheus
func (specConfig *PrometheusSpecConfig) GetPrometheusConfigMap() (*v1.ConfigMap, error) {
	data := map[string]interface{}{
		"global": map[string]interface{}{
			"scrape_interval": "5s",
		},
		"scrape_configs": []interface{}{
			map[string]interface{}{
				"job_name":        "synopsys-operator-scrape",
				"scrape_interval": "5s",
				"static_configs": []interface{}{
					map[string]interface{}{
						"targets": []string{"synopsys-operator:8080"},
					},
				},
			},
		},
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		return nil, errors.Trace(err)
	}

	cmData := map[string]string{}
	cmData["prometheus.yml"] = string(bytes)
	cmData["Image"] = specConfig.Image
	cmData["Expose"] = specConfig.Expose

	prometheusConfigMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "prometheus",
			Namespace: specConfig.Namespace,
			Labels: map[string]string{
				"app":       "synopsys-operator",
				"component": "prometheus",
			},
		},
		Data: cmData,
	}
	return prometheusConfigMap, nil
}

// GetOpenShiftRoute creates the OpenShift route component for the prometheus
//func (specConfig *PrometheusSpecConfig) GetOpenShiftRoute() *api.Route {
//	if strings.ToUpper(specConfig.Expose) == util.OPENSHIFT {
//		return &api.Route{
//			Name:               "synopsys-operator-prometheus",
//			Namespace:          specConfig.Namespace,
//			Kind:               "Service",
//			ServiceName:        "prometheus",
//			PortName:           "prometheus",
//			Labels:             map[string]string{"app": "synopsys-operator", "component": "prometheus"},
//			TLSTerminationType: routev1.TLSTerminationEdge,
//		}
//	}
//	return nil
//}
