#!/bin/bash

set +x
NS=bds-perceptor

cleanup() {
	oc delete ns $NS
	while oc get ns | grep -q $NS ; do
	  echo "Waiting for deletion...`oc get ns | grep $NS` "
	  sleep 1
	done
}

cleanup

oc new-project $NS

cat << EOF > aux-config.json
{
	"Namespace": "$NS"
}
EOF

go run ./scannertester.go ./config.json aux-config.json
