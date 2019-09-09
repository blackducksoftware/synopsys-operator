package controllers

import (
	synopsysv1 "github.com/blackducksoftware/synopsys-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	appsv1 "k8s.io/api/extensions/v1beta1"
)

func updateServiceCoreContainerImage(containers *[]corev1.Container, serviceName *string, imageDetails *synopsysv1.ImageDetails) *[]corev1.Container {
	// pop the service core container
	for index, container := range *containers {
		if container.Name == *serviceName {
			// Update image details in container
			container.Image = imageDetails.Repository + "/" + imageDetails.Image + ":" + imageDetails.Tag
			// Replace the container object in list
			(*containers)[index] = container
		}
	}
	return containers
}

func PatchImageForService(imageDetails *synopsysv1.ImageDetails, deployment *appsv1.Deployment) *appsv1.Deployment {
	deployment.Spec.Template.Spec.Containers = *updateServiceCoreContainerImage(
		&deployment.Spec.Template.Spec.Containers,
		&deployment.ObjectMeta.Name,
		imageDetails,
	)
	return deployment
}
