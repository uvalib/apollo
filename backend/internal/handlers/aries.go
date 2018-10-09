package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/uvalib/apollo/backend/internal/models"
)

// AriesService contains details for a service reference
type AriesService struct {
	URL      string `json:"url"`
	Protocol string `json:"protocol"`
}

// Aries is a structure containing the data to be returned from an aries query
type Aries struct {
	Identifiers []string       `json:"identifier,omitempty"`
	AdminURL    []string       `json:"administrative_url,omitempty"`
	Services    []AriesService `json:"service_url,omitempty"`
}

// AriesLookup will query apollo for information on the supplied identifer
func (h *ApolloHandler) AriesLookup(c *gin.Context) {
	passedPID := c.Param("id")
	ids, err := h.DB.Lookup(passedPID)
	if err != nil {
		c.String(http.StatusNotFound, "%s not found", passedPID)
		return
	}

	// Get the referenced node and the containing collection. No need
	// for error handlign because the PID was already matched up to a
	// node ID; just getting the rest of the data
	node, _ := h.DB.GetChildren(ids.ID)
	collection, _ := h.DB.GetParentCollection(node)

	var out Aries
	out.Identifiers = append(out.Identifiers, ids.PID)
	if ids.PID != passedPID {
		out.Identifiers = append(out.Identifiers, passedPID)
	} else {
		// this passed PID was apollo. See if an external PID exists
		extPID := getExternalPID(node)
		if extPID != "" {
			out.Identifiers = append(out.Identifiers, extPID)
		}
	}

	if ids.PID != collection.PID {
		// This is not the collection level node. Use the admin URL that links
		// directly to the item and do not include an iiif presentation service
		out.AdminURL = append(out.AdminURL, fmt.Sprintf("%s/#/collections/%s?item=%s", h.URL, collection.PID, ids.PID))
	} else {
		out.AdminURL = append(out.AdminURL, fmt.Sprintf("%s/#/collections/%s", h.URL, collection.PID))
	}

	// Get the PID for adigital object owned directly by this node
	// If one is found, add the IIIF manifest URL as a service
	dObjPID := digitalObjectPID(node)
	if dObjPID != "" {
		out.Services = append(out.Services,
			AriesService{URL: fmt.Sprintf("%s/%s", h.IIIF, dObjPID), Protocol: "iiif-presentation"})
	}
	out.Services = append(out.Services,
		AriesService{URL: fmt.Sprintf("%s/api/collections/%s", h.URL, ids.PID), Protocol: "json-metadata"})

	c.JSON(http.StatusOK, out)
}

func getExternalPID(node *models.Node) string {
	for _, child := range node.Children {
		if child.Type.Name == "externalPID" {
			return child.Value
		}
	}
	return ""
}

func digitalObjectPID(node *models.Node) string {
	hasDigitalObject := false
	extPID := ""
	for _, child := range node.Children {
		if child.Type.Name == "externalPID" {
			extPID = child.Value
		}
		if child.Type.Name == "digitalObject" {
			hasDigitalObject = true
		}
	}
	if hasDigitalObject {
		return extPID
	}
	return ""
}
