package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// TypesIndex will return a list of controlled vocabulary types
//
func (app *ApolloHandler) TypesIndex(rw http.ResponseWriter, req *http.Request, params httprouter.Params) {
	types := app.DB.AllTypes()
	var buffer bytes.Buffer
	for _, name := range types {
		json, _ := json.MarshalIndent(name, "", "  ")
		if buffer.Len() > 0 {
			buffer.WriteString(",")
		}
		buffer.WriteString(string(json))
	}
	rw.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(rw, fmt.Sprintf("[%s]", buffer.String()))
}
