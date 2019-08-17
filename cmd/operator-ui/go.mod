module github.com/blackducksoftware/synopsys-operator/cmd/operator-ui

go 1.12

require (
	cloud.google.com/go v0.44.3 // indirect
	contrib.go.opencensus.io/exporter/ocagent v0.6.0 // indirect
	github.com/blackducksoftware/synopsys-operator v0.0.0-20190814123933-0fd5c95e9f03
	github.com/cockroachdb/apd v1.1.0 // indirect
	github.com/cockroachdb/cockroach-go v0.0.0-20181001143604-e0a95dfd547c // indirect
	github.com/elazarl/goproxy v0.0.0-20190711103511-473e67f1d7d2 // indirect
	github.com/elazarl/goproxy/ext v0.0.0-20190711103511-473e67f1d7d2 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/gobuffalo/buffalo v0.14.8
	github.com/gobuffalo/envy v1.7.0
	github.com/gobuffalo/mw-csrf v0.0.0-20190129204204-25460a055517
	github.com/gobuffalo/mw-forcessl v0.0.0-20190224202501-6d1ef7ffb276
	github.com/gobuffalo/mw-i18n v0.0.0-20190224203426-337de00e4c33
	github.com/gobuffalo/mw-paramlogger v0.0.0-20190224201358-0d45762ab655
	github.com/gobuffalo/packr v1.25.0 // indirect
	github.com/gobuffalo/packr/v2 v2.5.2
	github.com/gobuffalo/suite v2.8.1+incompatible
	github.com/gogo/protobuf v1.2.2-0.20190730201129-28a6bbf47e48 // indirect
	github.com/googleapis/gnostic v0.3.0 // indirect
	github.com/hashicorp/golang-lru v0.5.3 // indirect
	github.com/jackc/fake v0.0.0-20150926172116-812a484cc733 // indirect
	github.com/juju/loggo v0.0.0-20190526231331-6e530bcce5d8 // indirect
	github.com/juju/testing v0.0.0-20190723135506-ce30eb24acd2 // indirect
	github.com/magiconair/properties v1.8.1 // indirect
	github.com/pelletier/go-toml v1.4.0 // indirect
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_model v0.0.0-20190812154241-14fe0d1b01d4 // indirect
	github.com/satori/go.uuid v1.2.0 // indirect
	github.com/shopspring/decimal v0.0.0-20180709203117-cd690d0c9e24 // indirect
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/unrolled/secure v1.0.1
	golang.org/x/crypto v0.0.0-20190701094942-4def268fd1a4 // indirect
	golang.org/x/net v0.0.0-20190813141303-74dc4d7220e7 // indirect
	golang.org/x/sys v0.0.0-20190813064441-fde4db37ae7a // indirect
	google.golang.org/grpc v1.23.0 // indirect
	gopkg.in/mgo.v2 v2.0.0-20180705113604-9856a29383ce // indirect
	k8s.io/api v0.0.0-20190814101207-0772a1bdf941 // indirect
	k8s.io/apiextensions-apiserver v0.0.0-20190814104803-452bd7d41760 // indirect
	k8s.io/apimachinery v0.0.0-20190814100815-533d101be9a6
	k8s.io/client-go v11.0.0+incompatible
)

