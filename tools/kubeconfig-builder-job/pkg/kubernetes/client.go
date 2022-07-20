package kubernetes

import (
	"flag"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	log "github.com/sirupsen/logrus"

	//Vendor specific authentication (Azure/GCP)
	_ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

// Create new kubernetes client based on the runtime
// Create kubernetes client while running locally in machine using kubeconfig for incluster use service account
// Use vendor specific sdk for auth if you are not using cluster certificates in your kubeconfig
func NewClient() (client *kubernetes.Clientset, err error) {

	var kubeconfig *string

	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)

	if err != nil {
		// Warning to notify kubeconfig not found locally, fallback to in-cluster authentication
		log.Warnf("Unable to build kubernetes client config from flags - %s, using in-cluster service account.\n", err.Error())
		config, err = rest.InClusterConfig()

		if err != nil {
			log.Errorf("Kubernetes client Error fetching getting inclusterconfig - %s", err.Error())
			return client, err
		}
	}

	clientset, err := kubernetes.NewForConfig(config)

	if err != nil {
		log.Errorf("Error creating clientset %v \n", err.Error())
		return clientset, err
	}

	return clientset, nil
}
