#!/bin/bash
source ../common/parse-or-gather-user-input.sh "${@}"

oc new-project $_arg_pcp_namespace

source ../common/oadm-policy-init.sh $arg_pcp_namespace

#oc project $_arg_pcp_namespace

oc create -f config.yml
oc create -f proto.yml
