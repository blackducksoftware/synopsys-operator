// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

// Start Command Defaults
var start_synopsysOperatorImage = "docker.io/blackducksoftware/synopsys-operator:2018.12.0"
var start_prometheusImage = "docker.io/prom/prometheus:v2.1.0"
var start_blackduckRegistrationKey = ""
var start_dockerConfigPath = ""

var start_secretName = "blackduck-secret"
var start_secretType = "Opaque"
var start_secretAdminPassword = "YmxhY2tkdWNr"
var start_secretPostgresPassword = "YmxhY2tkdWNr"
var start_secretUserPassword = "YmxhY2tkdWNr"
var start_secretBlackduckPassword = "YmxhY2tkdWNr"

// Create Blackduck Defaults
var create_blackduck_size = 11
var create_blackduck_persistentStorage = true

var namespace = ""
