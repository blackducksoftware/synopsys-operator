#!/bin/sh
#
# Copyright (C) 2018 Synopsys, Inc.
#
# Licensed to the Apache Software Foundation (ASF) under one
# or more contributor license agreements. See the NOTICE file
# distributed with this work for additional information
# regarding copyright ownership. The ASF licenses this file
# to you under the Apache License, Version 2.0 (the
# "License"); you may not use this file except in compliance
# with the License. You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing,
# software distributed under the License is distributed on an
# "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
# KIND, either express or implied. See the License for the
# specific language governing permissions and limitations
# under the License.

function get_user_input() {
  # Prompt for configuration information
  clear
  read -p "Install Perceptor [y/N] ? " install_perceptor
  read -p "Install Alert [y/N] ? " install_alert
  read -p "Install Hub Federator [y/N] ? " install_hub_federator

  # Convert everything to uppercase
  install_perceptor=$(upper $install_perceptor)
  install_alert=$(upper $install_alert)
  install_hub_federator=$(upper $install_hub_federator)

  # Input needed for all cases
  read -p "Cluster config file: " cluster_config_file
  echo "ClusterConfigFile: $cluster_config_file" >> $bootstrapper_config

  # configure the apps
  configure_hub

  if [[ "$install_perceptor" == "Y" || "$install_perceptor" == "YES" ]]
  then
    configure_perceptor
  fi

  if [[ "$install_alert" == "Y" || "$install_alert" == "YES" ]]
  then
    configure_alert
  fi

  if [[ "$install_hub_federator" == "Y" || "$install_hub_federator" == "YES" ]]
  then
    configure_hub_federator
  fi
}

function upper() {
  input=$1
  output=$(echo $input | tr '[:lower:]' '[:upper:]')
  echo "$output"
}

function configure_hub() {
  printf "\n============================================\n"
  printf "Hub Configuration Information\n"
  printf "============================================\n"
  read -p "Hub server host (e.g. hub.mydomain.com): " hub_host
  read -p "Hub server port: " hub_port
  read -p "Hub user name: " hub_user
  read -sp "Hub user password: " hub_password

  echo "HubHost: $hub_host" >> $bootstrapper_config
  echo "HubUser: $hub_user" >> $bootstrapper_config
  echo "HubPort: $hub_port" >> $bootstrapper_config
  echo "HubUserPassword: $hub_password" >> $bootstrapper_config
  printf "\n"
}

function configure_perceptor() {
  printf "\n============================================\n"
  printf "Perceptor Configuration Information\n"
  printf "============================================\n"
  read -p "Is the target cluster running openshift? [y/N] " is_openshift
  is_openshift=$(upper $is_openshift)

  echo "AnnotatePods: true" >> $bootstrapper_config
  if [[ "$is_openshift" == "Y" || "$is_openshift" == "YES" ]]
  then
    echo "AnnotateImages: true" >> $bootstrapper_config
  fi

  read -p "Maximum concurrent scans: " max_scans
  echo "ConcurrentScanLimit: $max_scans" >> $bootstrapper_config
}

function configure_alert() {
#  printf "\n============================================\n"
#  printf "Alert Configuration Information\n"
#  printf "============================================\n"

  echo "AlertEnabled: true" >> $bootstrapper_config
}

function configure_hub_federator() {
  printf "\n============================================\n"
  printf "Hub Federator Configuration Information\n"
  printf "============================================\n"

  echo "HubFederatorEnabled: true" >> $bootstrapper_config

  read -p "Hub Registration Key: " key
  echo "HubFederatorRegistrationKey: $key" >> $bootstrapper_config
}

config_file=$1
config_created=0
cluster_config_file=""

REGISTRY=${REGISTRY:-"gcr.io"}
IMAGE_PATH=${IMAGE_PATH:-"gke-verification/blackducksoftware/protoform-bootstrapper"}
IMAGE_VERSION=${IMAGE_VERSION:-"master"}

# get input from user if no config provided
bootstrapper_config=$(mktemp /tmp/bootstrap_config-XXXXXX)
if [[ "$config_file" == "" ]]
then
  get_user_input
else
  cp $config_file $bootstrapper_config
  cluster_config_file=`grep ClusterConfigFile $config_file | cut -d':' -f 2 | tr -d [:space:]`
fi

# give the config a yaml extension so viper will read it
mv $bootstrapper_config "${bootstrapper_config}.yaml"
bootstrapper_config="${bootstrapper_config}.yaml"

# launch the bootstrapper
cluster_config_path=$(dirname $cluster_config_file)
config_dir=$(dirname $bootstrapper_config)
docker run -v $config_dir:/overrides:ro -v $cluster_config_path:/$cluster_config_path:ro $REGISTRY/$IMAGE_PATH:$IMAGE_VERSION ./bootstrapper ./defaults.yaml /overrides/$(basename $bootstrapper_config)

# cleanup
rm -f ${bootstrapper_config}
