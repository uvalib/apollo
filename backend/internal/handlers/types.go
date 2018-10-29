package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// TypesIndex will return a list of controlled vocabulary types
//
func (app *Apollo) TypesIndex(c *gin.Context) {
	types := app.DB.ListNodeTypes()
	c.JSON(http.StatusOK, types)
}
