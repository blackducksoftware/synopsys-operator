module github.com/blackducksoftware/synopsys-operator/cmd/operator-ui

go 1.12

replace (
	github.com/golang/lint => github.com/golang/lint v0.0.0-20190409202823-5614ed5bae6fb75893070bdc0996a68765fdd275
	k8s.io/api => k8s.io/api v0.0.0-20190313235455-40a48860b5ab
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190315093550-53c4693659ed
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190313205120-d7deff9243b1
	sourcegraph.com/sourcegraph/go-diff => sourcegraph.com/sourcegraph/go-diff v0.5.0
)

require (
	contrib.go.opencensus.io/exporter/ocagent v0.5.0 // indirect
	github.com/Azure/go-autorest/autorest v0.2.0 // indirect
	github.com/blackducksoftware/synopsys-operator v0.0.0-20190619142920-09b2da2fed54
	github.com/coreos/bbolt v1.3.3 // indirect
	github.com/coreos/etcd v3.3.13+incompatible // indirect
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/go-systemd v0.0.0-20190618135430-ff7011eec365 // indirect
	github.com/docker/spdystream v0.0.0-20181023171402-6480d4af844c // indirect
	github.com/elazarl/goproxy v0.0.0-20190421051319-9d40249d3c2f // indirect
	github.com/elazarl/goproxy/ext v0.0.0-20190421051319-9d40249d3c2f // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/gobuffalo/buffalo v0.14.6
	github.com/gobuffalo/buffalo-docker v1.1.0 // indirect
	github.com/gobuffalo/buffalo-pop v1.14.0 // indirect
	github.com/gobuffalo/envy v1.7.0
	github.com/gobuffalo/fizz v1.9.1 // indirect
	github.com/gobuffalo/github_flavored_markdown v1.1.0 // indirect
	github.com/gobuffalo/mw-csrf v0.0.0-20190129204204-25460a055517
	github.com/gobuffalo/mw-forcessl v0.0.0-20190224202501-6d1ef7ffb276
	github.com/gobuffalo/mw-i18n v0.0.0-20190224203426-337de00e4c33
	github.com/gobuffalo/mw-paramlogger v0.0.0-20190224201358-0d45762ab655
	github.com/gobuffalo/packr v1.26.0 // indirect
	github.com/gobuffalo/packr/v2 v2.4.0
	github.com/gobuffalo/suite v2.6.2+incompatible
	github.com/gobuffalo/x v0.0.0-20190614162758-d80e318e1bb4 // indirect
	github.com/golang/mock v1.3.1 // indirect
	github.com/google/gofuzz v1.0.0 // indirect
	github.com/google/pprof v0.0.0-20190515194954-54271f7e092f // indirect
	github.com/googleapis/gax-go/v2 v2.0.5 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.9.2 // indirect
	github.com/juju/loggo v0.0.0-20190526231331-6e530bcce5d8 // indirect
	github.com/karrick/godirwalk v1.10.12 // indirect
	github.com/kisielk/errcheck v1.2.0 // indirect
	github.com/kr/pty v1.1.5 // indirect
	github.com/magiconair/properties v1.8.1 // indirect
	github.com/markbates/deplist v1.2.0 // indirect
	github.com/nicksnyder/go-i18n v2.0.2+incompatible // indirect
	github.com/pelletier/go-toml v1.4.0 // indirect
	github.com/pkg/errors v0.8.1
	github.com/prometheus/common v0.6.0 // indirect
	github.com/rogpeppe/fastuuid v1.1.0 // indirect
	github.com/russross/blackfriday v2.0.0+incompatible // indirect
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/ugorji/go/codec v1.1.5-pre // indirect
	github.com/unrolled/secure v1.0.0
	go.etcd.io/bbolt v1.3.3 // indirect
	go.opencensus.io v0.22.0 // indirect
	golang.org/x/crypto v0.0.0-20190618222545-ea8f1a30c443 // indirect
	golang.org/x/exp v0.0.0-20190510132918-efd6b22b2522 // indirect
	golang.org/x/image v0.0.0-20190618124811-92942e4437e2 // indirect
	golang.org/x/mobile v0.0.0-20190607214518-6fa95d984e88 // indirect
	golang.org/x/mod v0.1.0 // indirect
	golang.org/x/net v0.0.0-20190619014844-b5b0513f8c1b // indirect
	golang.org/x/sys v0.0.0-20190618155005-516e3c20635f // indirect
	golang.org/x/tools v0.0.0-20190618233249-04b924abaa25 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/mgo.v2 v2.0.0-20180705113604-9856a29383ce // indirect
	honnef.co/go/tools v0.0.0-20190614002413-cb51c254f01b // indirect
	k8s.io/api v0.0.0-20190615205754-1d1b8b084b30 // indirect
	k8s.io/apiextensions-apiserver v0.0.0-20190615210511-390f8f388302 // indirect
	k8s.io/apimachinery v0.0.0-20190612125636-6a5db36e93ad
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/kube-openapi v0.0.0-20190603182131-db7b694dc208 // indirect
)
