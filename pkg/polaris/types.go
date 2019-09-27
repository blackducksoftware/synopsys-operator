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

// Polaris configures all Polaris specifications
type Polaris struct {
	Namespace           string               `json:"namespace,omitempty"`
	EnvironmentDNS      string               `json:"environment_address"`
	ImagePullSecrets    string               `json:"image_pull_secrets"`
	Version             string               `json:"version"`
	EnableReporting     bool                 `json:"enable_reporting"`
	PolarisDBSpec       *PolarisDBSpec       `json:"polaris_db_spec,omitempty"`
	PolarisSpec         *PolarisSpec         `json:"polaris_spec,omitempty"`
	ReportingSpec       *ReportingSpec       `json:"reporting_spec,omitempty"`
	Repository          string               `json:"repository,omitempty"`
	OrganizationDetails *OrganizationDetails `json:"organization_details"`
	Licenses            *Licenses            `json:"licenses"`
}

// PolarisDBSpec configures Polaris DB specifications
type PolarisDBSpec struct {
	SMTPDetails          SMTPDetails         `json:"smtp_details"`
	PostgresInstanceType string              `json:"postgres_instance_type"`
	PostgresDetails      PostgresDetails     `json:"postgres_details"`
	EventstoreDetails    EventstoreDetails   `json:"eventstore_details,omitempty"`
	UploadServerDetails  UploadServerDetails `json:"upload_server_details,omitempty"`
	MongoDBDetails       MongoDBDetails      `json:"mongodb_details,omitempty"`
}

// PolarisSpec configure Polaris Specifications
type PolarisSpec struct {
	AuthServerDetails            AuthServerDetails            `json:"auth_server_details,omitempty"`
	ConfigsServiceDetails        ConfigsServiceDetails        `json:"configs_service_details,omitempty"`
	CosServerDetails             CosServerDetails             `json:"cos_server_details,omitempty"`
	DesktopMetricsServiceDetails DesktopMetricsServiceDetails `json:"desktop_metrics_service_details,omitempty"`
	DownloadServerDetails        DownloadServerDetails        `json:"download_server_details,omitempty"`
	IssueServerDetails           IssueServerDetails           `json:"issue_server_details,omitempty"`
	JobsControllerServiceDetails JobsControllerServiceDetails `json:"jobs_controller_service_details,omitempty"`
	JobsServiceDetails           JobsServiceDetails           `json:"jobs_service_details,omitempty"`
	LogsServiceDetails           LogsServiceDetails           `json:"logs_service_details,omitempty"`
	PericlesSwaggerUIDetails     PericlesSwaggerUIDetails     `json:"pericles_swagger_ui_detail,omitempty"`
	TaxonomyServerDeails         TaxonomyServerDeails         `json:"taxonomy_server_details,omitempty"`
	TDSCodeAnalysisDetails       TDSCodeAnalysisDetails       `json:"tds_code_analysis_details,omitempty"`
	ToolsServiceDetails          ToolsServiceDetails          `json:"tools_service_details,omitempty"`
	TriageCommandHandlerDetails  TriageCommandHandlerDetails  `json:"triage_command_handler,omitempty"`
	TriageQueryDetails           TriageQueryDetails           `json:"triage_query_details,omitempty"`
	VinylServerDetails           VinylServerDetails           `json:"vinyl_server_details,omitempty"`
	WebCoreDetails               WebCoreDetails               `json:"web_core_details,omitempty"`
	WebHelpDetails               WebHelpDetails               `json:"web_help_details,omitempty"`
}

// ReportingSpec configure Polaris reporting Specifications
type ReportingSpec struct {
	RPFrontendDetails              RPFrontendDetails              `json:"rp_frontend_details,omitempty"`
	RPIssueManager                 RPIssueManager                 `json:"rp_issue_manager,omitemtpy"`
	RPPolarisAgentServiceDetails   RPPolarisAgentServiceDetails   `json:"rp_polaris_agent_service_details,omitempty"`
	RPPortfolioServiceDetails      RPPortfolioServiceDetails      `json:"rp_portfolio_service_details,omitempty"`
	RPReportServiceDetails         RPReportServiceDetails         `json:"rp_report_service_details,omitempty"`
	RPSwaggerDocDetails            RPSwaggerDocDetails            `json:"rp_swagger_doc_details,omitempty"`
	RPToolsPortfolioServiceDetails RPToolsPortfolioServiceDetails `json:"rp_tools_portfolio_service_details,omitempty"`
	ReportStorageDetails           ReportStorageDetails           `json:"report_storage_details,omitempty"`
}

/*
------------
Common types
------------
Shared structures across components should be defined here
*/

