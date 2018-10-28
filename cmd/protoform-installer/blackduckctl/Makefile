build:
	go build -o blackduckctl ./cmd
test:
	go test ./cmd
	go test ./pkg
publish: build
	git add blackduckctl
	git commit -m "blackduckctl auto commit"
	git push jayunit100 master