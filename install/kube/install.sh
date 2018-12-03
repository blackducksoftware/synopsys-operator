#!/bin/bash

unset DYLD_INSERT_LIBRARIES

source ../common/args.sh "${@}"

echo "Using the secret encoded in ../common/secret.json.  Edit the file before running, or press enter to continue with the defaults."
read x

kubectl create ns $_arg_namespace

kubectl create -f ../common/secret.json -n $_arg_namespace

cat ../common/synopsys-operator.yaml | \
sed 's/${REGISTRATION_KEY}/'$_arg_blackduck_registration_key'/g' | \
sed 's/${NAMESPACE}/'$_arg_namespace'/g' | \
sed 's/${IMAGE}/'$(echo $_arg_image | sed -e 's/\\/\\\\/g; s/\//\\\//g; s/&/\\\&/g')'/g' | \
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
