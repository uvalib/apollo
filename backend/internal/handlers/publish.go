package handlers

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"

	"github.com/julienschmidt/httprouter"
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

	ids, err := app.DB.GetCollectionItemIDs(nodeID)
	if err != nil {
		out := fmt.Sprintf("Unable to retrieve collection items %s", err.Error())
		http.Error(rw, out, http.StatusInternalServerError)
		return
	}

	ids = append(ids, nodeID)

	// kick off the walk of tree and generate of solr in a goroutine
	go app.publishItems(ids, nodeID)

	fmt.Fprintf(rw, "Publication of collection %s started", params.ByName("pid"))
}

func (app *ApolloHandler) publishItems(IDs []int64, rootID int64) {
	// set up a wait group as big as the number of IDs to process
	var wg sync.WaitGroup
	wg.Add(len(IDs))

	// Kick off each generation in a goroutine
	for _, ID := range IDs {
		go func(id int64, dropoffDir string, iiif string, wg *sync.WaitGroup) {
			xml, err := app.DB.GetSolrXML(id, iiif)
			if err != nil {
				log.Printf("ERROR: Unable to generate solr xml for %d: %s", id, err.Error())
			} else {
				filename := fmt.Sprintf("%s/%d.xml", dropoffDir, id)
				ioutil.WriteFile(filename, []byte(xml), 0644)
			}
			wg.Done()
		}(ID, app.SolrDir, app.IIIF, &wg)
	}

	wg.Wait()
	log.Printf("All goroutines done; publication complete")
	app.DB.NodePublished(rootID, app.AuthComputingID)
}
