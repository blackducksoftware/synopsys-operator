#!/bin/bash
SCC="add-scc-to-user"
ROLE="add-role-to-user"
CLUSTER="add-cluster-role-to-user"
NS=$_arg_pcp_namespace

# Create the perceiver service account
oc create serviceaccount perceiver -n $NS

# Protoform has its own SA.
oc create serviceaccount protoform -n $NS
oc adm policy $CLUSTER cluster-admin system:serviceaccount:$NS:protoform

# following allows us to write cluster level metadata for imagestreams
oc adm policy $CLUSTER cluster-admin system:serviceaccount:$NS:perceiver

# Create the serviceaccount for perceptor-scanner to talk with Docker
oc create sa perceptor-scanner-sa -n $NS

# allows launching of privileged containers for Docker machine access
oc adm policy $SCC privileged system:serviceaccount:$NS:perceptor-scanner-sa

# these allow us to pull images
oc adm policy $CLUSTER cluster-admin system:serviceaccount:$NS:perceptor-scanner-sa
oc policy $ROLE view system:serviceaccount::perceptor-scanner-sa

_arg_private_registry_token=$(oc sa get-token perceptor-scanner-sa)
