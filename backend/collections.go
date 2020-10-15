package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ListCollections returns a json array containg all collection in tghe system
func (app *Apollo) ListCollections(c *gin.Context) {
	log.Printf("Get all collections")
	collections := getCollections(&app.DB)
	c.JSON(http.StatusOK, collections)
}

// GetCollection finds a collection by PID and returns details as json
func (app *Apollo) GetCollection(c *gin.Context) {
	pid := c.Param("pid")
	log.Printf("Get collection for PID %s", pid)
	rootID, dbErr := lookupIdentifier(&app.DB, pid)
	if dbErr != nil {
		log.Printf("ERROR: %s", dbErr.Error())
		c.String(http.StatusNotFound, dbErr.Error())
		return
	}

	root, dbErr := getTree(&app.DB, rootID.ID)
	if dbErr != nil {
		log.Printf("ERROR: %s", dbErr.Error())
		c.String(http.StatusInternalServerError, dbErr.Error())
		return
	}
	log.Printf("Collection tree retrieved from DB; sending to client")
	c.JSON(http.StatusOK, root)
}

// getCollections returns a list of all collections. Data is ID/PID/Title
func getCollections(db *DB) []Collection {
	var IDs []NodeIdentifier
	var out []Collection
	qs := "select id,pid from nodes where parent_id is null"
	db.Select(&IDs, qs)

	tq := "select value from nodes where ancestry=? and node_type_id=? order by id asc limit 1"
	for _, val := range IDs {
		var title string
		db.QueryRow(tq, val.ID, 2).Scan(&title)
		out = append(out, Collection{ID: val.ID, PID: val.PID, Title: title})
	}
	return out
}
