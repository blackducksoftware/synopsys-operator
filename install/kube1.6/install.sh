source ../create-install-files.sh "${@}"

kubectl create ns $arg_pcp_namespace

source ../kube-rbac-sa-init.sh $arg_pcp_namespace

kubectl create -f config.yml -n $NAMESPACE
kubectl create -f proto.yml -n $NAMESPACE
