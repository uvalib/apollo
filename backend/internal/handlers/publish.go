package handlers

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/uvalib/apollo/backend/internal/models"
)

// PublishCollection generates the solr documents for all sections of the collection
// and tags the collection as having been published
func (app *Apollo) PublishCollection(c *gin.Context) {
	log.Printf("Publish collection '%s' to %s", c.Param("pid"), app.SolrDir)
	svc := app.InitServices(c)
	collectionIDs, err := svc.LookupIdentifier(c.Param("pid"))
	if err != nil {
		out := fmt.Sprintf("Collection %s not found", c.Param("pid"))
		c.String(http.StatusNotFound, out)
		return
	}

	// Get a list of identifters for all items in this collection. This
	// is a struct containing both PID and DB ID
	itemIDs, err := app.DB.GetCollectionItemIdentifiers(collectionIDs.ID, "all")
	if err != nil {
		out := fmt.Sprintf("Unable to retrieve collection items %s", err.Error())
		c.String(http.StatusInternalServerError, out)
		return
	}

	itemIDs = append(itemIDs, models.NodeIdentifier{ID: collectionIDs.ID, PID: c.Param("pid")})

	// setup a subdir for the dropoff, if it doesn already exist
	tgtPath := fmt.Sprintf("%s/%s", app.SolrDir, c.Param("pid"))
	if _, err := os.Stat(tgtPath); os.IsNotExist(err) {
		os.Mkdir(tgtPath, 0644)
	} else {
		os.Chown(tgtPath, 118698, 10708) // libsnlocal:	libr-snlocal
	}

	// Kick off the publication of all the items in the list in a goroutine
	go svc.PublishSolrForItems(tgtPath, itemIDs, collectionIDs.ID)

	c.String(http.StatusOK, "Publication of collection %s started", c.Param("pid"))
}
