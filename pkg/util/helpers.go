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

package util

import (
	"fmt"
	"strings"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	opssightapi "github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
)

// Container defines the configuration for a container
type Container struct {
	ContainerConfig       *horizonapi.ContainerConfig
	EnvConfigs            []*horizonapi.EnvConfig
	VolumeMounts          []*horizonapi.VolumeMountConfig
	PortConfig            []*horizonapi.PortConfig
	ActionConfig          *horizonapi.ActionConfig
	ReadinessProbeConfigs []*horizonapi.ProbeConfig
	LivenessProbeConfigs  []*horizonapi.ProbeConfig
	PreStopConfig         *horizonapi.ActionConfig
}

// PodConfig used for configuring the pod
type PodConfig struct {
	Name                   string
	Labels                 map[string]string
	ServiceAccount         string
	Containers             []*Container
	Volumes                []*components.Volume
	InitContainers         []*Container
	PodAffinityConfigs     map[horizonapi.AffinityType][]*horizonapi.PodAffinityConfig
	PodAntiAffinityConfigs map[horizonapi.AffinityType][]*horizonapi.PodAffinityConfig
	NodeAffinityConfigs    map[horizonapi.AffinityType][]*horizonapi.NodeAffinityConfig
	ImagePullSecrets       []string
}

// MergeEnvMaps will merge the source and destination environs. If the same value exist in both, source environ will given more preference
func MergeEnvMaps(source, destination map[string]string) map[string]string {
	// if the source key present in the destination map, it will overrides the destination value
	// if the source value is empty, then delete it from the destination
	for key, value := range source {
		if len(value) == 0 {
			delete(destination, key)
		} else {
			destination[key] = value
		}
	}
	return destination
}

// MergeEnvSlices will merge the source and destination environs. If the same value exist in both, source environ will given more preference
func MergeEnvSlices(source, destination []string) []string {
	// create a destination map
	destinationMap := make(map[string]string)
	for _, value := range destination {
		values := strings.SplitN(value, ":", 2)
		if len(values) == 2 {
			mapKey := strings.TrimSpace(values[0])
			mapValue := strings.TrimSpace(values[1])
			if len(mapKey) > 0 && len(mapValue) > 0 {
				destinationMap[mapKey] = mapValue
			}
		}
	}

	// if the source key present in the destination map, it will overrides the destination value
	// if the source value is empty, then delete it from the destination
	for _, value := range source {
		values := strings.SplitN(value, ":", 2)
		if len(values) == 2 {
			mapKey := strings.TrimSpace(values[0])
			mapValue := strings.TrimSpace(values[1])
			if len(mapValue) == 0 {
				delete(destinationMap, mapKey)
			} else {
				destinationMap[mapKey] = mapValue
			}
		}
	}

	// convert destination map to string array
	mergedValues := []string{}
	for key, value := range destinationMap {
		mergedValues = append(mergedValues, fmt.Sprintf("%s:%s", key, value))
	}
	return mergedValues
}

// GetDefaultPasswords returns admin,user,postgres,hub passwords for db maintainance tasks.  Should only be used during
// initialization, or for 'babysitting' ephemeral hub instances (which might have postgres restarts)
// MAKE SURE YOU SEND THE NAMESPACE OF THE SECRET SOURCE (operator), NOT OF THE new hub  THAT YOUR TRYING TO CREATE !
func GetDefaultPasswords(kubeClient *kubernetes.Clientset, nsOfSecretHolder string) (adminPassword string, userPassword string, postgresPassword string, hubPassword string, err error) {
	blackduckSecret, err := GetSecret(kubeClient, nsOfSecretHolder, "blackduck-secret")
	if err != nil {
		log.Infof("warning: You need to first create a 'blackduck-secret' in this namespace with ADMIN_PASSWORD, USER_PASSWORD, POSTGRES_PASSWORD")
		return "", "", "", "", err
	}
	adminPassword = string(blackduckSecret.Data["ADMIN_PASSWORD"])
	userPassword = string(blackduckSecret.Data["USER_PASSWORD"])
	postgresPassword = string(blackduckSecret.Data["POSTGRES_PASSWORD"])
	hubPassword = string(blackduckSecret.Data["HUB_PASSWORD"])

	// default named return
	return adminPassword, userPassword, postgresPassword, hubPassword, err
}

// AppendBlackDuckSecrets will append the secrets of external and internal Black Duck
func AppendBlackDuckSecrets(existingExternalBlackDucks map[string]*opssightapi.Host, oldInternalBlackDucks []*opssightapi.Host, newInternalBlackDucks []*opssightapi.Host) map[string]*opssightapi.Host {
	existingInternalBlackducks := make(map[string]*opssightapi.Host)
	for _, oldInternalBlackDuck := range oldInternalBlackDucks {
		existingInternalBlackducks[oldInternalBlackDuck.Domain] = oldInternalBlackDuck
	}

	currentInternalBlackducks := make(map[string]*opssightapi.Host)
	for _, newInternalBlackDuck := range newInternalBlackDucks {
		currentInternalBlackducks[newInternalBlackDuck.Domain] = newInternalBlackDuck
	}

	for _, currentInternalBlackduck := range currentInternalBlackducks {
		// check if external host contains the internal host
		if _, ok := existingExternalBlackDucks[currentInternalBlackduck.Domain]; ok {
			// if internal host contains an external host, then check whether it is already part of status,
			// if yes replace it with existing internal host else with new internal host
			if existingInternalBlackduck, ok1 := existingInternalBlackducks[currentInternalBlackduck.Domain]; ok1 {
				existingExternalBlackDucks[currentInternalBlackduck.Domain] = existingInternalBlackduck
			} else {
				existingExternalBlackDucks[currentInternalBlackduck.Domain] = currentInternalBlackduck
			}
		} else {
			// add new internal Black Duck
			existingExternalBlackDucks[currentInternalBlackduck.Domain] = currentInternalBlackduck
		}
	}

	return existingExternalBlackDucks
}

// AppendBlackDuckHosts will append the old and new internal Black Duck hosts
func AppendBlackDuckHosts(oldBlackDucks []*opssightapi.Host, newBlackDucks []*opssightapi.Host) []*opssightapi.Host {
	existingBlackDucks := make(map[string]*opssightapi.Host)
	for _, oldBlackDuck := range oldBlackDucks {
		existingBlackDucks[oldBlackDuck.Domain] = oldBlackDuck
	}

	finalBlackDucks := []*opssightapi.Host{}
	for _, newBlackDuck := range newBlackDucks {
		if existingBlackduck, ok := existingBlackDucks[newBlackDuck.Domain]; ok {
			// add the existing internal Black Duck from the final Black Duck list
			finalBlackDucks = append(finalBlackDucks, existingBlackduck)
		} else {
			// add the new internal Black Duck to the final Black Duck list
			finalBlackDucks = append(finalBlackDucks, newBlackDuck)
		}
	}

	return finalBlackDucks
}
