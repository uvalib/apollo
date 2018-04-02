package handlers

import (
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// NodesIndex : report the version of the serivce
//
func (h *ApolloHandler) NodesIndex(rw http.ResponseWriter, req *http.Request, params httprouter.Params) {
	fmt.Fprintf(rw, "NODE INDEX")
}
