package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ValuesForName will return a list of controlled values for a specific name
//
func (app *ApolloHandler) ValuesForName(c *gin.Context) {
	log.Printf("Get controlled values for '%s'", c.Param("name"))
	values := app.DB.ListControlledValuesFoName(c.Param("name"))
	c.JSON(http.StatusOK, values)
}
