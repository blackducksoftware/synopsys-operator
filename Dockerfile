FROM golang:1.8

WORKDIR /go/
COPY . /go/src/github.com/blackducksoftware/perceptor-protoform
WORKDIR /go/src/github.com/blackducksoftware/perceptor-protoform
RUN cd cmd/blackduckctl ; go build ./ ; cp blackduckctl /bin/blackduckctl
RUN cd cmd/operator ; go build ./ ; cp operator /bin/blackduck-operator

### ? Senthil, see what you can do to make buffalo build below?
# RUN cd cmd/operator-ui ; .... 
