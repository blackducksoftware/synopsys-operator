FROM gobuffalo/buffalo:v0.14.3 as builder

ARG BINARY_VERSION

# Set the environment
ENV GO111MODULE=on
ENV BUFFALO_PLUGIN_CACHE=off
ENV BP=$GOPATH/src/github.com/blackducksoftware/synopsys-operator

# Add the Synopsys Operator repository
ADD . $BP

# Setting the work directory
WORKDIR $BP

### Build the Synopsys Operator binary
RUN cd cmd/operator && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$BINARY_VERSION" -o /bin/operator

### Build the Synopsys Operator UI
RUN cd cmd/operator-ui && yarn install --no-progress && mkdir -p public/assets && buffalo build --static -o /bin/app

# Container catalog requirements
COPY ./LICENSE /bin/LICENSE 
COPY ./help.1 /bin/help.1

FROM scratch

MAINTAINER Synopsys Cloud Native Team

ARG VERSION
ARG BUILDTIME
ARG LASTCOMMIT

# RUN apk add --no-cache curl
# RUN apk add --no-cache bash
# RUN apk add --no-cache ca-certificates

# Uncomment to run the binary in "production" mode:
# ENV GO_ENV=production

# Bind the app to 0.0.0.0 so it can be seen from outside the container
# ENV ADDR=0.0.0.0

COPY --from=builder /bin/app .
COPY --from=builder /bin/operator .
COPY --from=builder /bin/LICENSE /licenses/
COPY --from=builder /bin/help.1 /help.1

LABEL name="Synopsys Operator" \
    vendor="Synopsys" \
    release.version="$VERSION" \
    summary="Synopsys Operator" \
    description="This container is used to deploy Synopsys Operators." \
    lastcommit="$LASTCOMMIT" \
    buildtime="$BUILDTIME" \
    license="apache" \
    release="$VERSION" \
    version="$VERSION"

CMD ./app
