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
