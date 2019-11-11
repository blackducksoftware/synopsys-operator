#!/bin/bash

# HOW TO RUN:
# ./create_k8s_cluster_and_run_synopsysctl.sh [KIND-IMAGE] [CLUSTER-NAME]

# https://news.ycombinator.com/item?id=10736584
set -o errexit -o nounset -o pipefail
# this line enables debugging
set -xv

#function create_cluster {
printf "\nCreating cluster with image: %s and name: %s\n" "$1" "$2";
# Create a new Kubernetes cluster using KinD
kind --loglevel debug create cluster --image="$1" --name "$2";
# if kind command fails, can't test anything else, fail immediately
if [ $? -eq 0 ]; then
  printf "\nkind cluster: %s created successfully\n" "$2"
else
  printf "\ncould not create kind cluster\n"
  exit 1
fi

# Set KUBECONFIG environment variable based on https://github.com/koalaman/shellcheck/wiki/SC2155
KUBECONFIG="$(kind get kubeconfig-path --name "$2")";
export KUBECONFIG
# cp "$(kind get kubeconfig-path)" /home/travis/.kube/config

kubectl version;

printf "\nCreating Namespace %s\n" "$NAMESPACE";
kubectl create ns "$NAMESPACE";
printf "\nSleeping for 15 seconds (current hack for default service account to come up in a new namespace [TODO: fix this])\n"
sleep 15s

printf "\nGoing to try and create polaris using synopsysctl\n"
synopsysctl --kubeconfig=$(kind get kubeconfig-path --name "$2") --verbose-level=debug create polaris --namespace "$NAMESPACE" --version "$POLARIS_VERSION" --gcp-service-account-path "$GCP_SERVICE_ACCOUNT_PATH" --polaris-license-path "$POLARIS_LICENSE_PATH" --coverity-license-path "$COVERITY_LICENSE_PATH" --environment-dns "travis-onprem-polaris.com" --smtp-host "smtp.sendgrid.net" --smtp-port "3000" --smtp-username "smtp-admin" --smtp-password "password" --smtp-sender-email "noreply@synopsys.com" --postgres-container  --postgres-password "postgres" --postgres-size "1Gi" --uploadserver-size "1Gi" --eventstore-size "1Gi" --mongodb-size "1Gi" --downloadserver-size "1Gi" --enable-reporting --reportstorage-size "1Gi" --organization-description "Cloud Native Eng" --organization-admin-email "polaris-admin@synopsys.com" --organization-admin-name "admin" --organization-admin-username "test123" --ingress-class "nginx" --yaml-url "https://raw.githubusercontent.com/blackducksoftware/releases/Development"
# if synopsysctl command fails, can't test anything else, fail immediately
if [ $? -eq 0 ]; then
  printf "\n\nsynopsysctl command completed, \n"
  #TODO: Add tests here
else
  printf "\n\nsynopsysctl command failed, please check\n"
  exit 1
fi

#printf "let's delete the cluster"

#function delete_cluster {
#kind delete cluster --name "$2"