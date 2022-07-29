package constants

// Define all the global constants required for environment configurations
const (

	// Service Cluster projectID in the GCP
	GCPServiceClusterProjectID string = "SERVICE_CLUSTER_PROJECTID"

	// Expiry for the New token to be generated
	ProwBuildClusterTokenExpiry string = "TOKEN_DURATION"

	// Prow Build Cluster Name, this name will appear in the kubeconfig
	ProwBuildClusterKubeConfigName string = "KUBECONFIG_NAME"

	// The Service account name in the Build cluster, token will be generated on behalf of the service account
	ProwBuildClusterServiceAccount string = "SERVICE_ACCOUNT_NAME"

	// Local Path of the gcp service account credentials
	GCPSecretManagerCredetialsPath string = "GOOGLE_APPLICATION_CREDENTIALS"

	// Specify api server address type - public/private
	APIServerAddressType string = "APISERVER_ADDRESS_TYPE"

	// GCP project ID of the Builc cluster
	GKEClusterProjectID string = "BUILD_CLUSTER_PROJECTID"
	ß
	// GCP Location of the build cluster
	GKEClusterLocation string = "BUILD_CLUSTER_LOCATION"

	// Build Cluster name
	GKEClusterName string = "BUILD_CLUSTER_NAME"
)
