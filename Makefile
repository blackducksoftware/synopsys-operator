OUTDIR = _output

# Supported platforms
LINUX = linux
MAC = darwin
WINDOWS = windows

BUILD_TIME:=$(shell date)
LAST_COMMIT:=$(shell git rev-parse HEAD)
CURRENT_DIR := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
CONTROLLER_GEN=/go/bin/controller-gen
BINARY_TARGET=binary
BINARIES=manager synopsysctl
SCTL_BINARY_NAME=synopsysctl

# Set the release version information
TAG=$(shell cat build.properties | cut -d'=' -f 2)
ifdef IMAGE_TAG
TAG="$(IMAGE_TAG)"
endif

SHA_SUM_CMD=/usr/bin/shasum -a 256
ifdef SHA_SUM
SHA_SUM_CMD="$(SHA_SUM)"
endif

# Image URL to use all building/pushing image targets
IMG ?= blackducksoftware/synopsys-operator:dev
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

# Determine what synopsysctl binaries to build
ifneq ($(MAKECMDGOALS),${BINARY_TARGET})
  ifeq ($(OS),Windows_NT)
    PLATFORM=${WINDOWS}
  else
    UNAME:=$(shell uname)
    ifeq ($(UNAME), Linux)
      PLATFORM=${LINUX}
    else ifeq (Darwin,$(UNAME))
      PLATFORM=${MAC}
    endif
  endif
else
  PLATFORM = ${MAC} ${LINUX} ${WINDOWS}
endif

# Function to build synopsysctl
define sctl_build 
	@echo "creating synopsysctl binary for $(1) platform"

$(if $(findstring $(WINDOWS), $(1)), $(eval SCTL_BINARY_NAME=synopsysctl.exe))

$(if $(findstring ${BINARY_TARGET},$(MAKECMDGOALS)),
	docker run --rm -e CGO_ENABLED=0 -e GOOS=$(1) -e GOARCH=amd64 -e GO111MODULE=on -v "${CURRENT_DIR}":/go/src/github.com/blackducksoftware/synopsys-operator -w /go/src/github.com/blackducksoftware/synopsys-operator golang:1.13 go build -i -ldflags "-X main.version=${TAG}" -o /go/src/github.com/blackducksoftware/synopsys-operator/${OUTDIR}/$(1)/$(SCTL_BINARY_NAME) cmd/synopsysctl/synopsysctl.go,\
	CGO_ENABLED=0 GOOS=$(1) GOARCH=amd64 GO111MODULE=on go build -i -ldflags "-X main.version=${TAG}" -o ${OUTDIR}/$(1)/$(SCTL_BINARY_NAME) cmd/synopsysctl/synopsysctl.go)

	@echo "completed synopsysctl binary for $(1) platform"
endef

all: ${BINARIES}

$(BINARY_TARGET): ${BINARIES}

# Run tests
test: generate fmt vet manifests
	GO111MODULE=on go test ./... -coverprofile cover.out

# Build manager binary
manager: clean generate fmt vet
ifeq ($(MAKECMDGOALS),${BINARY_TARGET})
	docker run --rm -e GO111MODULE=on -v "${CURRENT_DIR}":/go/src/github.com/blackducksoftware/synopsys-operator -w /go/src/github.com/blackducksoftware/synopsys-operator golang:1.13 go build -o /go/src/github.com/blackducksoftware/synopsys-operator/${OUTDIR}/bin/manager main.go
else
	GO111MODULE=on go build -o ${OUTDIR}/bin/manager main.go
endif

# Build synopsysctl binaries
synopsysctl: clean generate fmt vet
	$(foreach p,${PLATFORM}, $(call sctl_build,$(p)))

# Build a release package information
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
		cd .. \
	)

container:
	docker build . --pull -t $(REGISTRY)/synopsys-operator:$(TAG) --build-arg VERSION=${TAG} --build-arg BINARY_VERSION=${TAG} --build-arg 'BUILDTIME=$(BUILD_TIME)' --build-arg LASTCOMMIT=$(LAST_COMMIT)

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	GO111MODULE=on go run ./main.go

# Install CRDs into a cluster
install: manifests
	kustomize build config/crd | kubectl apply -f -

# Delete CRDs from a cluster
delete-crd:
	kustomize build config/crd | kubectl delete -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests
	cd config/manager && kustomize edit set image controller=${IMG}
	kustomize build config/default | kubectl apply -f -

# Remove the controller in the configured Kubernetes cluster in ~/.kube/config
destroy: manifests
	cd config/manager && kustomize edit set image controller=${IMG}
	kustomize build config/default | kubectl delete -f -

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	GO111MODULE=on $(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases


# Run go fmt against code
fmt:
ifeq ($(MAKECMDGOALS),${BINARY_TARGET})
	docker run -e GO111MODULE=on -v "${CURRENT_DIR}":/go/src/github.com/blackducksoftware/synopsys-operator -w /go/src/github.com/blackducksoftware/synopsys-operator golang:1.13 go fmt ./...
else
	GO111MODULE=on go fmt ./...
endif

# Run go vet against code
vet:
ifeq ($(MAKECMDGOALS),${BINARY_TARGET})
	docker run -e GO111MODULE=on -v "${CURRENT_DIR}":/go/src/github.com/blackducksoftware/synopsys-operator -w /go/src/github.com/blackducksoftware/synopsys-operator golang:1.13 go vet ./...
else
	GO111MODULE=on go vet ./...
endif

# Generate code
generate: controller-gen
ifeq ($(MAKECMDGOALS),${BINARY_TARGET})
	docker run -e GO111MODULE=on -v "${CURRENT_DIR}":/go/src/github.com/blackducksoftware/synopsys-operator -w /go/src/github.com/blackducksoftware/synopsys-operator golang:1.13 /bin/bash -c "go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.1 && $(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt paths=\"./...\""
else
	GO111MODULE=on $(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt paths="./..."
endif

# Build the docker image
docker-build: test
	docker build . -t ${IMG}

# Push the docker image
docker-push:
	docker push ${IMG}

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifneq ($(MAKECMDGOALS),${BINARY_TARGET})
	GO111MODULE=on go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.1
CONTROLLER_GEN=$(shell which controller-gen)
endif

clean:
	rm -rf ${OUTDIR}
