#!/bin/bash
source ../create-install-files.sh "${@}"

oc new-project $_arg_pcp_namespace

source ../oadm-policy-init.sh $arg_pcp_namespace

oc project $_arg_pcp_namespace
../common/oadm-init.sh
oc create -f config.yml
oc create -f proto.yml
