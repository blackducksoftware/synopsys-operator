#!/bin/bash

unset DYLD_INSERT_LIBRARIES

source ../common/args.sh "${@}"

echo "Using the secret encoded in ../common/secret.json.  Edit the file before running, or press enter to continue with the defaults."
read x

oc new-project $_arg_namespace

oc create -f ../common/secret.json -n $_arg_namespace

cat ../common/synopsys-operator.yaml | \
sed 's/${REGISTRATION_KEY}/'$_arg_blackduck_registration_key'/g' | \
sed 's/${NAMESPACE}/'$_arg_namespace'/g' | \
sed 's/${IMAGE}/'$(echo $_arg_image | sed -e 's/\\/\\\\/g; s/\//\\\//g; s/&/\\\&/g')'/g' | \
oc create --namespace=$_arg_namespace -f -

if [[ ! -z "$_arg_docker_config" ]]; then
  oc create secret generic custom-registry-pull-secret --from-file=.dockerconfigjson="$_arg_docker_config" --type=kubernetes.io/dockerconfigjson
  oc secrets link default custom-registry-pull-secret --for=pull
  oc secrets link synopsys-operator custom-registry-pull-secret --for=pull; 
  oc scale rc synopsys-operator --replicas=0
  oc scale rc synopsys-operator --replicas=1
fi

echo "Done deploying!"
echo
oc get pods -n $_arg_namespace 
echo
echo "Press any key to expose a route to the Synopsys Operator. (This will only work in supported openshift clouds.)"
read x

oc expose rc synopsys-operator --port=80 --target-port=3000 --name=synopsys-operator-tcp --type=LoadBalancer --namespace=${_arg_namespace}
oc create route edge --service=synopsys-operator-tcp -n $_arg_namespace

oc get svc -n $_arg_namespace
