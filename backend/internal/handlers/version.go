package handlers

import (
	"github.com/gin-gonic/gin"
)

// VersionInfo : report the version of the serivce
//
func (h *Apollo) VersionInfo(c *gin.Context) {
	c.String(200, "Apollo version %s", h.Version)
}
