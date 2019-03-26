package containers

import (
	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
)

// GetServiceAccount will return the service account
func (c *Creater) GetServiceAccount() *components.ServiceAccount {
	return components.NewServiceAccount(horizonapi.ServiceAccountConfig{
		Name:      c.hubSpec.Namespace,
		Namespace: c.hubSpec.Namespace,
	})
}

// GetClusterRoleBinding will return the cluster role binding
func (c *Creater) GetClusterRoleBinding() *components.ClusterRoleBinding {
	clusterRoleBinding := components.NewClusterRoleBinding(horizonapi.ClusterRoleBindingConfig{
		Name:       c.hubSpec.Namespace,
		APIVersion: "rbac.authorization.k8s.io/v1",
	})

	clusterRoleBinding.AddSubject(horizonapi.SubjectConfig{
		Kind:      "ServiceAccount",
		Name:      c.hubSpec.Namespace,
		Namespace: c.hubSpec.Namespace,
	})
	clusterRoleBinding.AddRoleRef(horizonapi.RoleRefConfig{
		APIGroup: "",
		Kind:     "ClusterRole",
		Name:     "synopsys-operator-admin",
	})

	return clusterRoleBinding
}
