package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Aries is a structure containing the data to be returned from an aries query
type Aries struct {
	Identifiers []string       `json:"identifier,omitempty"`
	AdminURL    []string       `json:"administrative_url,omitempty"`
	Services    []AriesService `json:"service_url,omitempty"`
}

// AriesService contains details for a service reference
type AriesService struct {
	URL      string `json:"url"`
	Protocol string `json:"protocol"`
}

// AriesPing handles requests to the aries endpoint with no params.
// Just returns and alive message
func (app *Apollo) AriesPing(c *gin.Context) {
	c.String(http.StatusOK, "Apollo Aries API")
}

// AriesLookup will query apollo for information on the supplied identifer
func (app *Apollo) AriesLookup(c *gin.Context) {
	passedPID := c.Param("id")
	log.Printf("INFO: aries lookup %s", passedPID)
	ids, err := lookupIdentifier(&app.DB, passedPID)
	if err != nil {
		c.String(http.StatusNotFound, "%s not found", passedPID)
		return
	}

	// Get the referenced node and the containing collection. No need
	// for error handlign because the PID was already matched up to a
	// node ID; just getting the rest of the data
	node, _ := getNode(&app.DB, ids.ID)
	collection, _ := getNodeCollection(&app.DB, node)

	var out Aries
	out.Identifiers = append(out.Identifiers, ids.PID)
	extIds := getExternalIdentifiers(node)
	if len(extIds) > 0 {
		out.Identifiers = append(out.Identifiers, extIds...)
	}

	if ids.PID != collection.PID {
		// This is not the collection level node. Use the admin URL that links
		// directly to the item and do not include an iiif presentation service
		out.AdminURL = append(out.AdminURL, fmt.Sprintf("%s/collections/%s?item=%s", app.ApolloURL, collection.PID, ids.PID))
	} else {
		out.AdminURL = append(out.AdminURL, fmt.Sprintf("%s/collections/%s", app.ApolloURL, collection.PID))
	}

	out.Services = append(out.Services,
		AriesService{URL: fmt.Sprintf("%s/api/collections/%s", app.ApolloURL, ids.PID), Protocol: "json-metadata"})

	c.JSON(http.StatusOK, out)
}

func getExternalIdentifiers(node *Node) []string {
	var identifiers []string
	keys := []string{"externalPID", "barcode", "catalogKey", "callNumber", "wslsID"}
	for _, child := range node.Children {
		for _, key := range keys {
			if child.Type.Name == key {
				identifiers = append(identifiers, child.Value)
			}
		}
	}
	return identifiers
}
