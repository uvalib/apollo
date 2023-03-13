package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// CollectionHit contains all of the search hits grouped by collection
type CollectionHit struct {
	ID    int64        `json:"-"`
	PID   string       `json:"collection_pid"`
	Title string       `json:"collection_title"`
	URL   string       `json:"collection_url"`
	Hits  *[]SearchHit `json:"hits"`
}

// SearchHit is one match found in the search
type SearchHit struct {
	PID     string `json:"pid"`
	Title   string `json:"title,omitempty"`
	Type    string `json:"match_type"`
	Match   string `json:"match"`
	ItemURL string `json:"item_url"`
}

// SearchResults contains all of the results for a search operation
type SearchResults struct {
	Hits           int             `json:"total"`
	ResponseTimeMS int64           `json:"response_time_ms"`
	Results        []CollectionHit `json:"collections"`
	Status         int             `json:"-"`
	Message        string          `json:"-"`
}

// SearchHandler will search for the terms included in the query string in all collections
func (app *Apollo) SearchHandler(c *gin.Context) {
	qs := c.Query("q")
	if qs == "" {
		c.String(http.StatusBadRequest, "missing query term")
		return
	}
	res := app.searchAll(qs)
	c.JSON(http.StatusOK, res)
}

// Search will search node values for the query string and return a struct containing match results
func (app *Apollo) searchAll(query string) *SearchResults {
	query = strings.ToLower(query)
	start := time.Now()
	searchQ := `select n.id,n.pid,n.parent_id,np.pid as parent_pid,n.ancestry,nt.name as type,n.value,cv.value as controlled_value from nodes n
		inner join node_types nt on nt.id=n.node_type_id
		inner join nodes np on np.id = n.parent_id
		left join controlled_values cv on cv.id=n.value
		where n.node_type_id != 6 and (n.value REGEXP ? or (cv.value REGEXP ? and nt.controlled_vocab = 1))`
	rows, err := app.DB.Queryx(searchQ, query, query)
	if err != nil {
		log.Printf("ERROR: Search for %s failed: %s", query, err.Error())
		elapsed := time.Since(start)
		elapsedMS := int64(elapsed / time.Millisecond)
		return &SearchResults{Hits: 0, ResponseTimeMS: elapsedMS, Status: http.StatusNotFound, Message: err.Error()}
	}

	// init blank search resutls and PID/ID scoreboard
	collections := make([]CollectionHit, 0)
	hits := 0
	// pidMap := make(map[int64]string)

	// get minimal info on all collections; OID, PID and TItle. Only a few exist right
	// now, so this brute force grab is OK
	for _, coll := range getCollections(&app.DB) {
		hits := make([]SearchHit, 0)
		collInfo := CollectionHit{ID: coll.ID, PID: coll.PID, Title: coll.Title,
			URL: fmt.Sprintf("%s/collections/%s", app.ApolloURL, coll.PID), Hits: &hits}
		collections = append(collections, collInfo)
	}

	type hitRow struct {
		ID              int64  `db:"id"`
		PID             string `db:"pid"`
		ParentID        int64  `db:"parent_id"`
		ParentPID       string `db:"parent_pid"`
		Type            string `db:"type"`
		Ancestry        string `db:"ancestry"`
		Value           string `db:"value"`
		ControlledValue string `db:"controlled_value"`
	}

	// Walk the rows from the search query and generate hit list for response
	var hitCollection *CollectionHit
	for rows.Next() {
		// Parse the hit row, figure out which collection it was from (first part of ancestry)
		// and find the matching CollectionHits object. It will be used to track this hit.
		var hr hitRow
		rows.StructScan(&hr)
		hit := SearchHit{Type: hr.Type}
		collID, _ := strconv.ParseInt(strings.Split(hr.Ancestry, "/")[0], 10, 64)
		for _, coll := range collections {
			if coll.ID == collID {
				hitCollection = &coll
				break
			}
		}

		// // Lookup the PID of the parent container of the hit node
		// parentPID := pidMap[hr.ParentID]
		// if parentPID == "" {
		// 	// Mapping not found, look it up in DB and cache it in the map
		// 	log.Printf("LOOKUP PID")
		// 	pq := "select pid from nodes where id=?"
		// 	app.DB.Get(&parentPID, pq, hr.ParentID)
		// 	pidMap[hr.ParentID] = parentPID
		// 	hit.PID = parentPID
		// } else {
		// 	continue
		// }

		// For non-top-level items, add a query param that allows a link directly to that item
		if hr.ParentPID != hitCollection.PID {
			hit.ItemURL = fmt.Sprintf("%s/collections/%s?item=%s", app.ApolloURL, hitCollection.PID, hr.ParentPID)
			if hit.Type != "title" {
				// Non-title hit, grab the title for some context
				pq := "select value from nodes where parent_id=? and node_type_id=2"
				app.DB.Get(&hit.Title, pq, hr.ParentID)
			}
		} else {
			hit.ItemURL = fmt.Sprintf("%s/collections/%s", app.ApolloURL, hitCollection.PID)
		}

		// see if the hit was in value or controlled value...
		if strings.Contains(strings.ToLower(hr.Value), query) {
			hit.Match = hr.Value
		} else {
			hit.Match = hr.ControlledValue
		}

		hits++
		*hitCollection.Hits = append(*hitCollection.Hits, hit)
	}

	// Fill in the final results
	out := SearchResults{}
	if hits > 0 {
		for _, collResults := range collections {
			if collResults.Hits != nil && len(*collResults.Hits) > 0 {
				out.Results = append(out.Results, collResults)
			}
		}
	} else {
		log.Printf("INFO: search for %s found no matches", query)
		out.Results = []CollectionHit{}
	}
	elapsed := time.Since(start)
	elapsedMS := int64(elapsed / time.Millisecond)
	out.ResponseTimeMS = elapsedMS
	out.Hits = hits
	return &out
}

// lookupIdentifier will accept any sort of known identifier and find a matching
// Apollo ItemID which includes internal ID and PID
func lookupIdentifier(db *DB, identifier string) (*NodeIdentifier, error) {
	log.Printf("INFO: lookup identifier %s", identifier)

	// First easy case; the identifier is an apollo PID
	var nodeID int64
	db.QueryRow("select id from nodes where pid=?", identifier).Scan(&nodeID)
	if nodeID > 0 {
		log.Printf("INFO: %s is an ApolloPID. ID: %d", identifier, nodeID)
		return &NodeIdentifier{PID: identifier, ID: nodeID}, nil
	}

	// Next case; See if it is an externalPID, barcode, catalog key, call number or WSLS ID
	var apolloPID string
	var idType string
	qs := `SELECT t.name, np.id, np.pid FROM nodes ns INNER JOIN nodes np ON np.id = ns.parent_id
			 inner join node_types t on t.id = ns.node_type_id
	 		 WHERE ns.value=? and t.id in (5,9,10,13,23)`
	db.QueryRow(qs, identifier).Scan(&idType, &nodeID, &apolloPID)
	if apolloPID != "" {
		log.Printf("INFO: %s matches type %s. ApolloPID: %s ID: %d",
			identifier, idType, apolloPID, nodeID)
		return &NodeIdentifier{PID: apolloPID, ID: nodeID}, nil
	}

	return nil, fmt.Errorf("%s was not found", identifier)
}
