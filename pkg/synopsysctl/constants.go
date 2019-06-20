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

package synopsysctl

// DefaultOperatorImage is Synopsys Operator image that is deployed by default
const DefaultOperatorImage string = "docker.io/blackducksoftware/synopsys-operator:2019.6.0"

// DefaultMetricsImage is the Metrics image deployed with Synopsys Operator by default
const DefaultMetricsImage string = "docker.io/prom/prometheus:v2.1.0"

// DefaultOperatorNamespace is the default namespace of Synopsys Operator
const DefaultOperatorNamespace string = "synopsys-operator"

// Default Base Specs for Create
const defaultBaseAlertSpec string = "default"
const defaultBaseBlackDuckSpec string = "persistentStorageLatest"
const defaultBaseOpsSightSpec string = "default"
