package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AriesLookup : query apollo for information on the supplied identifer
//
func (h *ApolloHandler) AriesLookup(c *gin.Context) {
	c.String(http.StatusOK, "")
}
