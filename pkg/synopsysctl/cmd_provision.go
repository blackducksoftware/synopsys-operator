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

package synopsysctl

import (
	"fmt"
	"github.com/blackducksoftware/synopsys-operator/pkg/polaris"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/spf13/cobra"
	v12 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type polarisProvisionJob struct {
	coverityLicensePath                          string
	organizationProvisionOrganizationDescription string
	organizationProvisionOrganizationName        string
	organizationProvisionAdminName               string
	organizationProvisionAdminUsername           string
	organizationProvisionAdminEmail              string
}

var ctlProvisionJob polarisProvisionJob

var provisionCmd = &cobra.Command{
	Use:   "provision",
	Short: "Provision a Polaris instance",
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("must specify a sub-command")
	},
}

var provisionPolarisCmd = &cobra.Command{
	Use:           "polaris",
	Example:       "synopsysctl provision polaris -n <namespace>",
	Aliases:       []string{"polaris"},
	Short:         "Provision polaris",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 0 {
			cmd.Help()
			return fmt.Errorf("this command takes 0 arguments")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		p, err := getPolarisFromSecret()
		if err != nil {
			return err
		}

		jobConfig := polaris.ProvisionJob{
			Namespace:        p.Namespace,
			EnvironmentName:  p.EnvironmentName,
			EnvironmentDNS:   p.EnvironmentDNS,
			ImagePullSecrets: p.ImagePullSecrets,
			Repository:       p.Repository,
			Version:          p.Version,
			OrganizationProvisionOrganizationDescription: ctlProvisionJob.organizationProvisionOrganizationDescription,
			OrganizationProvisionOrganizationName:        ctlProvisionJob.organizationProvisionOrganizationName,
			OrganizationProvisionAdminName:               ctlProvisionJob.organizationProvisionAdminName,
			OrganizationProvisionAdminUsername:           ctlProvisionJob.organizationProvisionAdminUsername,
			OrganizationProvisionAdminEmail:              ctlProvisionJob.organizationProvisionAdminEmail,
			OrganizationProvisionLicenseSeatCount:        "100",
			OrganizationProvisionLicenseType:             "internal",
			OrganizationProvisionResultsStartDate:        "",
			OrganizationProvisionResultsEndDate:          "",
			OrganizationProvisionRetentionStartDate:      "",
			OrganizationProvisionRetentionEndDate:        "",
		}

		license, err := util.ReadFileData(ctlProvisionJob.coverityLicensePath)
		if err != nil {
			return err
		}

		if _, err := kubeClient.CoreV1().Secrets(namespace).Create(&v12.Secret{
			ObjectMeta: v1.ObjectMeta{
				Name:      "coverity-licensed",
				Namespace: namespace,
			},
			Data: map[string][]byte{
				"license": []byte(license),
			},
			Type: v12.SecretTypeOpaque,
		}); err != nil {
			return err
		}

		job, err := polaris.GetPolarisProvisionJob(baseUrl, jobConfig)
		if err != nil {
			return err
		}
		if _, err := kubeClient.BatchV1().Jobs(namespace).Create(job); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(provisionCmd)

	provisionPolarisCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", namespace, "Namespace of the instance(s)")
	provisionPolarisCmd.PersistentFlags().StringVarP(&ctlProvisionJob.organizationProvisionOrganizationDescription, "organization-description", "", ctlProvisionJob.organizationProvisionOrganizationDescription, "Organization description")
	provisionPolarisCmd.PersistentFlags().StringVarP(&ctlProvisionJob.organizationProvisionOrganizationName, "organization-name", "", ctlProvisionJob.organizationProvisionOrganizationName, "Organization name")
	provisionPolarisCmd.PersistentFlags().StringVarP(&ctlProvisionJob.organizationProvisionAdminName, "organization-admin-name", "", ctlProvisionJob.organizationProvisionAdminName, "Organization admin name")
	provisionPolarisCmd.PersistentFlags().StringVarP(&ctlProvisionJob.organizationProvisionAdminUsername, "organization-admin-username", "", ctlProvisionJob.organizationProvisionAdminUsername, "Organization admin username")
	provisionPolarisCmd.PersistentFlags().StringVarP(&ctlProvisionJob.organizationProvisionAdminEmail, "organization-admin-email", "", ctlProvisionJob.organizationProvisionAdminEmail, "Organization admin username")
	provisionPolarisCmd.PersistentFlags().StringVarP(&ctlProvisionJob.coverityLicensePath, "coverity-license-path", "", ctlProvisionJob.coverityLicensePath, "Path to the coverity license")
	cobra.MarkFlagRequired(provisionPolarisCmd.PersistentFlags(), "namespace")
	cobra.MarkFlagRequired(provisionPolarisCmd.PersistentFlags(), "organization-description")
	cobra.MarkFlagRequired(provisionPolarisCmd.PersistentFlags(), "organization-name")
	cobra.MarkFlagRequired(provisionPolarisCmd.PersistentFlags(), "organization-admin-name")
	cobra.MarkFlagRequired(provisionPolarisCmd.PersistentFlags(), "organization-admin-username")
	cobra.MarkFlagRequired(provisionPolarisCmd.PersistentFlags(), "organization-admin-email")
	cobra.MarkFlagRequired(provisionPolarisCmd.PersistentFlags(), "coverity-license-path")

	addBaseUrlFlag(provisionPolarisCmd)

	provisionCmd.AddCommand(provisionPolarisCmd)

}
