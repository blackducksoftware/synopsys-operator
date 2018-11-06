FROM gobuffalo/buffalo:v0.13.2 as builder

# Set the environment
ENV BP=$GOPATH/src/github.com/blackducksoftware/perceptor-protoform

# Add the whole directory
ADD . $BP

### BUILD THE BINARIES...
WORKDIR $BP

RUN cd cmd/blackduckctl ; CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/blackduckctl
RUN cd cmd/operator ; CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/operator

### BUILD THE UI
WORKDIR $BP/cmd/operator-ui
RUN yarn install --no-progress
RUN buffalo build --static -o /bin/app

FROM alpine

RUN apk add --no-cache curl
RUN apk add --no-cache bash
RUN apk add --no-cache ca-certificates

# Uncomment to run the binary in "production" mode:
# ENV GO_ENV=production

# Bind the app to 0.0.0.0 so it can be seen from outside the container
# ENV ADDR=0.0.0.0

COPY --from=builder /bin/app /bin/
COPY --from=builder /bin/blackduckctl /bin/
COPY --from=builder /bin/operator /bin/

RUN chmod 777 /bin/app
RUN chmod 777 /bin/blackduckctl
RUN chmod 777 /bin/operator
RUN ls -althr /bin/

EXPOSE 3000

CMD /bin/app
