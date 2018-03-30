package handlers

import (
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// UsersIndex : report the version of the serivce
//
func (h *SmsHandler) UsersIndex(rw http.ResponseWriter, req *http.Request, params httprouter.Params) {
	fmt.Fprintf(rw, "USERS INDEX")
}
