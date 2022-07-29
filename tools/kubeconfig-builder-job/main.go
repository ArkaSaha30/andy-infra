package main

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/vmware-tanzu-openbtr/gencred-build/pkg/gcloud"
	"github.com/vmware-tanzu-openbtr/gencred-build/pkg/kubernetes"
	"github.com/vmware-tanzu-openbtr/gencred-build/pkg/utility"
	"github.com/vmware-tanzu-openbtr/gencred-build/pkg/utility/constants"
)

// main entry point
func main() {

	// Create a Kubernetes clientset for interacting with the cluster.
	clientset, err := kubernetes.NewClient()

	if err != nil {
		log.WithError(err).Errorf("FATAL ---- Error connecting to k8s cluster ---- Exiting...")
		os.Exit(1)
	}

	// Generate the kubeconfig file from the new token and cluster certificate
	kubeconfig, err := kubernetes.CreateClusterServiceAccountCredentials(clientset)

	if err != nil {
		log.WithError(err).Errorf("Failed to create kubeconfig file with the new token for the cluster")
		os.Exit(1)
	}

	// Update the kubeconfig in the secretmanager
	err = UpdateKubeconfigToSecretManager(kubeconfig)

	if err != nil {
		log.WithError(err).Errorf("Failed to update newly created kubeconfig in the secretmanager")
		os.Exit(1)
	}

}

// Upgrade the newly created kubeconfig file in the secret Manager
func UpdateKubeconfigToSecretManager(kubeconfig []byte) error {

	secretName := fmt.Sprintf("kubeconfig-%s", utility.GetEnv(constants.ProwBuildClusterKubeConfigName, ""))

	secretResult, err := gcloud.StoreSecretToSecretManager(secretName, &kubeconfig)

	if err != nil {
		log.Errorf("Error Updating kubeconfig in the secret store")
		return err
	}

	log.Infof("Succesfully updated the new kubeconfig and upgraded the version %s", *secretResult)

	return nil
}
