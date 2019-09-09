package controllers

import (
	"fmt"

	b64 "encoding/base64"

	synopsysv1 "github.com/blackducksoftware/synopsys-operator/meta-builder/api/v1"
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
		p.patchSMTPDetails,
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
	// patch cloudsql secret
	SecretUniqueID := "Secret." + "cloudsql"
	secretRuntimeObject, ok := p.mapOfUniqueIdToBaseRuntimeObject[SecretUniqueID]
	if !ok {
		return nil
	}
	secretInstance := secretRuntimeObject.(*corev1.Secret)
	secretInstance.Data = map[string][]byte{
		"username":              []byte(b64.StdEncoding.EncodeToString([]byte(p.polarisDbCr.Spec.PostgresDetails.Username))),
		"password":              []byte(b64.StdEncoding.EncodeToString([]byte(p.polarisDbCr.Spec.PostgresDetails.Password))),
		"reporting_db_username": []byte(b64.StdEncoding.EncodeToString([]byte(p.polarisDbCr.Spec.PostgresDetails.Username))),
		"reporting_db_password": []byte(b64.StdEncoding.EncodeToString([]byte(p.polarisDbCr.Spec.PostgresDetails.Password))),
		"host":                  []byte(b64.StdEncoding.EncodeToString([]byte(p.polarisDbCr.Spec.PostgresDetails.Host))),
		"port":                  []byte(b64.StdEncoding.EncodeToString([]byte(string(p.polarisDbCr.Spec.PostgresDetails.Port)))),
	}
	if p.polarisDbCr.Spec.PostgresInstanceType == "internal" {
		// patch postgresql-config secret
		SecretUniqueID = "Secret." + "postgresql-config"
		secretRuntimeObject, ok = p.mapOfUniqueIdToBaseRuntimeObject[SecretUniqueID]
		if !ok {
			return nil
		}
		secretInstance = secretRuntimeObject.(*corev1.Secret)
		secretInstance.Data = map[string][]byte{
			"POSTGRES_USER":     []byte(b64.StdEncoding.EncodeToString([]byte(p.polarisDbCr.Spec.PostgresDetails.Username))),
			"POSTGRES_PASSWORD": []byte(b64.StdEncoding.EncodeToString([]byte(p.polarisDbCr.Spec.PostgresDetails.Password))),
		}
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
