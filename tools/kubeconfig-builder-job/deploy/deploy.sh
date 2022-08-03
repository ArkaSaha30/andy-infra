#!/bin/sh

# Build and Push Image for the job
docker build --platform amd64  -t kubeconfig-builder-job:3.0 -f Dockerfile .
docker tag kubeconfig-builder-job:3.0 us-west1-docker.pkg.dev/prow-open-btr/prow-sandbox-registry/kubeconfig-builder-job:3.0
docker push us-west1-docker.pkg.dev/prow-open-btr/prow-sandbox-registry/kubeconfig-builder-job:3.0

# Create the CR, CRD and service account for the cluster authentication in the default namespace
kubectl apply -f cluster/clusterrole.yaml

# Create the artifact registry secrets in the namespace to acces images from
kubectl create secret docker-registry artifact-registry \
    --docker-server=https://us-west1-docker.pkg.dev \
    --docker-email=711165799992-compute@developer.gserviceaccount.com \
    --docker-username=_json_key \
    --docker-password="$(cat prow-tkg-build-sa.json)"

# Secret for authenticating with the secret manager
#
## Manually create the secret for authenticating with the gcp secret manager
kubectl create secret generic secret-manager-credentials --from-file=serviceaccount.json

# Deploy the configmap and the cronjob
kubectl apply -f cluster/cronjob.yaml
