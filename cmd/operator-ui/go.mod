module github.com/blackducksoftware/synopsys-operator/cmd/operator-ui

go 1.12

require (
	github.com/blackducksoftware/synopsys-operator v0.0.0-20190604155518-8cf99e3a95dc
	github.com/codegangsta/negroni v1.0.0 // indirect
	github.com/gobuffalo/buffalo v0.14.4
	github.com/gobuffalo/envy v1.7.0
	github.com/gobuffalo/makr v1.1.5 // indirect
	github.com/gobuffalo/mw-csrf v0.0.0-20190129204204-25460a055517
	github.com/gobuffalo/mw-forcessl v0.0.0-20190224202501-6d1ef7ffb276
	github.com/gobuffalo/mw-i18n v0.0.0-20190224203426-337de00e4c33
	github.com/gobuffalo/mw-paramlogger v0.0.0-20190224201358-0d45762ab655
	github.com/gobuffalo/packr/v2 v2.3.1
	github.com/gobuffalo/suite v2.6.2+incompatible
	github.com/jackc/pgx v3.4.0+incompatible // indirect
	github.com/mattn/go-colorable v0.1.2 // indirect
	github.com/pkg/errors v0.8.1
	github.com/sirupsen/logrus v1.4.2
	github.com/unrolled/secure v1.0.0
	golang.org/x/tools v0.0.0-20190603231351-8aaa1484dc10 // indirect
	k8s.io/apimachinery v0.0.0-20190602183612-63a6072eb563
	k8s.io/client-go v11.0.0+incompatible
)

replace (
	github.com/golang/lint => github.com/golang/lint v0.0.0-20190409202823-5614ed5bae6fb75893070bdc0996a68765fdd275
	k8s.io/api => k8s.io/api v0.0.0-20190313235455-40a48860b5ab
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190315093550-53c4693659ed
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190313205120-d7deff9243b1
	sourcegraph.com/sourcegraph/go-diff => sourcegraph.com/sourcegraph/go-diff v0.5.0
)
