package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// NamesIndex will return a list of controlled vocabulary names
//
func (app *ApolloHandler) NamesIndex(rw http.ResponseWriter, req *http.Request, params httprouter.Params) {
	names := app.DB.AllNames()
	var buffer bytes.Buffer
	for _, name := range names {
		json, _ := json.MarshalIndent(name, "", "  ")
		if buffer.Len() > 0 {
			buffer.WriteString(",")
		}
		buffer.WriteString(string(json))
	}
	rw.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(rw, fmt.Sprintf("[%s]", buffer.String()))
}
