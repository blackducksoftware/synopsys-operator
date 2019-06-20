package main

import (
	"os"

	"github.com/blackducksoftware/synopsys-operator/cmd/operator-ui/actions"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	log "github.com/sirupsen/logrus"
)

var version string

// main is the starting point for your Buffalo application.
// You can feel free and add to this `main` method, change
// what it does, etc...
// All we ask is that, at some point, you make sure to
// call `app.Serve()`, unless you don't want to start your
// application that is. :)
func main() {
	var configPath string
	var ok bool
	if configPath, ok = os.LookupEnv("CONFIG_FILE_PATH"); ok {
		log.Infof("Config path: %s", configPath)
	} else {
		log.Warn("no config file sent. running operator with environment variable and default settings")
	}

	if len(version) == 0 {
		if version, ok = os.LookupEnv("SYNOPSYS_OPERATOR_VERSION"); !ok {
			log.Warn("version is not set. please set the version in OPERATOR_VERSION environment variable")
		}
	} else {
		os.Setenv("SYNOPSYS_OPERATOR_VERSION", version)
	}
	log.Infof("version: %s", version)

	config, err := protoform.GetConfig(configPath, version)
	if err != nil {
		log.Panicf("unable to get the configuration due to %+v", err)
	}
	app := actions.App(config)
	if err := app.Serve(); err != nil {
		log.Fatal(err)
	}
}

/*
# Notes about `main.go`

## SSL Support

We recommend placing your application behind a proxy, such as
Apache or Nginx and letting them do the SSL heavy lifting
for you. https://gobuffalo.io/en/docs/proxy

## Buffalo Build

When `buffalo build` is run to compile your binary, this `main`
function will be at the heart of that binary. It is expected
that your `main` function will start your application using
the `app.Serve()` method.

*/
