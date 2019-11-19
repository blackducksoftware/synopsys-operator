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

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// CRSpecBuilderFromCobraFlags uses Cobra commands, Cobra flags and other
// values to create an BDBA CR's Spec.
//
// The fields in the CRSpecBuilderFromCobraFlags represent places where the values of the Cobra flags are stored.
//
// Usage: Use CRSpecBuilderFromCobraFlags to add flags to your Cobra Command for making an BDBA Spec.
// When flags are used the correspoding value in this struct will by set. You can then
// generate the spec by telling CRSpecBuilderFromCobraFlags what flags were changed.
type CRSpecBuilderFromCobraFlags struct {
	spec    BDBA
	Version string

	Hostname    string
	IngressHost string

	MinioAccessKey string
	MinioSecretKey string

	WorkerReplicas int

	AdminEmail string

	BrokerURL string

	PGPPassword string

	RabbitMQULimitNoFiles string

	HideLicenses      string
	LicensingPassword string
	LicensingUsername string

	InsecureCookies  string
	SessionCookieAge string

	URL       string
	Actual    string
	Expected  string
	StartFlag string
	Result    string
}

// NewCRSpecBuilderFromCobraFlags creates a new CRSpecBuilderFromCobraFlags type
func NewCRSpecBuilderFromCobraFlags() *CRSpecBuilderFromCobraFlags {
	return &CRSpecBuilderFromCobraFlags{
		spec: BDBA{},
	}
}

// GetCRSpec returns a pointer to the BDBASpec as an interface{}
func (ctl *CRSpecBuilderFromCobraFlags) GetCRSpec() interface{} {
	return ctl.spec
}

// SetCRSpec sets the BDBASpec in the struct
func (ctl *CRSpecBuilderFromCobraFlags) SetCRSpec(spec interface{}) error {
	convertedSpec, ok := spec.(BDBA)
	if !ok {
		return fmt.Errorf("error setting BDBA spec")
	}

	ctl.spec = convertedSpec
	return nil
}

// SetPredefinedCRSpec sets the Spec to a predefined spec
func (ctl *CRSpecBuilderFromCobraFlags) SetPredefinedCRSpec(specType string) error {
	ctl.spec = *GetBDBADefault()
	return nil
}

// AddCRSpecFlagsToCommand adds flags to a Cobra Command that are need for Spec.
// The flags map to fields in the CRSpecBuilderFromCobraFlags struct.
// master - if false, doesn't add flags that all Users shouldn't use
func (ctl *CRSpecBuilderFromCobraFlags) AddCRSpecFlagsToCommand(cmd *cobra.Command, master bool) {
	cmd.Flags().StringVar(&ctl.Version, "version", ctl.Version, "Version of BDBA")

	cmd.Flags().StringVar(&ctl.Hostname, "hostname", ctl.Hostname, "TODO - describe flag")

	cmd.Flags().StringVar(&ctl.IngressHost, "ingress-host", ctl.IngressHost, "TODO - describe flag")

	cmd.Flags().StringVar(&ctl.MinioAccessKey, "minio-acceskey", ctl.MinioAccessKey, "TODO - describe flag")
	cmd.Flags().StringVar(&ctl.MinioSecretKey, "minio-secret-key", ctl.MinioSecretKey, "TODO - describe flag")

	cmd.Flags().IntVar(&ctl.WorkerReplicas, "worker-replicas", ctl.WorkerReplicas, "TODO - describe flag")

	cmd.Flags().StringVar(&ctl.AdminEmail, "admin-email", ctl.AdminEmail, "TODO - describe flag")

	cmd.Flags().StringVar(&ctl.BrokerURL, "broker-url", ctl.BrokerURL, "TODO - describe flag")

	cmd.Flags().StringVar(&ctl.PGPPassword, "pgp-password", ctl.PGPPassword, "TODO - describe flag")

	cmd.Flags().StringVar(&ctl.RabbitMQULimitNoFiles, "rabbitmq-ulimit-no-files", ctl.RabbitMQULimitNoFiles, "TODO - describe flag")

	cmd.Flags().StringVar(&ctl.HideLicenses, "hide-licenses", ctl.HideLicenses, "TODO - describe flag")
	cmd.Flags().StringVar(&ctl.LicensingPassword, "licensing-password", ctl.LicensingPassword, "TODO - describe flag")
	cmd.Flags().StringVar(&ctl.LicensingUsername, "licensing-username", ctl.LicensingUsername, "TODO - describe flag")

	cmd.Flags().StringVar(&ctl.InsecureCookies, "insecure-cookies", ctl.InsecureCookies, "TODO - describe flag")
	cmd.Flags().StringVar(&ctl.SessionCookieAge, "session-cookie-age", ctl.SessionCookieAge, "TODO - describe flag")

	cmd.Flags().StringVar(&ctl.URL, "url", ctl.URL, "TODO - describe flag")
	cmd.Flags().StringVar(&ctl.Actual, "actual", ctl.Actual, "TODO - describe flag")
	cmd.Flags().StringVar(&ctl.Expected, "expected", ctl.Expected, "TODO - describe flag")
	cmd.Flags().StringVar(&ctl.StartFlag, "startFlag", ctl.StartFlag, "TODO - describe flag")
	cmd.Flags().StringVar(&ctl.Result, "result", ctl.Result, "TODO - describe flag")
}

