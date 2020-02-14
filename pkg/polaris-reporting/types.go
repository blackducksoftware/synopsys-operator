/*
 * Copyright (C) 2020 Synopsys, Inc.
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

package polarisreporting

/*
 ------------
 Common types
 ------------
 Shared structures across components should be defined here
*/

type SMTPTLSMode string

const (
	SMTPTLSModeDisable         SMTPTLSMode = "disable"
	SMTPTLSModeTryStartTLS     SMTPTLSMode = "try-starttls"
	SMTPTLSModeRequireStartTLS SMTPTLSMode = "require-starttls"
	SMTPTLSModeRequireTLS      SMTPTLSMode = "require-tls"
)

type PostgresSSLMode string

const (
	PostgresSSLModeDisable PostgresSSLMode = "disable"
	//PostgresSSLModeAllow   PostgresSSLMode = "allow"
	// Not supported???
	//PostgresSSLModePrefer  PostgresSSLMode = "prefer"
	PostgresSSLModeRequire PostgresSSLMode = "require"

	// Not supported on-prem yet
	//PostgresSSLModeVerifyCA PostgresSSLMode = "verify-ca"
	//PostgresSSLModeVerifyFull PostgresSSLMode = "verify-full"
)
