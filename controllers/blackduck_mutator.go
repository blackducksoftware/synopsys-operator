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

package controllers

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/go-logr/logr"

	synopsysv1 "github.com/blackducksoftware/synopsys-operator/api/v1"
	controllers_utils "github.com/blackducksoftware/synopsys-operator/controllers/util"
)

func patchBlackduck(client client.Client, blackDuckCr *synopsysv1.Blackduck, mapOfUniqueIdToBaseRuntimeObject map[string]runtime.Object, isDryRun bool, log logr.Logger, isOpenShift bool) map[string]runtime.Object {
	patcher := BlackduckPatcher{
		Client:                           client,
		blackDuckCr:                      blackDuckCr,
		mapOfUniqueIdToBaseRuntimeObject: mapOfUniqueIdToBaseRuntimeObject,
		isDryRun:                         isDryRun,
		log:                              log,
		isOpenShift:                      isOpenShift,
	}
	return patcher.patch()
}

type BlackduckPatcher struct {
	client.Client
	blackDuckCr                      *synopsysv1.Blackduck
	mapOfUniqueIdToBaseRuntimeObject map[string]runtime.Object
	isDryRun                         bool
	log                              logr.Logger
	isOpenShift                      bool
}

func (p *BlackduckPatcher) patch() map[string]runtime.Object {
	// TODO JD: Patching this way is costly. Consider iterating over the mapOfUniqueIdToBaseRuntimeObject only once
	// and apply the necessary changes

	patches := []func() error{
		p.patchNamespace,
		p.patchStorage,
		p.patchLiveness,
		p.patchEnvirons,
		p.patchWebserverCertificates,
		p.patchPostgresConfig,
		p.patchImages,
		p.patchAuthCert,
		p.patchProxyCert,
		p.patchExposeService,
		p.patchBDBA,
		p.patchSealKey,
		p.patchWithSize,
		p.patchReplicas,
		p.patchOpenshift,
	}
	for _, f := range patches {
		err := f()
		if err != nil {
			fmt.Printf("%s\n", err)
		}
	}

	return p.mapOfUniqueIdToBaseRuntimeObject
}

func (p *BlackduckPatcher) patchOpenshift() error {
	// TODO uncomment once we pass protoform config
	//if p.config.IsOpenshift {
	//	for _, v := range p.mapOfUniqueIdToBaseRuntimeObject {
	//		switch v.(type) {
	//		case *corev1.ReplicationController:
	//			for i := range v.(*corev1.ReplicationController).Spec.Template.Spec.Containers {
	//				v.(*corev1.ReplicationController).Spec.Template.Spec.SecurityContext.FSGroup = nil
	//			}
	//		}
	//	}
	//}
	return nil
}

func (p *BlackduckPatcher) patchBDBA() error {
	for _, e := range p.blackDuckCr.Spec.Environs {
		vals := strings.Split(e, ":")
		if len(vals) != 2 {
			continue
		}
		if strings.Compare(vals[0], "USE_BINARY_UPLOADS") == 0 {
			if strings.Compare(vals[1], "1") != 0 {
				delete(p.mapOfUniqueIdToBaseRuntimeObject, fmt.Sprintf("ReplicationController.%s-blackduck-rabbitmq", p.blackDuckCr.Name))
				delete(p.mapOfUniqueIdToBaseRuntimeObject, fmt.Sprintf("Service.%s-blackduck-rabbitmq", p.blackDuckCr.Name))
				delete(p.mapOfUniqueIdToBaseRuntimeObject, fmt.Sprintf("ReplicationController.%s-blackduck-binaryscanner", p.blackDuckCr.Name))
			}
			break
		}
	}
	return nil
}

func (p *BlackduckPatcher) patchSealKey() error {
	if p.isDryRun {
		return nil
	}

	var secret corev1.Secret
	if err := p.Client.Get(context.TODO(), types.NamespacedName{
		Namespace: "synopsys-operator", // <<< TODO Get this from protoform
		Name:      "blackduck-secret",
	}, &secret); err != nil {
		return err
	}

	sealKey, ok := secret.Data["SEAL_KEY"]
	if !ok {
		return fmt.Errorf("SEAL_KEY key couldn't be found inside blackduck-secret")
	}

	runtimeObject, ok := p.mapOfUniqueIdToBaseRuntimeObject[fmt.Sprintf("Secret.%s-blackduck-upload-cache", p.blackDuckCr.Name)]
	if !ok {
		return nil
	}

	runtimeObject.(*corev1.Secret).Data["SEAL_KEY"] = sealKey
	return nil
}

