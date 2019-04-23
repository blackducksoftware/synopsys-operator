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

package containers

import (
	"strings"
)

const envOptions = `
IPV4_ONLY=0
USE_ALERT=0
USE_BINARY_UPLOADS=0
RABBIT_MQ_HOST=rabbitmq
RABBIT_MQ_PORT=5671
BROKER_URL=amqps://rabbitmq/protecodesc
BROKER_USE_SSL=yes
CFSSL=cfssl:8888
HUB_LOGSTASH_HOST=logstash
SCANNER_CONCURRENCY=1
HTTPS_VERIFY_CERTS=yes
RABBITMQ_DEFAULT_VHOST=protecodesc
RABBITMQ_SSL_FAIL_IF_NO_PEER_CERT=false
CLIENT_CERT_CN=binaryscanner
ENABLE_SOURCE_UPLOADS=false
DATA_RETENTION_IN_DAYS=180
MAX_TOTAL_SOURCE_SIZE_MB=4000`

// GetHubKnobs returns the default environs
func GetHubKnobs() (env map[string]string) {
	env = map[string]string{}
	for _, val := range strings.Split(envOptions, "\n") {
		if strings.Contains(val, "=") {
			keyval := strings.Split(val, "=")
			env[keyval[0]] = keyval[1]
		}
	}
	return env
}
