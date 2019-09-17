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

package polaris

type Polaris struct {
	Namespace        string `json:"namespace,omitempty"`
	EnvironmentName  string `json:"environment"`
	EnvironmentDNS   string `json:"environment_address"`
	ImagePullSecrets string `json:"image_pull_secrets"`
	Version          string `json:"version"`
	EnableReporting  bool   `json:"enable_reporting"`
	//ReportingSpec *ReportingSpec `json:"reporting_spec,omitempty"`
	PolarisDBSpec *PolarisDBSpec `json:"polaris_db_spec,omitempty"`
	PolarisSpec   *PolarisSpec   `json:"polaris_spec,omitempty"`
}

type ResourcesSpec struct {
	RequestsSpec RequestsSpec `json:"requests,omitempty"`
	LimitsSpec   LimitsSpec   `json:"limits,omitempty"`
}

type RequestsSpec struct {
	CPU    string `json:"cpu,omitempty"`
	Memory string `json:"memory,omitempty"`
}

type LimitsSpec struct {
	CPU    string `json:"cpu,omitempty"`
	Memory string `json:"memory,omitempty"`
}

type ImageDetails struct {
	Repository string `json:"repository"`
	Image      string `json:"image"`
	Tag        string `json:"tag"`
}

type PolarisSpec struct {
}

type PolarisDBSpec struct {
	SMTPDetails            SMTPDetails            `json:"smtp_details"`
	PostgresInstanceType   string                 `json:"postgres_instance_type"`
	PostgresStorageDetails PostgresStorageDetails `json:"postgres_storage_details,omitempty"`
	PostgresDetails        PostgresDetails        `json:"postgres"`
	EventstoreDetails      EventstoreDetails      `json:"eventstore_details,omitempty"`
	UploadServerDetails    UploadServerDetails    `json:"upload_server_details,omitempty"`
}

type EventstoreDetails struct {
	Replicas *int32  `json:"replicas,omitempty"`
	Storage  Storage `json:"storage_size,omitempty"`
}

type UploadServerDetails struct {
	Replicas      *int32        `json:"replicas,omitempty"`
	ResourcesSpec ResourcesSpec `json:"resources,omitempty"`
	Storage       Storage       `json:"storage,omitempty"`
}

type Storage struct {
	StorageSize string `json:"size,omitempty"`
}

type SMTPDetails struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

type PostgresDetails struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type PostgresStorageDetails struct {
	StorageSize  string  `json:"size,omitempty"`
	StorageClass *string `json:"storage_class,omitempty"`
}

type AuthServerSpec struct {
}
