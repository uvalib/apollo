package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthCheck : report health of this and associated services
//
func (h *Apollo) HealthCheck(c *gin.Context) {
	err := h.DB.Ping()
	if err != nil {
		log.Printf("Healthcheck failure: %s", err)
		// gin.H is a shortcut for map[string]interface{}
		c.JSON(http.StatusInternalServerError, gin.H{"alive": "true", "mysql": "false"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"alive": "true", "mysql": "true"})
}
