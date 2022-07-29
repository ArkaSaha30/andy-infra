# kubeconfig-builder-job

`Kubeconfig-builder-job` is a tool that enables automatic rotation of kubeconfig files used by the Prow Service clusters to connect to the build clusters. The job is supposed to be run as a cronjob in the build cluster and generate timebounded cluster authentication tokens on behalf of a service account and update the same to the Secret Manager where the Prow service cluster is connected. The user can specify the token expiry and the rotation frequencey on need basis.

## Usage

### Prerequisites

- Current version of the tool is designed in a way that it assumes the Prow service cluster is deployed in GCP connected to the secret manager via external-secrets operator to sync the secrets in realtime.
- The job image is currently present in the private container registry `prow-prod-registry` in the `prow-tkg-build` GCP Project, so the build cluster should have the necessary credentials to connect to the registry.

We are also using ESO to sync the service account secret to the default namespace for use by the job to access the cluster and write to the Google Secret Manager.  Configuration steps to set this up follows:

```s
#create service account for secret manager
gcloud iam service-accounts create prow-secret-writer \
    --project=prow-tkg-build

# add roles to GSA
gcloud projects add-iam-policy-binding prow-tkg-build \
    --member "serviceAccount:prow-secret-writer@prow-tkg-build.iam.gserviceaccount.com" \
    --role roles/secretmanager.admin

gcloud projects add-iam-policy-binding prow-tkg-build \
    --member "serviceAccount:prow-service-secrets@prow-tkg-build.iam.gserviceaccount.com" \
    --role roles/secretmanager.secretAccessor

gcloud projects add-iam-policy-binding prow-tkg-build \
    --member "serviceAccount:prow-build-secrets@prow-tkg-build.iam.gserviceaccount.com" \
    --role roles/secretmanager.secretAccessor

# download the service account json

# create secret for service account
# Name the secret: secretmanager-cred


# on build cluster

# create service account
kubectl create serviceaccount prow-build-secrets-sa \
    --namespace default

# bind KSA to GSA
gcloud iam service-accounts add-iam-policy-binding prow-build-secrets@prow-tkg-build.iam.gserviceaccount.com \
    --role roles/iam.workloadIdentityUser \
    --member "serviceAccount:prow-tkg-build.svc.id.goog[default/prow-build-secrets-sa]"

# annotate KSA
kubectl annotate serviceaccount prow-build-secrets-sa \
    --namespace default \
    iam.gke.io/gcp-service-account=prow-build-secrets@prow-tkg-build.iam.gserviceaccount.com

# install ESO
kubectl -n default apply -f prow-build-cluster-secret-store-default.yaml
kubectl -n default apply -f prow-build-external-secrets-default.yaml


# on service cluster

# create service account
kubectl create serviceaccount prow-service-secrets-sa \
    --namespace default

# bind KSA to GSA
gcloud iam service-accounts add-iam-policy-binding prow-service-secrets@prow-tkg-build.iam.gserviceaccount.com \
    --role roles/iam.workloadIdentityUser \
    --member "serviceAccount:prow-tkg-build.svc.id.goog[default/prow-service-secrets-sa]"

# annotate KSA
kubectl annotate serviceaccount prow-service-secrets-sa \
    --namespace default \
    iam.gke.io/gcp-service-account=prow-service-secrets@prow-tkg-build.iam.gserviceaccount.com

# install ESO
kubectl -n default apply -f prow-service-cluster-secret-store-default.yaml
kubectl -n default apply -f prow-service-external-secrets-default.yaml

```
After the above has been implemented, check that the `secretmanager-cred` secret has been synced to the namespace.


### Configurations

There are few environment variables the job is expecting, these environment variables are passed as configmap and later mounted to the cronjob. Please update the configuration files in the configmap with the relevant values.

