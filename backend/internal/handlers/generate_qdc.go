package handlers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"text/template"

	"github.com/julienschmidt/httprouter"
)

// qdcControlledValue is a controllev value and the source URI for
// a QDC entry value
type qdcControlledValue struct {
	Value    string
	ValueURI string
}

// wslsQdcData holds all of the data needed to populate the QDC XML template for an item
// in the collection.
// NOTE: much ow WSLS has hardoced values, so for now, this code is specific to that collection
// and simplified. Once new collections need this functionality, it will have to be generalized
type wslsQdcData struct {
	PID         string
	Title       string
	Description string
	DateCreated string
	RightsURI   string
	Duration    string
	Color       string
	Tag         string
	Places      []qdcControlledValue
	Topics      []qdcControlledValue
	Preview     string
}

// GenerateQDC generates the QDC XML documents needed to publish to the DPLA
func (app *ApolloHandler) GenerateQDC(rw http.ResponseWriter, req *http.Request, params httprouter.Params) {
	var data wslsQdcData
	data.PID = params.ByName("pid")
	data.Topics = append(data.Topics, qdcControlledValue{Value: "FakeTest", ValueURI: "http://reddit.com"})
	data.Topics = append(data.Topics, qdcControlledValue{Value: "NoURITest"})
	log.Printf("Generate QDC for collection %s", data.PID)

	destFilename := fmt.Sprintf("%s/%s.xml", app.QdcDir, data.PID)
	log.Printf("Open QDC results file: %s", destFilename)
	outFile, err := os.OpenFile(destFilename, os.O_CREATE|os.O_WRONLY, 0777)
	if err != nil {
		log.Printf("Unable to open destination QDC file %s: %s", destFilename, err.Error())
		return
	}
	defer outFile.Close()

	// NOTE text/templte must be used becatse html/template doesn't handle XML properly
	// Must escape all values added
	log.Printf("Render results")
	qdcTemplate := template.Must(template.ParseFiles("./templates/wsls_qdc.xml"))
	qdcTemplate.Execute(outFile, data)
	log.Printf("DONE")
}
