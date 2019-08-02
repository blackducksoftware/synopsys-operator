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

	"github.com/blackducksoftware/horizon/pkg/components"
	blackduckapi "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/components/rc/utils"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/database/postgres"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/store"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/types"
	apputils "github.com/blackducksoftware/synopsys-operator/pkg/apps/utils"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"k8s.io/client-go/kubernetes"
)

// BdReplicationController holds the Black Duck RC configuration
type BdReplicationController struct {
	*types.PodResource
	config     *protoform.Config
	kubeClient *kubernetes.Clientset
	blackDuck  *blackduckapi.Blackduck
}

func init() {
	store.Register(types.BlackDuckPostgresRCV1, NewBdReplicationController)
}

// NewBdReplicationController returns the Black Duck RC configuration
func NewBdReplicationController(podResource *types.PodResource, config *protoform.Config, kubeClient *kubernetes.Clientset, cr interface{}) (types.ReplicationControllerInterface, error) {
	blackDuck, ok := cr.(*blackduckapi.Blackduck)
	if !ok {
		return nil, fmt.Errorf("unable to cast the interface to Black Duck object")
	}
	return &BdReplicationController{PodResource: podResource, config: config, kubeClient: kubeClient, blackDuck: blackDuck}, nil
}

// GetRc returns the RC
func (c *BdReplicationController) GetRc() (*components.ReplicationController, error) {
	containerConfig, ok := c.Containers[types.PostgresContainerName]
	if !ok {
		return nil, fmt.Errorf("couldn't find container %s", types.PostgresContainerName)
	}

	name := apputils.GetResourceName(c.blackDuck.Name, util.BlackDuckName, "postgres")

	var pvcName string
	if c.blackDuck.Spec.PersistentStorage {
		pvcName = utils.GetPVCName("postgres", c.blackDuck)
	}

	p := &postgres.Postgres{
		Name:                   name,
		Namespace:              c.blackDuck.Spec.Namespace,
		PVCName:                pvcName,
		Port:                   int32(5432),
		Image:                  containerConfig.Image,
		MinCPU:                 util.Int32ToInt(containerConfig.MinCPU),
		MaxCPU:                 util.Int32ToInt(containerConfig.MaxCPU),
		MinMemory:              util.Int32ToInt(containerConfig.MinMem),
		MaxMemory:              util.Int32ToInt(containerConfig.MaxMem),
		Database:               "blackduck",
		User:                   "blackduck",
		PasswordSecretName:     apputils.GetResourceName(c.blackDuck.Name, util.BlackDuckName, "db-creds"),
		UserPasswordSecretKey:  "HUB_POSTGRES_ADMIN_PASSWORD_FILE",
		AdminPasswordSecretKey: "HUB_POSTGRES_POSTGRES_PASSWORD_FILE",
		MaxConnections:         300,
		SharedBufferInMB:       1024,
		EnvConfigMapRefs:       []string{apputils.GetResourceName(c.blackDuck.Name, util.BlackDuckName, "db-config")},
		Labels:                 apputils.GetLabel("postgres", c.blackDuck.Name),
		IsOpenshift:            c.config.IsOpenshift,
	}

	return p.GetPostgresReplicationController()
}
