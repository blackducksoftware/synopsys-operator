package kubebuilder

import (
	"k8s.io/api/apps/v1beta1"
	"k8s.io/api/core/v1"
)

type Resources interface {
	GetConfigMaps() []*v1.ConfigMap
	GetServices() []*v1.Service
	GetSecrets() []*v1.Secret
	GetReplicationControllers() []*v1.ReplicationController
	GetDeployments() []*v1beta1.Deployment
}
