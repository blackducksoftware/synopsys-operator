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
	Namespace        string         `json:"namespace,omitempty"`
	EnvironmentName  string         `json:"environment"`
	EnvironmentDNS   string         `json:"environment_address"`
	ImagePullSecrets string         `json:"image_pull_secrets"`
	Version          string         `json:"version"`
	EnableReporting  bool           `json:"enable_reporting"`
	PolarisDBSpec    *PolarisDBSpec `json:"polaris_db_spec,omitempty"`
	PolarisSpec      *PolarisSpec   `json:"polaris_spec,omitempty"`
	ReportingSpec    *ReportingSpec `json:"reporting_spec,omitempty"`
}

type PolarisDBSpec struct {
	SMTPDetails          SMTPDetails         `json:"smtp_details"`
	PostgresInstanceType string              `json:"postgres_instance_type"`
	PostgresDetails      PostgresDetails     `json:"postgres_details"`
	EventstoreDetails    EventstoreDetails   `json:"eventstore_details,omitempty"`
	UploadServerDetails  UploadServerDetails `json:"upload_server_details,omitempty"`
	MongoDBDetails       MongoDBDetails      `json:"mongodb_details,omitempty"`
}

type PolarisSpec struct {
	AuthServerDetails            *AuthServerDetails            `json:"auth_server_details,omitempty"`
	ConfigsServiceDetails        *ConfigsServiceDetails        `json:"configs_service_details,omitempty"`
	CosServerDetails             *CosServerDetails             `json:"cos_server_details,omitempty"`
	DesktopMetricsServiceDetails *DesktopMetricsServiceDetails `json:"desktop_metrics_service_details,omitempty"`
	DownloadServerDetails        *DownloadServerDetails        `json:"download_server_details,omitempty"`
	IssueServerDetails           *IssueServerDetails           `json:"issue_server_details,omitempty"`
	JobsControllerServiceDetails *JobsControllerServiceDetails `json:"jobs_controller_service_details,omitempty"`
	JobsServiceDetails           *JobsServiceDetails           `json:"jobs_service_details,omitempty"`
	LogsServiceDetails           *LogsServiceDetails           `json:"logs_service_details,omitempty"`
	PericlesSwaggerUIDetails     *PericlesSwaggerUIDetails     `json:"pericles_swagger_ui_detail,omitempty"`
	TaxonomyServerDeails         *TaxonomyServerDeails         `json:"taxonomy_server_details,omitempty"`
	TDSCodeAnalysisDetails       *TDSCodeAnalysisDetails       `json:"tds_code_analysis_details,omitempty"`
	ToolsServiceDetails          *ToolsServiceDetails          `json:"tools_service_details,omitempty"`
	TriageCommandHandlerDetails  *TriageCommandHandlerDetails  `json:"triage_command_handler,omitempty"`
	TriageQueryDetails           *TriageQueryDetails           `json:"triage_query_details,omitempty"`
	VinylServerDetails           *VinylServerDetails           `json:"vinyl_server_details,omitempty"`
	WebCoreDetails               *WebCoreDetails               `json:"web_core_details,omitempty"`
	WebHelpDetails               *WebHelpDetails               `json:"web_help_details,omitempty"`
}

type ReportingSpec struct {
	RPFrontendDetails              *RPFrontendDetails              `json:"rp_frontend_details,omitempty"`
	RPIssueManager                 *RPIssueManager                 `json:"rp_issue_manager,omitemtpy"`
	RPPolarisAgentServiceDetails   *RPPolarisAgentServiceDetails   `json:"rp_polaris_agent_service_details,omitempty"`
	RPPortfolioServiceDetails      *RPPortfolioServiceDetails      `json:"rp_portfolio_service_details,omitempty"`
	RPReportServiceDetails         *RPReportServiceDetails         `json:"rp_report_service_details,omitempty"`
	RPSwaggerDocDetails            *RPSwaggerDocDetails            `json:"rp_swagger_doc_details,omitempty"`
	RPToolsPortfolioServiceDetails *RPToolsPortfolioServiceDetails `json:"rp_tools_portfolio_service_details,omitempty"`
	ReportStorageDetails           *ReportStorageDetails           `json:"report_storage_details,omitempty"`
}

/*
------------
Common types
------------
Shared structures across components should be defined here
*/

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
}

type Storage struct {
	StorageSize string `json:"size,omitempty"`
}

