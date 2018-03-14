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
    image: ${_arg_imagepath}/perceptor-protoform:${_arg_protoform_version}
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
EOF