// ResourcesSpec configures resource specifications
type ResourcesSpec struct {
	RequestsSpec RequestsSpec `json:"requests,omitempty"`
	LimitsSpec   LimitsSpec   `json:"limits,omitempty"`
}

// RequestsSpec configures pod request specifications
type RequestsSpec struct {
	CPU    string `json:"cpu,omitempty"`
	Memory string `json:"memory,omitempty"`
}

// LimitsSpec configures pod limit specifications
type LimitsSpec struct {
	CPU    string `json:"cpu,omitempty"`
	Memory string `json:"memory,omitempty"`
}

// Storage configures volume storage specifications
type Storage struct {
	StorageSize string `json:"size,omitempty"`
}

/*
-------------------
Polaris DB Services
-------------------
Specification for PolarisDB services goes here
*/

// EventstoreDetails configures event store specifications
type EventstoreDetails struct {
	Replicas *int32  `json:"replicas,omitempty"`
	Storage  Storage `json:"storage,omitempty"`
}

// MongoDBDetails configures Mongo DB specifications
type MongoDBDetails struct {
	Storage Storage `json:"storage,omitempty"`
}

// SMTPDetails configures SMTP specifications
type SMTPDetails struct {
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Username    string `json:"username,omitempty"`
	Password    string `json:"password,omitempty"`
	SenderEmail string `json:"sender_email,omitempty"`
}

// PostgresDetails configures postgres details specifications
type PostgresDetails struct {
	Host     string  `json:"host"`
	Port     int     `json:"port"`
	Username string  `json:"username"`
	Password string  `json:"password"`
	Storage  Storage `json:"storage,omitempty"`
}

// UploadServerDetails configures upload server specifications
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

// AuthServerDetails configures authentication server specifications
type AuthServerDetails struct {
}

// ConfigsServiceDetails configures configuration service specifications
type ConfigsServiceDetails struct {
}

// CosServerDetails configures cos server specifications
type CosServerDetails struct {
}

// DesktopMetricsServiceDetails configures desktop metrics specifications
type DesktopMetricsServiceDetails struct {
}

// DownloadServerDetails configures download server specifications
type DownloadServerDetails struct {
	Storage Storage `json:"storage,omitempty"`
}

// IssueServerDetails configures issue server specifications
type IssueServerDetails struct {
}

// JobsControllerServiceDetails configures job controller specifications
type JobsControllerServiceDetails struct {
}

// JobsServiceDetails configures job service specifications
type JobsServiceDetails struct {
}

// LogsServiceDetails configures log service specifications
type LogsServiceDetails struct {
}

// PericlesSwaggerUIDetails configures pericles swagger UI specifications
type PericlesSwaggerUIDetails struct {
}

// TaxonomyServerDeails configures taxonomy server specifications
type TaxonomyServerDeails struct {
}

// TDSCodeAnalysisDetails configures TDS code analysis specifications
type TDSCodeAnalysisDetails struct {
}

// ToolsServiceDetails configures tools service specifications
type ToolsServiceDetails struct {
}

// TriageCommandHandlerDetails configures triage command handler specifications
type TriageCommandHandlerDetails struct {
}

// TriageQueryDetails configures triage query specifications
type TriageQueryDetails struct {
}

// VinylServerDetails configures vinyl server specifications
type VinylServerDetails struct {
}

// WebCoreDetails configures web core specifications
type WebCoreDetails struct {
}

// WebHelpDetails configures web help specifications
type WebHelpDetails struct {
}

/*
------------------
Reporting Services
------------------
Specification for Reporting services goes here
*/

// RPFrontendDetails configures report front end specifications
type RPFrontendDetails struct {
}

// RPIssueManager configures report issue manager specifications
type RPIssueManager struct {
}

// RPPolarisAgentServiceDetails configures report polaris agent service specifications
type RPPolarisAgentServiceDetails struct {
}

// RPPortfolioServiceDetails configures report portfolio specifications
type RPPortfolioServiceDetails struct {
}

// RPReportServiceDetails configures report service specifications
type RPReportServiceDetails struct {
}

// RPSwaggerDocDetails configures report swagger specifications
type RPSwaggerDocDetails struct {
}

// RPToolsPortfolioServiceDetails configures report tools portfolio service specifications
type RPToolsPortfolioServiceDetails struct {
}

// ReportStorageDetails configures report storage specifications
type ReportStorageDetails struct {
	Storage Storage `json:"storage,omitempty"`
}

// OrganizationDetails configures organization details specifications
type OrganizationDetails struct {
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

// Licenses configures license specifications
type Licenses struct {
	Coverity string
}
