module github.com/blackducksoftware/synopsys-operator

go 1.12

require (
	github.com/Azure/go-autorest/autorest v0.5.0 // indirect
	github.com/blackducksoftware/horizon v0.0.0-20190603173136-e141457f7a80
	github.com/docker/spdystream v0.0.0-20181023171402-6480d4af844c // indirect
	github.com/evanphx/json-patch v4.5.0+incompatible // indirect
	github.com/gin-gonic/gin v1.4.0
	github.com/go-logfmt/logfmt v0.4.0 // indirect
	github.com/google/go-cmp v0.3.0
	github.com/google/gofuzz v1.0.0 // indirect
	github.com/gophercloud/gophercloud v0.3.0 // indirect
	github.com/imdario/mergo v0.3.7
	github.com/juju/errors v0.0.0-20190207033735-e65537c515d7
	github.com/juju/testing v0.0.0-20190723135506-ce30eb24acd2 // indirect
	github.com/lib/pq v1.1.1
	github.com/mitchellh/go-homedir v1.1.0
	github.com/onsi/ginkgo v1.8.0
	github.com/onsi/gomega v1.5.0
	github.com/openshift/api v3.9.0+incompatible
	github.com/openshift/client-go v3.9.0+incompatible
	github.com/prometheus/client_golang v0.9.4
	github.com/prometheus/common v0.4.1
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.3
	github.com/spf13/viper v1.4.0
	github.com/stretchr/testify v1.3.0
	gopkg.in/inf.v0 v0.9.1 // indirect
	k8s.io/api v0.0.0-20190612125737-db0771252981
	k8s.io/apiextensions-apiserver v0.0.0-20190726024412-102230e288fd
	k8s.io/apimachinery v0.0.0-20190727130956-f97a4e5b4abc
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/kube-openapi v0.0.0-20190722073852-5e22f3d471e6 // indirect
	k8s.io/utils v0.0.0-20190712204705-3dccf664f023 // indirect
	sigs.k8s.io/yaml v1.1.0
)

replace (
	k8s.io/api => k8s.io/api v0.0.0-20190313235455-40a48860b5ab
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190315093550-53c4693659ed
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190313205120-d7deff9243b1
)