// TODO: common with Alert
func (p *BlackduckPatcher) patchExposeService() error {

	// TODO use contansts
	routeID := fmt.Sprintf("Route.%s-blackduck-webserver-exposed", p.blackDuckCr.Name)
	serviceID := fmt.Sprintf("Service.%s-blackduck-webserver-exposed", p.blackDuckCr.Name)
	serviceRuntimeObject, ok := p.mapOfUniqueIdToBaseRuntimeObject[serviceID]
	if !ok {
		return nil
	}

	switch strings.ToUpper(p.blackDuckCr.Spec.ExposeService) {
	case "LOADBALANCER":
		serviceRuntimeObject.(*corev1.Service).Spec.Type = corev1.ServiceTypeLoadBalancer
		delete(p.mapOfUniqueIdToBaseRuntimeObject, routeID)
	case "NODEPORT":
		serviceRuntimeObject.(*corev1.Service).Spec.Type = corev1.ServiceTypeNodePort
		delete(p.mapOfUniqueIdToBaseRuntimeObject, routeID)
	case "OPENSHIFT":
		delete(p.mapOfUniqueIdToBaseRuntimeObject, serviceID)
		if !p.isOpenShift {
			p.log.Error(fmt.Errorf("cluster is not Openshift"), "removing route runtime object")
			delete(p.mapOfUniqueIdToBaseRuntimeObject, routeID)
		}
	default:
		delete(p.mapOfUniqueIdToBaseRuntimeObject, serviceID)
		delete(p.mapOfUniqueIdToBaseRuntimeObject, routeID)
	}

	return nil
}

func (p *BlackduckPatcher) patchAuthCert() error {
	if len(p.blackDuckCr.Spec.AuthCustomCA) == 0 {
		for _, v := range p.mapOfUniqueIdToBaseRuntimeObject {
			switch v.(type) {
			case *corev1.ReplicationController:
				removeVolumeAndVolumeMountFromRC(v.(*corev1.ReplicationController), fmt.Sprintf("%s-blackduck-auth-custom-ca", p.blackDuckCr.Name))
			}
		}
	} else {
		secret, ok := p.mapOfUniqueIdToBaseRuntimeObject[fmt.Sprintf("Secret.%s-blackduck-auth-custom-ca", p.blackDuckCr.Name)]
		if !ok {
			return nil
		}

		if secret.(*corev1.Secret).Data == nil {
			secret.(*corev1.Secret).Data = make(map[string][]byte)
		}

		secret.(*corev1.Secret).Data["AUTH_CUSTOM_CA"] = []byte(p.blackDuckCr.Spec.AuthCustomCA)
	}
	return nil
}

func (p *BlackduckPatcher) patchProxyCert() error {
	if len(p.blackDuckCr.Spec.ProxyCertificate) == 0 {
		for _, v := range p.mapOfUniqueIdToBaseRuntimeObject {
			switch v.(type) {
			case *corev1.ReplicationController:
				removeVolumeAndVolumeMountFromRC(v.(*corev1.ReplicationController), fmt.Sprintf("%s-blackduck-proxy-certificate", p.blackDuckCr.Name))
			}
		}
	} else {
		secret, ok := p.mapOfUniqueIdToBaseRuntimeObject[fmt.Sprintf("Secret.%s-blackduck-proxy-certificate", p.blackDuckCr.Name)]
		if !ok {
			return nil
		}

		if secret.(*corev1.Secret).Data == nil {
			secret.(*corev1.Secret).Data = make(map[string][]byte)
		}

		secret.(*corev1.Secret).Data["HUB_PROXY_CERT_FILE"] = []byte(p.blackDuckCr.Spec.ProxyCertificate)
	}
	return nil
}

