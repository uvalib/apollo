package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// ValuesForName will return a list of controlled values for a specific name
//
func (app *ApolloHandler) ValuesForName(rw http.ResponseWriter, req *http.Request, params httprouter.Params) {
	log.Printf("Get controlled values for '%s'", params.ByName("name"))
	values := app.DB.ListControlledValues(params.ByName("name"))
	var buffer bytes.Buffer
	for _, val := range values {
		json, _ := json.MarshalIndent(val, "", "  ")
		if buffer.Len() > 0 {
			buffer.WriteString(",")
		}
		buffer.WriteString(string(json))
	}
	rw.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(rw, fmt.Sprintf("[%s]", buffer.String()))
}
