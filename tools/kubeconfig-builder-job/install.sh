#!/bin/sh

##
# Specify the environment variables for kubeconfig-builder
#
##
export PROW_BUILD_CLUSTER_TOKEN_EXPIRY=48h                                                   # Token expiry for the Build cluster
export PROW_CLUSTER_CONFIG_NAME=prow-gke-build                                # Prow Build cluster config name
export GCP_SERVICE_CLUSTER_PROJECTID=prow-open-btr                                  # GCP PROJECTID where the service cluster lies
export PROW_CLUSTER_KUBECONFIG_SECRETNAME=kubeconfig-prow-gke-build         # Prow build cluster kubeconfig secret name
export GOOGLE_APPLICATION_CREDENTIALS="/etc/secretmanagercred/service-account.json"                                   # Credentials path for secretmanager
export PROW_CLUSTER_SERVICE_ACCOUNT_NAME=serviceaccount-cluster-admin                 # Build Cluster Service Account Details

go build -o kubeconfig-builder-job

./kubeconfig-builder-job