// TODO: common with Alert
func (p *BlackduckPatcher) patchWithSize() error {
	if p.isDryRun {
		return nil
	}
	var size synopsysv1.Size
	if len(p.blackDuckCr.Spec.Size) > 0 {
		if err := p.Client.Get(context.TODO(), types.NamespacedName{
			Namespace: p.blackDuckCr.Namespace,
			Name:      strings.ToLower(p.blackDuckCr.Spec.Size),
		}, &size); err != nil {

			if !apierrs.IsNotFound(err) {
				return err
			}
			if apierrs.IsNotFound(err) {
				return fmt.Errorf("blackduck instance [%s] is configured to use a Size [%s] that doesn't exist", p.blackDuckCr.Namespace, p.blackDuckCr.Spec.Size)
			}
		}

		for _, v := range p.mapOfUniqueIdToBaseRuntimeObject {
			switch v.(type) {
			case *corev1.ReplicationController:
				componentName, ok := v.(*corev1.ReplicationController).GetLabels()["component"]
				if !ok {
					return fmt.Errorf("component name is missing in %s", v.(*corev1.ReplicationController).Name)
				}

				sizeConf, ok := size.Spec.PodResources[componentName]
				if !ok {
					return fmt.Errorf("blackDuckCr instance [%s] is configured to use a Size [%s] but the size doesn't contain an entry for [%s]", p.blackDuckCr.Namespace, p.blackDuckCr.Spec.Size, v.(*corev1.ReplicationController).Name)
				}
				v.(*corev1.ReplicationController).Spec.Replicas = func(i int) *int32 { j := int32(i); return &j }(sizeConf.Replica)
				for containerIndex, container := range v.(*corev1.ReplicationController).Spec.Template.Spec.Containers {
					containerConf, ok := sizeConf.ContainerLimit[container.Name]
					if !ok {
						return fmt.Errorf("blackDuckCr instance [%s] is configured to use a Size [%s]. The size oesn't contain an entry for pod [%s] container [%s]", p.blackDuckCr.Namespace, p.blackDuckCr.Spec.Size, v.(*corev1.ReplicationController).Name, container.Name)
					}
					resourceRequirements, err := controllers_utils.GenResourceRequirementsFromContainerSize(containerConf)
					if err != nil {
						return err
					}
					v.(*corev1.ReplicationController).Spec.Template.Spec.Containers[containerIndex].Resources = *resourceRequirements

					for envIndex, env := range v.(*corev1.ReplicationController).Spec.Template.Spec.Containers[containerIndex].Env {
						if strings.Compare(env.Name, "HUB_MAX_MEMORY") == 0 {
							v.(*corev1.ReplicationController).Spec.Template.Spec.Containers[containerIndex].Env[envIndex].Value = fmt.Sprintf("%dm", *containerConf.MaxMem-512)
							break
						}
					}
				}

			}
		}
	}
	return nil
}

// TODO: common with Alert
func (p *BlackduckPatcher) patchReplicas() error {
	for _, v := range p.mapOfUniqueIdToBaseRuntimeObject {
		switch v.(type) {
		case *corev1.ReplicationController:
			switch strings.ToUpper(p.blackDuckCr.Spec.DesiredState) {
			case "STOP":
				v.(*corev1.ReplicationController).Spec.Replicas = func(i int32) *int32 { return &i }(0)
			case "DBMIGRATE":
				if value, ok := v.(*corev1.ReplicationController).GetLabels()["component"]; !ok || strings.Compare(value, "postgres") != 0 {
					v.(*corev1.ReplicationController).Spec.Replicas = func(i int32) *int32 { return &i }(0)
				}
			}
		}
	}
	return nil
}

// TODO: common with Alert
func (p *BlackduckPatcher) patchImages() error {
	if p.blackDuckCr.Spec.RegistryConfiguration != nil && (len(p.blackDuckCr.Spec.RegistryConfiguration.Registry) > 0 || len(p.blackDuckCr.Spec.ImageRegistries) > 0) {
		for _, v := range p.mapOfUniqueIdToBaseRuntimeObject {
			switch v.(type) {
			case *corev1.ReplicationController:
				for i := range v.(*corev1.ReplicationController).Spec.Template.Spec.Containers {
					v.(*corev1.ReplicationController).Spec.Template.Spec.Containers[i].Image = controllers_utils.GenerateImageTag(v.(*corev1.ReplicationController).Spec.Template.Spec.Containers[i].Image, p.blackDuckCr.Spec.ImageRegistries, *p.blackDuckCr.Spec.RegistryConfiguration)
				}
			}
		}
	}
	return nil
}

