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

package v1

import (
	"fmt"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	opssightapi "github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/opssight/components/configmap/opssight"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/store"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/types"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/utils"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/juju/errors"
	"k8s.io/client-go/kubernetes"
)

// OpsSightConfigmap holds the OpsSight config map configuration
type OpsSightConfigmap struct {
	config     *protoform.Config
	kubeClient *kubernetes.Clientset
	opsSight   *opssightapi.OpsSight
}

func init() {
	store.Register(types.OpsSightConfigMapV1, NewOpsSightConfigmap)
}

// NewOpsSightConfigmap returns the OpsSight config map configuration
func NewOpsSightConfigmap(config *protoform.Config, kubeClient *kubernetes.Clientset, cr interface{}) (types.ConfigMapInterface, error) {
	opsSight, ok := cr.(*opssightapi.OpsSight)
	if !ok {
		return nil, fmt.Errorf("unable to cast the interface to OpsSight object")
	}
	return &OpsSightConfigmap{config: config, kubeClient: kubeClient, opsSight: opsSight}, nil
}

// GetCM returns the config map
func (o *OpsSightConfigmap) GetCM() (*components.ConfigMap, error) {
	name := utils.GetResourceName(o.opsSight.Name, util.OpsSightName, o.opsSight.Spec.ConfigMapName)
	cm := components.NewConfigMap(horizonapi.ConfigMapConfig{
		Name:      name,
		Namespace: o.opsSight.Spec.Namespace,
	})

	configMapString, err := o.getCM()
	if err != nil {
		return nil, errors.Trace(err)
	}

	cm.AddLabels(map[string]string{"app": "opssight", "component": name, "name": o.opsSight.Name})
	cm.AddData(map[string]string{fmt.Sprintf("%s.json", o.opsSight.Spec.ConfigMapName): configMapString})

	return cm, nil
}

func (o *OpsSightConfigmap) getCM() (string, error) {
	opssightSpec := &o.opsSight.Spec
	configMap := &opssight.MainOpssightConfigMap{
		LogLevel: opssightSpec.LogLevel,
		BlackDuck: &opssight.BlackDuckConfig{
			ConnectionsEnvironmentVariableName: opssightSpec.Blackduck.ConnectionsEnvironmentVariableName,
			TLSVerification:                    opssightSpec.Blackduck.TLSVerification,
		},
		ImageFacade: &opssight.ImageFacadeConfig{
			CreateImagesOnly: false,
			Host:             "localhost",
			Port:             opssightSpec.ScannerPod.ImageFacade.Port,
			ImagePullerType:  opssightSpec.ScannerPod.ImageFacade.ImagePullerType,
		},
		Perceiver: &opssight.PerceiverConfig{
			Image: &opssight.ImagePerceiverConfig{},
			Pod: &opssight.PodPerceiverConfig{
				NamespaceFilter: opssightSpec.Perceiver.PodPerceiver.NamespaceFilter,
			},
			AnnotationIntervalSeconds: opssightSpec.Perceiver.AnnotationIntervalSeconds,
			DumpIntervalMinutes:       opssightSpec.Perceiver.DumpIntervalMinutes,
			Port:                      opssightSpec.Perceiver.Port,
		},
		Perceptor: &opssight.PerceptorConfig{
			Timings: &opssight.PerceptorTimingsConfig{
				CheckForStalledScansPauseHours: opssightSpec.Perceptor.CheckForStalledScansPauseHours,
				ClientTimeoutMilliseconds:      opssightSpec.Perceptor.ClientTimeoutMilliseconds,
				ModelMetricsPauseSeconds:       opssightSpec.Perceptor.ModelMetricsPauseSeconds,
				StalledScanClientTimeoutHours:  opssightSpec.Perceptor.StalledScanClientTimeoutHours,
				UnknownImagePauseMilliseconds:  opssightSpec.Perceptor.UnknownImagePauseMilliseconds,
			},
			Host:        utils.GetResourceName(o.opsSight.Name, util.OpsSightName, opssightSpec.Perceptor.Name),
			Port:        opssightSpec.Perceptor.Port,
			UseMockMode: false,
		},
		Scanner: &opssight.ScannerConfig{
			BlackDuckClientTimeoutSeconds: opssightSpec.ScannerPod.Scanner.ClientTimeoutSeconds,
			ImageDirectory:                opssightSpec.ScannerPod.ImageDirectory,
			Port:                          opssightSpec.ScannerPod.Scanner.Port,
		},
		Skyfire: &opssight.SkyfireConfig{
			BlackDuckClientTimeoutSeconds: opssightSpec.Skyfire.HubClientTimeoutSeconds,
			BlackDuckDumpPauseSeconds:     opssightSpec.Skyfire.HubDumpPauseSeconds,
			KubeDumpIntervalSeconds:       opssightSpec.Skyfire.KubeDumpIntervalSeconds,
			PerceptorDumpIntervalSeconds:  opssightSpec.Skyfire.PerceptorDumpIntervalSeconds,
			Port:                          opssightSpec.Skyfire.Port,
			PrometheusPort:                opssightSpec.Skyfire.PrometheusPort,
			UseInClusterConfig:            true,
		},
	}

	return opssight.JSONString(configMap)
}
