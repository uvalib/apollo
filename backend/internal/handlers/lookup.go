package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ExternalPIDLookup find an external system PID and return the corresponding Apollo PID
//
func (app *ApolloHandler) ExternalPIDLookup(c *gin.Context) {
	pid, err := app.DB.ExternalPIDLookup(c.Param("pid"))
	if err != nil {
		c.String(http.StatusNotFound, err.Error())
		return
	}
	c.String(http.StatusOK, pid)
}
