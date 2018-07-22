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

package model

type CreateHubRequest struct {
	Namespace        string `json:"namespace"`
	Flavor           string `json:"flavor"`
	DockerRegistry   string `json:"dockerRegistry"`
	DockerRepo       string `json:"dockerRepo"`
	HubVersion       string `json:"hubVersion"`
	AdminPassword    string `json:"adminPassword"`
	UserPassword     string `json:"userPassword"`
	PostgresPassword string `json:"postgresPassword"`
	IsRandomPassword bool   `json:"isRandomPassword"`
}

type CreateHub struct {
	Namespace        string `json:"namespace"`
	DockerRegistry   string `json:"dockerRegistry"`
	DockerRepo       string `json:"dockerRepo"`
	HubVersion       string `json:"hubVersion"`
	Flavor           string `json:"flavor"`
	AdminPassword    string `json:"adminPassword"`
	UserPassword     string `json:"userPassword"`
	PostgresPassword string `json:"postgresPassword"`
	IsRandomPassword bool   `json:"isRandomPassword"`
	Status           string `json:"status"`
}

type DeleteHubRequest struct {
	Namespace string `json:"namespace"`
}
