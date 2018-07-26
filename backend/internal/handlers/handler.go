package handlers

import "github.com/uvalib/apollo/backend/internal/models"

// ApolloHandler is the basic handler for all serials requests.
// It contains common config information and services, like the DB
type ApolloHandler struct {
	Version         string
	DB              *models.DB
	DevAuthUser     string
	AuthComputingID string
	IIIF            string
	FedoraURL       string
	SolrDir         string
	QdcDir          string
}
