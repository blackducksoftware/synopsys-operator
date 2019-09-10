OUTDIR = _output
LINUX = linux
MAC = darwin
WINDOWS = windows
PLATFORM := ${MAC} ${LINUX} ${WINDOWS}
BUILD_TIME:=$(shell date)
LAST_COMMIT=$(shell git rev-parse HEAD)

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

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

all: binary

binary: manager synopsysctl

# Run tests
test: generate fmt vet manifests
	GO111MODULE=on go test ./... -coverprofile cover.out

# Build manager binary
manager: clean generate fmt vet
	GO111MODULE=on go build -o ${OUTDIR}/bin/manager main.go

# Build synopsysctl binaries
synopsysctl: clean ${OUTDIR} 
	$(foreach p,${PLATFORM}, \
		echo "creating synopsysctl binary for $(p) platform" && \
		if [[ $(p) = ${WINDOWS} ]]; then \
			CGO_ENABLED=0 GOOS=$(p) GOARCH=amd64 GO111MODULE=on go build -i -ldflags "-X main.version=${TAG}" -o ${OUTDIR}/$(p)/synopsysctl.exe cmd/synopsysctl/synopsysctl.go ; \
		else \
			CGO_ENABLED=0 GOOS=$(p) GOARCH=amd64 GO111MODULE=on go build -i -ldflags "-X main.version=${TAG}" -o ${OUTDIR}/$(p)/synopsysctl cmd/synopsysctl/synopsysctl.go; \
		fi && \
		echo "completed synopsysctl binary for $(p) platform" \
	)

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
	GO111MODULE=on go fmt ./...

# Run go vet against code
vet:
	GO111MODULE=on go vet ./...

# Generate code
generate: controller-gen
	GO111MODULE=on $(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt paths="./..."

# Build the docker image
docker-build: test
	docker build . -t ${IMG}

# Push the docker image
docker-push:
	docker push ${IMG}

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	GO111MODULE=on go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.1
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

clean:
	rm -rf ${OUTDIR}
