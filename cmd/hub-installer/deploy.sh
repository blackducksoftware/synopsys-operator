#!/bin/bash

NS=$1

REG_KEY=$2

kubectl create ns $NS

cat hub-installer.yaml | sed 's/${REGISTRATION_KEY}/'$REG_KEY'/g' | sed 's/${NAMESPACE}/'$NS'/g' | kubectl create --namespace=$NS -f -

kubectl expose rc hub-installer --port=8080 --target-port=8080 --name=hub-installer-exposed --type=NodePort --namespace=$NS
