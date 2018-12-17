#!/bin/bash

unset DYLD_INSERT_LIBRARIES

source ../common/args.sh "${@}"

echo "Using the secret encoded in ../common/secret.json.  Edit the file before running, or press enter to continue with the defaults."
read x

CLUSTER_BINDING_NS=$(kubectl get clusterrolebindings synopsys-operator-admin -o go-template='{{range .subjects}}{{.namespace}}{{" "}}{{end}}' 2> /dev/null)
if [[ $? -eq 0 ]]; then
    SCRIPT_DIR=$(dirname "$0")
    echo "You have already installed the operator in namespace $CLUSTER_BINDING_NS. Please run the cleanup script ($SCRIPT_DIR/cleanup.sh) before attempting to install the operator in $_arg_namespace"
    exit 1
fi

kubectl create ns $_arg_namespace

kubectl create -f ../common/secret.json -n $_arg_namespace

cat ../common/synopsys-operator.yaml | \
sed 's/${REGISTRATION_KEY}/'$_arg_blackduck_registration_key'/g' | \
sed 's/${NAMESPACE}/'$_arg_namespace'/g' | \
sed 's/${IMAGE}/'$(echo $_arg_synopsys_operator_image | sed -e 's/\\/\\\\/g; s/\//\\\//g; s/&/\\\&/g')'/g' | \
sed 's/${PROMETHEUS_IMAGE}/'$(echo $_arg_prometheus_image | sed -e 's/\\/\\\\/g; s/\//\\\//g; s/&/\\\&/g')'/g' | \
kubectl create --namespace=$_arg_namespace -f -

if [[ ! -z "$_arg_docker_config" ]]; then
  kubectl create secret generic custom-registry-pull-secret --from-file=.dockerconfigjson="$_arg_docker_config" --type=kubernetes.io/dockerconfigjson
  kubectl secrets link default custom-registry-pull-secret --for=pull
  kubectl secrets link synopsys-operator custom-registry-pull-secret --for=pull; 
  kubectl scale rc synopsys-operator --replicas=0
  kubectl scale rc synopsys-operator --replicas=1
fi

echo "Done deploying!"
echo
kubectl get pods -n $_arg_namespace 
echo
echo "Click a key to expose the LoadBalancer. (This will only work in supported kubernetes clouds.)"
read x

kubectl expose rc synopsys-operator --port=80 --target-port=3000 --name=synopsys-operator-tcp --type=LoadBalancer --namespace=${_arg_namespace}

kubectl get svc -n $_arg_namespace
