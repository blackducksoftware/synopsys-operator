package controllers

import (
	"fmt"
	"strconv"

	b64 "encoding/base64"

	synopsysv1 "github.com/blackducksoftware/synopsys-operator/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func patchPolarisDB(client client.Client, polarisDbCr *synopsysv1.PolarisDB, mapOfUniqueIdToBaseRuntimeObject map[string]runtime.Object) map[string]runtime.Object {
	patcher := PolarisDBPatcher{
		polarisDbCr:                      polarisDbCr,
		mapOfUniqueIdToBaseRuntimeObject: mapOfUniqueIdToBaseRuntimeObject,
		Client:                           client,
	}
	return patcher.patch()
}

type PolarisDBPatcher struct {
	polarisDbCr                      *synopsysv1.PolarisDB
	mapOfUniqueIdToBaseRuntimeObject map[string]runtime.Object
	client.Client
}

func (p *PolarisDBPatcher) patch() map[string]runtime.Object {
	patches := []func() error{
		p.patchNamespace,
		p.patchSMTPSecretDetails,
		p.patchSMTPConfigMapDetails,
		p.patchPostgresDetails,
		p.patchEventstoreDetails,
		p.patchUploadServerDetails,
	}
	for _, f := range patches {
		err := f()
		if err != nil {
			fmt.Printf("%s\n", err)
		}
	}
	return p.mapOfUniqueIdToBaseRuntimeObject
}

func (p *PolarisDBPatcher) patchNamespace() error {
	accessor := meta.NewAccessor()
	for _, runtimeObject := range p.mapOfUniqueIdToBaseRuntimeObject {
		accessor.SetNamespace(runtimeObject, p.polarisDbCr.Spec.Namespace)
	}
	return nil
}

func (p *PolarisDBPatcher) patchSMTPSecretDetails() error {
	SecretUniqueID := "Secret." + "smtp"
	secretRuntimeObject, ok := p.mapOfUniqueIdToBaseRuntimeObject[SecretUniqueID]
	if !ok {
		return nil
	}
	secretInstance := secretRuntimeObject.(*corev1.Secret)
	secretInstance.Data = map[string][]byte{
		"username": []byte(b64.StdEncoding.EncodeToString([]byte(p.polarisDbCr.Spec.SMTPDetails.Username))),
		"passwd":   []byte(b64.StdEncoding.EncodeToString([]byte(p.polarisDbCr.Spec.SMTPDetails.Password))),
	}
	return nil
}

func (p *PolarisDBPatcher) patchSMTPConfigMapDetails() error {
	ConfigMapUniqueID := "ConfigMap." + "smtp"
	configmapRuntimeObject, ok := p.mapOfUniqueIdToBaseRuntimeObject[ConfigMapUniqueID]
	if !ok {
		return nil
	}
	configMapInstance := configmapRuntimeObject.(*corev1.ConfigMap)
	configMapInstance.Data = map[string]string{
		"host": p.polarisDbCr.Spec.SMTPDetails.Host,
		"port": strconv.Itoa(p.polarisDbCr.Spec.SMTPDetails.Port),
	}
	return nil
}

func (p *PolarisDBPatcher) patchSMTPDetails() error {
	SecretUniqueID := "Secret." + "smtp"
	secretRuntimeObject, ok := p.mapOfUniqueIdToBaseRuntimeObject[SecretUniqueID]
	if !ok {
		return nil
	}
	secretInstance := secretRuntimeObject.(*corev1.Secret)
	secretInstance.Data = map[string][]byte{
		"username": []byte(b64.StdEncoding.EncodeToString([]byte(p.polarisDbCr.Spec.SMTPDetails.Username))),
		"passwd":   []byte(b64.StdEncoding.EncodeToString([]byte(p.polarisDbCr.Spec.SMTPDetails.Password))),
		"host":     []byte(b64.StdEncoding.EncodeToString([]byte(p.polarisDbCr.Spec.SMTPDetails.Host))),
		"port":     []byte(b64.StdEncoding.EncodeToString([]byte(string(p.polarisDbCr.Spec.SMTPDetails.Port)))),
	}
	return nil
}

func (p *PolarisDBPatcher) patchPostgresDetails() error {
	// patch postgresql-config secret
	ConfigMapUniqueID := "ConfigMap." + "postgresql-config"
	configmapRuntimeObject, ok := p.mapOfUniqueIdToBaseRuntimeObject[ConfigMapUniqueID]
	if !ok {
		return nil
	}
	configMapInstance := configmapRuntimeObject.(*corev1.ConfigMap)
	configMapInstance.Data = map[string]string{
		"POSTGRESQL_ADMIN_PASSWORD": p.polarisDbCr.Spec.PostgresDetails.Password,
		"POSTGRESQL_DATABASE":       p.polarisDbCr.Spec.PostgresDetails.Username,
		"POSTGRESQL_PASSWORD":       p.polarisDbCr.Spec.PostgresDetails.Password,
		"POSTGRESQL_USER":           p.polarisDbCr.Spec.PostgresDetails.Username,
		"POSTGRESQL_HOST":           p.polarisDbCr.Spec.PostgresDetails.Host,
		"POSTGRESQL_PORT":           strconv.Itoa(p.polarisDbCr.Spec.PostgresDetails.Port),
	}
	if p.polarisDbCr.Spec.PostgresInstanceType == "internal" {
		// patch storage
		PostgresPVCUniqueID := "PersistentVolumeClaim." + "postgresql-pv-claim"
		PostgresPVCRuntimeObject, ok := p.mapOfUniqueIdToBaseRuntimeObject[PostgresPVCUniqueID]
		if !ok {
			return nil
		}
		PostgresPVCInstance := PostgresPVCRuntimeObject.(*corev1.PersistentVolumeClaim)
		UpdatePersistentVolumeClaim(PostgresPVCInstance, p.polarisDbCr.Spec.PostgresStorageDetails.StorageSize)
	}
	return nil
}

func (p *PolarisDBPatcher) patchEventstoreDetails() error {
	StatefulSetUniqueID := "StatefulSet." + "eventstore"
	statefulSetRuntimeObject, ok := p.mapOfUniqueIdToBaseRuntimeObject[StatefulSetUniqueID]
	if !ok {
		return nil
	}
	statefulsetInstance := statefulSetRuntimeObject.(*appsv1.StatefulSet)
	if size, err := resource.ParseQuantity(p.polarisDbCr.Spec.EventstoreDetails.Storage.StorageSize); err == nil {
		statefulsetInstance.Spec.VolumeClaimTemplates[0].Spec.Resources.Requests[v1.ResourceStorage] = size
	}
	return nil
}

func (p *PolarisDBPatcher) patchUploadServerDetails() error {
	UploadServerPVCUniqueID := "PersistentVolumeClaim." + "upload-server-pv-claim"
	UploadServerPVCRuntimeObject, ok := p.mapOfUniqueIdToBaseRuntimeObject[UploadServerPVCUniqueID]
	if !ok {
		return nil
	}
	UploadServerPVCInstance := UploadServerPVCRuntimeObject.(*corev1.PersistentVolumeClaim)
	UpdatePersistentVolumeClaim(UploadServerPVCInstance, p.polarisDbCr.Spec.UploadServerDetails.Storage.StorageSize)
	return nil
}
