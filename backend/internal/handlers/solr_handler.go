package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// GenerateSolr generates a Solr Add document for ingest into Virgo3
// The general format is: <add><doc><field name="name"></field>, <field/>, ... </doc></add>
// If a field has multiple values, just add multiple field elements with
// the same name attribute
func (app *ApolloHandler) GenerateSolr(rw http.ResponseWriter, req *http.Request, params httprouter.Params) {
	// Assume the PID is external... find it; fail if not found
	log.Printf("Generate Solr Add for '%s'", params.ByName("pid"))
	apolloPID, err := app.DB.ExternalPIDLookup(params.ByName("pid"))
	if err != nil {
		out := fmt.Sprintf("Unable to find PID %s : %s", params.ByName("pid"), err.Error())
		http.Error(rw, out, http.StatusNotFound)
		return
	}

	xmlOut, err := app.DB.GetSolrXML(apolloPID)
	if err != nil {
		out := fmt.Sprintf("Unable to generate Solr doc for %s: %s", apolloPID, err.Error())
		http.Error(rw, out, http.StatusNotFound)
		return
	}

	// Note: using text/html because XML out is human-readable because
	// of the use of MarshalIndent above
	rw.Header().Set("Content-Type", "text/xml")
	fmt.Fprintf(rw, string(xmlOut))
}
