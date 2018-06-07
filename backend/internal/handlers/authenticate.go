package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// Authenticate will authenticate a user based on Shibboleth headers. Ths can be used in
// future to return auth tokens
func (app *ApolloHandler) Authenticate(rw http.ResponseWriter, req *http.Request, params httprouter.Params) {
	computingID := req.Header.Get("remote_user")
	if app.DevAuthUser != "" {
		computingID = app.DevAuthUser
	}
	if computingID == "" {
		http.Error(rw, "You are not authorized to access this site", http.StatusForbidden)
		return
	}
	log.Printf("Authenticating remote_user [%s]", computingID)
	user, err := app.DB.FindUserBy("computing_id", computingID)
	if err != nil {
		http.Error(rw, "You are not authorized to access this site", http.StatusForbidden)
		return
	}

	// TODO generate an auth token and include with user?

	log.Printf("User %s has successfully authorized", user.ComputingID)
	json, _ := json.Marshal(user)
	rw.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(rw, string(json))
}
