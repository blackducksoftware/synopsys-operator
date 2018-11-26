#!/bin/bash

unset DYLD_INSERT_LIBRARIES

echo "args = Namespace, Reg_key, Version of Operator"

NS=$1
REG_KEY=$2
VERSION=$3

echo "Using the secret encoded in ../common/secret.yaml.  Edit the file before running, or press enter to continue with the defaults."
read x

oc new-project $NS

oc create -f ../common/secret.yaml -n $NS

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
echo "Press any key to expose a route to the Synopsys Operator. (This will only work in supported openshift clouds.)"
read x

oc expose rc synopsys-operator --port=80 --target-port=3000 --name=synopsys-operator-tcp --type=LoadBalancer --namespace=${NS}

oc get svc -n $NS
