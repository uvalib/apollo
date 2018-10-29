package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// UsersIndex returns a list of  Apollo users.
//
func (app *Apollo) UsersIndex(c *gin.Context) {
	users := app.DB.ListUsers()
	c.JSON(http.StatusOK, users)
}

// UsersShow : return json detail for a user
//
func (app *Apollo) UsersShow(c *gin.Context) {
	user, err := app.DB.FindUserBy("id", c.Param("id"))
	if err != nil {
		c.String(http.StatusNotFound, "User %s not found", c.Param("id"))
		return
	}

	c.JSON(http.StatusOK, user)
}
