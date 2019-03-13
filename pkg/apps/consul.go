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

package apps

import (
	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
)

// Consul stores the consul configuration
type Consul struct {
	namespace    string
	storageClass string
}

// NewConsul returns the consul configuration
func NewConsul(namespace string, storageClass string) *Consul {
	return &Consul{namespace: namespace, storageClass: storageClass}
}

// GetConsulStatefulSet will return the consul statefulset
func (c *Consul) GetConsulStatefulSet() *components.StatefulSet {
	envs := c.getConsulEnvConfigs()
	volumes := c.getConsulVolumes()
	volumeMounts := c.getConsulVolumeMounts()

	var containers []*util.Container

	containers = append(containers, &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{
			Name:       "consul",
			Image:      "consul:1.0.0",
			PullPolicy: horizonapi.PullAlways,
			MinMem:     "",
			MaxMem:     "",
			MinCPU:     "",
			MaxCPU:     "",
			Command: []string{
				"/bin/sh",
				"-ec",
				`IP=$(hostname -i)

            if [ -e /etc/consul/secrets/gossip-key ]; then
              echo "{\"encrypt\": \"$(base64 /etc/consul/secrets/gossip-key)\"}" > /etc/consul/encrypt.json
              GOSSIP_KEY="-config-file /etc/consul/encrypt.json"
            fi

            for i in $(seq 0 $((${INITIAL_CLUSTER_SIZE} - 1))); do
                while true; do
                    echo "Waiting for ${STATEFULSET_NAME}-${i}.${STATEFULSET_NAME} to come up"
                    ping -W 1 -c 1 ${STATEFULSET_NAME}-${i}.${STATEFULSET_NAME}.${STATEFULSET_NAMESPACE}.svc.cluster.local > /dev/null && break
                    sleep 1s
                done
            done

            PEERS=""
            for i in $(seq 0 $((${INITIAL_CLUSTER_SIZE} - 1))); do
              NEXT_PEER="$(ping -c 1 ${STATEFULSET_NAME}-${i}.${STATEFULSET_NAME}.${STATEFULSET_NAMESPACE}.svc.cluster.local | awk -F'[()]' '/PING/{print $2}')"
              if [ "${NEXT_PEER}" != "${POD_IP}" ]; then
                PEERS="${PEERS}${PEERS:+ } -retry-join ${STATEFULSET_NAME}-${i}.${STATEFULSET_NAME}.${STATEFULSET_NAMESPACE}.svc.cluster.local"
              fi
            done

            exec /bin/consul agent \
              -ui \
              -domain=consul \
              -data-dir=/var/lib/consul \
              -server \
              -bootstrap-expect=${INITIAL_CLUSTER_SIZE} \
              -disable-keyring-file \
              -bind=0.0.0.0 \
              -advertise=${IP} \
              ${PEERS} \
              ${GOSSIP_KEY} \
              -client=0.0.0.0 \
              -dns-port=${DNSPORT} \
              -http-port=8500`,
			},
		},
		EnvConfigs:   envs,
		VolumeMounts: volumeMounts,
		PortConfig: []*horizonapi.PortConfig{
			{ContainerPort: "8500", Protocol: horizonapi.ProtocolTCP},
			{ContainerPort: "8400", Protocol: horizonapi.ProtocolTCP},
			{ContainerPort: "8301", Protocol: horizonapi.ProtocolTCP},
			{ContainerPort: "8301", Protocol: horizonapi.ProtocolUDP},
			{ContainerPort: "8302", Protocol: horizonapi.ProtocolTCP},
			{ContainerPort: "8302", Protocol: horizonapi.ProtocolUDP},
			{ContainerPort: "8300", Protocol: horizonapi.ProtocolUDP},
			{ContainerPort: "8600", Protocol: horizonapi.ProtocolUDP},
			{ContainerPort: "8600", Protocol: horizonapi.ProtocolTCP},
		},
		LivenessProbeConfigs: []*horizonapi.ProbeConfig{
			{
				ActionConfig: horizonapi.ActionConfig{
					Command: []string{"consul", "members"},
				},
				Delay:   300,
				Timeout: 5,
			},
		},
	})

	stateFulSetConfig := &horizonapi.StatefulSetConfig{
		Name:      "consul",
		Namespace: c.namespace,
		Replicas:  util.IntToInt32(3),
		Service:   "consul",
	}

	stateFulSet := util.CreateStateFulSetFromContainer(stateFulSetConfig, "", containers, volumes, nil, nil)

	claim, _ := util.CreatePersistentVolumeClaim("datadir", c.namespace, "1Gi", c.storageClass, horizonapi.ReadWriteOnce)
	stateFulSet.AddVolumeClaimTemplate(*claim)
	return stateFulSet
}

