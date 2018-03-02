#!/bin/bash

if env | grep -q SKIP_PREINSTALL ; then
  echo "skipping preinstall"
else
  ./pre-install.sh
fi

CONCURRENT_SCAN=2
DEF_HUBPORT=443
DEF_HUBUSER="sysadmin"

clear
echo " "
echo "============================================"
echo "Black Duck Hub Configuration Information"
echo "============================================"
read -p "Hub server host (e.g. hub.mydomain.com): " hubHost
read -p "Hub server port [$DEF_HUBPORT]: " hubPort
read -p "Hub user name [$DEF_HUBUSER]: " hubUser
read -sp "Hub user password: " hubPassword
echo " "
read -p "Maximum concurrent scans [$CONCURRENT_SCAN]: " noOfConcurrentScan

#apply defaults
hubPort="${hubPort:-$DEF_HUBPORT}"
hubUser="${hubUser:-$DEF_HUBUSER}"
noOfConcurrentScan="${noOfConcurrentScan:-$CONCURRENT_SCAN}"

DOCKER_PASSWORD=$(oc sa get-token perceptor-scanner-sa)

cat << EOF > config.yml
apiVersion: v1
kind: List
metadata:
  name: viper-input
items:
- apiVersion: v1
  kind: ConfigMap
  metadata:
    name: viper-input
  data:
    protoform.yaml: |
      DockerPasswordOrToken: "$DOCKER_PASSWORD"
      HubHost: "$hubHost"
      HubPort: "$hubPort"
      HubUser: "$hubUser"
      HubUserPassword: "$hubPassword"
      ConcurrentScanLimit: "$noOfConcurrentScan"
      DockerUsername: "admin"
EOF

oc create -f config.yml

echo "Your configuration is at config.yml -- hit return to proceed installing, or edit it before continuing"
read -s

cat << EOF > protoform.yml
apiVersion: v1
kind: Pod
metadata:
  name: protoform
spec:
  volumes:
  - name: viper-input
    configMap:
      name: viper-input
  containers:
  - name: protoform
    image: gcr.io/gke-verification/blackducksoftware/perceptor-protoform:latest
    imagePullPolicy: Always
    command: [ ./protoform ]
    ports:
    - containerPort: 3001
      protocol: TCP
    volumeMounts:
    - name: viper-input
      mountPath: /etc/protoform/
  restartPolicy: Never
  serviceAccountName: perceptor-protoform-sa
  serviceAccount: perceptor-protoform-sa
EOF

oc create -f protoform.yml

#./post-hack.sh
