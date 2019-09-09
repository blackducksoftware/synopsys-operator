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

package utils

import (
	synopsysv1 "github.com/blackducksoftware/synopsys-operator/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"time"
)

// CreateAlert will create alert in the cluster
func CreateAlert(restClient *rest.RESTClient, obj *synopsysv1.Alert) (*synopsysv1.Alert, error) {
	result := &synopsysv1.Alert{}
	req := restClient.Post().Resource("alerts").Body(obj)

	if len(obj.Namespace) > 0 {
		req = req.Namespace(obj.Namespace)
	}

	err := req.Do().Into(result)
	return result, err
}

// ListAlerts will list all alerts in the cluster
func ListAlerts(restClient *rest.RESTClient, namespace string, opts metav1.ListOptions) (*synopsysv1.AlertList, error) {
	result := &synopsysv1.AlertList{}
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}

	req := restClient.Get().Resource("alerts").VersionedParams(&opts, scheme.ParameterCodec).Timeout(timeout)

	if len(namespace) > 0 {
		req = req.Namespace(namespace)
	}

	err := req.Do().Into(result)
	return result, err
}

// GetAlert will get Alert in the cluster
func GetAlert(restClient *rest.RESTClient, namespace string, name string, options metav1.GetOptions) (*synopsysv1.Alert, error) {
	result := &synopsysv1.Alert{}
	req := restClient.Get().Resource("alerts").Name(name).VersionedParams(&options, scheme.ParameterCodec)

	if len(namespace) > 0 {
		req = req.Namespace(namespace)
	}

	err := req.Do().Into(result)
	return result, err
}

// UpdateAlert will update an Alert in the cluster
func UpdateAlert(restClient *rest.RESTClient, obj *synopsysv1.Alert) (*synopsysv1.Alert, error) {
	result := &synopsysv1.Alert{}
	req := restClient.Put().Resource("alerts").Body(obj).Name(obj.Name)

	if len(obj.Namespace) > 0 {
		req = req.Namespace(obj.Namespace)
	}

	err := req.Do().Into(result)
	return result, err
}

// DeleteAlert will delete Alert in the cluster
func DeleteAlert(restClient *rest.RESTClient, name string, namespace string, options *metav1.DeleteOptions) error {
	req := restClient.Delete().Resource("alerts").Name(name).Body(options)

	if len(namespace) > 0 {
		req = req.Namespace(namespace)
	}
	return req.Do().Error()
}

// CreateBlackduck will create alert in the cluster
func CreateBlackduck(restClient *rest.RESTClient, obj *synopsysv1.Blackduck) (*synopsysv1.Blackduck, error) {
	result := &synopsysv1.Blackduck{}
	req := restClient.Post().Resource("blackducks").Body(obj)

	if len(obj.Namespace) > 0 {
		req = req.Namespace(obj.Namespace)
	}

	err := req.Do().Into(result)
	return result, err
}

// ListBlackduck will list all alerts in the cluster
func ListBlackduck(restClient *rest.RESTClient, namespace string, opts metav1.ListOptions) (*synopsysv1.BlackduckList, error) {
	result := &synopsysv1.BlackduckList{}
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}

	req := restClient.Get().Resource("blackducks").VersionedParams(&opts, scheme.ParameterCodec).Timeout(timeout)

	if len(namespace) > 0 {
		req = req.Namespace(namespace)
	}

	err := req.Do().Into(result)
	return result, err
}

// GetBlackduck will get Alert in the cluster
func GetBlackduck(restClient *rest.RESTClient, namespace string, name string, options metav1.GetOptions) (*synopsysv1.Blackduck, error) {
	result := &synopsysv1.Blackduck{}
	req := restClient.Get().Resource("blackducks").Name(name).VersionedParams(&options, scheme.ParameterCodec)

	if len(namespace) > 0 {
		req = req.Namespace(namespace)
	}

	err := req.Do().Into(result)
	return result, err
}

// UpdateBlackduck will update an Alert in the cluster
func UpdateBlackduck(restClient *rest.RESTClient, obj *synopsysv1.Blackduck) (*synopsysv1.Blackduck, error) {
	result := &synopsysv1.Blackduck{}
	req := restClient.Put().Resource("blackducks").Body(obj).Name(obj.Name)

	if len(obj.Namespace) > 0 {
		req = req.Namespace(obj.Namespace)
	}

	err := req.Do().Into(result)
	return result, err
}

// DeleteBlackduck will delete Alert in the cluster
func DeleteBlackduck(restClient *rest.RESTClient, name string, namespace string, options *metav1.DeleteOptions) error {
	req := restClient.Delete().Resource("blackducks").Name(name).Body(options)

	if len(namespace) > 0 {
		req = req.Namespace(namespace)
	}
	return req.Do().Error()
}

