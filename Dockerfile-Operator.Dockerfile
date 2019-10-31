# Run: "docker run --rm -i hadolint/hadolint < Dockerfile" to ensure best practices!

# [STAGE: BUILD OPERATOR]
FROM golang:1.13-alpine as operatorbuilder

# Set the environment
ENV BP=$GOPATH/src/github.com/blackducksoftware/synopsys-operator

# Add the whole directory
COPY . ${BP}

# Container catalog requirements
COPY ./LICENSE /bin/LICENSE
COPY ./help.1 /bin/help.1

WORKDIR ${BP}/cmd/operator
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/operator


# [FINAL STAGE]
FROM scratch

LABEL maintainer="Synopsys Cloud Native Team"

ARG VERSION
ARG BUILDTIME
ARG LASTCOMMIT

COPY --from=operatorbuilder /bin/operator .
COPY --from=operatorbuilder /bin/LICENSE /licenses/
COPY --from=operatorbuilder /bin/help.1 /help.1

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

CMD ["./operator"]
