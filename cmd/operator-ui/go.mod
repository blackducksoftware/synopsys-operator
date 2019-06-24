module github.com/blackducksoftware/synopsys-operator/cmd/operator-ui

go 1.12

replace (
	github.com/Azure/go-autorest/autorest => github.com/Azure/go-autorest/autorest v0.2.0
	github.com/blackducksoftware/synopsys-operator => github.com/blackducksoftware/synopsys-operator v0.0.0-20190624162144-87d94ffcf2b5
	github.com/docker/spdystream => github.com/docker/spdystream v0.0.0-20181023171402-6480d4af844c
	github.com/gobuffalo/buffalo => github.com/gobuffalo/buffalo v0.14.6
	github.com/gobuffalo/envy => github.com/gobuffalo/envy v1.7.0
	github.com/gobuffalo/mw-csrf => github.com/gobuffalo/mw-csrf v0.0.0-20190129204204-25460a055517
	github.com/gobuffalo/mw-forcessl => github.com/gobuffalo/mw-forcessl v0.0.0-20190224202501-6d1ef7ffb276
	github.com/gobuffalo/mw-i18n => github.com/gobuffalo/mw-i18n v0.0.0-20190224203426-337de00e4c33
	github.com/gobuffalo/mw-paramlogger => github.com/gobuffalo/mw-paramlogger v0.0.0-20190224201358-0d45762ab655
	github.com/gobuffalo/packr/v2 => github.com/gobuffalo/packr/v2 v2.4.0
	github.com/golang/lint => github.com/golang/lint v0.0.0-20190409202823-5614ed5bae6fb75893070bdc0996a68765fdd275
	github.com/google/gofuzz => github.com/google/gofuzz v1.0.0
	github.com/pkg/errors => github.com/pkg/errors v0.8.1
	github.com/sirupsen/logrus => github.com/sirupsen/logrus v1.4.2
	github.com/unrolled/secure => github.com/unrolled/secure v1.0.0
	gopkg.in/inf.v0 => gopkg.in/inf.v0 v0.9.1
	k8s.io/api => k8s.io/api v0.0.0-20190313235455-40a48860b5ab
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190315093550-53c4693659ed
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190313205120-d7deff9243b1
	k8s.io/client-go => k8s.io/client-go v11.0.0+incompatible
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20190603182131-db7b694dc208
	sourcegraph.com/sourcegraph/go-diff => sourcegraph.com/sourcegraph/go-diff v0.5.0
)

require (
	contrib.go.opencensus.io/exporter/ocagent v0.5.0 // indirect
	github.com/Azure/go-autorest/autorest v0.2.0 // indirect
	github.com/blackducksoftware/synopsys-operator v0.0.0-20190624162144-87d94ffcf2b5
	github.com/docker/spdystream v0.0.0-20181023171402-6480d4af844c // indirect
	github.com/elazarl/goproxy v0.0.0-20190421051319-9d40249d3c2f // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/gobuffalo/buffalo v0.14.6
	github.com/gobuffalo/envy v1.7.0
	github.com/gobuffalo/mw-csrf v0.0.0-20190129204204-25460a055517
	github.com/gobuffalo/mw-forcessl v0.0.0-20190224202501-6d1ef7ffb276
	github.com/gobuffalo/mw-i18n v0.0.0-20190224203426-337de00e4c33
	github.com/gobuffalo/mw-paramlogger v0.0.0-20190224201358-0d45762ab655
	github.com/gobuffalo/packr/v2 v2.4.0
	github.com/kr/pty v1.1.3 // indirect
	github.com/magiconair/properties v1.8.1 // indirect
	github.com/markbates/going v1.0.3 // indirect
	github.com/pelletier/go-toml v1.4.0 // indirect
	github.com/pkg/errors v0.8.1
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/ugorji/go v1.1.5-pre // indirect
	github.com/unrolled/secure v1.0.0
	go.opencensus.io v0.22.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	k8s.io/apimachinery v0.0.0-20190624085041-961b39a1baa0
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/kube-openapi v0.0.0-20190603182131-db7b694dc208 // indirect
)
