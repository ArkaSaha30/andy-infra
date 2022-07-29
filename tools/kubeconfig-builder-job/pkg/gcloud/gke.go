package gcloud

import (
	"context"
	"errors"
	"fmt"

	container "cloud.google.com/go/container/apiv1"
	log "github.com/sirupsen/logrus"
	"github.com/vmware-tanzu-openbtr/gencred-build/pkg/utility"
	"github.com/vmware-tanzu-openbtr/gencred-build/pkg/utility/constants"
	containerpb "google.golang.org/genproto/googleapis/container/v1"
)

type clusterConn struct {
	Project   string
	Location  string
	ClusterID string
}

type ClusterConn interface {
	GetClusterInfo() (*containerpb.Cluster, error)
	GetClusterPrivateEndpoint() (*string, error)
}

func NewGkeClient() (ClusterConn, error) {

	projectID := utility.GetEnv(constants.GKEClusterProjectID, "")
	Location := utility.GetEnv(constants.GKEClusterLocation, "")
	ClusterID := utility.GetEnv(constants.GKEClusterName, "")

	if projectID == "" || Location == "" || ClusterID == "" {
		return nil, errors.New(fmt.Sprintf("Unable to find necessary environment variables for gke cluster connection. Please make sure the runtime has the all the 3 required envioronment values set - %s, %s, %s", constants.GKEClusterName, constants.GKEClusterLocation, constants.GKEClusterProjectID))
	}

	return &clusterConn{Project: projectID, Location: Location, ClusterID: ClusterID}, nil
}

// getClusterInfo - Provides cluster details of the specified gke cluster
func (conn *clusterConn) GetClusterInfo() (*containerpb.Cluster, error) {

	ctx := context.Background()

	c, err := container.NewClusterManagerClient(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	req := &containerpb.GetClusterRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s/clusters/%s", conn.Project, conn.Location, conn.ClusterID),
	}
	resp, err := c.GetCluster(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// GetClusterPrivateEndpoint - Provides the private endpoint of the the API Server
func (conn *clusterConn) GetClusterPrivateEndpoint() (*string, error) {

	clusterDetails, err := conn.GetClusterInfo()

	if err != nil {
		log.Errorf("Error connecting to the gke cluster to fetch cluster details")
		return nil, err
	}

	if clusterDetails.PrivateClusterConfig != nil && clusterDetails.PrivateClusterConfig.PrivateEndpoint != "" {
		return &clusterDetails.PrivateClusterConfig.PrivateEndpoint, nil
	}

	return nil, errors.New("Cluster not configured with private endpoints, use the public endpoint for accessing the api server.")
}
