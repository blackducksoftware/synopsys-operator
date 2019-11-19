/*
 * Copyright (C) 2019 Synopsys, Inc.
 *
 *  Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements. See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership. The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 *  with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 *  under the License.
 */

package bdba

// BDBA configures all BDBA specifications
type BDBA struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Version   string `json:"version"`

	Hostname    string `json:"hostname"`
	IngressHost string `json:"ingressHost"`

	MinioAccessKey string `json:"minioAccessKey"`
	MinioSecretKey string `json:"minioSecretKey"`

	WorkerReplicas int `json:"workerReplicas"`

	AdminEmail string `json:"adminEmail"`

	BrokerURL string `json:"brokerURL"`

	PGPPassword string `json:"pgpPassword"`

	RabbitMQULimitNoFiles string `json:"rabbitMQULimitNoFiles"`

	HideLicenses      string `json:"hideLicenses"`
	LicensingPassword string `json:"licensingPassword"`
	LicensingUsername string `json:"licensingUsername"`

	InsecureCookies  string `json:"insecureCookies"`
	SessionCookieAge string `json:"sessionCookieAge"`

	URL       string `json:"url"`
	Actual    string `json:"actual"`
	Expected  string `json:"expected"`
	StartFlag string `json:"startFlag"`
	Result    string `json:"result"`
}
