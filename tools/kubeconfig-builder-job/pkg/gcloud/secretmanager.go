// Sample quickstart is a basic program that uses Secret Manager.
package gcloud

import (
	"context"
	"errors"
	"fmt"
	"os"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	log "github.com/sirupsen/logrus"
	"github.com/vmware-tanzu-openbtr/gencred-build/pkg/utility"
	"github.com/vmware-tanzu-openbtr/gencred-build/pkg/utility/constants"
	"google.golang.org/api/iterator"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
)

// StoreSecretToSecretManager Updates the specified secret in GCP secretmanager with the new payload,
//
// Updates the existing secret with the new value and upgrade the version. If secret doesnt exists, creates the new secret and update the value and version.
//
func StoreSecretToSecretManager(secretName string, secretPayload *[]byte) (*string, error) {

	// Verify the credentials path has the relevant files for authenticating with secret manager
	if path, err := VerifySecretManagercreedExists(); err != nil {
		log.Errorf("Unable to find the service account credentials path in the specified path %s", *path)
		return nil, err
	}

	// Create the client for authentication.
	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		log.WithError(err).Errorf("Failed to setup secret manager client, check the authentcation...")
		return nil, err
	}
	defer client.Close()

	projectID := utility.GetEnv(constants.GCPServiceClusterProjectID, "")

	log.Infof("Updating GCP secret Manager secret %s in the project %s with new kubeconfig...", secretName, projectID)

	getSecretReq := &secretmanagerpb.GetSecretRequest{
		Name: fmt.Sprintf("projects/%s/secrets/%s", projectID, secretName),
	}

	secret, err := client.GetSecret(ctx, getSecretReq)

	if err != nil {
		log.Warnf("Unable to find provided secret %s in the secretmanager, creating new secret..", secretName)

		// Create the request to create the secret.
		createSecretReq := &secretmanagerpb.CreateSecretRequest{
			Parent:   fmt.Sprintf("projects/%s", projectID),
			SecretId: secretName,
			Secret: &secretmanagerpb.Secret{
				Replication: &secretmanagerpb.Replication{
					Replication: &secretmanagerpb.Replication_Automatic_{
						Automatic: &secretmanagerpb.Replication_Automatic{},
					},
				},
			},
		}

		secret, err = client.CreateSecret(ctx, createSecretReq)

		if err != nil {
			log.WithError(err).Errorf("Error creating/fetching the requested secret from the secret manager.")
			return nil, err
		}
	}

	log.Infof("Updating the secret version with the newly added secret...")

	addSecretVersionReq := &secretmanagerpb.AddSecretVersionRequest{
		Parent: secret.Name,
		Payload: &secretmanagerpb.SecretPayload{
			Data: *secretPayload,
		},
	}

	// Upgrading the secret version to the newly created secret
	version, err := client.AddSecretVersion(ctx, addSecretVersionReq)
	if err != nil {
		log.Errorf("failed to add secret to new version: %v", err)
		return nil, err
	}

	log.Debugf("Updated secret with the new value and upgraded version %v", version.Name)

	return &version.Name, nil
}

// VerifySecretManagercreedExists check whether the provided Service credntials path is valid, check whether the auth file exits in the path.
func VerifySecretManagercreedExists() (*string, error) {

	credentialsPath := utility.GetEnv(constants.GCPSecretManagerCredetialsPath, "")

	if _, err := os.Stat(credentialsPath); err == nil {
		return &credentialsPath, nil

	} else if errors.Is(err, os.ErrNotExist) {
		return &credentialsPath, err

	} else {
		return &credentialsPath, err
	}
}

// cleanDeprecatedSecretversions - cleans up old secret versions which are deprecated.
func cleanDeprecatedSecretversions(client *secretmanager.Client, projectID string, secret string) {

	ctx := context.Background()

	listSecretVersionsRequest := &secretmanagerpb.ListSecretVersionsRequest{
		Parent: fmt.Sprintf("projects/%s/secrets/%s", projectID, secret),
	}

	versions := client.ListSecretVersions(ctx, listSecretVersionsRequest)

	for {
		resp, err := versions.Next()

		if err == iterator.Done {
			break
		}

		if err != nil {
			log.Errorf("failed to list secret versions: %v", err)
			return
		}

		log.Printf("Found secret version %s with state %s\n",
			resp.Name, resp.State)
	}

	// destroySecretVersionRequest := &secretmanagerpb.DestroySecretVersionRequest{
	// 	Name: fmt.Sprintf("projects/%s/secrets/%s/versions/%s", projectID, secret, "2"),
	// }

	// res, err := client.DestroySecretVersion(ctx, destroySecretVersionRequest)

	// if err != nil {
	// 	log.Warnf("Unable to cleanup the old secret Versions %v", err)
	// 	return
	// }

	// log.Debugf("Cleaned up old secret Versions %v", res)

}
