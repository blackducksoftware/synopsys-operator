#!/bin/bash

# gcloud auth print-access-token
# docker login -u oauth2accesstoken -p https://gcr.io

docker pull gcr.io/snps-swip-staging/reporting-frontend-service:0.0.673
docker pull gcr.io/snps-swip-staging/reporting-polaris-service:0.0.111
docker pull gcr.io/snps-swip-staging/reporting-report-service:0.0.450
docker pull gcr.io/snps-swip-staging/reporting-rp-issue-manager:0.0.487
docker pull gcr.io/snps-swip-staging/reporting-rp-portfolio-service:0.0.663
docker pull gcr.io/snps-swip-staging/reporting-tools-portfolio-service:0.0.974
docker pull gcr.io/snps-swip-staging/swip_auth-server:latest
docker pull gcr.io/snps-swip-staging/swip_eventstore:latest
docker pull gcr.io/snps-swip-staging/eventstore-util:latest
docker pull gcr.io/snps-swip-staging/reporting-clamav:latest
docker pull gcr.io/snps-swip-staging/swip_mongodb:latest
docker pull gcr.io/snps-swip-staging/vault-util:latest
