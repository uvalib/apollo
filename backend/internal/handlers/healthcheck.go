package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthCheck : report health of this and associated services
//
func (h *ApolloHandler) HealthCheck(c *gin.Context) {
	_, err := h.DB.Query("SELECT 1")
	if err != nil {
		// gin.H is a shortcut for map[string]interface{}
		c.JSON(http.StatusInternalServerError, gin.H{"alive": "true", "mysql": "false"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"alive": "true", "mysql": "true"})
}
