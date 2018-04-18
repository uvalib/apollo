package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// UsersIndex returns a list of  Apollo users.
//
func (app *ApolloHandler) UsersIndex(rw http.ResponseWriter, req *http.Request, params httprouter.Params) {
	users := app.DB.AllUsers()
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
func (app *ApolloHandler) UsersShow(rw http.ResponseWriter, req *http.Request, params httprouter.Params) {
	user, err := app.DB.FindUserBy("id", params.ByName("id"))
	if err != nil {
		out := fmt.Sprintf("User %s not found", params.ByName("id"))
		http.Error(rw, out, http.StatusNotFound)
		return
	}

	json, _ := json.Marshal(user)
	rw.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(rw, string(json))
}
