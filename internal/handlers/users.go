package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// UsersIndex : report the version of the serivce
//
func (h *ApolloHandler) UsersIndex(rw http.ResponseWriter, req *http.Request, params httprouter.Params) {
	users := h.DB.AllUsers()
	var buffer bytes.Buffer
	for _, user := range users {
		json := fmt.Sprintf(`{"id": %d, "email": "%s"}`, user.ID, user.Email)
		if buffer.Len() > 0 {
			buffer.WriteString(",")
		}
		buffer.WriteString(json)
	}
	rw.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(rw, fmt.Sprintf("[%s]", buffer.String()))
}

// UsersShow : return json detail for a user
//
func (h *ApolloHandler) UsersShow(rw http.ResponseWriter, req *http.Request, params httprouter.Params) {
	user, err := h.DB.FindUserBy("id", params.ByName("id"))
	if err != nil {
		http.Error(rw, "User not found", http.StatusNotFound)
		return
	}

	json, err := json.Marshal(user)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(rw, string(json))
}
