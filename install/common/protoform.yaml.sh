#!/bin/bash

cat << EOF > protoform.yaml
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
    image: ${_arg_pcp_container_registry}/perceptor-protoform:${_arg_pcp_container_version}
    imagePullPolicy: Always
    command: [ ./protoform ]
    ports:
    - containerPort: 3001
      protocol: TCP
    volumeMounts:
    - name: viper-input
      mountPath: /etc/protoform/
  restartPolicy: Never
  serviceAccountName: protoform
  serviceAccount: protoform
---
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
      DockerPasswordOrToken: "$_arg_scanned_registry_token"
      HubHost: "$_arg_hub_host"
      HubUser: "$_arg_hub_user"
      HubPort: "$_arg_hub_port"
      # TODO, inject as secret.
      HubUserPassword: "$_arg_hub_password"
      ConcurrentScanLimit: "$_arg_hub_max_concurrent_scans"
      DockerUsername: "admin"
      Namespace: "$_arg_pcp_namespace"
      Openshift: "false"
      Registry: "$_arg_scanned_registry"
      ImagePath: "$_arg_imagepath"

      # TODO: Assuming for now that we run the same version of everything
      # For the curated installers.  For developers ? You might want to
      # hard code one of these values if using this script for dev/test.
      PerceptorContainerVersion: "$_arg_pcp_container_version"
      ScannerContainerVersion: "$_arg_pcp_container_version"
      PerceiverContainerVersion: "$_arg_pcp_container_version"
      ImageFacadeContainerVersion: "$_arg_pcp_container_version"

      DefaultCPU: "$_arg_pcp_container_default_cpu"
      DefaultMem: "$_arg_pcp_container_default_memory"
EOF