// CheckValuesFromFlags returns an error if a value stored in the struct will not be able to be
// used in the spec
func (ctl *CRSpecBuilderFromCobraFlags) CheckValuesFromFlags(flagset *pflag.FlagSet) error {
	return nil
}

// GenerateCRSpecFromFlags checks if a flag was changed and updates the spec with the value that's stored
// in the corresponding struct field
func (ctl *CRSpecBuilderFromCobraFlags) GenerateCRSpecFromFlags(flagset *pflag.FlagSet) (interface{}, error) {
	err := ctl.CheckValuesFromFlags(flagset)
	if err != nil {
		return nil, err
	}
	flagset.VisitAll(ctl.SetCRSpecFieldByFlag)
	return ctl.spec, nil
}

// SetCRSpecFieldByFlag updates a field in the spec if the flag was set by the user. It gets the
// value from the corresponding struct field
func (ctl *CRSpecBuilderFromCobraFlags) SetCRSpecFieldByFlag(f *pflag.Flag) {
	if f.Changed {
		log.Debugf("flag '%s': CHANGED", f.Name)
		switch f.Name {
		case "version":
			ctl.spec.Version = ctl.Version
		case "hostname":
			ctl.spec.Hostname = ctl.Hostname
		case "ingress-host":
			ctl.spec.IngressHost = ctl.IngressHost
		case "minio-acceskey":
			ctl.spec.MinioAccessKey = ctl.MinioAccessKey
		case "minio-secret-key":
			ctl.spec.MinioSecretKey = ctl.MinioSecretKey
		case "worker-replicas":
			ctl.spec.WorkerReplicas = ctl.WorkerReplicas
		case "admin-email":
			ctl.spec.AdminEmail = ctl.AdminEmail
		case "broker-url":
			ctl.spec.BrokerURL = ctl.BrokerURL
		case "pgp-password":
			ctl.spec.PGPPassword = ctl.PGPPassword
		case "rabbitmq-ulimit-no-files":
			ctl.spec.RabbitMQULimitNoFiles = ctl.RabbitMQULimitNoFiles
		case "hide-licenses":
			ctl.spec.HideLicenses = ctl.HideLicenses
		case "licensing-password":
			ctl.spec.LicensingPassword = ctl.LicensingPassword
		case "licensing-username":
			ctl.spec.LicensingUsername = ctl.LicensingUsername
		case "insecure-cookies":
			ctl.spec.InsecureCookies = ctl.InsecureCookies
		case "session-cookie-age":
			ctl.spec.SessionCookieAge = ctl.SessionCookieAge
		case "url":
			ctl.spec.URL = ctl.URL
		case "actual":
			ctl.spec.Actual = ctl.Actual
		case "expected":
			ctl.spec.Expected = ctl.Expected
		case "startFlag":
			ctl.spec.StartFlag = ctl.StartFlag
		case "result":
			ctl.spec.Result = ctl.Result
		default:
			log.Debugf("flag '%s': NOT FOUND", f.Name)
		}
	} else {
		log.Debugf("flag '%s': UNCHANGED", f.Name)
	}
}
