#!/bin/bash

unset DYLD_INSERT_LIBRARIES

DEFAULT_FILE_PATH="../common/default-values.json"

if [[ ! -z "$1" ]]; then
  DEFAULT_FILE_PATH="$1"
fi

array=( $(sed -n '/{/,/}/{s/[^:]*:[^"]*"\([^"]*\).*/\1/p;}' "$DEFAULT_FILE_PATH") ) 
NS=${array[0]}
REG_KEY=${array[1]}

echo "Using the secret encoded in ../common/secret.json and default values in ../common/default-values.json.  Edit the file before running, or press enter to continue with the defaults."
read x

kubectl create ns $NS

kubectl create -f ../common/secret.json -n $NS

cat ../common/synopsys-operator.yaml | \
sed 's/${REGISTRATION_KEY}/'$REG_KEY'/g' | \
sed 's/${NAMESPACE}/'$NS'/g' | \
sed 's/${IMAGE}/'$(echo ${array[2]} | sed -e 's/\\/\\\\/g; s/\//\\\//g; s/&/\\\&/g')'/g' | \
kubectl create --namespace=$NS -f -

if [[ ! -z "${array[3]}" ]]; then
  kubectl create secret generic custom-registry-pull-secret --from-file=.dockerconfigjson="${array[3]}" --type=kubernetes.io/dockerconfigjson
  kubectl secrets link default custom-registry-pull-secret --for=pull
  kubectl secrets link synopsys-operator custom-registry-pull-secret --for=pull; 
  kubectl scale rc synopsys-operator --replicas=0
  kubectl scale rc synopsys-operator --replicas=1
fi

echo "Done deploying!"
echo
kubectl get pods -n $NS 
echo
echo "Click a key to expose the LoadBalancer. (This will only work in supported kubernetes clouds.)"
read x

kubectl expose rc synopsys-operator --port=80 --target-port=3000 --name=synopsys-operator-tcp --type=LoadBalancer --namespace=${NS}

kubectl get svc -n $NS
