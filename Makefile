TAG=$(shell cat build.properties | cut -d'=' -f 2)
ifdef IMAGE_TAG
TAG="$(IMAGE_TAG)"
endif

SHA_SUM_CMD=/usr/bin/shasum -a 256
ifdef SHA_SUM
SHA_SUM_CMD="$(SHA_SUM)"
endif

CURRENT_DIR := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

OUTDIR = _output
LINUX = linux
WINDOWS = windows
PLATFORM := darwin linux windows

UI_BRANCH := master

binary: clean ${OUTDIR} 
	$(foreach p,${PLATFORM}, \
		echo "creating synopsysctl binary for $(p) platform" && \
		if [[ $(p) = ${WINDOWS} ]]; then \
			docker run --rm -e CGO_ENABLED=0 -e GOOS=$(p) -e GOARCH=amd64 -v "${CURRENT_DIR}":/go/src/github.com/blackducksoftware/synopsys-operator -w /go/src/github.com/blackducksoftware/synopsys-operator/cmd/synopsysctl golang:1.13 go build -ldflags "-X main.version=${TAG}" -o /go/src/github.com/blackducksoftware/synopsys-operator/${OUTDIR}/$(p)/synopsysctl.exe; \
		else \
			docker run --rm -e CGO_ENABLED=0 -e GOOS=$(p) -e GOARCH=amd64 -v "${CURRENT_DIR}":/go/src/github.com/blackducksoftware/synopsys-operator -w /go/src/github.com/blackducksoftware/synopsys-operator/cmd/synopsysctl golang:1.13 go build -ldflags "-X main.version=${TAG}" -o /go/src/github.com/blackducksoftware/synopsys-operator/${OUTDIR}/$(p)/synopsysctl; \
		fi && \
		echo "completed synopsysctl binary for $(p) platform"; \
	)

local-binary: clean ${OUTDIR} 
	$(foreach p,${PLATFORM}, \
		echo "creating synopsysctl binary for $(p) platform" && \
		if [[ $(p) = ${WINDOWS} ]]; then \
			CGO_ENABLED=0 GOOS=$(p) GOARCH=amd64 go build -ldflags "-X main.version=${TAG}" -o ./${OUTDIR}/$(p)/synopsysctl.exe ./cmd/synopsysctl; \
		else \
			CGO_ENABLED=0 GOOS=$(p) GOARCH=amd64 go build -ldflags "-X main.version=${TAG}" -o ./${OUTDIR}/$(p)/synopsysctl ./cmd/synopsysctl; \
		fi && \
		echo "completed synopsysctl binary for $(p) platform"; \
	) 

coverity: clean ${OUTDIR}
	mkdir -p /go/src/github.com/blackducksoftware && ln -s `pwd` /go/src/github.com/blackducksoftware/synopsys-operator
	$(foreach p,${PLATFORM}, \
		echo "creating synopsysctl binary for $(p) platform at `date`" && \
		cd /go/src/github.com/blackducksoftware/synopsys-operator/cmd/synopsysctl && \
		if [ $(p) = ${WINDOWS} ]; then \
			go build -ldflags "-X main.version=${TAG}" -o /go/src/github.com/blackducksoftware/synopsys-operator/${OUTDIR}/$(p)/synopsysctl.exe; \
		else \
			go build -ldflags "-X main.version=${TAG}" -o /go/src/github.com/blackducksoftware/synopsys-operator/${OUTDIR}/$(p)/synopsysctl; \
		fi && \
		echo "completed synopsysctl binary for $(p) platform at `date`" \
	)
	rm /go/src/github.com/blackducksoftware/synopsys-operator

package:
	$(foreach p,${PLATFORM}, \
		echo "creating synopsysctl package for $(p) platform" && \
		cd ${OUTDIR}/$(p) && \
		if [[ $(p) = ${LINUX} ]]; then \
			tar -zcvf synopsysctl-$(p)-amd64.tar.gz synopsysctl && mv synopsysctl-$(p)-amd64.tar.gz .. && cd .. && $(SHA_SUM_CMD) synopsysctl-$(p)-amd64.tar.gz >> CHECKSUM && rm -rf $(p); \
		elif [[ $(p) = ${WINDOWS} ]]; then \
			zip synopsysctl-$(p)-amd64.zip synopsysctl.exe && mv synopsysctl-$(p)-amd64.zip .. && cd .. && $(SHA_SUM_CMD) synopsysctl-$(p)-amd64.zip >> CHECKSUM && rm -rf $(p); \
		else \
			zip synopsysctl-$(p)-amd64.zip synopsysctl && mv synopsysctl-$(p)-amd64.zip .. && cd .. && $(SHA_SUM_CMD) synopsysctl-$(p)-amd64.zip >> CHECKSUM && rm -rf $(p); \
		fi && \
		echo "completed synopsysctl package for $(p) platform" && \
		cd ..; \
	)

clean:
	rm -rf ${OUTDIR}

${OUTDIR}:
	mkdir -p $(foreach p,${PLATFORM}, ${OUTDIR}/$(p))

lint:
	./hack/verify-gofmt.sh
	./hack/verify-golint.sh
	./hack/verify-govet.sh

build:
	go build -ldflags "-X main.version=${TAG}" -o synopsysctl ./cmd/synopsysctl

test:
	go test -ldflags "-X main.version=${TAG}" -o synopsysctl ./cmd/synopsysctl
