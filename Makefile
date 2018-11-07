build:
	docker build ./
dev:
	hack/local-up-perceptor.sh

lint:
	./hack/verify-gofmt.sh
	./hack/verify-golint.sh
	./hack/verify-govet.sh
