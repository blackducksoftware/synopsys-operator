init:
	brew install clang
	brew install dep
	brew install gcc
	brew install npm
jay:	
	brew install zsh
	brew install tmux
build:
	docker build ./
dev:
	hack/local-up-perceptor.sh
lint:
	./hack/verify-gofmt.sh
	./hack/verify-golint.sh
	./hack/verify-govet.sh
