package types

import (
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"k8s.io/client-go/kubernetes"
)

// ComponentName denotes the component/resource name
type ComponentName string

// ContainerName denotes the container name
type ContainerName string

// ReplicationControllerCreater refers to Replication Controller creater
type ReplicationControllerCreater func(*ReplicationController, *protoform.Config, *kubernetes.Clientset, interface{}) (ReplicationControllerInterface, error)

// ServiceCreater refers to Service creater
type ServiceCreater func(*protoform.Config, *kubernetes.Clientset, interface{}) (ServiceInterface, error)

// ConfigmapCreater refers to Replication Controller creater
type ConfigmapCreater func(*protoform.Config, *kubernetes.Clientset, interface{}) (ConfigMapInterface, error)

// PvcCreater refers to Replication Controller creater
type PvcCreater func(*protoform.Config, *kubernetes.Clientset, interface{}) (PVCInterface, error)

// SecretCreater refers to Replication Controller creater
type SecretCreater func(*protoform.Config, *kubernetes.Clientset, interface{}) (SecretInterface, error)

// TagOrImage refers to Image and Tag
type TagOrImage struct {
	Tag   string
	Image string
}

// ConfigMapInterface refers to Config Map related interface
type ConfigMapInterface interface {
	GetCM() []*components.ConfigMap
}

// PVCInterface refers to PVC related interface
type PVCInterface interface {
	GetPVCs() ([]*components.PersistentVolumeClaim, error)
	// TODO add deployment, rc
}

// ReplicationController refers to replication controller configuration
type ReplicationController struct {
	Namespace  string
	Replicas   int
	Containers map[ContainerName]Container
}

// Container refers to container configuration
type Container struct {
	Image  string
	MinCPU *int32
	MaxCPU *int32
	MinMem *int32
	MaxMem *int32
}

// ReplicationControllerInterface refers to replication controller related interface
type ReplicationControllerInterface interface {
	GetRc() (*components.ReplicationController, error)
	// TODO add deployment, rc
}

// SecretInterface refers to secret related interface
type SecretInterface interface {
	GetSecrets() []*components.Secret
}

// ServiceInterface refers to service related interface
type ServiceInterface interface {
	GetService() *components.Service
	// TODO add deployment, rc
}

// SizeInterface refers to size related interface
type SizeInterface interface {
	GetSize(name string) map[string]*Size
}

// ContainerSize refers to container size configuration
type ContainerSize struct {
	MinCPU *int32
	MaxCPU *int32
	MinMem *int32
	MaxMem *int32
}

// Size refers to size configuration
type Size struct {
	Replica    int
	Containers map[ContainerName]ContainerSize
}
