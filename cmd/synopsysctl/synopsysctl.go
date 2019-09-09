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

package main

import (
	synopsysv1 "github.com/blackducksoftware/synopsys-operator/meta-builder/api/v1"
	"github.com/blackducksoftware/synopsys-operator/meta-builder/synopsysctl"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var version string

func main() {
	synopsysv1.AddToScheme(scheme.Scheme)

	//

	log.Debugf("version: %s", version)
	synopsysctl.Execute(version)
}
