module github.com/blackducksoftware/synopsys-operator/cmd/operator-ui

go 1.12

replace (
	github.com/Azure/go-autorest/autorest => github.com/Azure/go-autorest/autorest v0.2.0
	github.com/blackducksoftware/synopsys-operator => github.com/blackducksoftware/synopsys-operator v0.0.0-20190813193533-b923eff16b81
	github.com/codegangsta/negroni => github.com/codegangsta/negroni v1.0.0
	github.com/gobuffalo/buffalo => github.com/gobuffalo/buffalo v0.14.6
	github.com/gobuffalo/envy => github.com/gobuffalo/envy v1.7.0
	github.com/gobuffalo/mw-csrf => github.com/gobuffalo/mw-csrf v0.0.0-20190129204204-25460a055517
	github.com/gobuffalo/mw-forcessl => github.com/gobuffalo/mw-forcessl v0.0.0-20190224202501-6d1ef7ffb276
	github.com/gobuffalo/mw-i18n => github.com/gobuffalo/mw-i18n v0.0.0-20190224203426-337de00e4c33
	github.com/gobuffalo/mw-paramlogger => github.com/gobuffalo/mw-paramlogger v0.0.0-20190224201358-0d45762ab655
	github.com/gobuffalo/packr/v2 => github.com/gobuffalo/packr/v2 v2.4.0
	github.com/gobuffalo/suite => github.com/gobuffalo/suite v2.8.1+incompatible
	github.com/golang/lint => github.com/golang/lint v0.0.0-20190409202823-5614ed5bae6fb75893070bdc0996a68765fdd275
	github.com/google/gofuzz => github.com/google/gofuzz v1.0.0
	github.com/kr/pty => github.com/kr/pty v1.1.3
	github.com/markbates/going => github.com/markbates/going v1.0.3
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
	cloud.google.com/go v0.44.3 // indirect
	contrib.go.opencensus.io/exporter/ocagent v0.6.0 // indirect
	github.com/Azure/go-autorest/autorest v0.8.0 // indirect
	github.com/Azure/go-autorest/autorest/adal v0.4.0 // indirect
	github.com/Azure/go-autorest/autorest/mocks v0.2.0 // indirect
	github.com/NYTimes/gziphandler v1.1.1 // indirect
	github.com/blackducksoftware/horizon v0.0.0-20190625151958-16cafa9109a3 // indirect
	github.com/blackducksoftware/synopsys-operator v0.0.0-20190813193533-b923eff16b81
	github.com/codegangsta/negroni v1.0.0 // indirect
	github.com/coreos/bbolt v1.3.3 // indirect
	github.com/coreos/etcd v3.3.13+incompatible // indirect
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/go-systemd v0.0.0-20190719114852-fd7a80b32e1f // indirect
	github.com/emicklei/go-restful v2.9.6+incompatible // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/go-kit/kit v0.9.0 // indirect
	github.com/go-openapi/spec v0.19.2 // indirect
	github.com/go-openapi/swag v0.19.4 // indirect
	github.com/gobuffalo/buffalo v0.14.6
	github.com/gobuffalo/envy v1.7.0
	github.com/gobuffalo/mw-csrf v0.0.0-20190129204204-25460a055517
	github.com/gobuffalo/mw-forcessl v0.0.0-20190224202501-6d1ef7ffb276
	github.com/gobuffalo/mw-i18n v0.0.0-20190224203426-337de00e4c33
	github.com/gobuffalo/mw-paramlogger v0.0.0-20190224201358-0d45762ab655
	github.com/gobuffalo/packr/v2 v2.5.1
	github.com/gobuffalo/suite v2.8.1+incompatible
	github.com/gogo/protobuf v1.2.2-0.20190730201129-28a6bbf47e48 // indirect
	github.com/golang/groupcache v0.0.0-20190702054246-869f871628b6 // indirect
	github.com/google/go-cmp v0.3.1 // indirect
	github.com/google/pprof v0.0.0-20190723021845-34ac40c74b70 // indirect
	github.com/gophercloud/gophercloud v0.3.0 // indirect
	github.com/hashicorp/golang-lru v0.5.3 // indirect
	github.com/juju/errors v0.0.0-20190806202954-0232dcc7464d // indirect
	github.com/juju/testing v0.0.0-20190723135506-ce30eb24acd2 // indirect
	github.com/kr/pty v1.1.8 // indirect
	github.com/lib/pq v1.2.0 // indirect
	github.com/magiconair/properties v1.8.1 // indirect
	github.com/mailru/easyjson v0.0.0-20190626092158-b2ccc519800e // indirect
	github.com/markbates/going v1.0.3 // indirect
	github.com/munnerz/goautoneg v0.0.0-20190414153302-2ae31c8b6b30 // indirect
	github.com/mwitkow/go-conntrack v0.0.0-20190716064945-2f068394615f // indirect
	github.com/pelletier/go-toml v1.4.0 // indirect
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v1.1.0 // indirect
	github.com/prometheus/client_model v0.0.0-20190812154241-14fe0d1b01d4 // indirect
	github.com/rogpeppe/fastuuid v1.2.0 // indirect
	github.com/rogpeppe/go-charset v0.0.0-20190617161244-0dc95cdf6f31 // indirect
	github.com/russross/blackfriday v2.0.0+incompatible // indirect
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/ugorji/go v1.1.7 // indirect
	github.com/unrolled/secure v1.0.0
	go.etcd.io/bbolt v1.3.3 // indirect
	golang.org/x/crypto v0.0.0-20190701094942-4def268fd1a4 // indirect
	golang.org/x/mobile v0.0.0-20190806162312-597adff16ade // indirect
	golang.org/x/net v0.0.0-20190813141303-74dc4d7220e7 // indirect
	golang.org/x/sys v0.0.0-20190813064441-fde4db37ae7a // indirect
	golang.org/x/tools v0.0.0-20190813142322-97f12d73768f // indirect
	google.golang.org/grpc v1.23.0 // indirect
	honnef.co/go/tools v0.0.1-2019.2.2 // indirect
	k8s.io/api v0.0.0-20190813020757-36bff7324fb7 // indirect
	k8s.io/apiextensions-apiserver v0.0.0-20190810101755-ebc439d6a67b
	k8s.io/apimachinery v0.0.0-20190813060636-0c17871ad6fd
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/gengo v0.0.0-20190813173942-955ffa8fcfc9 // indirect
	k8s.io/klog v0.4.0 // indirect
	k8s.io/kube-openapi v0.0.0-20190722073852-5e22f3d471e6 // indirect
	k8s.io/utils v0.0.0-20190809000727-6c36bc71fc4a // indirect
	sigs.k8s.io/structured-merge-diff v0.0.0-20190724202554-0c1d754dd648 // indirect
)
