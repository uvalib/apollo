package handlers

import (
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/uvalib/apollo/backend/internal/models"
)

// PublishCollection generates the solr documents for all sections of the collection
// and tags the collection as having been published
func (app *ApolloHandler) PublishCollection(c *gin.Context) {
	log.Printf("Publish collection '%s' to %s", c.Param("pid"), app.SolrDir)
	nodeID, err := app.DB.GetNodeIDFromPID(c.Param("pid"))
	if err != nil {
		out := fmt.Sprintf("Collection %s not found", c.Param("pid"))
		c.String(http.StatusNotFound, out)
		return
	}

	// Get a list of identifters for all items in this collection. This
	// is a struct containing both PID and DB ID
	ids, err := app.DB.GetCollectionContainerIdentifiers(nodeID, "all")
	if err != nil {
		out := fmt.Sprintf("Unable to retrieve collection items %s", err.Error())
		c.String(http.StatusInternalServerError, out)
		return
	}

	ids = append(ids, models.ItemIDs{ID: nodeID, PID: c.Param("pid")})

	// kick off the walk of tree and generate of solr in a goroutine
	go app.publishItems(ids, nodeID)

	c.String(http.StatusOK, "Publication of collection %s started", c.Param("pid"))
}

func (app *ApolloHandler) publishItems(IDs []models.ItemIDs, rootID int64) {
	// chop up id list into blocks chunks that can be executed concurrenty
	// limit the maximum number of concurrrent generation threads to 50
	// to avoid choking the DB, tracksys or IIIF manifest service
	var chunks [][]models.ItemIDs
	var maxConcurrent = 10
	var chunkSize = int(math.Round(float64(len(IDs)) / float64(maxConcurrent)))
	if chunkSize == 0 {
		chunkSize = 1
	}
	for i := 0; i < len(IDs); i += chunkSize {
		endIdx := i + chunkSize
		if endIdx > len(IDs) {
			endIdx = len(IDs)
		}
		chunks = append(chunks, IDs[i:endIdx])
	}

	// set up a wait group as big as the number of IDs to process
	var wg sync.WaitGroup
	wg.Add(len(chunks))

	// Kick off  generation of each block of IDs in a goroutine
	for _, chunk := range chunks {
		go app.processIDs(chunk, &wg)
	}

	wg.Wait()
	log.Printf("All goroutines done; flagging publication complete by [%s]", app.AuthComputingID)
	app.DB.NodePublished(rootID, app.AuthComputingID)
	log.Printf("Publication COMPLETE")
}

func (app *ApolloHandler) processIDs(IDs []models.ItemIDs, wg *sync.WaitGroup) {
	log.Printf("GOROUTINE: Process %v", IDs)
	for _, ID := range IDs {
		xml, err := app.DB.GetSolrXML(ID.ID, app.IIIF)
		if err != nil {
			log.Printf("ERROR: Unable to generate solr xml for %d: %s", ID.ID, err.Error())
		} else {
			filename := fmt.Sprintf("%s/%s.xml", app.SolrDir, ID.PID)
			log.Printf("Write file %s", filename)
			ioutil.WriteFile(filename, []byte(xml), 0644)
		}
	}
	wg.Done()
}
