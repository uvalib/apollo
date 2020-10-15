package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetNodeTypes will return a list of controlled vocabulary types
func (app *Apollo) GetNodeTypes(c *gin.Context) {
	types := []NodeType{}
	err := app.DB.Select(&types, "select * from node_types order by name asc")
	if err != nil {
		log.Printf("ERROR: unable to get node types: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, types)
}

// GeControlledValues returns the controlled values for a type name
func (app *Apollo) GeControlledValues(c *gin.Context) {
	tgtName := c.Param("name")
	log.Printf("Get controlled values for '%s'", tgtName)
	var vals []ControlledValue
	err := app.DB.Select(&vals,
		"SELECT cv.* FROM controlled_values cv inner join node_types nt on nt.id = cv.node_type_id WHERE nt.name=?", tgtName)
	if err != nil {
		log.Printf("ERROR: Unable to get all controlled values for %s: %s", tgtName, err.Error())
		c.String(http.StatusNotFound, fmt.Sprintf("%s not found", tgtName))
		return
	}
	c.JSON(http.StatusOK, vals)
}

// GetControlledValueByName finds a controlled value record by name
func getControlledValueByName(db *DB, name string) (*ControlledValue, error) {
	cv := ControlledValue{}
	err := db.Get(&cv, "SELECT * FROM controlled_values WHERE value=?", name)
	if err != nil {
		return nil, err
	}
	return &cv, nil
}

// GetControlledValueByID finds a controlled value record by ID
func getControlledValueByID(db *DB, id int64) (*ControlledValue, error) {
	cv := ControlledValue{}
	err := db.Get(&cv, "SELECT * FROM controlled_values WHERE id=?", id)
	if err != nil {
		return nil, err
	}
	return &cv, nil
}