// CreateOpsSight will create alert in the cluster
func CreateOpsSight(restClient *rest.RESTClient, obj *synopsysv1.OpsSight) (*synopsysv1.OpsSight, error) {
	result := &synopsysv1.OpsSight{}
	req := restClient.Post().Resource("opssights").Body(obj)

	if len(obj.Namespace) > 0 {
		req = req.Namespace(obj.Namespace)
	}

	err := req.Do().Into(result)
	return result, err
}

// ListOpsSight will list all alerts in the cluster
func ListOpsSight(restClient *rest.RESTClient, namespace string, opts metav1.ListOptions) (*synopsysv1.OpsSightList, error) {
	result := &synopsysv1.OpsSightList{}
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}

	req := restClient.Get().Resource("opssights").VersionedParams(&opts, scheme.ParameterCodec).Timeout(timeout)

	if len(namespace) > 0 {
		req = req.Namespace(namespace)
	}

	err := req.Do().Into(result)
	return result, err
}

// GetOpsSight will get Alert in the cluster
func GetOpsSight(restClient *rest.RESTClient, namespace string, name string, options metav1.GetOptions) (*synopsysv1.OpsSight, error) {
	result := &synopsysv1.OpsSight{}
	req := restClient.Get().Resource("opssights").Name(name).VersionedParams(&options, scheme.ParameterCodec)

	if len(namespace) > 0 {
		req = req.Namespace(namespace)
	}

	err := req.Do().Into(result)
	return result, err
}

// UpdateOpsSight will update an Alert in the cluster
func UpdateOpsSight(restClient *rest.RESTClient, obj *synopsysv1.OpsSight) (*synopsysv1.OpsSight, error) {
	result := &synopsysv1.OpsSight{}
	req := restClient.Put().Resource("opssights").Body(obj).Name(obj.Name)

	if len(obj.Namespace) > 0 {
		req = req.Namespace(obj.Namespace)
	}

	err := req.Do().Into(result)
	return result, err
}

// DeleteOpsSight will delete Alert in the cluster
func DeleteOpsSight(restClient *rest.RESTClient, name string, namespace string, options *metav1.DeleteOptions) error {
	req := restClient.Delete().Resource("opssights").Name(name).Body(options)

	if len(namespace) > 0 {
		req = req.Namespace(namespace)
	}
	return req.Do().Error()
}

// CreatePolaris will create alert in the cluster
func CreatePolaris(restClient *rest.RESTClient, obj *synopsysv1.Polaris) (*synopsysv1.Polaris, error) {
	result := &synopsysv1.Polaris{}
	req := restClient.Post().Resource("polaris").Body(obj)

	if len(obj.Namespace) > 0 {
		req = req.Namespace(obj.Namespace)
	}

	err := req.Do().Into(result)
	return result, err
}

// ListPolaris will list all alerts in the cluster
func ListPolaris(restClient *rest.RESTClient, namespace string, opts metav1.ListOptions) (*synopsysv1.PolarisList, error) {
	result := &synopsysv1.PolarisList{}
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}

	req := restClient.Get().Resource("polaris").VersionedParams(&opts, scheme.ParameterCodec).Timeout(timeout)

	if len(namespace) > 0 {
		req = req.Namespace(namespace)
	}

	err := req.Do().Into(result)
	return result, err
}

// GetPolaris will get Alert in the cluster
func GetPolaris(restClient *rest.RESTClient, namespace string, name string, options metav1.GetOptions) (*synopsysv1.Polaris, error) {
	result := &synopsysv1.Polaris{}
	req := restClient.Get().Resource("polaris").Name(name).VersionedParams(&options, scheme.ParameterCodec)

	if len(namespace) > 0 {
		req = req.Namespace(namespace)
	}

	err := req.Do().Into(result)
	return result, err
}

// UpdatePolaris will update an Alert in the cluster
func UpdatePolaris(restClient *rest.RESTClient, obj *synopsysv1.Polaris) (*synopsysv1.Polaris, error) {
	result := &synopsysv1.Polaris{}
	req := restClient.Put().Resource("polaris").Body(obj).Name(obj.Name)

	if len(obj.Namespace) > 0 {
		req = req.Namespace(obj.Namespace)
	}

	err := req.Do().Into(result)
	return result, err
}

// DeletePolaris will delete Alert in the cluster
func DeletePolaris(restClient *rest.RESTClient, name string, namespace string, options *metav1.DeleteOptions) error {
	req := restClient.Delete().Resource("polaris").Name(name).Body(options)

	if len(namespace) > 0 {
		req = req.Namespace(namespace)
	}
	return req.Do().Error()
}

