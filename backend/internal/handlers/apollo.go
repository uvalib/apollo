package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/uvalib/apollo/backend/internal/models"
	"github.com/uvalib/apollo/backend/internal/services"
)

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

// InitServices will initailize a service context for running Apollo services
func (app *Apollo) InitServices(c *gin.Context) *services.Apollo {
	return &services.Apollo{DB: app.DB, Hostname: app.ApolloHost, 
		HTTPS: c.Request.TLS != nil, IIIFManifestURL: app.IIIF, AuthComputingID: app.AuthComputingID}
}
