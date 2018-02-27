.DEFAULT_GOAL := compile 

# protoform is the name of the secret WW2 people in captain america.
compile:
	docker run -t -i --rm -v $(shell pwd):/go/src/github.com/blackducksoftware/perceptor/ -w /go/src/github.com/blackducksoftware/perceptor -e CGO_ENABLED=0 -e GOOS=linux -e GOARCH=amd64 golang:1.9 go build -o build/protoform ./contrib/cmd/protoform_install/

test:
	go test ./pkg/...
