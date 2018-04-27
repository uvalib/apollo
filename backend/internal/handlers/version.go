package handlers

import (
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// VersionInfo : report the version of the serivce
//
func (h *ApolloHandler) VersionInfo(rw http.ResponseWriter, req *http.Request, params httprouter.Params) {
	fmt.Fprintf(rw, "Apollo version %s", h.Version)
}