// CreatePolarisDB will create alert in the cluster
func CreatePolarisDB(restClient *rest.RESTClient, obj *synopsysv1.PolarisDB) (*synopsysv1.PolarisDB, error) {
	result := &synopsysv1.PolarisDB{}
	req := restClient.Post().Resource("polarisdbs").Body(obj)

	if len(obj.Namespace) > 0 {
		req = req.Namespace(obj.Namespace)
	}

	err := req.Do().Into(result)
	return result, err
}

// ListPolarisDB will list all alerts in the cluster
func ListPolarisDB(restClient *rest.RESTClient, namespace string, opts metav1.ListOptions) (*synopsysv1.PolarisDBList, error) {
	result := &synopsysv1.PolarisDBList{}
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}

	req := restClient.Get().Resource("polarisdbs").VersionedParams(&opts, scheme.ParameterCodec).Timeout(timeout)

	if len(namespace) > 0 {
		req = req.Namespace(namespace)
	}

	err := req.Do().Into(result)
	return result, err
}

// GetPolarisDB will get Alert in the cluster
func GetPolarisDB(restClient *rest.RESTClient, namespace string, name string, options metav1.GetOptions) (*synopsysv1.PolarisDB, error) {
	result := &synopsysv1.PolarisDB{}
	req := restClient.Get().Resource("polarisdbs").Name(name).VersionedParams(&options, scheme.ParameterCodec)

	if len(namespace) > 0 {
		req = req.Namespace(namespace)
	}

	err := req.Do().Into(result)
	return result, err
}

// UpdatePolarisDB will update an Alert in the cluster
func UpdatePolarisDB(restClient *rest.RESTClient, obj *synopsysv1.PolarisDB) (*synopsysv1.PolarisDB, error) {
	result := &synopsysv1.PolarisDB{}
	req := restClient.Put().Resource("polarisdbs").Body(obj).Name(obj.Name)

	if len(obj.Namespace) > 0 {
		req = req.Namespace(obj.Namespace)
	}

	err := req.Do().Into(result)
	return result, err
}

// DeletePolarisDB will delete Alert in the cluster
func DeletePolarisDB(restClient *rest.RESTClient, name string, namespace string, options *metav1.DeleteOptions) error {
	req := restClient.Delete().Resource("polarisdbs").Name(name).Body(options)

	if len(namespace) > 0 {
		req = req.Namespace(namespace)
	}
	return req.Do().Error()
}

// CreateAuthServer will create alert in the cluster
func CreateAuthServer(restClient *rest.RESTClient, obj *synopsysv1.AuthServer) (*synopsysv1.AuthServer, error) {
	result := &synopsysv1.AuthServer{}
	req := restClient.Post().Resource("authservers").Body(obj)

	if len(obj.Namespace) > 0 {
		req = req.Namespace(obj.Namespace)
	}

	err := req.Do().Into(result)
	return result, err
}

// ListAuthServer will list all alerts in the cluster
func ListAuthServer(restClient *rest.RESTClient, namespace string, opts metav1.ListOptions) (*synopsysv1.AuthServerList, error) {
	result := &synopsysv1.AuthServerList{}
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}

	req := restClient.Get().Resource("authservers").VersionedParams(&opts, scheme.ParameterCodec).Timeout(timeout)

	if len(namespace) > 0 {
		req = req.Namespace(namespace)
	}

	err := req.Do().Into(result)
	return result, err
}

// GetAuthServer will get Alert in the cluster
func GetAuthServer(restClient *rest.RESTClient, namespace string, name string, options metav1.GetOptions) (*synopsysv1.AuthServer, error) {
	result := &synopsysv1.AuthServer{}
	req := restClient.Get().Resource("authservers").Name(name).VersionedParams(&options, scheme.ParameterCodec)

	if len(namespace) > 0 {
		req = req.Namespace(namespace)
	}

	err := req.Do().Into(result)
	return result, err
}

// UpdateAuthServer will update an Alert in the cluster
func UpdateAuthServer(restClient *rest.RESTClient, obj *synopsysv1.AuthServer) (*synopsysv1.AuthServer, error) {
	result := &synopsysv1.AuthServer{}
	req := restClient.Put().Resource("authservers").Body(obj).Name(obj.Name)

	if len(obj.Namespace) > 0 {
		req = req.Namespace(obj.Namespace)
	}

	err := req.Do().Into(result)
	return result, err
}

// DeleteAuthServer will delete Alert in the cluster
func DeleteAuthServer(restClient *rest.RESTClient, name string, namespace string, options *metav1.DeleteOptions) error {
	req := restClient.Delete().Resource("authservers").Name(name).Body(options)

	if len(namespace) > 0 {
		req = req.Namespace(namespace)
	}
	return req.Do().Error()
}
