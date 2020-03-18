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

// SMTPTLSMode configures the SMTP TLS mode
type SMTPTLSMode string

const (
	// SMTPTLSModeDisable denotes disabled SMTP TLS mode
	SMTPTLSModeDisable SMTPTLSMode = "disable"
	// SMTPTLSModeTryStartTLS denotes try-starttls SMTP TLS mode
	SMTPTLSModeTryStartTLS SMTPTLSMode = "try-starttls"
	// SMTPTLSModeRequireStartTLS denotes require-starttls SMTP TLS mode
	SMTPTLSModeRequireStartTLS SMTPTLSMode = "require-starttls"
	// SMTPTLSModeRequireTLS denotes require-tls SMTP TLS mode
	SMTPTLSModeRequireTLS SMTPTLSMode = "require-tls"
)

// SMTPDetails configures SMTP specifications
type SMTPDetails struct {
	Host                   string      `json:"host"`
	Port                   int         `json:"port"`
	Username               string      `json:"username,omitempty"`
	Password               string      `json:"password,omitempty"`
	SenderEmail            string      `json:"sender_email,omitempty"`
	TLSMode                SMTPTLSMode `json:"tlsMode"`
	TLSCheckServerIdentity bool        `json:"tlsCheckServerIdentity"`
	TLSTrustedHosts        string      `json:"tlsTrustedHosts"`
}

// PostgresSSLMode configures the Postgres SSL mode
type PostgresSSLMode string

const (
	// PostgresSSLModeDisable denotes disabled postgres SSL mode
	PostgresSSLModeDisable PostgresSSLMode = "disable"
	//PostgresSSLModeAllow   PostgresSSLMode = "allow"
	// Not supported???
	//PostgresSSLModePrefer  PostgresSSLMode = "prefer"

	// PostgresSSLModeRequire denotes require postgres SSL mode
	PostgresSSLModeRequire PostgresSSLMode = "require"

	// Not supported on-prem yet
	//PostgresSSLModeVerifyCA PostgresSSLMode = "verify-ca"
	//PostgresSSLModeVerifyFull PostgresSSLMode = "verify-full"
)
