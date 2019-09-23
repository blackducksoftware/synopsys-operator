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
	Repository       string         `json:"repository,omitempty"`
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
}

type ConfigsServiceDetails struct {
}

type CosServerDetails struct {
}

type DesktopMetricsServiceDetails struct {
}

type DownloadServerDetails struct {
	Storage *Storage `json:"storage,omitempty"`
}

type IssueServerDetails struct {
}

type JobsControllerServiceDetails struct {
}

type JobsServiceDetails struct {
}

type LogsServiceDetails struct {
}

type PericlesSwaggerUIDetails struct {
}

type TaxonomyServerDeails struct {
}

type TDSCodeAnalysisDetails struct {
}

type ToolsServiceDetails struct {
}

type TriageCommandHandlerDetails struct {
}

type TriageQueryDetails struct {
}

type VinylServerDetails struct {
}

type WebCoreDetails struct {
}

type WebHelpDetails struct {
}

/*
------------------
Reporting Services
------------------
Specification for Reporting services goes here
*/
type RPFrontendDetails struct {
}

type RPIssueManager struct {
}

type RPPolarisAgentServiceDetails struct {
}

type RPPortfolioServiceDetails struct {
}

type RPReportServiceDetails struct {
}

type RPSwaggerDocDetails struct {
}

type RPToolsPortfolioServiceDetails struct {
}

type ReportStorageDetails struct {
	Storage *Storage `json:"storage,omitempty"`
}

type ProvisionJob struct {
	Namespace                                    string
	EnvironmentName                              string
	EnvironmentDNS                               string
	ImagePullSecrets                             string
	Repository                                   string
	Version                                      string
	OrganizationProvisionOrganizationDescription string
	OrganizationProvisionOrganizationName        string
	OrganizationProvisionAdminName               string
	OrganizationProvisionAdminUsername           string
	OrganizationProvisionAdminEmail              string
	OrganizationProvisionLicenseSeatCount        string
	OrganizationProvisionLicenseType             string
	OrganizationProvisionResultsStartDate        string
	OrganizationProvisionResultsEndDate          string
	OrganizationProvisionRetentionStartDate      string
	OrganizationProvisionRetentionEndDate        string
}
