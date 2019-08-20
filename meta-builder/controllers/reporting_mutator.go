package controllers

import (
	"fmt"

	synopsysv1 "github.com/blackducksoftware/synopsys-operator/meta-builder/api/v1"
	corev1 "k8s.io/api/core/v1"
  appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
)

func patchReporting(reportingCr *synopsysv1.Reporting, mapOfUniqueIdToBaseRuntimeObject map[string]runtime.Object, accessor meta.MetadataAccessor) map[string]runtime.Object {
	patcher := ReportingPatcher{
		reportingCr:                          reportingCr,
		mapOfUniqueIdToBaseRuntimeObject: mapOfUniqueIdToBaseRuntimeObject,
		accessor:                         accessor,
	}
	return patcher.patch()
}

type ReportingPatcher struct {
	reportingCr                          *synopsysv1.Reporting
	mapOfUniqueIdToBaseRuntimeObject map[string]runtime.Object
	accessor                         meta.MetadataAccessor
}

func (p *ReportingPatcher) patch() map[string]runtime.Object {
	patches := []func() error{
		p.patchNamespace,
		p.patchEnvironmentName,
		p.patchEnvironmentDNS,
		p.patchImagePullSecrets,
	}
	for _, f := range patches {
		err := f()
		if err != nil {
			fmt.Printf("%s\n", err)
		}
	}
	return p.mapOfUniqueIdToBaseRuntimeObject
}

func (p *ReportingPatcher) patchNamespace() error {
	for _, runtimeObject := range p.mapOfUniqueIdToBaseRuntimeObject {
		p.accessor.SetNamespace(runtimeObject, p.reportingCr.Spec.Namespace)
	}
	return nil
}

func (p *ReportingPatcher) patchEnvironmentName() error {
	// Patch instances of environment name
	return nil
}

func (p *ReportingPatcher) patchEnvironmentDNS() error {
	// Patch instances of environment dns
	return nil
}

func (p *ReportingPatcher) patchImagePullSecrets() error {
	// improve logic to get these objects directly from dependency manual
	deployments := []string {
		"rp-issue-manager",
		"rp-portfolio-service",
		"rp-report-service",
		"rp-swagger-doc",
		"rp-tools-portfolio-service",
		"report-storage",
		"rp-frontend",
		"rp-polaris-agent-service",
		}
	for _, deployment := range deployments{
		DeploymentUniqueID := "Deployment."+deployment
		deploymentRuntimeObject, ok := p.mapOfUniqueIdToBaseRuntimeObject[DeploymentUniqueID]
		if !ok {
			return nil
		}
		deploymentInstance := deploymentRuntimeObject.(*appsv1.Deployment)
		deploymentInstance.Spec.Template.Spec.ImagePullSecrets = []corev1.LocalObjectReference{
			{
				Name: p.reportingCr.Spec.ImagePullSecrets,
			},
		}
	}
	return nil
}