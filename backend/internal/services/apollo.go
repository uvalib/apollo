package services

import "github.com/uvalib/apollo/backend/internal/models"

// Apollo is the service context for all services. It contains common config parameters
// that may be needed for any service
type Apollo struct {
	HTTPS           bool
	Hostname        string
	DB              *models.DB
	IIIFManifestURL string
	AuthComputingID string
}
