/*
Copyright (C) 2019 Synopsys, Inc.

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

package synopsysctl

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// Flags for using mock mode - doesn't deploy
var mockFormat string
var nativeFormat = "json"

const (
	clusterTypeKubernetes = "KUBERNETES"
	clusterTypeOpenshift  = "OPENSHIFT"
)

var nativeClusterType = clusterTypeKubernetes

func verifyClusterType(cType string) error {
	if strings.EqualFold(strings.ToUpper(cType), clusterTypeKubernetes) || strings.EqualFold(strings.ToUpper(cType), clusterTypeOpenshift) {
		return nil
	}
	return fmt.Errorf("invalid cluster type '%s'", cType)
}

func addNativeFormatFlag(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&nativeFormat, "format", "f", nativeFormat, "Output format (json|yaml)")
	cmd.Flags().StringVar(&nativeClusterType, "target", nativeClusterType, "Type of cluster to generate the resources for (KUBERNETES|OPENSHIFT)")
}

func addMockFlag(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&mockFormat, "mock", "o", mockFormat, "Print the resource spec in the specified format instead of creating it (json|yaml)")
}
