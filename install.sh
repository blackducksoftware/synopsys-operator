#!/bin/bash

./pre-install.sh

echo "Enter your password for the hub:"
read -s HUB_PASSWORD
echo "Done..."

DOCKER_PASSWORD=$(oc sa get-token perceptor-scanner-sa)

#
# Note that in production, you will want to encrypt the password, rather then inject it as a yml parameter.
# Instructions will be added shortly.
#
cat << EOF > config.yml
apiVersion: v1
kind: List
metadata:
  name: perceptor-configs
items:
- apiVersion: v1
  kind: ConfigMap
  metadata:
    name: prometheus
  data:
    prometheus.yml: |
      global:
        scrape_interval: 5s
      scrape_configs:
      - job_name: 'perceptor-scrape'
        scrape_interval: 5s
        static_configs:
        - targets: ['perceptor:3001', 'perceptor-scanner:3003'] # TODO Add perciever metrics here...
- apiVersion: v1
  kind: ConfigMap
  metadata:
    name: perceptor-scanner-config
  data:
    perceptor_scanner_conf.yaml: |
      HubHost: "34.227.106.252.xip.io"
      HubUser: "sysadmin"
      HubUserPassword: "$HUB_PASSWORD"
- apiVersion: v1
  kind: ConfigMap
  metadata:
    name: perceptor-imagefacade-config
  data:
    perceptor_imagefacade_conf.yaml: |
    DockerUser: "admin"
    DockerPassword: "$DOCKER_PASSWORD"
- apiVersion: v1
  kind: ConfigMap
  metadata:
    name: kube-generic-perceiver-config
  data:
    perceiver.yaml: |
      PerceptorHost: "perceptor"
      PerceptorPort: 3001
      AnnotationIntervalSeconds: 30
      DumpIntervalMinutes: 30
# TODO Replace w/ Secret creation.
- apiVersion: v1
  kind: ConfigMap
  metadata:
    name: perceptor-config
  data:
    perceptor_conf.yaml: |
      HubHost: "34.227.106.252.xip.io"
      HubUser: "sysadmin"
      HubUserPassword: "$HUB_PASSWORD"
- apiVersion: v1
  kind: ConfigMap
  metadata:
    name: openshift-perceiver-config
  data:
    perceiver.yaml: |
      PerceptorHost: "perceptor"
      PerceptorPort: 3001
      AnnotationIntervalSeconds: 30
      DumpIntervalMinutes: 30
EOF

oc create -f config.yml

echo "Your configuration is at config.yml, click enter to proceed installing, or edit it bbefore continuing"
cat << EOF > protoform.yml
apiVersion: v1
kind: Pod
metadata:
  name: protoform
spec:
  containers:
  - name: protoform
    image: gcr.io/gke-verification/blackducksoftware/perceptor-protoform:latest
    imagePullPolicy: Always
    command: [ ./protoform ]
    ports:
    - containerPort: 3001
      protocol: TCP
  restartPolicy: Never
  serviceAccountName: openshift-perceiver
  serviceAccount: openshift-perceiver
EOF

oc create -f protoform.yml

#./post-hack.sh
