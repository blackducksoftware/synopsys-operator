# Run: "docker run --rm -i hadolint/hadolint < Dockerfile" to ensure best practices!

ARG BINARY_VERSION

# [STAGE: BUILD OPERATOR]
FROM golang:1.13-alpine as operatorbuilder

ENV BP=$GOPATH/src/github.com/blackducksoftware/synopsys-operator

# Add the whole Synopsys Operator repository
COPY . ${BP}

# Container catalog requirements
COPY ./LICENSE /bin/LICENSE
COPY ./help.1 /bin/help.1

# Build the Synopsys Operator binary
WORKDIR ${BP}/cmd/operator
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$BINARY_VERSION" -o /bin/operator

# [STAGE: BUILD OPERATOR-UI]

FROM gobuffalo/buffalo:v0.15.0 as operatoruibuilder

ENV GO111MODULE=on
ENV BUFFALO_PLUGIN_CACHE=off
ENV BP=$GOPATH/src/github.com/blackducksoftware/synopsys-operator

# Add the whole Synopsys Operator repository
COPY . $BP

# Build the Synopsys Operator UI
WORKDIR ${BP}/cmd/operator-ui
RUN yarn install --no-progress && mkdir -p public/assets && buffalo build --ldflags "-X main.version=$BINARY_VERSION" --static -o /bin/app

# [FINAL STAGE]
FROM scratch

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

COPY --from=operatorbuilder /bin/operator .
COPY --from=operatorbuilder /bin/LICENSE /licenses/
COPY --from=operatorbuilder /bin/help.1 /help.1

COPY --from=operatoruibuilder /bin/app .

LABEL name="Synopsys Operator" \
    maintainer="Synopsys Cloud Native Team" \
    vendor="Synopsys" \
    release.version="$VERSION" \
    summary="Synopsys Operator" \
    description="This image is used to deploy Synopsys Operator and Synopsy Operator user interface." \
    lastcommit="$LASTCOMMIT" \
    buildtime="$BUILDTIME" \
    license="apache" \
    release="$VERSION" \
    version="$VERSION"

CMD ["./app"]
