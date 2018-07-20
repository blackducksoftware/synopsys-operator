package hub

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type HubCreator struct {
	Config *rest.Config
	Client *kubernetes.Clientset
}

type Hub struct {
	Namespace      string
	DockerRegistry string
	DockerRepo     string
	HubVersion     string
	Flavor         string
	AdminPassword  string
	UserPassword   string
}
