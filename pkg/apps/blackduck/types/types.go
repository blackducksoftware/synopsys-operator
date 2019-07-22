package types

import (
	"github.com/blackducksoftware/horizon/pkg/components"
	v1 "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"k8s.io/client-go/kubernetes"
)

type ReplicationControllerCreater func(*ReplicationController, *protoform.Config, *kubernetes.Clientset, *v1.Blackduck) ReplicationControllerInterface
type ServiceCreater func(config *protoform.Config, kubeClient *kubernetes.Clientset, blackduck *v1.Blackduck) ServiceInterface
type ConfigmapCreater func(config *protoform.Config, kubeClient *kubernetes.Clientset, blackduck *v1.Blackduck) ConfigMapInterface
type PvcCreater func(blackduck *v1.Blackduck) PVCInterface
type SecretCreater func(config *protoform.Config, kubeClient *kubernetes.Clientset, blackduck *v1.Blackduck) SecretInterface

type TagOrImage struct {
	Tag   string
	Image string
}

type ConfigMapInterface interface {
	GetCM() []*components.ConfigMap
}

type PVCInterface interface {
	GetPVCs() []*components.PersistentVolumeClaim
	// TODO add deployment, rc
}

type ReplicationController struct {
	Namespace  string
	Replicas   int
	Containers map[ContainerName]Container
}

type Container struct {
	Image  string
	MinCPU *int32
	MaxCPU *int32
	MinMem *int32
	MaxMem *int32
}

type ReplicationControllerInterface interface {
	GetRc() (*components.ReplicationController, error)
	// TODO add deployment, rc
}

type SecretInterface interface {
	GetSecrets() []*components.Secret
}

type ServiceInterface interface {
	GetService() *components.Service
	// TODO add deployment, rc
}

type SizeInterface interface {
	GetSize(name string) map[string]*Size
}

type ContainerSize struct {
	MinCPU *int32
	MaxCPU *int32
	MinMem *int32
	MaxMem *int32
}

type Size struct {
	Replica    int
	Containers map[ContainerName]ContainerSize
}
