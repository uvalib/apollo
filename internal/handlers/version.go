package handlers

import (
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// VersionInfo : report the version of the serivce
//
func (h *SmsHandler) VersionInfo(rw http.ResponseWriter, req *http.Request, params httprouter.Params) {
	fmt.Fprintf(rw, "UVA Serials Manager version %s", h.Version)
}
