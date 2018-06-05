package handlers

import (
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// AuthMiddleware is middleware that will enforce user authentication based on Shibboleth headers
//
func (app *ApolloHandler) AuthMiddleware(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		// log.Printf("HEADERS: %s", req.Header)
		app.AuthComputingID = req.Header.Get("remote_user")
		if len(app.DevAuthUser) > 0 && app.AuthComputingID == "" {
			log.Printf("Authenticating using devMode user")
			app.AuthComputingID = app.DevAuthUser
		}
		log.Printf("Authenticating request; remote_user [%s]", app.AuthComputingID)
		if app.AuthComputingID == "" {
			http.Error(w, "You are not authorized to access this site", http.StatusForbidden)
			return
		}
		user, err := app.DB.FindUserBy("computing_id", app.AuthComputingID)
		if err != nil {
			http.Error(w, "You are not authorized to access this site", http.StatusForbidden)
			return
		}
		log.Printf("User %s is authorized for %s", user.ComputingID, req.RequestURI)
		w.Header().Set("cache-control", "private, max-age=0, no-cache")

		next(w, req, ps)
	}
}
