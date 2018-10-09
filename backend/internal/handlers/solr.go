package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GenerateSolr generates a Solr Add document for ingest into Virgo3
// The general format is: <add><doc><field name="name"></field>, <field/>, ... </doc></add>
// If a field has multiple values, just add multiple field elements with
// the same name attribute
func (app *ApolloHandler) GenerateSolr(c *gin.Context) {
	log.Printf("Generate Solr Add for '%s'", c.Param("pid"))
	ids, err := app.DB.Lookup(c.Param("pid"))
	if err != nil {
		out := fmt.Sprintf("Unable to find PID %s : %s", c.Param("pid"), err.Error())
		c.String(http.StatusNotFound, out)
		return
	}

	xmlOut, err := app.DB.GetSolrXML(ids.ID, app.IIIF)
	if err != nil {
		out := fmt.Sprintf("Unable to generate Solr doc for %s: %s", c.Param("pid"), err.Error())
		c.String(http.StatusNotFound, out)
		return
	}

	// Note: using text/html because XML out is human-readable because
	// of the use of MarshalIndent above
	c.Header("Content-Type", "text/xml")
	c.String(http.StatusOK, xmlOut)
}
