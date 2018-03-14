#!/bin/bash
source ../common/parse-or-gather-user-input.sh "${@}"

_arg_openshift_perceiver="on"

oc new-project $_arg_pcp_namespace

source ../common/oadm-policy-init.sh $arg_pcp_namespace

source ../common/protoform.yaml.sh
#oc project $_arg_pcp_namespace
oc create -f protoform.yaml