// GetConsulServices will return the consul service
func (c *Consul) GetConsulServices() *components.Service {
	// Consul service
	consul := components.NewService(horizonapi.ServiceConfig{
		Name:          "consul",
		Namespace:     c.namespace,
		IPServiceType: horizonapi.ClusterIPServiceTypeDefault,
	})
	consul.AddSelectors(map[string]string{
		"app": "consul",
	})
	consul.AddPort(horizonapi.ServicePortConfig{Name: "http", Port: 8500})
	consul.AddPort(horizonapi.ServicePortConfig{Name: "rpc", Port: 8400})
	consul.AddPort(horizonapi.ServicePortConfig{Name: "serflan-tcp", Port: 8301, Protocol: horizonapi.ProtocolTCP})
	consul.AddPort(horizonapi.ServicePortConfig{Name: "serflan-udp", Port: 8301, Protocol: horizonapi.ProtocolUDP})
	consul.AddPort(horizonapi.ServicePortConfig{Name: "serfwan-tcp", Port: 8302, Protocol: horizonapi.ProtocolTCP})
	consul.AddPort(horizonapi.ServicePortConfig{Name: "serfwan-udp", Port: 8302, Protocol: horizonapi.ProtocolUDP})
	consul.AddPort(horizonapi.ServicePortConfig{Name: "server", Port: 8300})
	consul.AddPort(horizonapi.ServicePortConfig{Name: "consuldns-tcp", Port: 8600, Protocol: horizonapi.ProtocolTCP})
	consul.AddPort(horizonapi.ServicePortConfig{Name: "consuldns-udp", Port: 8600, Protocol: horizonapi.ProtocolUDP})

	return consul
}

// getConsulVolumes will return the postgres volumes
func (c *Consul) getConsulVolumes() []*components.Volume {
	var volumes []*components.Volume
	volumes = append(volumes, components.NewSecretVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      "gossip-key",
		MapOrSecretName: "consul-gossip-key",
	}))
	return volumes
}

// getConsulVolumeMounts will return the postgres volume mounts
func (c *Consul) getConsulVolumeMounts() []*horizonapi.VolumeMountConfig {
	var volumeMounts []*horizonapi.VolumeMountConfig
	volumeMounts = append(volumeMounts, &horizonapi.VolumeMountConfig{Name: "datadir", MountPath: "/var/lib/consul"})
	volumeMounts = append(volumeMounts, &horizonapi.VolumeMountConfig{Name: "gossip-key", MountPath: "/etc/consul/secrets"})
	return volumeMounts
}

// getConsulEnvConfigs will return the postgres environment config maps
func (c *Consul) getConsulEnvConfigs() []*horizonapi.EnvConfig {
	var envs []*horizonapi.EnvConfig
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "INITIAL_CLUSTER_SIZE", KeyOrVal: "3"})
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "DNSPORT", KeyOrVal: "8600"})
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "STATEFULSET_NAME", KeyOrVal: "consul"})
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvFromPodIP, NameOrPrefix: "POD_IP"})
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvFromNamespace, NameOrPrefix: "STATEFULSET_NAMESPACE"})
	return envs
}

// GetConsulSecrets returns the consul secret
func (c *Consul) GetConsulSecrets() *components.Secret {
	gossipKey := components.NewSecret(horizonapi.SecretConfig{
		Name:      "consul-gossip-key",
		Namespace: c.namespace,
		Type:      horizonapi.SecretTypeOpaque,
	})

	rand, _ := util.RandomString(24)
	gossipKey.AddStringData(map[string]string{
		"gossip-key": rand,
	})

	return gossipKey
}