replace (
	cloud.google.com/go => cloud.google.com/go v0.44.0 // indirect
	contrib.go.opencensus.io/exporter/ocagent => contrib.go.opencensus.io/exporter/ocagent v0.6.0 // indirect
	github.com/NYTimes/gziphandler => github.com/NYTimes/gziphandler v1.1.1 // indirect
	github.com/blackducksoftware/synopsys-operator => github.com/blackducksoftware/synopsys-operator v0.0.0-20190815124934-88a08276dccb
	github.com/cockroachdb/cockroach-go => github.com/cockroachdb/cockroach-go v0.0.0-20181001143604-e0a95dfd547c // indirect
	github.com/coreos/bbolt => github.com/coreos/bbolt v1.3.3 // indirect
	github.com/coreos/etcd => github.com/coreos/etcd v3.3.13+incompatible // indirect
	github.com/coreos/go-semver => github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/go-systemd => github.com/coreos/go-systemd v0.0.0-20190719114852-fd7a80b32e1f // indirect
	github.com/emicklei/go-restful => github.com/emicklei/go-restful v2.9.6+incompatible // indirect
	github.com/gin-contrib/sse => github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/go-kit/kit => github.com/go-kit/kit v0.9.0 // indirect
	github.com/go-openapi/spec => github.com/go-openapi/spec v0.19.2 // indirect
	github.com/go-openapi/swag => github.com/go-openapi/swag v0.19.4 // indirect
	github.com/gobuffalo/buffalo => github.com/gobuffalo/buffalo v0.14.8
	github.com/gobuffalo/envy => github.com/gobuffalo/envy v1.7.0
	github.com/gobuffalo/mw-csrf => github.com/gobuffalo/mw-csrf v0.0.0-20190129204204-25460a055517
	github.com/gobuffalo/mw-forcessl => github.com/gobuffalo/mw-forcessl v0.0.0-20190224202501-6d1ef7ffb276
	github.com/gobuffalo/mw-i18n => github.com/gobuffalo/mw-i18n v0.0.0-20190224203426-337de00e4c33
	github.com/gobuffalo/mw-paramlogger => github.com/gobuffalo/mw-paramlogger v0.0.0-20190224201358-0d45762ab655
	github.com/gobuffalo/packr => github.com/gobuffalo/packr v1.25.0 // indirect
	github.com/gobuffalo/packr/v2 => github.com/gobuffalo/packr/v2 v2.5.2
	github.com/gogo/protobuf => github.com/gogo/protobuf v1.2.2-0.20190730201129-28a6bbf47e48 // indirect
	github.com/golang/groupcache => github.com/golang/groupcache v0.0.0-20190702054246-869f871628b6 // indirect
	github.com/golang/lint => github.com/golang/lint v0.0.0-20190409202823-5614ed5bae6fb75893070bdc0996a68765fdd275
	github.com/google/pprof => github.com/google/pprof v0.0.0-20190723021845-34ac40c74b70 // indirect
	github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.3.0 // indirect
	github.com/hashicorp/golang-lru => github.com/hashicorp/golang-lru v0.5.3 // indirect
	github.com/juju/testing => github.com/juju/testing v0.0.0-20190723135506-ce30eb24acd2 // indirect
	github.com/kr/pty => github.com/kr/pty v1.1.8 // indirect
	github.com/magiconair/properties => github.com/magiconair/properties v1.8.1 // indirect
	github.com/mailru/easyjson => github.com/mailru/easyjson v0.0.0-20190626092158-b2ccc519800e // indirect
	github.com/munnerz/goautoneg => github.com/munnerz/goautoneg v0.0.0-20190414153302-2ae31c8b6b30 // indirect
	github.com/mwitkow/go-conntrack => github.com/mwitkow/go-conntrack v0.0.0-20190716064945-2f068394615f // indirect
	github.com/pelletier/go-toml => github.com/pelletier/go-toml v1.4.0 // indirect
	github.com/pkg/errors => github.com/pkg/errors v0.8.1
	github.com/prometheus/client_model => github.com/prometheus/client_model v0.0.0-20190812154241-14fe0d1b01d4 // indirect
	github.com/rogpeppe/fastuuid => github.com/rogpeppe/fastuuid v1.2.0 // indirect
	github.com/russross/blackfriday => github.com/russross/blackfriday v2.0.0+incompatible // indirect
	github.com/sirupsen/logrus => github.com/sirupsen/logrus v1.4.2
	github.com/spf13/jwalterweatherman => github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/ugorji/go => github.com/ugorji/go v1.1.7 // indirect
	github.com/unrolled/secure => github.com/unrolled/secure v1.0.1
	go.etcd.io/bbolt => go.etcd.io/bbolt v1.3.3 // indirect
	golang.org/x/build => golang.org/x/build v0.0.0-20190111050920-041ab4dc3f9d // indirect
	golang.org/x/crypto => golang.org/x/crypto v0.0.0-20190701094942-4def268fd1a4 // indirect
	golang.org/x/mobile => golang.org/x/mobile v0.0.0-20190806162312-597adff16ade // indirect
	golang.org/x/net => golang.org/x/net v0.0.0-20190724013045-ca1201d0de80 // indirect
	golang.org/x/sys => golang.org/x/sys v0.0.0-20190812172437-4e8604ab3aff // indirect
	golang.org/x/tools => golang.org/x/tools v0.0.0-20190812191214-4147ede4f82b // indirect
	google.golang.org/grpc => google.golang.org/grpc v1.22.1 // indirect
	honnef.co/go/tools => honnef.co/go/tools v0.0.1-2019.2.2 // indirect
	k8s.io/api => k8s.io/api v0.0.0-20190313235455-40a48860b5ab
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190315093550-53c4693659ed
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190313205120-d7deff9243b1
	k8s.io/client-go => k8s.io/client-go v11.0.0+incompatible
	k8s.io/gengo => k8s.io/gengo v0.0.0-20190327210449-e17681d19d3a // indirect
	sigs.k8s.io/structured-merge-diff => sigs.k8s.io/structured-merge-diff v0.0.0-20190724202554-0c1d754dd648 // indirect
)
