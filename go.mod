module github.com/blackducksoftware/synopsys-operator

go 1.12

require (
	github.com/Azure/go-autorest/autorest v0.8.0 // indirect
	github.com/blackducksoftware/horizon v0.0.0-20190625151958-16cafa9109a3
	github.com/docker/spdystream v0.0.0-20181023171402-6480d4af844c // indirect
	github.com/evanphx/json-patch v4.5.0+incompatible // indirect
	github.com/gin-gonic/gin v1.4.0
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/mock v1.2.0 // indirect
	github.com/google/go-cmp v0.3.0
	github.com/google/gofuzz v1.0.0
	github.com/gophercloud/gophercloud v0.3.0 // indirect
	github.com/imdario/mergo v0.3.7
	github.com/juju/errors v0.0.0-20190806202954-0232dcc7464d
	github.com/lib/pq v1.2.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/onsi/ginkgo v1.8.0
	github.com/onsi/gomega v1.5.0
	github.com/openshift/api v3.9.0+incompatible
	github.com/openshift/client-go v3.9.0+incompatible
	github.com/prometheus/client_golang v1.1.0
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.3
	github.com/spf13/viper v1.4.0
	github.com/stretchr/testify v1.3.0
	golang.org/x/net v0.0.0-20190613194153-d28f0bde5980
	gopkg.in/inf.v0 v0.9.1 // indirect
	k8s.io/api v0.0.0-20190808180749-077ce48e77da
	k8s.io/apiextensions-apiserver v0.0.0-20190809061809-636e76ffcf57
	k8s.io/apimachinery v0.0.0-20190809020650-423f5d784010
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/klog v0.4.0 // indirect
	k8s.io/kube-openapi v0.0.0-20190722073852-5e22f3d471e6 // indirect
	k8s.io/utils v0.0.0-20190809000727-6c36bc71fc4a // indirect
	sigs.k8s.io/yaml v1.1.0
)

replace (
	git.apache.org/thrift.git => github.com/apache/thrift v0.0.0-20180902110319-2566ecd5d999
	github.com/Azure/go-autorest/autorest => github.com/Azure/go-autorest/autorest v0.8.0
	github.com/blackducksoftware/horizon => github.com/blackducksoftware/horizon v0.0.0-20190625151958-16cafa9109a3
	github.com/docker/spdystream => github.com/docker/spdystream v0.0.0-20181023171402-6480d4af844c // indirect
	github.com/evanphx/json-patch => github.com/evanphx/json-patch v4.5.0+incompatible // indirect
	github.com/gin-gonic/gin => github.com/gin-gonic/gin v1.4.0
	github.com/golang/mock => github.com/golang/mock v1.2.0 // indirect
	github.com/google/go-cmp => github.com/google/go-cmp v0.3.0
	github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.3.0 // indirect
	github.com/gophercloud/gophercloud => github.com/gophercloud/gophercloud v0.3.0 // indirect
	github.com/imdario/mergo => github.com/imdario/mergo v0.3.7
	github.com/juju/errors => github.com/juju/errors v0.0.0-20190806202954-0232dcc7464d
	github.com/lib/pq => github.com/lib/pq v1.2.0
	github.com/mitchellh/go-homedir => github.com/mitchellh/go-homedir v1.1.0
	github.com/onsi/ginkgo => github.com/onsi/ginkgo v1.7.0
	github.com/onsi/gomega => github.com/onsi/gomega v1.4.3
	github.com/openshift/api => github.com/openshift/api v3.9.0+incompatible
	github.com/openshift/client-go => github.com/openshift/client-go v3.9.0+incompatible
	github.com/prometheus/client_golang => github.com/prometheus/client_golang v1.1.0
	github.com/sirupsen/logrus => github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra => github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag => github.com/spf13/pflag v1.0.3
	github.com/spf13/viper => github.com/spf13/viper v1.4.0
	github.com/stretchr/testify => github.com/stretchr/testify v1.3.0
	gopkg.in/inf.v0 => gopkg.in/inf.v0 v0.9.1 // indirect
	k8s.io/api => k8s.io/api v0.0.0-20190313235455-40a48860b5ab
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190315093550-53c4693659ed
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190313205120-d7deff9243b1
	k8s.io/client-go => k8s.io/client-go v11.0.0+incompatible
	k8s.io/klog => k8s.io/klog v0.4.0 // indirect
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20190722073852-5e22f3d471e6 // indirect
	k8s.io/utils => k8s.io/utils v0.0.0-20190809000727-6c36bc71fc4a // indirect
	sigs.k8s.io/yaml => sigs.k8s.io/yaml v1.1.0
)
