#!/bin/bash

#function create_cluster {
echo "Creating cluster with image $1 and name $2";
# Create a new Kubernetes cluster using KinD
kind create cluster --image="$1" --name "$2" --loglevel debug;

# Set KUBECONFIG environment variable based on https://github.com/koalaman/shellcheck/wiki/SC2155
KUBECONFIG="$(kind get kubeconfig-path --name "$2")";
export KUBECONFIG
# cp "$(kind get kubeconfig-path)" /home/travis/.kube/config

kubectl version;

echo "Creating Namespace $NAMESPACE";
kubectl create ns "$NAMESPACE";

echo "Going to try and create polaris"

#function run_synopsysctl {
echo "Sleeping for 15 seconds (current hack for default service account to come up in a new namespace [TODO: fix this])"
sleep 15s
synopsysctl --verbose-level=debug create polaris --kubeconfig=$(kind get kubeconfig-path --name "$2") --namespace "$NAMESPACE" --version "$POLARIS_VERSION" --environment-dns "travis-onprem-polaris.com" --smtp-host "smtp.sendgrid.net" --smtp-port "3000" --smtp-username "smtp-admin" --smtp-password "password" --smtp-sender-email "noreply@synopsys.com" --postgres-username "postgres-admin" --postgres-password "postgres" --organization-admin-email "polaris-admin@synopsys.com" --organization-admin-name "Polaris Admin" --organization-admin-username "test123" --organization-description "Cloud Native Eng" --organization-name "Polaris" --gcp-service-account-path "$GCP_SERVICE_ACCOUNT_PATH" --coverity-license-path "$COVERITY_LICENSE_PATH" --yaml-url "https://raw.githubusercontent.com/blackducksoftware/releases/Development"

#TODO: Add tests here
echo "Tests ran successfully, let's delete the cluster"

#function delete_cluster {
kind delete cluster --name "$2"