CURRENT_DIR:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

.PHONY: perceptor

perceptor:
	docker run --rm -e CGO_ENABLED=0 -e GOOS=linux -e GOARCH=amd64 -v "${CURRENT_DIR}":/go/src/github.com/blackducksoftware/perceptor-protoform/contrib/hydra -w /go/src/github.com/blackducksoftware/perceptor-protoform/contrib/hydra/standardperceptor golang:1.10 go build -o perceptor
