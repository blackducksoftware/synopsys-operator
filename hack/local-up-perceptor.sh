#!/bin/bash
BDS=blackducksoftware

function verify() {
    if [[ "$GOPATH" == "" ]] ; 
        then echo "gopath not found: $GOPATHs" ; 
        exit 2
    fi

    if which go ; then
        echo "go found."
    else
        exit 2
    fi
}

function run() {
    pushd $GOPATH/src/github.com/$BDS/perceptor-protoform/
    go run cmd/protoform-installer/deployer.go
    popd
}

# startup
verify
run
