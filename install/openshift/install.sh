#!/bin/bash

NS=$1
REG_KEY=$2
BRANCH=$3

kubectl create ns $NS

cat ../blackduck-protoform.yaml | sed 's/${REGISTRATION_KEY}/'$REG_KEY'/g' | sed 's/${NAMESPACE}/'$NS'/g' |sed 's/${BCH}/'${BRANCH}'/g' | kubectl create --namespace=$NS -f -

#kubectl expose rc blackduck-protoform --port=8080 --target-port=8080 --name=blackduck-protoform-np --type=NodePort --namespace=$NS

#kubectl expose rc blackduck-protoform --port=8080 --target-port=8080 --name=blackduck-protoform-lb --type=LoadBalancer --namespace=$NS
