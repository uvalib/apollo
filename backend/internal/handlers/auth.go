package handlers

import (
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// AuthHandler is middleware that will enforce user authentication based on Shibboleth headers
//
func (app *ApolloHandler) AuthHandler(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		computingID := req.Header.Get("remote_user")
		if len(app.DevAuthUser) > 0 {
			computingID = app.DevAuthUser
		}
		log.Printf("Authenticating request; remote_user [%s]", computingID)
		if len(computingID) == 0 {
			http.Error(w, "You are not authorized to access this site", http.StatusForbidden)
			return
		}
		user, err := app.DB.FindUserBy("computing_id", computingID)
		if err != nil {
			http.Error(w, "You are not authorized to access this site", http.StatusForbidden)
			return
		}
		log.Printf("User %s is authorized", user.ComputingID)
		next(w, req, ps)
	}
}
