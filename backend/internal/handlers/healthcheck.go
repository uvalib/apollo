package handlers

import (
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// HealthCheck : report health of this and associated services
//
func (h *ApolloHandler) HealthCheck(rw http.ResponseWriter, req *http.Request, params httprouter.Params) {
	rw.Header().Set("Content-Type", "application/json")
	_, err := h.DB.Query("SELECT 1")
	if err != nil {
		http.Error(rw, `{"alive": true, "mysql": false}`, http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(rw, `{"alive": true, "mysql": true}`)
}
