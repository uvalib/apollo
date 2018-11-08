package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// PublishCollectionQDC generates the QDC XML documents needed to publish to the DPLA
// NOTE: Test with this: curl -X POST http://localhost:8085/api/qdc/[PID]
func (app *Apollo) PublishCollectionQDC(c *gin.Context) {
	pid := c.Param("pid")
	limit, err := strconv.Atoi(c.Query("limit"))
	if err != nil {
		limit = -1
	}

	// HACK for now, only WSLS is an option. Choke on all other pids
	if pid != "uva-an109873" {
		out := fmt.Sprintf("QDC generation is not supported for %s", pid)
		c.String(http.StatusBadRequest, out)
		return
	}

	// lookup identifiers for the passed PID
	svc := app.InitServices(c)
	ids, err := svc.LookupIdentifier(pid)
	if err != nil {
		c.String(http.StatusNotFound, "%s not found", pid)
		return
	}

	// Get a list of identifters for all items in this collection. This
	// is a struct containing both PID and DB ID. Items are the only thing
	// that goes to DPLA
	itemIDs, err := app.DB.GetCollectionItemIdentifiers(ids.ID, "item")
	if err != nil {
		out := fmt.Sprintf("Unable to retrieve collection items %s", err.Error())
		c.String(http.StatusInternalServerError, out)
		return
	}

	// kick off the generation of QDC in a goroutine. Results written to QDC out dir
	go svc.PublishQDCForItems(app.QdcDir, ids.ID, itemIDs, limit)
	c.String(http.StatusOK, "QDC is being generated to %s...", app.QdcDir)
}

// GenerateQDC generates the QDC XML document for a single item
func (app *Apollo) GenerateQDC(c *gin.Context) {
	pid := c.Param("pid")
	svc := app.InitServices(c)
	ids, err := svc.LookupIdentifier(pid)
	if err != nil {
		c.String(http.StatusNotFound, "%s not found", pid)
		return
	}

	qdc, err := svc.GenerateQDC(ids)
	if err != nil {
		c.String(http.StatusNotFound, err.Error())
	} else {
		c.Header("Content-Type", "text/xml")
		c.String(http.StatusOK, qdc)
	}
}
