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
	"encoding/json"
	"fmt"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	opssightapi "github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
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
	store.Register(types.OpsSightMetricsConfigMapV1, NewOpsSightConfigmap)
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
	if !o.opsSight.Spec.EnableMetrics {
		return nil, nil
	}

	configMap := components.NewConfigMap(horizonapi.ConfigMapConfig{
		Name:      utils.GetResourceName(o.opsSight.Name, util.OpsSightName, "prometheus"),
		Namespace: o.opsSight.Spec.Namespace,
	})

	targets := []string{
		fmt.Sprintf("%s:%d", utils.GetResourceName(o.opsSight.Name, util.OpsSightName, o.opsSight.Spec.Perceptor.Name), o.opsSight.Spec.Perceptor.Port),
		fmt.Sprintf("%s:%d", utils.GetResourceName(o.opsSight.Name, util.OpsSightName, o.opsSight.Spec.ScannerPod.Scanner.Name), o.opsSight.Spec.ScannerPod.Scanner.Port),
		fmt.Sprintf("%s:%d", utils.GetResourceName(o.opsSight.Name, util.OpsSightName, o.opsSight.Spec.ScannerPod.ImageFacade.Name), o.opsSight.Spec.ScannerPod.ImageFacade.Port),
	}
	if o.opsSight.Spec.Perceiver.EnableImagePerceiver {
		targets = append(targets, fmt.Sprintf("%s:%d", utils.GetResourceName(o.opsSight.Name, util.OpsSightName, o.opsSight.Spec.Perceiver.ImagePerceiver.Name), o.opsSight.Spec.Perceiver.Port))
	}
	if o.opsSight.Spec.Perceiver.EnablePodPerceiver {
		targets = append(targets, fmt.Sprintf("%s:%d", utils.GetResourceName(o.opsSight.Name, util.OpsSightName, o.opsSight.Spec.Perceiver.PodPerceiver.Name), o.opsSight.Spec.Perceiver.Port))
	}
	if o.opsSight.Spec.EnableSkyfire {
		targets = append(targets, fmt.Sprintf("%s:%d", utils.GetResourceName(o.opsSight.Name, util.OpsSightName, o.opsSight.Spec.Skyfire.Name), o.opsSight.Spec.Skyfire.PrometheusPort))
	}
	data := map[string]interface{}{
		"global": map[string]interface{}{
			"scrape_interval": "5s",
		},
		"scrape_configs": []interface{}{
			map[string]interface{}{
				"job_name":        "perceptor-scrape",
				"scrape_interval": "5s",
				"static_configs": []interface{}{
					map[string]interface{}{
						"targets": targets,
					},
				},
			},
		},
	}
	bytes, err := json.Marshal(data)
	if err != nil {
		return nil, errors.Trace(err)
	}
	configMap.AddLabels(map[string]string{"app": "opssight", "name": o.opsSight.Name, "component": "prometheus"})
	configMap.AddData(map[string]string{"prometheus.yml": string(bytes)})

	return configMap, nil
}
