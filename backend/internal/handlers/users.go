package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// UsersIndex returns a list of  Apollo users.
//
func (app *ApolloHandler) UsersIndex(c *gin.Context) {
	users := app.DB.AllUsers()
	c.JSON(http.StatusOK, users)
}

// UsersShow : return json detail for a user
//
func (app *ApolloHandler) UsersShow(c *gin.Context) {
	user, err := app.DB.FindUserBy("id", c.Param("id"))
	if err != nil {
		c.String(http.StatusNotFound, "User %s not found", c.Param("id"))
		return
	}

	c.JSON(http.StatusOK, user)
}
