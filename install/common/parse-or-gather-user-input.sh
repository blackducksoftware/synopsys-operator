#!/bin/bash
source `dirname ${BASH_SOURCE}`/args.sh "${@}"

function prompt() {
  if [[ $_arg_proto_prompty == "on" ]]; then
    clear
    echo "============================================"
    echo "Blackduck Hub Configuration Information"
    echo "============================================"
    echo "Interactive"
    echo "============================================"
    read -p "Hub server host (e.g. hub.mydomain.com:443): " _arg_hub_host_port
    read -p "Hub user name (e.g. blackduck): " _arg_hub_user
    read -sp "Hub user password : " _arg_hub_password
    echo " "
    read -p "Maximum concurrent scans: " _arg_hub_max_concurrent_scans
    echo "============================================"
  else
    echo "Skipping prompts, --proto_prompty was turned off."
  fi
}

function create_protoform() {
cat << EOF > proto.yml
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
EOF
}

prompt
source `dirname ${BASH_SOURCE}`/protoform.yaml.sh
source `dirname ${BASH_SOURCE}`/rbac.yaml.sh
