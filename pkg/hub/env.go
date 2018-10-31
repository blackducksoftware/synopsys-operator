package hub

import (
	"strings"

	"github.com/sirupsen/logrus"
)

const envOptions = `
# CATALINA_OPTS="-Dparameter1=value1 -Dparameter2=value2",
# Change this to 1 if you are also using Hub Alert.
# Change this to 1 if you are also using Hub Binary uploads
# Cross Origin Resource Sharing (CORS) configuration
# Define TLS_PROTOCOLS in order to override supported protocols (default: TLSv1.2)
# Do not change
# For NTLM Proxies
# HUB_PROXY_PASSWORD=
# Hosts that won't go through a proxy.
# If IPv6 is disabled for a host machine, this flag should be set to '1' so that NGiNX
# If the container port is changed, the Hub webserver port value should be equally changed.
# If the host port is changed, the public Hub webserver port value should be equally changed.
# JOBRUNNER_OPTS="-Dparameter1=value1 -Dparameter2=value2"
# See Hub Alert documentation for details.
# Set to true or false
# Specific settings for Tomcat-based containers can set via:
# Specify any property-specific overrides here
# TO disable hub dashboard
# The proxy password can be specified here if it is not specified in a
# The public Hub webserver host will be served as a subject alternative name (SAN) within the 
# This will enable Alert routes in 'webserver'
# To access API documentation when using a reverse proxy with Hub mounted under a sub-path
# Whether or not to include ignored components in the reports
# default, served certificate.   This value should be changed to the publically-facing 
# for instance:
# hostname that users will enter in their browser in order to access Hub.
# separate mounted file or secret.
# specify where RabbitMQ is at
# used by uploadcache service
# used by worker and rabbitmq
# while those for the Job Runner can be set via:
# will not listen on an IPv6 address. By default NGiNX will listen on IPv4 and IPv6.
#TLS_PROTOCOLS=TLSv1.1 TLSv1.2
#hub-bdba.env
#hub-postgres.env
#hub-proxy.env
ALLOW_ACCESS_DIRECTIVES=
BLACKDUCK_CORS_ALLOWED_HEADERS_PROP_NAME=
BLACKDUCK_CORS_ALLOWED_ORIGINS_PROP_NAME=
BLACKDUCK_CORS_EXPOSED_HEADERS_PROP_NAME=
BLACKDUCK_HUB_CORS_ENABLED=
BLACKDUCK_REPORT_IGNORED_COMPONENTS=false
BLACKDUCK_SWAGGER_DISPLAYALL=
BLACKDUCK_SWAGGER_PROXY_PREFIX=
BROKER_URL=amqps://rabbitmq/protecodesc
BROKER_USE_SSL=yes
CFSSL=cfssl:8888
CLIENT_CERT_CN=binaryscanner
DENY_ACCESS_DIRECTIVES=
DISABLE_HUB_DASHBOARD=#hub-webserver.env
HTTPS_VERIFY_CERTS=yes
HUB_LOGSTASH_HOST=logstash
HUB_POSTGRES_ADMIN=blackduck
HUB_POSTGRES_ENABLE_SSL="false"
HUB_POSTGRES_HOST=
HUB_POSTGRES_PORT=
HUB_POSTGRES_USER=blackduck_user
HUB_PROXY_DOMAIN=
HUB_PROXY_HOST=
HUB_PROXY_NON_PROXY_HOSTS=solr
HUB_PROXY_PORT=
HUB_PROXY_SCHEME=
HUB_PROXY_USER=
HUB_PROXY_WORKSTATION=
HUB_VERSION=5.0.2
HUB_WEBSERVER_PORT=8443
IPV4_ONLY=0
PUBLIC_HUB_WEBSERVER_HOST=localhost
PUBLIC_HUB_WEBSERVER_PORT=443
RABBITMQ_DEFAULT_VHOST=protecodesc
RABBITMQ_SSL_FAIL_IF_NO_PEER_CERT=false
RABBIT_MQ_HOST=rabbitmq
RABBIT_MQ_PORT=5671
SCANNER_CONCURRENCY=1
USE_ALERT=0
USE_BINARY_UPLOADS=1
~/go/src/github.com/blackducksoftware/perceptor-protoform/hack
~/go/src/github.com/blackducksoftware/perceptor-protoform/hack/hub/docker-compose ~/go/src/github.com/blackducksoftware/perceptor-protoform/hack
image: blackducksoftware/appcheck-worker:1.0.1
image: blackducksoftware/blackduck-upload-cache:1.0.2
image: blackducksoftware/hub-authentication:5.0.2
image: blackducksoftware/hub-cfssl:5.0.2
image: blackducksoftware/hub-documentation:5.0.2
image: blackducksoftware/hub-jobrunner:5.0.2
image: blackducksoftware/hub-logstash:5.0.2
image: blackducksoftware/hub-nginx:5.0.2
image: blackducksoftware/hub-postgres:5.0.2
image: blackducksoftware/hub-registration:5.0.2
image: blackducksoftware/hub-scan:5.0.2
image: blackducksoftware/hub-solr:5.0.2
image: blackducksoftware/hub-webapp:5.0.2
image: blackducksoftware/hub-zookeeper:5.0.2
image: blackducksoftware/rabbitmq:1.0.0`

func GetHubKnobs() (env map[string]string, images []string) {
	env = map[string]string{}
	images = []string{}
	logrus.Infof("%v", len(strings.Split(envOptions, "\n")))

	for _, val := range strings.Split(envOptions, "\n") {
		if strings.Contains(val, "=") {
			keyval := strings.Split(val, "=")
			env[keyval[0]] = keyval[1]
		} else if strings.Contains(val, "image") {
			fullImage := strings.Split(val, ": ")
			images = append(images, fullImage[1])
		} else {
			logrus.Infof("Skipping line %v", val)
		}
	}
	logrus.Infof("%v \n %v", images, env)
	return env, images
}
