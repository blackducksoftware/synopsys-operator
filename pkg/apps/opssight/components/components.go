package components

import (

	// Cluster Role
	_ "github.com/blackducksoftware/synopsys-operator/pkg/apps/opssight/components/clusterrole/imageprocessor/v1"
	_ "github.com/blackducksoftware/synopsys-operator/pkg/apps/opssight/components/clusterrole/podprocessor/v1"
	_ "github.com/blackducksoftware/synopsys-operator/pkg/apps/opssight/components/clusterrole/skyfire/v1"

	// Cluster Role Binding
	_ "github.com/blackducksoftware/synopsys-operator/pkg/apps/opssight/components/clusterrolebinding/imageprocessor/v1"
	_ "github.com/blackducksoftware/synopsys-operator/pkg/apps/opssight/components/clusterrolebinding/podprocessor/v1"
	_ "github.com/blackducksoftware/synopsys-operator/pkg/apps/opssight/components/clusterrolebinding/scanner/v1"
	_ "github.com/blackducksoftware/synopsys-operator/pkg/apps/opssight/components/clusterrolebinding/skyfire/v1"

	// Configmap
	_ "github.com/blackducksoftware/synopsys-operator/pkg/apps/opssight/components/configmap/metrics/v1"
	_ "github.com/blackducksoftware/synopsys-operator/pkg/apps/opssight/components/configmap/opssight/v1"

	// Deployments
	_ "github.com/blackducksoftware/synopsys-operator/pkg/apps/opssight/components/deployment/prometheus/v1"

	// RCs
	_ "github.com/blackducksoftware/synopsys-operator/pkg/apps/opssight/components/rc/core/v1"
	_ "github.com/blackducksoftware/synopsys-operator/pkg/apps/opssight/components/rc/imageprocessor/v1"
	_ "github.com/blackducksoftware/synopsys-operator/pkg/apps/opssight/components/rc/podprocessor/v1"
	_ "github.com/blackducksoftware/synopsys-operator/pkg/apps/opssight/components/rc/scanner/v1"
	_ "github.com/blackducksoftware/synopsys-operator/pkg/apps/opssight/components/rc/skyfire/v1"

	// Route
	_ "github.com/blackducksoftware/synopsys-operator/pkg/apps/opssight/components/route/core/v1"
	_ "github.com/blackducksoftware/synopsys-operator/pkg/apps/opssight/components/route/metrics/v1"

	// Secrets
	_ "github.com/blackducksoftware/synopsys-operator/pkg/apps/opssight/components/secret/v1"

	// Services
	_ "github.com/blackducksoftware/synopsys-operator/pkg/apps/opssight/components/service/core/v1"
	_ "github.com/blackducksoftware/synopsys-operator/pkg/apps/opssight/components/service/exposecore/v1"
	_ "github.com/blackducksoftware/synopsys-operator/pkg/apps/opssight/components/service/exposemetrics/v1"
	_ "github.com/blackducksoftware/synopsys-operator/pkg/apps/opssight/components/service/imagegetter/v1"
	_ "github.com/blackducksoftware/synopsys-operator/pkg/apps/opssight/components/service/imageprocessor/v1"
	_ "github.com/blackducksoftware/synopsys-operator/pkg/apps/opssight/components/service/podprocessor/v1"
	_ "github.com/blackducksoftware/synopsys-operator/pkg/apps/opssight/components/service/scanner/v1"
	_ "github.com/blackducksoftware/synopsys-operator/pkg/apps/opssight/components/service/skyfire/v1"

	// Service Account
	_ "github.com/blackducksoftware/synopsys-operator/pkg/apps/opssight/components/serviceaccount/imageprocessor/v1"
	_ "github.com/blackducksoftware/synopsys-operator/pkg/apps/opssight/components/serviceaccount/podprocessor/v1"
	_ "github.com/blackducksoftware/synopsys-operator/pkg/apps/opssight/components/serviceaccount/scanner/v1"
	_ "github.com/blackducksoftware/synopsys-operator/pkg/apps/opssight/components/serviceaccount/skyfire/v1"
)
