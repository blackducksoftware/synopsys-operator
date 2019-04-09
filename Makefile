TAG="latest"
ifdef IMAGE_TAG
TAG="$(IMAGE_TAG)"
endif


ifneq (, $(findstring gcr.io,$(REGISTRY)))
PREFIX_CMD="gcloud"
DOCKER_OPTS="--"
endif

CURRENT_DIR := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
BUILD_TIME:=$(shell date)
LAST_COMMIT=$(shell git rev-parse HEAD)

OUTDIR = _output
LINUX = linux
WINDOWS = windows
PLATFORM := darwin linux windows

binary: clean ${OUTDIR} 
	$(foreach p,${PLATFORM}, \
		echo "creating synopsysctl binary for $(p) platform"; \
		if [[ $(p) = ${WINDOWS} ]]; then \
			docker run --rm -e CGO_ENABLED=0 -e GOOS=$(p) -e GOARCH=amd64 -v "${CURRENT_DIR}":/go/src/github.com/blackducksoftware/synopsys-operator -w /go/src/github.com/blackducksoftware/synopsys-operator/cmd/synopsysctl golang:1.11 go build -o /go/src/github.com/blackducksoftware/synopsys-operator/${OUTDIR}/synopsysctl.exe; \
		else \
			docker run --rm -e CGO_ENABLED=0 -e GOOS=$(p) -e GOARCH=amd64 -v "${CURRENT_DIR}":/go/src/github.com/blackducksoftware/synopsys-operator -w /go/src/github.com/blackducksoftware/synopsys-operator/cmd/synopsysctl golang:1.11 go build -o /go/src/github.com/blackducksoftware/synopsys-operator/${OUTDIR}/synopsysctl; \
		fi; \
		cd ${OUTDIR}; \
		if [[ $(p) = ${LINUX} ]]; then \
			tar -zcvf synopsysctl-$(p)-amd64.tar.gz synopsysctl; \
			shasum -a 256 synopsysctl-$(p)-amd64.tar.gz >> CHECKSUM; \
			rm -f synopsysctl; \
		elif [[ $(p) = ${WINDOWS} ]]; then \
			zip synopsysctl-$(p)-amd64.zip synopsysctl.exe; \
			shasum -a 256 synopsysctl-$(p)-amd64.zip >> CHECKSUM; \
			rm -f synopsysctl.exe; \
		else \
			zip synopsysctl-$(p)-amd64.zip synopsysctl; \
			shasum -a 256 synopsysctl-$(p)-amd64.zip >> CHECKSUM; \
			rm -f synopsysctl; \
		fi; \
		echo "completed synopsysctl binary for $(p) platform"; \
		cd ..; \
	)

clean:
	rm -rf ${OUTDIR}

${OUTDIR}:
	mkdir -p ${OUTDIR}
	touch ${OUTDIR}/CHECKSUM

init:
	brew install clang
	brew install dep
	brew install gcc
	brew install npm

container:
	docker build . --pull -t $(REGISTRY)/synopsys-operator:$(TAG) --build-arg VERSION=$(TAG) --build-arg 'BUILDTIME=$(BUILD_TIME)' --build-arg LASTCOMMIT=$(LAST_COMMIT)

push: container
	$(PREFIX_CMD) docker $(DOCKER_OPTS) push $(REGISTRY)/synopsys-operator:$(TAG)

dev:
	hack/local-up-perceptor.sh

lint:
	./hack/verify-gofmt.sh
	./hack/verify-golint.sh
	./hack/verify-govet.sh

build:
	docker run --rm -e CGO_ENABLED=0 -e GOOS=linux -e GOARCH=amd64 -v "${CURRENT_DIR}":/go/src/github.com/blackducksoftware/synopsys-operator -w /go/src/github.com/blackducksoftware/synopsys-operator golang:1.11 go build ./cmd/... ./pkg/...

test:
	docker run --rm -e CGO_ENABLED=0 -e GOOS=linux -e GOARCH=amd64 -v "${CURRENT_DIR}":/go/src/github.com/blackducksoftware/synopsys-operator -w /go/src/github.com/blackducksoftware/synopsys-operator golang:1.11 go test ./cmd/... ./pkg/...
