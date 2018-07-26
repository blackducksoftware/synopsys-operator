/*
Copyright (C) 2018 Synopsys, Inc.

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

package api

// ProtoformDefaults defines default values for Protoform.
// These fields need to be named the same as those in
// protoformConfig in order for defaults to be applied
// properly.  A field that exists in ProtoformDefaults
// but does not exist in protoformConfig will be ignored
type ProtoformDefaults struct {
	PerceptorPort                         int
	PerceiverPort                         int
	ScannerPort                           int
	ImageFacadePort                       int
	SkyfirePort                           int
	AnnotationIntervalSeconds             int
	DumpIntervalMinutes                   int
	HubClientTimeoutPerceptorMilliseconds int
	HubClientTimeoutScannerSeconds        int
	HubHost                               string
	HubUser                               string
	HubUserPassword                       string
	HubPort                               int
	DockerUsername                        string
	DockerPasswordOrToken                 string
	ConcurrentScanLimit                   int
	InternalDockerRegistries              []string
	DefaultVersion                        string
	Registry                              string
	ImagePath                             string
	PerceptorImageName                    string
	ScannerImageName                      string
	ImagePerceiverImageName               string
	PodPerceiverImageName                 string
	ImageFacadeImageName                  string
	SkyfireImageName                      string
	PerceptorImageVersion                 string
	ScannerImageVersion                   string
	PerceiverImageVersion                 string
	ImageFacadeImageVersion               string
	SkyfireImageVersion                   string
	LogLevel                              string
	Namespace                             string
	DefaultCPU                            string // Should be passed like: e.g. "300m"
	DefaultMem                            string // Should be passed like: e.g "1300Mi"
	ImagePerceiver                        bool
	PodPerceiver                          bool
	Metrics                               bool
}
