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

// GetHubKnobs returns the default environs
func GetHubKnobs() map[string]string {
	return map[string]string{
		"IPV4_ONLY":                         "0",
		"USE_ALERT":                         "0",
		"USE_BINARY_UPLOADS":                "0",
		"RABBIT_MQ_PORT":                    "5671",
		"BROKER_USE_SSL":                    "yes",
		"SCANNER_CONCURRENCY":               "1",
		"HTTPS_VERIFY_CERTS":                "yes",
		"RABBITMQ_DEFAULT_VHOST":            "protecodesc",
		"RABBITMQ_SSL_FAIL_IF_NO_PEER_CERT": "false",
	}
}
