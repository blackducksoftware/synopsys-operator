export GOPATH=/tmp/go
mkdir -p /tmp/go/src/github.com/blackducksoftware/
git clone github.com/blackducksoftware/horizon-hacking.git
cd $GOPATH/src/github.com/blackducksoftware/horizon-hacking/
go run cmd/main.go
