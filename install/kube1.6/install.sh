source ../common/parse-or-gather-user-input.sh "${@}"

kubectl create ns $_arg_pcp_namespace

source ../common/rbac.yaml.sh $_arg_pcp_namespace
set -x
kubectl create -f protoform.yaml -n $_arg_pcp_namespace
kubectl create -f rbac.yaml -n $_arg_pcp_namespace
