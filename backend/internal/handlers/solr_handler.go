package handlers

import (
	"encoding/xml"
	"fmt"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type solrAdd struct {
	XMLName xml.Name `xml:"add"`
	Doc     solrDoc
}

type solrDoc struct {
	XMLName xml.Name `xml:"doc"`
	Fields  *[]solrField
}

type solrField struct {
	XMLName xml.Name `xml:"field"`
	Name    string   `xml:"name,attr"`
	Value   string   `xml:",chardata"`
}

// GenerateSolr generates a Solr Add document for ingest into Virgo3
// The general format is: <add><doc><field name="name"></field>, <field/>, ... </doc></add>
// If a field has multiple values, just add multiple field elements with
// the same name attribute
func (app *ApolloHandler) GenerateSolr(rw http.ResponseWriter, req *http.Request, params httprouter.Params) {
	log.Printf("Generate Solr Add for '%s'", params.ByName("pid"))

	var add solrAdd
	var fields []solrField
	fields = append(fields, solrField{Name: "id", Value: params.ByName("pid")})
	add.Doc.Fields = &fields

	xmlOut, err := xml.MarshalIndent(add, "", "  ")
	if err != nil {
		out := fmt.Sprintf("Unable to generate Solr doc for %s: %s", params.ByName("pid"), err.Error())
		http.Error(rw, out, http.StatusNotFound)
		return
	}

	// Note: using text/html because XML out is human-readable because
	// of the use of MarshalIndent above
	rw.Header().Set("Content-Type", "text/xml")
	fmt.Fprintf(rw, string(xmlOut))
}