/*
-------------------
Polaris DB Services
-------------------
Specification for PolarisDB services goes here
*/

type EventstoreDetails struct {
	Replicas *int32  `json:"replicas,omitempty"`
	Storage  Storage `json:"storage,omitempty"`
}

type MongoDBDetails struct {
	Storage Storage `json:"storage,omitempty"`
}

type SMTPDetails struct {
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Username    string `json:"username,omitempty"`
	Password    string `json:"password,omitempty"`
	SenderEmail string `json:"sender_email,omitempty"`
}

type PostgresDetails struct {
	Host     string  `json:"host"`
	Port     int     `json:"port"`
	Username string  `json:"username"`
	Password string  `json:"password"`
	Storage  Storage `json:"storage,omitempty"`
}

type UploadServerDetails struct {
	Replicas      *int32        `json:"replicas,omitempty"`
	ResourcesSpec ResourcesSpec `json:"resources,omitempty"`
	Storage       Storage       `json:"storage,omitempty"`
}

/*
----------------
Polaris Services
----------------
Specification for Polaris services goes here
*/

type AuthServerDetails struct {
	ImageDetails *ImageDetails `json:"image_details,omitempty"`
}

type ConfigsServiceDetails struct {
	ImageDetails *ImageDetails `json:"image_details,omitempty"`
}

type CosServerDetails struct {
	ImageDetails *ImageDetails `json:"image_details,omitempty"`
}

type DesktopMetricsServiceDetails struct {
	ImageDetails *ImageDetails `json:"image_details,omitempty"`
}

type DownloadServerDetails struct {
	ImageDetails *ImageDetails `json:"image_details,omitempty"`
	Storage      *Storage      `json:"storage,omitempty"`
}

type IssueServerDetails struct {
	ImageDetails *ImageDetails `json:"image_details,omitempty"`
}

type JobsControllerServiceDetails struct {
	ImageDetails *ImageDetails `json:"image_details,omitempty"`
}

type JobsServiceDetails struct {
	ImageDetails *ImageDetails `json:"image_details,omitempty"`
}

type LogsServiceDetails struct {
	ImageDetails *ImageDetails `json:"image_details,omitempty"`
}

type PericlesSwaggerUIDetails struct {
	ImageDetails *ImageDetails `json:"image_details,omitempty"`
}

type TaxonomyServerDeails struct {
	ImageDetails *ImageDetails `json:"image_details,omitempty"`
}

type TDSCodeAnalysisDetails struct {
	ImageDetails *ImageDetails `json:"image_details,omitempty"`
}

type ToolsServiceDetails struct {
	ImageDetails *ImageDetails `json:"image_details,omitempty"`
}

type TriageCommandHandlerDetails struct {
	ImageDetails *ImageDetails `json:"image_details,omitempty"`
}

type TriageQueryDetails struct {
	ImageDetails *ImageDetails `json:"image_details,omitempty"`
}

type VinylServerDetails struct {
	ImageDetails *ImageDetails `json:"image_details,omitempty"`
}

type WebCoreDetails struct {
	ImageDetails *ImageDetails `json:"image_details,omitempty"`
}

type WebHelpDetails struct {
	ImageDetails *ImageDetails `json:"image_details,omitempty"`
}

/*
------------------
Reporting Services
------------------
Specification for Reporting services goes here
*/
type RPFrontendDetails struct {
	ImageDetails *ImageDetails `json:"image_details,omitempty"`
}

type RPIssueManager struct {
	ImageDetails *ImageDetails `json:"image_details,omitempty"`
}

type RPPolarisAgentServiceDetails struct {
	ImageDetails *ImageDetails `json:"image_details,omitempty"`
}

type RPPortfolioServiceDetails struct {
	ImageDetails *ImageDetails `json:"image_details,omitempty"`
}

type RPReportServiceDetails struct {
	ImageDetails *ImageDetails `json:"image_details,omitempty"`
}

type RPSwaggerDocDetails struct {
	ImageDetails *ImageDetails `json:"image_details,omitempty"`
}

type RPToolsPortfolioServiceDetails struct {
	ImageDetails *ImageDetails `json:"image_details,omitempty"`
}

type ReportStorageDetails struct {
	ImageDetails *ImageDetails `json:"image_details,omitempty"`
	Storage      *Storage      `json:"storage,omitempty"`
}
