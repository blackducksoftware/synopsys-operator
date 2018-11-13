#!/bin/bash

unset DYLD_INSERT_LIBRARIES

echo "args = Namespace, Reg_key, branch"

NS=$1
REG_KEY=$2
BRANCH=$3

echo "Using the secret encoded in this file.  Change it before running, or press enter..."
read x

cat << EOF > /tmp/secret
apiVersion: v1
data:
  ADMIN_PASSWORD: YmxhY2tkdWNr
  POSTGRES_PASSWORD: YmxhY2tkdWNr
  USER_PASSWORD: YmxhY2tkdWNr
  HUB_PASSWORD: YmxhY2tkdWNr
kind: Secret
metadata:
  name: blackduck-secret
type: Opaque
EOF

oc new-project $NS

oc create -f /tmp/secret -n $NS

DOCKER_REGISTRY=gcr.io
DOCKER_REPO=saas-hub-stg/blackducksoftware

cat ../synopsys-operator.yaml | \
sed 's/${REGISTRATION_KEY}/'$REG_KEY'/g' | \
sed 's/${NAMESPACE}/'$NS'/g' | \
sed 's/${TAG}/'${VERSION}'/g' | \
sed 's/${DOCKER_REGISTRY}/'$DOCKER_REGISTRY'/g' | \
sed 's/${DOCKER_REPO}/'$(echo $DOCKER_REPO | sed -e 's/\\/\\\\/g; s/\//\\\//g; s/&/\\\&/g')'/g' | \
oc create --namespace=$NS -f -

echo "Done deploying!"
echo
oc get pods -n $NS 
echo
echo "Click a key to expose the LoadBalancer. (This will only work in supported kubernetes clouds.)"
read x

oc expose rc synopsys-operator --port=80 --target-port=3000 --name=synopsys-operator-tcp --type=LoadBalancer --namespace=${NS}

oc get svc -n $NS
