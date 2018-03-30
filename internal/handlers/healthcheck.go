package handlers

import (
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// HealthCheck : report health of this and associated services
//
func (h *SmsHandler) HealthCheck(rw http.ResponseWriter, req *http.Request, params httprouter.Params) {
	rw.Header().Set("Content-Type", "application/json")
	err := h.DB.Ping()
	if err != nil {
		http.Error(rw, `{"alive": true, "mysql": false}`, http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(rw, `{"alive": true, "mysql": true}`)
}
