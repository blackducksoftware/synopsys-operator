FROM gobuffalo/buffalo:v0.13.5 as builder

# Set the environment
ENV BP=$GOPATH/src/github.com/blackducksoftware/synopsys-operator

# Add the whole directory
ADD . $BP

### BUILD THE BINARIES...
WORKDIR $BP

# RUN cd cmd/blackduckctl && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/blackduckctl
RUN cd cmd/operator && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/operator

### BUILD THE UI
WORKDIR $BP/cmd/operator-ui
RUN yarn install --no-progress && mkdir -p public/assets && go get $(go list ./... | grep -v /vendor/) && buffalo build --static -o /bin/app

# Container catalog requirements
COPY ./LICENSE /bin/LICENSE 
COPY ./help.1 /bin/help.1

FROM alpine

MAINTAINER Synopsys Cloud Native Team

ARG VERSION
ARG BUILDTIME
ARG LASTCOMMIT

RUN apk add --no-cache curl
RUN apk add --no-cache bash
RUN apk add --no-cache ca-certificates

# Uncomment to run the binary in "production" mode:
# ENV GO_ENV=production

# Bind the app to 0.0.0.0 so it can be seen from outside the container
# ENV ADDR=0.0.0.0

COPY --from=builder /bin/app .
# COPY --from=builder /bin/blackduckctl .
COPY --from=builder /bin/operator .
COPY --from=builder /bin/LICENSE /licenses/
COPY --from=builder /bin/help.1 /help.1

RUN chmod 777 ./app && chmod 777 ./operator

LABEL name="Synopsys Operator" \
      vendor="Synopsys" \
      release.version="$VERSION" \
      summary="Synopsys Operator" \
      description="This container is used to deploy the Synopsys Operators." \
      lastcommit="$LASTCOMMIT" \
      buildtime="$BUILDTIME" \
      license="apache" \
      release="$VERSION" \
      version="$VERSION"

CMD ./app
