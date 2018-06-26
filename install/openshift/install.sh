#!/bin/bash
source ../common/parse-or-gather-user-input.sh "${@}"

_arg_image_perceiver="on"

oc new-project $_arg_pcp_namespace

source ../common/oadm-policy-init.sh $arg_pcp_namespace

source ../common/parse-image-registry.sh "../openshift/image-registry.json"

source ../common/protoform.yaml.sh

oc create -f protoform.yaml
