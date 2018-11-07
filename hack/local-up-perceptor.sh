#!/bin/bash
BDS=blackducksoftware

function verify() {
    if [[ "$OSTYPE" == "darwin"* ]]; then
        echo "OS looks ok : $OSTYPE"
    else
        echo "this is a utility for mac dev only.  If you are hacking on linux, run the raw containers w/ makefile on a kube cluster!"
        exit 2
    fi
    if [[ "$GOPATH" == "" ]] ; 
        then echo "gopath not found: $GOPATHs" ; 
        exit 2
    fi

    if which go ; then
        echo "go found."
    else
        exit 2
    fi

    if ! [ -x "$(command -v buffalo)" ]; then
        echo "setting up buffalo"
        brew install gobuffalo/tap/buffalo    
    fi
}

function run() {
    pushd $GOPATH/src/github.com/blackducksoftware/perceptor-protoform/cmd/operator-ui
        buffalo dev
    popd
}

verify
run
