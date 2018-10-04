package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware is middleware that will enforce user authentication based on Shibboleth headers
//
func (app *ApolloHandler) AuthMiddleware(c *gin.Context) {
	// log.Printf("HEADERS: %s", req.Header)
	app.AuthComputingID = c.GetHeader("remote_user")
	if len(app.DevAuthUser) > 0 && app.AuthComputingID == "" {
		log.Printf("Authenticating using devMode user")
		app.AuthComputingID = app.DevAuthUser
	}
	log.Printf("Authenticating request; remote_user [%s]", app.AuthComputingID)
	if app.AuthComputingID == "" {
		c.String(http.StatusForbidden, "You are not authorized to access this site")
		return
	}
	user, err := app.DB.FindUserBy("computing_id", app.AuthComputingID)
	if err != nil {
		c.String(http.StatusForbidden, "You are not authorized to access this site")
		return
	}
	log.Printf("User %s is authorized for %s", user.ComputingID, c.Request.RequestURI)
	c.Header("cache-control", "private, max-age=0, no-cache")

	c.Next()
}
