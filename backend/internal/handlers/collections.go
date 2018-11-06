package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/uvalib/apollo/backend/internal/services"
)

// CollectionsIndex will report the version of the serivce
func (app *Apollo) CollectionsIndex(c *gin.Context) {
	collections := app.DB.GetCollections()
	c.JSON(http.StatusOK, collections)
}

// CollectionsShow will get details of a collection by PID
func (app *Apollo) CollectionsShow(c *gin.Context) {
	pid := c.Param("pid")
	log.Printf("Get collection for PID %s", pid)
	rootID, dbErr := services.LookupIdentifier(app.DB, pid)
	if dbErr != nil {
		log.Printf("ERROR: %s", dbErr.Error())
		c.String(http.StatusNotFound, dbErr.Error())
		return
	}

	root, dbErr := app.DB.GetTree(rootID.ID)
	if dbErr != nil {
		log.Printf("ERROR: %s", dbErr.Error())
		c.String(http.StatusInternalServerError, dbErr.Error())
		return
	}
	root.PublishedAt = app.DB.GetLatestPublication(rootID.ID)
	c.JSON(http.StatusOK, root)
}
