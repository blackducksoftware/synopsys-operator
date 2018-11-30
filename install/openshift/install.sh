#!/bin/bash

unset DYLD_INSERT_LIBRARIES

echo "args = Namespace, Reg_key, Version of Operator"

DEFAULT_FILE_PATH="../common/default-values.json"

if [[ ! -z "$1" ]]; then
  DEFAULT_FILE_PATH="$1"
fi

array=( $(sed -n '/{/,/}/{s/[^:]*:[^"]*"\([^"]*\).*/\1/p;}' "$DEFAULT_FILE_PATH") ) 
NS=${array[0]}
REG_KEY=${array[1]}

echo "Using the secret encoded in ../common/secret.json and default values in ../common/default-values.json.  Edit the file before running, or press enter to continue with the defaults."
read x

CLUSTER_BINDING_NS=$(oc get clusterrolebindings synopsys-operator-admin -o go-template='{{range .subjects}}{{.namespace}}{{" "}}{{end}}' 2> /dev/null)
if [[ $? -eq 0 ]]; then
    SCRIPT_DIR=$(dirname "$0")
    echo "You have already installed the operator in namespace $CLUSTER_BINDING_NS. Please run the cleanup script ($SCRIPT_DIR/cleanup.sh) before attempting to install the operator in $1"
    exit 1
fi

oc new-project $NS

oc create -f ../common/secret.json -n $NS

cat ../common/synopsys-operator.yaml | \
sed 's/${REGISTRATION_KEY}/'$REG_KEY'/g' | \
sed 's/${NAMESPACE}/'$NS'/g' | \
sed 's/${IMAGE}/'$(echo ${array[2]} | sed -e 's/\\/\\\\/g; s/\//\\\//g; s/&/\\\&/g')'/g' | \
oc create --namespace=$NS -f -

if [[ ! -z "${array[3]}" ]]; then
  oc create secret generic custom-registry-pull-secret --from-file=.dockerconfigjson="${array[3]}" --type=kubernetes.io/dockerconfigjson
  oc secrets link default custom-registry-pull-secret --for=pull
  oc secrets link synopsys-operator custom-registry-pull-secret --for=pull; 
  oc scale rc synopsys-operator --replicas=0
  oc scale rc synopsys-operator --replicas=1
fi

echo "Done deploying!"
echo
oc get pods -n $NS 
echo
echo "Press any key to expose a route to the Synopsys Operator. (This will only work in supported openshift clouds.)"
read x

oc expose rc synopsys-operator --port=80 --target-port=3000 --name=synopsys-operator-tcp --type=LoadBalancer --namespace=${NS}

oc get svc -n $NS