func (p *BlackduckPatcher) patchPostgresConfig() error {
	cmConf, ok := p.mapOfUniqueIdToBaseRuntimeObject[fmt.Sprintf("ConfigMap.%s-blackduck-db-config", p.blackDuckCr.Name)]
	if !ok {
		return nil
	}

	secretConf, ok := p.mapOfUniqueIdToBaseRuntimeObject[fmt.Sprintf("Secret.%s-blackduck-db-creds", p.blackDuckCr.Name)]
	if !ok {
		return nil
	}

	if cmConf.(*corev1.ConfigMap).Data == nil {
		cmConf.(*corev1.ConfigMap).Data = make(map[string]string)
	}

	if secretConf.(*corev1.Secret).Data == nil {
		secretConf.(*corev1.Secret).Data = make(map[string][]byte)
	}

	if p.blackDuckCr.Spec.ExternalPostgres != nil {
		cmConf.(*corev1.ConfigMap).Data["HUB_POSTGRES_ADMIN"] = p.blackDuckCr.Spec.ExternalPostgres.PostgresAdmin
		cmConf.(*corev1.ConfigMap).Data["HUB_POSTGRES_ENABLE_SSL"] = strconv.FormatBool(p.blackDuckCr.Spec.ExternalPostgres.PostgresSsl)
		cmConf.(*corev1.ConfigMap).Data["HUB_POSTGRES_HOST"] = p.blackDuckCr.Spec.ExternalPostgres.PostgresHost
		cmConf.(*corev1.ConfigMap).Data["HUB_POSTGRES_PORT"] = strconv.Itoa(p.blackDuckCr.Spec.ExternalPostgres.PostgresPort)
		cmConf.(*corev1.ConfigMap).Data["HUB_POSTGRES_USER"] = p.blackDuckCr.Spec.ExternalPostgres.PostgresUser

		secretConf.(*corev1.Secret).Data["HUB_POSTGRES_ADMIN_PASSWORD_FILE"] = []byte(p.blackDuckCr.Spec.ExternalPostgres.PostgresAdminPassword)
		secretConf.(*corev1.Secret).Data["HUB_POSTGRES_USER_PASSWORD_FILE"] = []byte(p.blackDuckCr.Spec.ExternalPostgres.PostgresUserPassword)

		// Delete the component required when deploying internal postgres
		delete(p.mapOfUniqueIdToBaseRuntimeObject, fmt.Sprintf("PersistentVolumeClaim.%s-blackduck-postgres", p.blackDuckCr.Name))
		delete(p.mapOfUniqueIdToBaseRuntimeObject, fmt.Sprintf("Job.%s-blackduck-init-postgres", p.blackDuckCr.Name))
		delete(p.mapOfUniqueIdToBaseRuntimeObject, fmt.Sprintf("ConfigMap.%s-blackduck-postgres-init-config", p.blackDuckCr.Name))
		delete(p.mapOfUniqueIdToBaseRuntimeObject, fmt.Sprintf("Service.%s-blackduck-postgres", p.blackDuckCr.Name))
		delete(p.mapOfUniqueIdToBaseRuntimeObject, fmt.Sprintf("ReplicationController.%s-blackduck-postgres", p.blackDuckCr.Name))
	} else {
		cmConf.(*corev1.ConfigMap).Data["HUB_POSTGRES_ADMIN"] = "blackduck"
		cmConf.(*corev1.ConfigMap).Data["HUB_POSTGRES_ENABLE_SSL"] = "false"
		cmConf.(*corev1.ConfigMap).Data["HUB_POSTGRES_HOST"] = fmt.Sprintf("%s-blackduck-postgres", p.blackDuckCr.Name)
		cmConf.(*corev1.ConfigMap).Data["HUB_POSTGRES_PORT"] = "5432"
		cmConf.(*corev1.ConfigMap).Data["HUB_POSTGRES_USER"] = "blackduck_user"

		secretConf.(*corev1.Secret).Data["HUB_POSTGRES_ADMIN_PASSWORD_FILE"] = []byte(p.blackDuckCr.Spec.AdminPassword)
		secretConf.(*corev1.Secret).Data["HUB_POSTGRES_USER_PASSWORD_FILE"] = []byte(p.blackDuckCr.Spec.UserPassword)
		secretConf.(*corev1.Secret).Data["HUB_POSTGRES_POSTGRES_PASSWORD_FILE"] = []byte(p.blackDuckCr.Spec.PostgresPassword)
	}

	return nil
}

func (p *BlackduckPatcher) patchWebserverCertificates() error {

	if len(p.blackDuckCr.Spec.Certificate) > 0 && len(p.blackDuckCr.Spec.CertificateKey) > 0 {
		runtimeObject, ok := p.mapOfUniqueIdToBaseRuntimeObject[fmt.Sprintf("Secret.%s-blackduck-webserver-certificate", p.blackDuckCr.Name)]
		if !ok {
			return nil
		}
		runtimeObject.(*corev1.Secret).Data["WEBSERVER_CUSTOM_CERT_FILE"] = []byte(p.blackDuckCr.Spec.Certificate)
		runtimeObject.(*corev1.Secret).Data["WEBSERVER_CUSTOM_KEY_FILE"] = []byte(p.blackDuckCr.Spec.CertificateKey)

	}

	return nil
}

