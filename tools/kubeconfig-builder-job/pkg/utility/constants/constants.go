package constants

// Define all the global constants required for environment configurations
const (

	// Service Cluster projectID in the GCP
	GCPServiceClusterProjectID string = "GCP_SERVICE_CLUSTER_PROJECTID"

	// Secret name in the secretmanager to keep the kubeconfig credentails
	ProwBuildClusterKubeconfigSecretname string = "PROW_CLUSTER_KUBECONFIG_SECRETNAME"

	// Expiry for the New token to be generated
	ProwBuildClusterTokenExpiry string = "PROW_CLUSTER_TOKEN_EXPIRY"

	// Prow Build Cluster Name, this name will appear in the kubeconfig
	ProwBuildClusterConfigName string = "PROW_CLUSTER_CONFIG_NAME"

	// The Service account name in the Build cluster, token will be generated on behalf of the service account
	ProwBuildClusterServiceAccount string = "PROW_CLUSTER_SERVICE_ACCOUNT_NAME"

	// Local Path of the gcp service account credentials
	GCPSecretManagerCredetialsPath string = "GOOGLE_APPLICATION_CREDENTIALS"
)
