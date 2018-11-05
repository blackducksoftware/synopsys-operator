FROM gobuffalo/buffalo:v0.13.2 as builder

# Set the environment
ENV BP=$GOPATH/src/github.com/blackducksoftware/perceptor-protoform
ENV CGO_ENABLED=0
ENV GOOS=linux 
ENV GOARCH=amd64

# Add the whole directory
ADD . $BP

### BUILD THE BINARIES...
# COPY . $GOPATH/src/github.com/blackducksoftware/perceptor-protoform
WORKDIR $BP

RUN cd cmd/blackduckctl ; go build -o /bin/blackduckctl
RUN cd cmd/operator ; go build -o /bin/operator

### BUILD THE UI
WORKDIR $BP/cmd/operator-ui
# RUN npm rebuild node-sass
RUN yarn install --no-progress
# RUN go get $(go list ./... | grep -v /vendor/) 
RUN buffalo build --static -o /bin/app

FROM alpine

# Uncomment to run the binary in "production" mode:
# ENV GO_ENV=production

# Bind the app to 0.0.0.0 so it can be seen from outside the container
# ENV ADDR=0.0.0.0

COPY --from=builder /bin/app .
COPY --from=builder /bin/blackduckctl .
COPY --from=builder /bin/operator .

EXPOSE 3000

CMD ./app
