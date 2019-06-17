module github.com/blackducksoftware/synopsys-operator/cmd/operator-ui

go 1.12

require (
	github.com/NYTimes/gziphandler v1.1.1 // indirect
	github.com/blackducksoftware/synopsys-operator v0.0.0-20190614224807-8d080a4e981c
	github.com/codegangsta/negroni v1.0.0 // indirect
	github.com/coreos/bbolt v1.3.3 // indirect
	github.com/coreos/etcd v3.3.13+incompatible // indirect
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/go-systemd v0.0.0-20190612170431-362f06ec6bc1 // indirect
	github.com/emicklei/go-restful v2.9.6+incompatible // indirect
	github.com/go-openapi/spec v0.19.2 // indirect
	github.com/gobuffalo/buffalo v0.14.4
	github.com/gobuffalo/envy v1.7.0
	github.com/gobuffalo/makr v1.1.5 // indirect
	github.com/gobuffalo/mw-csrf v0.0.0-20190129204204-25460a055517
	github.com/gobuffalo/mw-forcessl v0.0.0-20190224202501-6d1ef7ffb276
	github.com/gobuffalo/mw-i18n v0.0.0-20190224203426-337de00e4c33
	github.com/gobuffalo/mw-paramlogger v0.0.0-20190224201358-0d45762ab655
	github.com/gobuffalo/packr/v2 v2.3.1
	github.com/gobuffalo/suite v2.6.2+incompatible
	github.com/golang/mock v1.3.1 // indirect
	github.com/google/pprof v0.0.0-20190515194954-54271f7e092f // indirect
	github.com/googleapis/gax-go/v2 v2.0.5 // indirect
	github.com/googleapis/gnostic v0.3.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.9.1 // indirect
	github.com/jackc/pgx v3.4.0+incompatible // indirect
	github.com/kisielk/errcheck v1.2.0 // indirect
	github.com/mattn/go-colorable v0.1.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20190414153302-2ae31c8b6b30 // indirect
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v1.0.0 // indirect
	github.com/prometheus/tsdb v0.7.1 // indirect
	github.com/rogpeppe/fastuuid v1.1.0 // indirect
	github.com/russross/blackfriday v2.0.0+incompatible // indirect
	github.com/sirupsen/logrus v1.4.2
	github.com/unrolled/secure v1.0.0
	go.etcd.io/bbolt v1.3.3 // indirect
	golang.org/x/exp v0.0.0-20190510132918-efd6b22b2522 // indirect
	golang.org/x/image v0.0.0-20190616094056-33659d3de4f5 // indirect
	golang.org/x/mobile v0.0.0-20190607214518-6fa95d984e88 // indirect
	golang.org/x/mod v0.1.0 // indirect
	honnef.co/go/tools v0.0.0-20190614002413-cb51c254f01b // indirect
	k8s.io/api v0.0.0-20190615205754-1d1b8b084b30 // indirect
	k8s.io/apiextensions-apiserver v0.0.0-20190615210511-390f8f388302 // indirect
	k8s.io/apimachinery v0.0.0-20190612125636-6a5db36e93ad
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/gengo v0.0.0-20190327210449-e17681d19d3a // indirect
)

replace (
	github.com/golang/lint => github.com/golang/lint v0.0.0-20190409202823-5614ed5bae6fb75893070bdc0996a68765fdd275
	k8s.io/api => k8s.io/api v0.0.0-20190313235455-40a48860b5ab
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190315093550-53c4693659ed
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190313205120-d7deff9243b1
	sourcegraph.com/sourcegraph/go-diff => sourcegraph.com/sourcegraph/go-diff v0.5.0
)
