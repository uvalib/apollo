package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// CollectionsIndex : report the version of the serivce
//
func (app *ApolloHandler) CollectionsIndex(rw http.ResponseWriter, req *http.Request, params httprouter.Params) {
	collections := app.DB.GetCollections()
	outJSON, _ := json.Marshal(collections)
	rw.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(rw, string(outJSON))
}

// CollectionsShow : get details of a collection by PID
//
func (app *ApolloHandler) CollectionsShow(rw http.ResponseWriter, req *http.Request, params httprouter.Params) {
	pid := params.ByName("pid")
	log.Printf("Get collection for PID %s", pid)
	rootID, dbErr := app.DB.GetNodeIDFromPID(pid)
	if dbErr != nil {
		log.Printf("ERROR: %s", dbErr.Error())
		http.Error(rw, dbErr.Error(), http.StatusNotFound)
		return
	}

	root, dbErr := app.DB.GetTree(rootID)
	if dbErr != nil {
		log.Printf("ERROR: %s", dbErr.Error())
		http.Error(rw, dbErr.Error(), http.StatusInternalServerError)
		return
	}
	root.PublishedAt = app.DB.GetLatestPublication(rootID)

	log.Printf("Tree retrieved; Marshall to JSON...")
	json, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		log.Printf("MArshal problem %s", err.Error())
	}
	log.Printf("DONE")

	rw.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(rw, string(json))
}
