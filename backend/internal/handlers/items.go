package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// ItemShow will return a block of JSON metadata for the specified ITEM PID. This includes
// details of the specific item as well as some basic data amout the colection it
// belongs to.
//
func (app *ApolloHandler) ItemShow(rw http.ResponseWriter, req *http.Request, params httprouter.Params) {
	pid := params.ByName("pid")
	item, dbErr := app.DB.GetChildren(pid)
	if dbErr != nil {
		http.Error(rw, dbErr.Error(), http.StatusNotFound)
		return
	}

	// note: if above was successful, this will be as well
	parent, _ := app.DB.GetParentCollection(pid)

	jsonItem, _ := json.MarshalIndent(item, "", "  ")
	jsonParent, _ := json.MarshalIndent(parent, "", "  ")
	out := fmt.Sprintf("{\n\"collection\": %s,\n\"item\": %s}", jsonParent, jsonItem)
	rw.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(rw, out)
}
