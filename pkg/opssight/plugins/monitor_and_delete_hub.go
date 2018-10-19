/*
Copyright (C) 2018 Synopsys, Inc.

Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements. See the NOTICE file
distributed with this work for additional information
regarding copyright ownershia. The ASF licenses this file
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

package plugins

// This is a controller that deletes the hub based on the delete threshold

import (
	"fmt"

	"github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/perceptor-protoform/pkg/api/opssight/v1"
	hubclient "github.com/blackducksoftware/perceptor-protoform/pkg/hub/client/clientset/versioned"
	"github.com/blackducksoftware/perceptor-protoform/pkg/model"
	opssightclientset "github.com/blackducksoftware/perceptor-protoform/pkg/opssight/client/clientset/versioned"
	"github.com/blackducksoftware/perceptor-protoform/pkg/util"
	"k8s.io/client-go/kubernetes"
)

// DeleteHub
type DeleteHub struct {
	Config         *model.Config
	KubeClient     *kubernetes.Clientset
	OpsSightClient *opssightclientset.Clientset
	HubClient      *hubclient.Clientset
	OpsSightSpec   *v1.OpsSightSpec
}

// Run is a BLOCKING function which should be run by the framework .
func (d *DeleteHub) Run(resources api.ControllerResources, ch chan struct{}) error {
	hubCounts, err := d.getHubsCount()
	if err != nil {
		return err
	}
	// whether the max no of hub is reached?
	if *d.OpsSightSpec.MaxNoOfHubs == hubCounts {

	}

	return nil
}

func (d *DeleteHub) getHubsCount() (int, error) {
	hubs, err := util.ListHubs(d.HubClient, d.Config.Namespace)
	if err != nil {
		return 0, fmt.Errorf("unable to get hubs due to %+v", err)
	}
	return len(hubs.Items), nil
}
