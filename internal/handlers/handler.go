package handlers

import "github.com/uvalib/apollo/internal/models"

// SmsHandler is the basic handler for all serials requests.
// It contains common config information and services, like the DB
type SmsHandler struct {
	Version string
	DB      *models.DB
}
