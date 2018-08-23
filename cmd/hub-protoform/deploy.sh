#!/bin/bash

NS=$1

REG_KEY=$2

kubectl create ns $NS

cat hub-protoform.yaml | sed 's/${REGISTRATION_KEY}/'$REG_KEY'/g' | sed 's/${NAMESPACE}/'$NS'/g' | kubectl create --namespace=$NS -f -

kubectl expose rc hub-protoform --port=8080 --target-port=8080 --name=hub-protoform-np --type=NodePort --namespace=$NS

kubectl expose rc hub-protoform --port=8080 --target-port=8080 --name=hub-protoform-lb --type=LoadBalancer --namespace=$NS
