package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/uvalib/apollo/backend/internal/services"
)

// ItemShow will return a block of JSON metadata for the specified ITEM PID. This includes
// details of the specific item as well as some basic data amout the colection it
// belongs to.
func (app *Apollo) ItemShow(c *gin.Context) {
	pid := c.Param("pid")
	itemIDs, dbErr := services.LookupIdentifier(app.DB, pid)
	if dbErr != nil {
		log.Printf("ERROR: %s", dbErr.Error())
		c.String(http.StatusNotFound, dbErr.Error())
		return
	}

	item, dbErr := app.DB.GetItem(itemIDs.ID)
	if dbErr != nil {
		log.Printf("ERROR: %s", dbErr.Error())
		c.String(http.StatusNotFound, dbErr.Error())
		return
	}

	// note: if above was successful, this will be as well
	parent, _ := app.DB.GetParentCollection(item)

	jsonItem, _ := json.MarshalIndent(item, "", "  ")
	jsonParent, _ := json.MarshalIndent(parent, "", "  ")
	out := fmt.Sprintf("{\n\"collection\": %s,\n\"item\": %s}", jsonParent, jsonItem)
	c.Header("Content-Type", "application/json")
	c.String(http.StatusOK, out)
}
