#!/bin/bash

NS=$1

kubectl create ns $NS

kubectl create -f crd.yaml --namespace=$1

cat hub-installer.yaml | sed 's/${REGISTRATION_KEY}/'SETME'/g' | kubectl create --namespace=$NS -f -
