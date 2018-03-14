source ../common/parse-or-gather-user-input.sh "${@}"

kubectl create ns $_arg_pcp_namespace

source ../common/rbac.yaml.sh
source ../common/protoform.yaml.sh

set -x
kubectl create -f rbac.yaml -n $_arg_pcp_namespace
kubectl create -f protoform.yaml -n $_arg_pcp_namespace
