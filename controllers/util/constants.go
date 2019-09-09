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

package util

const (
	// OPENSHIFT denotes to create an OpenShift routes
	OPENSHIFT = "OPENSHIFT"
	// NONE denotes no exposed service
	NONE = "NONE"
	// NODEPORT denotes to create a NodePort service
	NODEPORT = "NODEPORT"
	// LOADBALANCER denotes to create a LoadBalancer service
	LOADBALANCER = "LOADBALANCER"

	// AlertCRDName is the name of an Alert CRD
	AlertCRDName = "alerts.synopsys.com"
	// BlackDuckCRDName is the name of the Black Duck CRD
	BlackDuckCRDName = "blackducks.synopsys.com"
	// OpsSightCRDName is the name of an OpsSight CRD
	OpsSightCRDName = "opssights.synopsys.com"
	// SizeCRDName is the name of the Size CRD
	SizeCRDName = "sizes.synopsys.com"

	// OPERATOR is the name of an Operator
	OPERATOR = "synopsys-operator"
	// ALERT is the name of an Alert app
	ALERT = "alert"
	// BLACKDUCK is the name of the Black Duck app
	BLACKDUCK = "blackduck"
	// OPSSIGHT is the name of an OpsSight app
	OPSSIGHT = "opssight"
	// Releases repo details
	POLARIS     = "polaris"
	POLARISDB   = "polarisdb"
	REPORTING   = "reporting"
	AUTHSERVICE = "polaris"
)
