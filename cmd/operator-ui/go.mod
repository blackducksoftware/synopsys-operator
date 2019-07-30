module github.com/blackducksoftware/synopsys-operator/cmd/operator-ui

go 1.12

require (
	contrib.go.opencensus.io/exporter/ocagent v0.5.0 // indirect
	github.com/blackducksoftware/synopsys-operator v0.0.0-20190724190330-6507f575685b
	github.com/codegangsta/negroni v1.0.0 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/gobuffalo/buffalo v0.14.6
	github.com/gobuffalo/envy v1.7.0
	github.com/gobuffalo/mw-csrf v0.0.0-20190129204204-25460a055517
	github.com/gobuffalo/mw-forcessl v0.0.0-20190224202501-6d1ef7ffb276
	github.com/gobuffalo/mw-i18n v0.0.0-20190224203426-337de00e4c33
	github.com/gobuffalo/mw-paramlogger v0.0.0-20190224201358-0d45762ab655
	github.com/gobuffalo/packr/v2 v2.5.1
	github.com/gobuffalo/suite v2.7.0+incompatible
	github.com/jackc/pgx v3.3.0+incompatible // indirect
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
	k8s.io/apiextensions-apiserver v0.0.0-20190612130911-80dacc8982f1
	k8s.io/apimachinery v0.0.0-20190624085041-961b39a1baa0
	k8s.io/client-go v11.0.0+incompatible
)

replace (
	github.com/golang/lint => github.com/golang/lint v0.0.0-20190409202823-5614ed5bae6fb75893070bdc0996a68765fdd275
	k8s.io/api => k8s.io/api v0.0.0-20190313235455-40a48860b5ab
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190315093550-53c4693659ed
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190313205120-d7deff9243b1
	sourcegraph.com/sourcegraph/go-diff => sourcegraph.com/sourcegraph/go-diff v0.5.0
)
