package handlers

import "github.com/uvalib/apollo/backend/internal/models"

// Apollo is the applicatin object through which all requests are handled.
// It contains common config information and services, like the DB
type Apollo struct {
	Version         string
	ApolloHost      string
	DB              *models.DB
	DevAuthUser     string
	AuthComputingID string
	IIIF            string
	FedoraURL       string
	SolrDir         string
	QdcDir          string
}