| Environment Variable | Description | Default | Remarks |
| --- | --- | --- | --- |
| `SERVICE_CLUSTER_PROJECTID` | specify the GCP projectId where the service cluster lies | "" | **required** |
| `TOKEN_DURATION` | The token expiry for the new token to be created | **48h** | use default |
| `KUBECONFIG_NAME` | kubeconfig name, this name is later used in the prowjob to specify cluster | "" | **required** |
| `SERVICE_ACCOUNT_NAME` | Service account name used for cluster authentication | **serviceaccount-cluster-admin** | use default |
| `GOOGLE_APPLICATION_CREDENTIALS` | path where the secret manager credentials are mounted | **/etc/secretmanagercred/serviceaccount.json** | use default |
| `APISERVER_ADDRESS_TYPE` | Specify the address type needed - private or public| **public** | **required** |
| `BUILD_CLUSTER_NAME` | Specify cluster Id as per the cloud portal | "" | **optional** Only needed for private address kubeconfigs |
| `BUILD_CLUSTER_LOCATION` | Specify cluster Location as per the cloud portal| "" | **optional** Only needed for private address kubeconfigs |
| `BUILD_CLUSTER_PROJECTID` | Specify cluster project Id as per the cloud portal| "" | **optional** Only needed for private address kubeconfigs |

### Deployment

#### 1. Create the service account and the clusterrole bindings in the default namespace

`kubectl apply -f deploy/cluster/clusterrole.yaml`

#### 2. Deploy the configmap with all the necessary configurations

```yaml
##
# Configmap with all the relevant environment variables for the kubeconfig-builder-job to work
# This configmap is mounted with the cronjob as envrionment variables
#
apiVersion: v1
kind: ConfigMap
metadata:
  name: kubeconfig-builder-config
data:
  # GCP PROJECTID where the service cluster lies
  SERVICE_CLUSTER_PROJECTID: prow-tkg-build  
  # Expiry for the new
  TOKEN_DURATION: 48h     
  # Prow Build cluster config name
  KUBECONFIG_NAME: gke-prow-gke-build
  # Build Cluster Service Account Details
  SERVICE_ACCOUNT_NAME: serviceaccount-cluster-admin                                   
  # Credentials path for secretmanager
  GOOGLE_APPLICATION_CREDENTIALS: /etc/secretmanagercred/service-account.json      
  # Apiserver address type - private/public
  APISERVER_ADDRESS_TYPE: private
  # Build cluster name
  BUILD_CLUSTER_NAME: prow-build
  # Build cluster location
  BUILD_CLUSTER_LOCATION: us-west1
  # Build cluster projectId
  BUILD_CLUSTER_PROJECTID: prow-tkg-build
```

#### 3. Deploy the cronjob to the cluster

```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: kubeconfig-prow-gke-build
spec:
  schedule: "0 0 * * *"
  concurrencyPolicy: Allow
  startingDeadlineSeconds: 100
  suspend: false
  successfulJobsHistoryLimit: 3
  failedJobsHistoryLimit: 1
  jobTemplate:
    spec:
      template:
        spec:
          imagePullSecrets:
            - name: artifact-registry
          serviceAccountName: serviceaccount-cluster-admin
          containers:
          - name: kubeconfig-prow-gke-build
            image: us-west1-docker.pkg.dev/prow-tkg-build/prow-prod-registry/kubeconfig-builder-job:3.0
            imagePullPolicy: IfNotPresent
            envFrom:
            - configMapRef:
                name: kubeconfig-prow-gke-build
            volumeMounts:
            - name: secretmanager-cred
              mountPath: /etc/secretmanagercred/
              readOnly: true
          restartPolicy: OnFailure
          volumes:
          - name: secretmanager-cred
            secret:
              secretName: secretmanager-cred
```

## Additional Information

### Create Docker registry secret to connect with GCP Artifact registry

```s
kubectl create secret docker-registry artifact-registry \
    --docker-server=https://us-west1-docker.pkg.dev \
    --docker-email=711165799992-compute@developer.gserviceaccount.com \
    --docker-username=_json_key \
    --docker-password="$(cat prow-tkg-build-sa.json)"
```

### Create IAM role for the Secret Manager in GCP

```s
# Create the service account and asscoiated role binding
gcloud iam service-accounts create NAME

gcloud projects add-iam-policy-binding PROJECT_ID --member="serviceAccount:SERVICE_ACCOUNT_NAME@PROJECT_ID.iam.gserviceaccount.com" --role=ROLE

gcloud iam service-accounts keys create FILE_NAME.json --iam-account=SERVICE_ACCOUNT_NAME@PROJECT_ID.iam.gserviceaccount.com
```

### Create role for accesing cluster Private IP

```s
gcloud projects add-iam-policy-binding <project-id> --role=roles/container.clusterViewer --member=serviceAccount:<sa-name>*.iam.gserviceaccount.com --project=<project-id>
```
