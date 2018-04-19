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

// CollectionsShow : get details of a collection by PID
//
func (app *ApolloHandler) CollectionsShow(rw http.ResponseWriter, req *http.Request, params httprouter.Params) {
	pid := params.ByName("pid")
	rw.Header().Set("Content-Type", "application/json")
	root, dbErr := app.DB.GetCollection(pid)
	if dbErr != nil {
		http.Error(rw, dbErr.Error(), http.StatusNotFound)
		return
	}
	json, _ := json.MarshalIndent(root, "", "  ")
	fmt.Fprintf(rw, string(json))
}
