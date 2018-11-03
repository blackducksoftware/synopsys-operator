FROM gobuffalo/buffalo:development as builder
ADD . $GOPATH/src/github.com/blackducksoftware/perceptor-protoform
ENV BP=$GOPATH/src/github.com/blackducksoftware/perceptor-protoform

### BUILD THE BINARIES...
WORKDIR $GOPATH
COPY . $GOPATH/src/github.com/blackducksoftware/perceptor-protoform
WORKDIR /go/src/github.com/blackducksoftware/perceptor-protoform
RUN cd cmd/blackduckctl ; go build ./ ; cp blackduckctl /bin/blackduckctl
RUN cd cmd/operator ; go build ./ ; cp operator /bin/blackduck-oper

### BUILD THE UI
WORKDIR $BP/cmd/operator-ui
RUN ls -altrh
RUN npm rebuild node-sass 
RUN buffalo build -v
#COPY --from=builder /bin/app .
EXPOSE 3000
CMD /bin/app
