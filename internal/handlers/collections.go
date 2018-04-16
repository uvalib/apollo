package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// CollectionsIndex : report the version of the serivce
//
func (app *ApolloHandler) CollectionsIndex(rw http.ResponseWriter, req *http.Request, params httprouter.Params) {
	pids := app.DB.GetCollectionPIDs()
	pidsJSON, _ := json.Marshal(pids)
	rw.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(rw, string(pidsJSON))
}
