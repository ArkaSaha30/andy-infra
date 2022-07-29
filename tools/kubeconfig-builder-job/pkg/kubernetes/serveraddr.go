package kubernetes

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/vmware-tanzu-openbtr/gencred-build/pkg/gcloud"
	"github.com/vmware-tanzu-openbtr/gencred-build/pkg/utility"
	"github.com/vmware-tanzu-openbtr/gencred-build/pkg/utility/constants"
	"k8s.io/client-go/kubernetes"
)

// getServerAddress - gets the address of the kubernetes cluster.
func getServerPublicAddressFromClientset(clientset kubernetes.Interface) *string {
	url := clientset.Discovery().RESTClient().Get().URL()
	address := fmt.Sprintf("%s://%s", url.Scheme, url.Host)
	return &address
}

// getServerPrivateAddressFromGcloud - gets the api server private address from gcloud.
func getServerPrivateAddressFromGcloud() (*string, error) {
	gkeClient, err := gcloud.NewGkeClient()

	if err != nil {
		return nil, err
	}

	endpoint, err := gkeClient.GetClusterPrivateEndpoint()

	if err != nil {
		return nil, err
	}

	address := fmt.Sprintf("https://%s:443", *endpoint)

	return &address, nil
}

// getServerAddress -
func getServerAddress(clientset kubernetes.Interface) string {

	var serverAddress *string

	if utility.GetEnv(constants.APIServerAddressType, "public") == "private" {

		privateEndpoint, err := getServerPrivateAddressFromGcloud()

		if err != nil {
			log.WithError(err).Warnf("Unable to find private Address for the cluster, using the public IP")
			serverAddress = getServerPublicAddressFromClientset(clientset)
			log.Infof("Using API Server public address in kubeconfig - %s", *serverAddress)
		} else {
			log.Infof("Using API Server private address in kubeconfig - %s", *privateEndpoint)
			serverAddress = privateEndpoint
		}

	} else {
		serverAddress = getServerPublicAddressFromClientset(clientset)
	}

	return *serverAddress
}
