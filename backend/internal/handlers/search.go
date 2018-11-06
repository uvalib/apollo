package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/uvalib/apollo/backend/internal/services"
)

// SearchHandler will search for the terms included in the query string in all collections
func (app *Apollo) SearchHandler(c *gin.Context) {
	qs := c.Query("q")
	if qs == "" {
		c.String(http.StatusBadRequest, "missing query term")
		return
	}
	res := services.Search(app.DB, qs)
	c.JSON(http.StatusOK, res)
}
