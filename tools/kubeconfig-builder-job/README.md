# kubeconfig-builder-job

`Kubeconfig-builder-job` is a tool that enables automatic rotation of kubeconfig files used by the Prow Service clusters to connect to the build clusters. The job is supposed to be run as a cronjob in the build cluster and generate timebounded cluster authentication tokens on behalf of a service account and update the same to the Secret Manager where the Prow service cluster is connected. The user can specify the token expiry and the rotation frequencey on need basis.

## Usage

### prerequisite

- Current version of the tool is designed in a way that it assumes the Prow service cluster is deployed in GCP connected to the secret manager via external-secrets operator to sync the secrets in realtime.
- The job image is currently present in the private container registry `prow-prod-registry` in the `prow-tkg-build` GCP Project, so the build cluster should have the necessary credentials to connect to the registry.

### Configurations

There are few environment variables the job is expecting, these environment variables are passed as configmap and later mounted to the cronjob. Please update the configuration files in the configmap with the relevant values.

| Environment Variable | Description | Default |
| --- | --- | --- |
| `GCP_SERVICE_CLUSTER_PROJECTID` | specify the GCP projectId where the service cluster lies | "" |
| `PROW_CLUSTER_KUBECONFIG_SECRETNAME` | Secretmanager secret name to which the kubeconfig to be updated | "" |
| `PROW_CLUSTER_TOKEN_EXPIRY` | The token expiry for the new token to be created | **48h** |
| `PROW_CLUSTER_CONFIG_NAME` | kubeconfig name, this name is later used in the prowjob to specify cluster | "" |
| `PROW_CLUSTER_SERVICE_ACCOUNT_NAME` | Service account name used for cluster authentication | **serviceaccount-cluster-admin** |
| `GOOGLE_APPLICATION_CREDENTIALS` | path where the secret manager credentials are mounted | **/etc/secretmanagercred/serviceaccount.json** |

### Deployment

#### 1. Create the serviceaccount and the clusterrole bindings in the default namespace

`kubectl apply -f deploy/cluster/clusterrole.yaml`

#### 2. Deploy the configmap with all the necessary configurations

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: kubeconfig-builder-config
  namespace: default
data:
  # GCP PROJECTID where the service cluster lies
  GCP_SERVICE_CLUSTER_PROJECTID: "prow-open-btr"
  # Prow build cluster kubeconfig secret name
  PROW_CLUSTER_KUBECONFIG_SECRETNAME: "kubeconfig-prow-gke-build"
  # Expiry for the new token to be generated
  PROW_CLUSTER_TOKEN_EXPIRY: "48h"
  # Prow Build cluster config name
  PROW_CLUSTER_CONFIG_NAME: "prow-gke-build"
  # Build Cluster Service Account Details
  PROW_CLUSTER_SERVICE_ACCOUNT_NAME: "serviceaccount-cluster-admin"   
  # Credentials path for secretmanager
  GOOGLE_APPLICATION_CREDENTIALS: /etc/secretmanagercred/serviceaccount.json      
```

#### 3. Create the secret for authenticating with the secret manager

`kubectl create secret generic secret-manager-credentials --from-file=serviceaccount.json`

#### 4. Deploy the cronjob to the cluster

```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: kubeconfig-builder-job
spec:
  schedule: "0 12 * * *"
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
          - name: kubeconfig-builder-job
            image: us-west1-docker.pkg.dev/prow-open-btr/prow-sandbox-registry/kubeconfig-builder-job:1.0
            imagePullPolicy: IfNotPresent
            envFrom:
            - configMapRef:
                name: kubeconfig-builder-config
            volumeMounts:
            - name: secretmanager-cred
              mountPath: /etc/secretmanagercred/
              readOnly: true
          restartPolicy: OnFailure
          volumes:
          - name: secretmanager-cred
            secret:
              secretName: secret-manager-credentials
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

## Notes on how to automate with service accounts and ESO synch:
```
# installing the kubeconfig job in sandbox
# on build cluster

#create service account for secret manager
gcloud iam service-accounts create prow-secret-writer \
    --project=prow-open-btr

# add roles to GSA
gcloud projects add-iam-policy-binding prow-open-btr \
    --member "serviceAccount:prow-secret-writer@prow-open-btr.iam.gserviceaccount.com" \
    --role roles/secretmanager.admin

gcloud projects add-iam-policy-binding prow-open-btr \
    --member "serviceAccount:prow-service-secrets@prow-open-btr.iam.gserviceaccount.com" \
    --role roles/secretmanager.secretAccessor

gcloud projects add-iam-policy-binding prow-open-btr \
    --member "serviceAccount:prow-build-secrets@prow-open-btr.iam.gserviceaccount.com" \
    --role roles/secretmanager.secretAccessor

# download the service account

# create secret for service account
secretmanager-cred

# create service account
kubectl create serviceaccount prow-build-secrets-sa \
    --namespace default

# bind KSA to GSA
gcloud iam service-accounts add-iam-policy-binding prow-build-secrets@prow-open-btr.iam.gserviceaccount.com \
    --role roles/iam.workloadIdentityUser \
    --member "serviceAccount:prow-open-btr.svc.id.goog[default/prow-build-secrets-sa]"

# annotate KSA
kubectl annotate serviceaccount prow-build-secrets-sa \
    --namespace default \
    iam.gke.io/gcp-service-account=prow-build-secrets@prow-open-btr.iam.gserviceaccount.com

# install ESO
kubectl -n default apply -f prow-build-cluster-secret-store-default.yaml
kubectl -n default apply -f prow-build-external-secrets-default.yaml


# on service cluster

# create service account
kubectl create serviceaccount prow-service-secrets-sa \
    --namespace default

# bind KSA to GSA
gcloud iam service-accounts add-iam-policy-binding prow-service-secrets@prow-open-btr.iam.gserviceaccount.com \
    --role roles/iam.workloadIdentityUser \
    --member "serviceAccount:prow-open-btr.svc.id.goog[default/prow-service-secrets-sa]"

# annotate KSA
kubectl annotate serviceaccount prow-service-secrets-sa \
    --namespace default \
    iam.gke.io/gcp-service-account=prow-service-secrets@prow-open-btr.iam.gserviceaccount.com

# install ESO
kubectl -n default apply -f prow-service-cluster-secret-store-default.yaml
kubectl -n default apply -f prow-service-external-secrets-default.yaml
```
