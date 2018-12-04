package actions

import (
	"fmt"

	"github.com/gobuffalo/buffalo"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// VasilyResource ...
type VasilyResource struct {
	buffalo.Resource
	kubeClient *kubernetes.Clientset
}

// NewVasilyResource ...
func NewVasilyResource(kubeConfig *rest.Config) (*VasilyResource, error) {
	kubeClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create kube client due to %+v", err)
	}
	return &VasilyResource{kubeClient: kubeClient}, nil
}

// Show ...
func (v VasilyResource) Show(c buffalo.Context) error {
	v1Client := v.kubeClient.CoreV1()
	nodeList, err := v1Client.Nodes().List(v1.ListOptions{})
	if err != nil {
		return c.Error(500, err)
	}
	return c.Render(200, r.JSON(nodeList))
}
