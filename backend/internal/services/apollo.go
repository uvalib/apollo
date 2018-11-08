package services

import "github.com/uvalib/apollo/backend/internal/models"

// ApolloSvc is the service context for all services. It contains common config parameters
// that may be needed for any service
type ApolloSvc struct {
	ApolloURL       string
	DB              *models.DB
	IIIFManifestURL string
	AuthComputingID string
	FedoraURL       string
}
