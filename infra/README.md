# Scripts for building PROW POC infra on AWS

Most of this scripting was taken from the TCE project and modified for our use.

`build_prow_infra_on_aws.sh` will build 3 cluster:
- management: prow-mgr - this is the TCE management cluster
- prow service: prow-service - this will run the prow infrastructure and trusted postsubmit jobs
- prow build: prow-build - this will run client repo untrusted presubmit jobs

Each cluster is made up of one control plane node and one worker node - both M5 size.  The clusters have a bastion host. They also have a load balancer which is targeted by kubeconfig for access to the cluster.

TCE package repo has been installed on the service and build clusters in the "tanzu-package-repo-global" namespace.

**Note:**
aws-nuke does not work and is commented out.  The PowerUser rights we have in our sandbox AWS accounts do not allow us to use Alias which is required for aws-nuke.

## AWS Infra Cluster Build variables
These variables need to be exported to the console for the build script to process.  The AWS_SESSION_TOKEN will expire so a managed account build is only useful for "short" tests or jobs that don't require long running AWS sessions.  Useful for: build cluster --> do tests --> destroy

**cloudgate**
```
export AWS_ACCESS_KEY_ID=<access key id>
export AWS_SECRET_ACCESS_KEY=<secret access key>
export AWS_SESSION_TOKEN=<session token if managed account>

```
**cluster variables**
```
export MGMT_CLUSTER_NAME="prow-mgr"
export MGMT_CLUSTER_PLAN="dev"
export MGMT_CONTROL_PLANE_MACHINE_TYPE="m5.large"
export MGMT_NODE_MACHINE_TYPE="m5.large"

export SERVICE_CLUSTER_NAME="prow-service"
export SERVICE_CLUSTER_PLAN="dev"
export SERVICE_CONTROL_PLANE_MACHINE_TYPE="m5.large"
export SERVICE_NODE_MACHINE_TYPE="m5.large"

export BUILD_CLUSTER_NAME="prow-build"
export BUILD_CLUSTER_PLAN="dev"
export BUILD_CONTROL_PLANE_MACHINE_TYPE="m5.large"
export BUILD_NODE_MACHINE_TYPE="m5.large"
```

**prow app variables**
```
export GITHUB_APP_ID=<github app id>
export GITHUB_ORG="AndyTauber"
export GITHUB_REPO1="andy-infra"
export GITHUB_REPO2="andy-test"
export MY_EMAIL="atauber@vmware.com"
export JOB_CONFIG_PATH="path-to-jobs/jobs/test-prow/test.yaml"
export GCS_BUCKET="andytauber-prow"
export PROW_FQDN="prow.andytauber.info"
export CERT_EMAIL="atauber@vmware.com"
export REGISTRY_USERNAME="AWS"
export REGISTRY_PUSH="public.ecr.aws/<registry address>"
```

**Secrets**
Use the following variables to create the secrets, this will be removed if using external secrets:
```
export GCS_CREDENTIAL_PATH="path-to-gcs-cred/service-account.json"
export HMAC_TOKEN_PATH="path-to-hmac/hmac-secret"
export GITHUB_TOKEN_PATH="path-to-github-token/private-key.pem"
export OAUTH_CONFIG_PATH="path-to-oath-config/github-oauth-config"
export COOKIE_PATH="path-to-cookie/cookie.txt"
```

## Required infra setup
You will need github repos setup, github app, AWS ECR, and GCP bucket for logs.  Also a domain name and ability to create a cname record in dns.

Use: https://github.com/rajaskakodkar/prow-on-tce and https://github.com/kubernetes/test-infra/blob/master/prow/getting_started_deploy.md to get required infra up and running so you can fill in the environment variables above.  

**Some notes of clarification:**

You need to create a kubeconfig.yaml file with gencred and apply the kubeconfig secret to prow namespace before you create the build cluster.  After creating the build cluster, update the kubeconfig secret once more.  This should be handled within the build_prow_on_aws.sh script.

Once you've created the Github app using the instructions: https://docs.github.com/en/developers/apps/building-github-apps/creating-a-github-app, don't forget to install the app to your org.  This is not spelled out in the documentation.
