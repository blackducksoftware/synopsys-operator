build:
	docker build ./
lint:
	./hack/verify-gofmt.sh
	./hack/verify-golint.sh
	./hack/verify-govet.sh