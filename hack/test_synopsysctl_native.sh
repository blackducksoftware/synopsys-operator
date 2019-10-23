#!/bin/bash

NATIVE_FILENAME=native.yml
synopsysctl --verbose-level=debug create polaris native --namespace "$NAMESPACE" --version "$POLARIS_VERSION" --environment-dns "travis-onprem-polaris.com" --smtp-host "smtp.sendgrid.net" --smtp-port "3000" --smtp-username "smtp-admin" --smtp-password "password" --smtp-sender-email "noreply@synopsys.com" --postgres-username "postgres-admin" --postgres-password "postgres" --organization-admin-email "polaris-admin@synopsys.com" --organization-admin-name "Polaris Admin" --organization-admin-username "test123" --organization-description "Cloud Native Eng" --organization-name "Polaris" --gcp-service-account-path "$GCP_SERVICE_ACCOUNT_PATH" --coverity-license-path "$COVERITY_LICENSE_PATH" --yaml-url "https://raw.githubusercontent.com/blackducksoftware/releases/Development" -o yaml >> $NATIVE_FILENAME

# Source: https://github.com/instrumenta/kubernetes-json-schema/blob/master/build.sh#L23
declare -a k8s_versions=("1.15.4"
                         "1.14.7")
#                         "1.13.11"
#                         "1.12.10"
#                         "1.11.9")

for k8s_version in "${k8s_versions[@]}"
do
  printf '\n\nRunning kubeval for Kubernetes version:\t%s\n' "$k8s_version"
  if ! kubeval $NATIVE_FILENAME -v "$k8s_version" --strict
  then
    printf '\n Yaml validation failed for Kubernetes version, exiting immediately:\t%s\n' "$k8s_version"
    exit 1
  fi
done

# Source: https://github.com/garethr/openshift-json-schema/blob/master/build.sh#L13
declare -a openshift_versions=("4.1.0"
                               "3.11.0")
#                               "3.10.0"
#                               "3.9.0")

for oc_version in "${openshift_versions[@]}"
do
  printf '\n\nRunning kubeval for Openshift version:\t%s\n' "$oc_version"
  if ! kubeval $NATIVE_FILENAME --openshift -v "$oc_version" --strict
  then
    printf '\n Yaml validation failed for Openshift version, ignoring and continuing for now [TODO: change me]:\t%s\n' "$oc_version"
  fi
done

exit 0