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

package components

import (
	// RCs
	_ "github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/components/rc/authentication/v1"
	_ "github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/components/rc/binaryscanner/v1"
	_ "github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/components/rc/cfssl/v1"
	_ "github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/components/rc/documentation/v1"
	_ "github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/components/rc/jobrunner/v1"
	_ "github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/components/rc/postgres/v1"
	_ "github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/components/rc/rabbitmq/v1"
	_ "github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/components/rc/registration/v1"
	_ "github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/components/rc/scan/v1"
	_ "github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/components/rc/solr/v1"
	_ "github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/components/rc/uploadcache/v1"
	_ "github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/components/rc/webapplogstash/v1"
	_ "github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/components/rc/webserver/v1"
	_ "github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/components/rc/zookeeper/v1"

	// Services
	_ "github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/components/service/authentication/v1"
	_ "github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/components/service/cfssl/v1"
	_ "github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/components/service/documentation/v1"
	_ "github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/components/service/expose/v1"
	_ "github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/components/service/logstash/v1"
	_ "github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/components/service/postgres/v1"
	_ "github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/components/service/rabbitmq/v1"
	_ "github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/components/service/registration/v1"
	_ "github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/components/service/scan/v1"
	_ "github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/components/service/solr/v1"
	_ "github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/components/service/uploadcache/v1"
	_ "github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/components/service/webapp/v1"
	_ "github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/components/service/webserver/v1"
	_ "github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/components/service/zookeeper/v1"

	// Configmap
	_ "github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/components/configmap/database/v1"
	_ "github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/components/configmap/global/v1"

	// Secrets
	_ "github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/components/secret/authcertificate/v1"
	_ "github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/components/secret/postgres/v1"
	_ "github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/components/secret/proxycertificate/v1"
	_ "github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/components/secret/uploadcache/v1"
	_ "github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/components/secret/webcertificate/v1"

	// PVCs
	_ "github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/components/pvc/v1"
	_ "github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/components/pvc/v2"
)
