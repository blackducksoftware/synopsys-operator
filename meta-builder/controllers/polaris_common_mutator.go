package controllers

import (
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
)

func UpdateImagePullSecretsForDeployment(objects map[string]runtime.Object, deployments []string, imagePullSecret string) error {
	for _, deployment := range deployments {
		DeploymentUniqueID := "Deployment." + deployment
		deploymentRuntimeObject, ok := objects[DeploymentUniqueID]
		if !ok {
			return nil
		}
		deploymentInstance := deploymentRuntimeObject.(*appsv1.Deployment)
		deploymentInstance.Spec.Template.Spec.ImagePullSecrets = []corev1.LocalObjectReference{
			{
				Name: imagePullSecret,
			},
		}
	}
	return nil
}

func UpdateImagePullSecretsForStatefulSets(objects map[string]runtime.Object, statefulsets []string, imagePullSecret string) error {
	for _, statefulset := range statefulsets {
		StatefulSetUniqueID := "StatefulSet." + statefulset
		statefulsetRuntimeObject, ok := objects[StatefulSetUniqueID]
		if !ok {
			return nil
		}
		statefulsetInstance := statefulsetRuntimeObject.(*appsv1.StatefulSet)
		statefulsetInstance.Spec.Template.Spec.ImagePullSecrets = []corev1.LocalObjectReference{
			{
				Name: imagePullSecret,
			},
		}
	}
	return nil
}

func UpdateImagePullSecretsForJobs(objects map[string]runtime.Object, jobs []string, imagePullSecret string) error {
	for _, job := range jobs {
		JobUniqueID := "Job." + job
		jobRuntimeObject, ok := objects[JobUniqueID]
		if !ok {
			return nil
		}
		jobInstance := jobRuntimeObject.(*batchv1.Job)
		jobInstance.Spec.Template.Spec.ImagePullSecrets = []corev1.LocalObjectReference{
			{
				Name: imagePullSecret,
			},
		}
	}
	return nil
}

func UpdatePersistentVolumeClaim(pvc *corev1.PersistentVolumeClaim, size string) error {
	if size, err := resource.ParseQuantity(size); err == nil {
		pvc.Spec.Resources.Requests[v1.ResourceStorage] = size
	}
	return nil
}