// TODO: common with Alert
func (p *BlackduckPatcher) patchEnvirons() error {
	configMapRuntimeObject, ok := p.mapOfUniqueIdToBaseRuntimeObject[fmt.Sprintf("ConfigMap.%s-blackduck-config", p.blackDuckCr.Name)]
	if !ok {
		return nil
	}
	configMap := configMapRuntimeObject.(*corev1.ConfigMap)
	for _, e := range p.blackDuckCr.Spec.Environs {
		vals := strings.Split(e, ":") // TODO - doesn't handle multiple colons
		if len(vals) != 2 {
			fmt.Printf("Could not split environ '%s' on ':'\n", e) // TODO change to log
			continue
		}
		environKey := strings.TrimSpace(vals[0])
		environVal := strings.TrimSpace(vals[1])
		configMap.Data[environKey] = environVal
	}
	return nil
}

// TODO: common with Alert
func (p *BlackduckPatcher) patchNamespace() error {
	accessor := meta.NewAccessor()
	for _, runtimeObject := range p.mapOfUniqueIdToBaseRuntimeObject {
		accessor.SetNamespace(runtimeObject, p.blackDuckCr.Spec.Namespace)
	}
	return nil
}

func (p *BlackduckPatcher) patchLiveness() error {
	// Removes liveness probes if Spec.LivenessProbes is set to false
	for _, v := range p.mapOfUniqueIdToBaseRuntimeObject {
		switch v.(type) {
		case *corev1.ReplicationController:
			if !p.blackDuckCr.Spec.LivenessProbes {
				for i := range v.(*corev1.ReplicationController).Spec.Template.Spec.Containers {
					v.(*corev1.ReplicationController).Spec.Template.Spec.Containers[i].LivenessProbe = nil
				}
			}
		}
	}
	return nil
}

// TODO: common with Alert
func (p *BlackduckPatcher) patchStorage() error {
	for k, v := range p.mapOfUniqueIdToBaseRuntimeObject {
		switch v.(type) {
		case *corev1.PersistentVolumeClaim:
			if !p.blackDuckCr.Spec.PersistentStorage {
				delete(p.mapOfUniqueIdToBaseRuntimeObject, k)
			} else {
				if len(p.blackDuckCr.Spec.PVCStorageClass) > 0 {
					v.(*corev1.PersistentVolumeClaim).Spec.StorageClassName = &p.blackDuckCr.Spec.PVCStorageClass
				}
				for _, pvc := range p.blackDuckCr.Spec.PVC {
					if strings.EqualFold(pvc.Name, v.(*corev1.PersistentVolumeClaim).Name) {
						v.(*corev1.PersistentVolumeClaim).Spec.VolumeName = pvc.VolumeName
						v.(*corev1.PersistentVolumeClaim).Spec.StorageClassName = &pvc.StorageClass
						if quantity, err := resource.ParseQuantity(pvc.Size); err == nil {
							v.(*corev1.PersistentVolumeClaim).Spec.Resources.Requests[corev1.ResourceStorage] = quantity
						}
					}
				}
			}
		case *corev1.ReplicationController:
			if !p.blackDuckCr.Spec.PersistentStorage {
				for i := range v.(*corev1.ReplicationController).Spec.Template.Spec.Volumes {
					// If PersistentVolumeClaim then we change it to emptyDir
					if v.(*corev1.ReplicationController).Spec.Template.Spec.Volumes[i].VolumeSource.PersistentVolumeClaim != nil {
						v.(*corev1.ReplicationController).Spec.Template.Spec.Volumes[i].VolumeSource = corev1.VolumeSource{
							EmptyDir: &corev1.EmptyDirVolumeSource{
								Medium:    corev1.StorageMediumDefault,
								SizeLimit: nil,
							},
						}
					}
				}
			}
		}

	}
	return nil
}

func removeVolumeAndVolumeMountFromRC(rc *corev1.ReplicationController, volumeName string) *corev1.ReplicationController {
	for volumeNb, volume := range rc.Spec.Template.Spec.Volumes {
		if volume.Secret != nil && strings.Compare(volume.Secret.SecretName, volumeName) == 0 {
			rc.Spec.Template.Spec.Volumes = append(rc.Spec.Template.Spec.Volumes[:volumeNb], rc.Spec.Template.Spec.Volumes[volumeNb+1:]...)
			for containerNb, container := range rc.Spec.Template.Spec.Containers {
				for volumeMountNb, volumeMount := range container.VolumeMounts {
					if strings.Compare(volumeMount.Name, volume.Name) == 0 {
						rc.Spec.Template.Spec.Containers[containerNb].VolumeMounts = append(rc.Spec.Template.Spec.Containers[containerNb].VolumeMounts[:volumeMountNb], rc.Spec.Template.Spec.Containers[containerNb].VolumeMounts[volumeMountNb+1:]...)
					}
				}
			}
		}
	}
	return rc
}
