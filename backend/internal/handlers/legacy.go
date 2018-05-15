package handlers

import (
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// LegacyLookup find a legacy tracksys PID and return the corresponding Apollo PID
//
func (app *ApolloHandler) LegacyLookup(rw http.ResponseWriter, req *http.Request, params httprouter.Params) {
	pid, err := app.DB.LegacyLookup(params.ByName("pid"))
	if err != nil {
		http.Error(rw, err.Error(), http.StatusNotFound)
		return
	}
	rw.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(rw, pid)
}
