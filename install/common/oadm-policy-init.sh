#!/bin/bash
CLUSTER="add-cluster-role-to-user"
NS=$_arg_pcp_namespace

# Protoform has its own SA.
oc create serviceaccount protoform -n $NS
oc adm policy $CLUSTER cluster-admin system:serviceaccount:$NS:protoform


# Get the default Docker Registry
route_docker_registry=$(oc get route docker-registry -n default -o jsonpath='{.spec.host}')
service_docker_registry=$(oc get svc docker-registry -n default -o jsonpath='{.spec.clusterIP}')
service_docker_registry_port=$(oc get svc docker-registry -n default -o jsonpath='{.spec.ports[0].port}')

defaultRegistries=()
if [[ ! -z "$route_docker_registry" ]]
then
  defaultRegistries+=("$route_docker_registry" "$route_docker_registry:443")
fi

if [[ ! -z "$service_docker_registry" ]]
then
  defaultRegistries+=("$service_docker_registry:$service_docker_registry_port" "docker-registry.default.svc:$service_docker_registry_port")
fi

i=0
_arg_private_registry="["
for dockerRegistry in "${defaultRegistries[@]}"
do
  if [ "$i" -eq "0" ]; then
    _arg_private_registry="$_arg_private_registry{\"Url\": \"$dockerRegistry\", \"User\": \"admin\", \"Password\": \"$_arg_private_registry_token\"}"
  else
    _arg_private_registry="$_arg_private_registry,{\"Url\": \"$dockerRegistry\", \"User\": \"admin\", \"Password\": \"$_arg_private_registry_token\"}"
  fi
  ((i++))
done
_arg_private_registry="$_arg_private_registry]"
