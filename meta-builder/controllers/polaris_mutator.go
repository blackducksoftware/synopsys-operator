package controllers

import (
	"fmt"

	synopsysv1 "github.com/blackducksoftware/synopsys-operator/meta-builder/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
)

func patchPolaris(polarisCr *synopsysv1.Polaris, mapOfUniqueIdToBaseRuntimeObject map[string]runtime.Object, accessor meta.MetadataAccessor) map[string]runtime.Object {
	patcher := PolarisPatcher{
		polarisCr:                        polarisCr,
		mapOfUniqueIdToBaseRuntimeObject: mapOfUniqueIdToBaseRuntimeObject,
		accessor:                         accessor,
	}
	return patcher.patch()
}

type PolarisPatcher struct {
	polarisCr                        *synopsysv1.Polaris
	mapOfUniqueIdToBaseRuntimeObject map[string]runtime.Object
	accessor                         meta.MetadataAccessor
}

func (p *PolarisPatcher) patch() map[string]runtime.Object {
	patches := []func() error{
		p.patchNamespace,
		p.patchEnvironmentName,
		p.patchEnvironmentDNS,
		p.patchImagePullSecrets,
		p.patchAuthServerSpec,
	}
	for _, f := range patches {
		err := f()
		if err != nil {
			fmt.Printf("%s\n", err)
		}
	}
	return p.mapOfUniqueIdToBaseRuntimeObject
}

func (p *PolarisPatcher) patchNamespace() error {
	for _, runtimeObject := range p.mapOfUniqueIdToBaseRuntimeObject {
		p.accessor.SetNamespace(runtimeObject, p.polarisCr.Spec.Namespace)
	}
	return nil
}

func (p *PolarisPatcher) patchEnvironmentName() error {
	// Patch instances of environment name
	return nil
}

func (p *PolarisPatcher) patchEnvironmentDNS() error {
	// Patch instances of environment dns
	return nil
}

func (p *PolarisPatcher) patchImagePullSecrets() error {
	// improve logic to get these objects directly from dependency manual
	deployments := []string{"auth-server"}
	for _, deployment := range deployments {
		DeploymentUniqueID := "Deployment." + deployment
		deploymentRuntimeObject, ok := p.mapOfUniqueIdToBaseRuntimeObject[DeploymentUniqueID]
		if !ok {
			return nil
		}
		deploymentInstance := deploymentRuntimeObject.(*appsv1.Deployment)
		deploymentInstance.Spec.Template.Spec.ImagePullSecrets = []corev1.LocalObjectReference{
			{
				Name: p.polarisCr.Spec.ImagePullSecrets,
			},
		}
	}
	return nil
}

func (p *PolarisPatcher) patchAuthServerSpec() error {
	// Patch auth-server spec with chagnes
	return nil
}
