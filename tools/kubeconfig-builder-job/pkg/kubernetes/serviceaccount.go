package kubernetes

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/vmware-tanzu-openbtr/gencred-build/pkg/utility"
	"github.com/vmware-tanzu-openbtr/gencred-build/pkg/utility/constants"
	authenticationv1 "k8s.io/api/authentication/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api/v1"
)

// generateNewTokenForServiceAccount Creates new token for the specified serviceaccount with the given duration
func generateNewTokenForServiceAccount(saObj *corev1.ServiceAccount, clientset kubernetes.Interface, duration metav1.Duration) (*authenticationv1.TokenRequest, error) {

	durationInSeconds := int64(duration.Seconds())
	tokenReq := &authenticationv1.TokenRequest{
		Spec: authenticationv1.TokenRequestSpec{
			ExpirationSeconds: &durationInSeconds,
		},
	}

	tokenResp, err := clientset.CoreV1().ServiceAccounts(saObj.Namespace).CreateToken(context.TODO(), saObj.Name, tokenReq, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("creating service account token: %w", err)
	}
	if tokenResp.Status.Token == "" {
		return nil, fmt.Errorf("no service account token returned: %v", err)
	}

	return tokenResp, nil
}

// getServiceAccountCredentials - fetch the existing service account and get new token
func getServiceAccountCredentials(clientset kubernetes.Interface, duration metav1.Duration) ([]byte, []byte, error) {
	client := clientset.CoreV1().ServiceAccounts(corev1.NamespaceDefault)

	// Get ServiceAccount.
	saObj, err := client.Get(context.TODO(), utility.GetEnv(constants.ProwBuildClusterServiceAccount, "serviceaccount-cluster-admin"), metav1.GetOptions{})
	if err != nil {
		return nil, nil, fmt.Errorf("get SA: %w", err)
	}

	tokenResp, err := generateNewTokenForServiceAccount(saObj, clientset, duration)

	if err != nil {
		log.WithError(err).Error("Unable to geenrate new topken for the given service account")
		return nil, nil, fmt.Errorf("Error generating new toke for the sa %s: %w", saObj.Name, err)
	}

	caConfigMap, err := clientset.CoreV1().ConfigMaps(corev1.NamespaceDefault).Get(context.TODO(), "kube-root-ca.crt", metav1.GetOptions{})
	if err != nil {
		return nil, nil, fmt.Errorf("locate root CA configmap: %w", err)
	}
	caPEM, ok := caConfigMap.Data["ca.crt"]
	if !ok {
		return nil, nil, errors.New("locate root CA data")
	}

	return []byte(tokenResp.Status.Token), []byte(caPEM), nil
}

// CreateClusterServiceAccountCredentials creates a service account to authenticate to a cluster API server.
func CreateClusterServiceAccountCredentials(clientset kubernetes.Interface) (kubeconfig []byte, err error) {

	// Get the Build cluster config name
	buildClusterConfigName := utility.GetEnv(constants.ProwBuildClusterConfigName, "build-cluster")

	// Setting up expiry for the token secret
	tokenExpiry := utility.GetEnv(constants.ProwBuildClusterTokenExpiry, "72h")
	tokenDuration, err := time.ParseDuration(tokenExpiry)

	if err != nil {
		log.WithError(err).Errorf("Error parsing token expiry based on the given input, please check the input - %s", tokenExpiry)
		os.Exit(1)
	}

	token, caPEM, err := getServiceAccountCredentials(clientset, metav1.Duration{Duration: tokenDuration})

	if err != nil {
		return nil, fmt.Errorf("get or create SA: %w", err)
	}

	authInfo := clientcmdapi.AuthInfo{
		Token: string(token),
	}

	kubeconfig, err = CreateKubeConfig(clientset, buildClusterConfigName, caPEM, authInfo)

	if err != nil {
		return nil, err
	}

	utility.WriteFileToLocal("kubeconfig.yaml", kubeconfig)

	log.Infof("Successfully generated kubeconfig using the new token...")

	return kubeconfig, nil
}
