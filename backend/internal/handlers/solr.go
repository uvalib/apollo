package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GenerateSolr generates a Solr Add document for ingest into Virgo.
func (app *Apollo) GenerateSolr(c *gin.Context) {
	log.Printf("Generate Solr Add for '%s'", c.Param("pid"))
	svc := app.InitServices(c)
	ids, err := svc.LookupIdentifier(c.Param("pid"))
	if err != nil {
		out := fmt.Sprintf("Unable to find PID %s : %s", c.Param("pid"), err.Error())
		c.String(http.StatusNotFound, out)
		return
	}

	xmlOut, err := svc.GetSolrXML(ids.ID)
	if err != nil {
		out := fmt.Sprintf("Unable to generate Solr doc for %s: %s", c.Param("pid"), err.Error())
		c.String(http.StatusNotFound, out)
		return
	}

	c.Header("Content-Type", "text/xml")
	c.String(http.StatusOK, xmlOut)
}
