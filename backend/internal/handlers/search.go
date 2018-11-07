package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// SearchHandler will search for the terms included in the query string in all collections
func (app *Apollo) SearchHandler(c *gin.Context) {
	qs := c.Query("q")
	if qs == "" {
		c.String(http.StatusBadRequest, "missing query term")
		return
	}
	svc := app.InitServices(c)
	res := svc.Search(qs)
	c.JSON(http.StatusOK, res)
}
