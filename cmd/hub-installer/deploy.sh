#!/bin/bash

NS=$1

kubectl create ns $NS

cat hub-installer.yaml | sed 's/${REGISTRATION_KEY}/'SETME'/g' | kubectl create --namespace=$NS -f -
