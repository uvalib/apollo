package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CollectionsIndex : report the version of the serivce
//
func (app *ApolloHandler) CollectionsIndex(c *gin.Context) {
	collections := app.DB.GetCollections()
	c.JSON(http.StatusOK, collections)
}

// CollectionsShow : get details of a collection by PID
//
func (app *ApolloHandler) CollectionsShow(c *gin.Context) {
	pid := c.Param("pid")
	log.Printf("Get collection for PID %s", pid)
	rootID, dbErr := app.DB.GetNodeIDFromPID(pid)
	if dbErr != nil {
		log.Printf("ERROR: %s", dbErr.Error())
		c.String(http.StatusNotFound, dbErr.Error())
		return
	}

	root, dbErr := app.DB.GetTree(rootID)
	if dbErr != nil {
		log.Printf("ERROR: %s", dbErr.Error())
		c.String(http.StatusInternalServerError, dbErr.Error())
		return
	}
	root.PublishedAt = app.DB.GetLatestPublication(rootID)
	c.JSON(http.StatusOK, root)
}
