#!/bin/sh

##
# Specify the environment variables for kubeconfig-builder
#
## 
export SERVICE_CLUSTER_PROJECTID=prow-open-btr                                  # GCP PROJECTID where the service cluster lies
export TOKEN_DURATION=1h                                                 # Token expiry for the Build cluster
export KUBECONFIG_NAME=gke-prow-build-cluster                          # Prow Build cluster config name
# 
export GOOGLE_APPLICATION_CREDENTIALS=serviceaccount.json                           # Credentials path for secretmanager
export SERVICE_ACCOUNT_NAME=serviceaccount-cluster-admin               # Build Cluster Service Account Details
#
export APISERVER_ADDRESS_TYPE=private
export BUILD_CLUSTER_NAME=prow-build
export BUILD_CLUSTER_LOCATION=us-west1
export BUILD_CLUSTER_PROJECTID=prow-open-btr

go build -o kubeconfig-builder

./kubeconfig-builder