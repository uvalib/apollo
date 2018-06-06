package handlers

import (
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"sync"

	"github.com/julienschmidt/httprouter"
	"github.com/uvalib/apollo/backend/internal/models"
)

// PublishCollection generates the solr documents for all sections of the collection
// and tags the collection as having been published
func (app *ApolloHandler) PublishCollection(rw http.ResponseWriter, req *http.Request, params httprouter.Params) {
	log.Printf("Publish collection '%s' to %s", params.ByName("pid"), app.SolrDir)
	nodeID, err := app.DB.GetNodeIDFromPID(params.ByName("pid"))
	if err != nil {
		out := fmt.Sprintf("Collection %s not found", params.ByName("pid"))
		http.Error(rw, out, http.StatusNotFound)
		return
	}

	// Get a list of identifters for all items in this collection. This
	// is a struct containing both PID and DB ID
	ids, err := app.DB.GetCollectionItemIdentifiers(nodeID)
	if err != nil {
		out := fmt.Sprintf("Unable to retrieve collection items %s", err.Error())
		http.Error(rw, out, http.StatusInternalServerError)
		return
	}

	ids = append(ids, models.ItemIDs{ID: nodeID, PID: params.ByName("pid")})

	// kick off the walk of tree and generate of solr in a goroutine
	go app.publishItems(ids, nodeID)

	fmt.Fprintf(rw, "Publication of collection %s started", params.ByName("pid"))
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
	log.Printf("All goroutines done; publication complete")
	app.DB.NodePublished(rootID, app.AuthComputingID)
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
