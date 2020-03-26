module github.com/blackducksoftware/synopsys-operator/cmd/operator-ui

go 1.13

require (
	github.com/blackducksoftware/synopsys-operator v0.0.0-20200306185505-e35a1ccfcce9
	github.com/gobuffalo/buffalo v0.15.5
	github.com/gobuffalo/envy v1.9.0
	github.com/gobuffalo/fizz v1.9.8 // indirect
	github.com/gobuffalo/mw-csrf v1.0.0
	github.com/gobuffalo/mw-forcessl v0.0.0-20200131175327-94b2bd771862
	github.com/gobuffalo/mw-i18n v1.0.0
	github.com/gobuffalo/mw-paramlogger v1.0.0
	github.com/gobuffalo/packr/v2 v2.7.1
	github.com/gobuffalo/suite v2.8.2+incompatible
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.4.2
	github.com/unrolled/secure v1.0.7
	k8s.io/apimachinery v0.17.3
	k8s.io/client-go v0.17.3
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.3+incompatible
	github.com/Azure/go-autorest/autorest/adal => github.com/Azure/go-autorest/autorest/adal v0.8.2
	github.com/blackducksoftware/synopsys-operator => github.com/blackducksoftware/synopsys-operator v0.0.0-20200306185505-e35a1ccfcce9
	github.com/gobuffalo/buffalo => github.com/gobuffalo/buffalo v0.15.5
	github.com/gobuffalo/envy => github.com/gobuffalo/envy v1.9.0
	github.com/gobuffalo/mw-csrf => github.com/gobuffalo/mw-csrf v1.0.0
	github.com/gobuffalo/mw-forcessl => github.com/gobuffalo/mw-forcessl v0.0.0-20200131175327-94b2bd771862
	github.com/gobuffalo/mw-i18n => github.com/gobuffalo/mw-i18n v1.0.0
	github.com/gobuffalo/mw-paramlogger => github.com/gobuffalo/mw-paramlogger v1.0.0
	github.com/gobuffalo/packr/v2 => github.com/gobuffalo/packr/v2 v2.7.1
	github.com/pkg/errors => github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus => github.com/sirupsen/logrus v1.4.2
	github.com/ugorji/go => github.com/ugorji/go v1.1.7 // indirect
	github.com/unrolled/secure => github.com/unrolled/secure v1.0.7
	k8s.io/api => k8s.io/api v0.17.3
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.17.3
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.3
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.17.3
	k8s.io/client-go => k8s.io/client-go v0.17.3
)
