package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Authenticate will authenticate a user based on Shibboleth headers. Ths can be used in
// future to return auth tokens
func (app *Apollo) Authenticate(c *gin.Context) {
	computingID := c.GetHeader("remote_user")
	// if app.DevAuthUser != "" {
	// 	computingID = app.DevAuthUser
	// }
	// if computingID == "" {
	// 	c.String(http.StatusForbidden, "You are not authorized to access this site")
	// 	return
	// }
	// log.Printf("Authenticating remote_user [%s]", computingID)
	// user, err := app.DB.FindUserBy("computing_id", computingID)
	// if err != nil {
	// 	c.String(http.StatusForbidden, "You are not authorized to access this site")
	// 	return
	// }

	// // TODO generate an auth token and include with user?

	// log.Printf("User %s has successfully authorized", user.ComputingID)
	// c.JSON(http.StatusOK, user)
	log.Printf("Auth has been disabled; %s access granted", computingID)
	c.String(http.StatusOK, "Auth has been disabled, all can access")
}
