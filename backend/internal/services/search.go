package services

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/uvalib/apollo/backend/internal/models"
)

// CollectionHit contains all of the search hits grouped by collection
type CollectionHit struct {
	ID    int64  `json:"-"`
	PID   string `json:"collection_pid"`
	Title string `json:"collection_title"`
	URL   string `json:"collection_url"`
	Hits  *[]Hit `json:"hits"`
}

// Hit is one match found in the search
type Hit struct {
	PID     string `json:"pid"`
	Title   string `json:"title,omitempty"`
	Type    string `json:"match_type"`
	Match   string `json:"match"`
	ItemURL string `json:"item_url"`
}

// SearchResults contains all of the results for a search operation
type SearchResults struct {
	Hits           int             `json:"hits"`
	ResponseTimeMS int64           `json:"response_time_ms"`
	Results        []CollectionHit `json:"results"`
}

type hitRow struct {
	ID              int64  `db:"id"`
	PID             string `db:"pid"`
	ParentID        int64  `db:"parent_id"`
	Type            string `db:"type"`
	Ancestry        string `db:"ancestry"`
	Value           string `db:"value"`
	ControlledValue string `db:"controlled_value"`
}

// Search will search node values for the query string and return a struct containing match results
func (svc *ApolloSvc) Search(query string) *SearchResults {
	query = strings.ToLower(query)
	start := time.Now()
	searchQ := `select n.id,n.pid,n.parent_id,n.ancestry,nt.name as type,n.value,cv.value as controlled_value from nodes n 
		inner join node_types nt on nt.id=n.node_type_id
		left join controlled_values cv on cv.id=n.value
		where n.node_type_id != 6 and (n.value REGEXP ? or (cv.value REGEXP ? and nt.controlled_vocab = 1))`
	rows, err := svc.DB.Queryx(searchQ, query, query)
	if err != nil {
		log.Printf("Query failed: %s", err.Error())
		elapsedNanoSec := time.Since(start)
		elapsedMS := int64(elapsedNanoSec / time.Millisecond)
		return &SearchResults{Hits: 0, ResponseTimeMS: elapsedMS}
	}

	// init blank search resutls and PID/ID scoreboard
	collections := make([]CollectionHit, 0)
	hits := 0
	pidMap := make(map[int64]string)

	// get minimal info on all collections; OID, PID and TItle. Only a few exist right
	// now, so this brute force grab is OK
	for _, coll := range svc.DB.GetCollections() {
		hits := make([]Hit, 0)
		collInfo := CollectionHit{ID: coll.ID, PID: coll.PID, Title: coll.Title,
			URL: fmt.Sprintf("%s/collections/%s", svc.ApolloURL, coll.PID), Hits: &hits}
		collections = append(collections, collInfo)
	}

	// Walk the rows from the search query and generate hit list for response
	var hitCollection *CollectionHit
	for rows.Next() {
		// Parse the hit row, figure out which collection it was from (first part of ancestry)
		// and find the matching CollectionHits object. It will be used to track this hit.
		var hitRow hitRow
		rows.StructScan(&hitRow)
		hit := Hit{Type: hitRow.Type}
		collID, _ := strconv.ParseInt(strings.Split(hitRow.Ancestry, "/")[0], 10, 64)
		for _, coll := range collections {
			if coll.ID == collID {
				hitCollection = &coll
				break
			}
		}

		// Lookup the PID of the parent container of the hit node
		parentPID := pidMap[hitRow.ParentID]
		if parentPID == "" {
			// Mapping not found, look it up in DB and cache it in the map
			pq := "select pid from nodes where id=?"
			svc.DB.Get(&parentPID, pq, hitRow.ParentID)
			pidMap[hitRow.ParentID] = parentPID
			hit.PID = parentPID
		} else {
			continue
		}

		// For non-top-level items, add a query param that allows a link directly to that item
		if parentPID != hitCollection.PID {
			hit.ItemURL = fmt.Sprintf("%s/collections/%s?item=%s", svc.ApolloURL, hitCollection.PID, parentPID)
			if hit.Type != "title" {
				// Non-title hit, grab the title for some context
				pq := "select value from nodes where parent_id=? and node_type_id=2"
				svc.DB.Get(&hit.Title, pq, hitRow.ParentID)
			}
		} else {
			hit.ItemURL = fmt.Sprintf("%s/collections/%s", svc.ApolloURL, hitCollection.PID)
		}

		// see if the hit was in value or controlled value...
		if strings.Contains(strings.ToLower(hitRow.Value), query) {
			hit.Match = hitRow.Value
		} else {
			hit.Match = hitRow.ControlledValue
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
		log.Printf("Search for %s found no matches", query)
		out.Results = []CollectionHit{}
	}
	elapsedNanoSec := time.Since(start)
	elapsedMS := int64(elapsedNanoSec / time.Millisecond)
	out.ResponseTimeMS = elapsedMS
	out.Hits = hits
	return &out
}

// LookupIdentifier will accept any sort of known identifier and find a matching
// Apollo ItemID which includes internal ID and PID
func (svc *ApolloSvc) LookupIdentifier(identifier string) (*models.NodeIdentifier, error) {
	log.Printf("Lookup identifier %s", identifier)

	// First easy case; the identifier is an apollo PID
	var nodeID int64
	svc.DB.QueryRow("select id from nodes where pid=?", identifier).Scan(&nodeID)
	if nodeID > 0 {
		log.Printf("%s is an ApolloPID. ID: %d", identifier, nodeID)
		return &models.NodeIdentifier{PID: identifier, ID: nodeID}, nil
	}

	// Next case; See if it is an externalPID, barcode, catalog key, call number or WSLS ID
	var apolloPID string
	var idType string
	qs := `SELECT t.name, np.id, np.pid FROM nodes ns INNER JOIN nodes np ON np.id = ns.parent_id
			 inner join node_types t on t.id = ns.node_type_id
	 		 WHERE ns.value=? and t.id in (5,9,10,13,23)`
	svc.DB.QueryRow(qs, identifier).Scan(&idType, &nodeID, &apolloPID)
	if apolloPID != "" {
		log.Printf("%s matches type %s. ApolloPID: %s ID: %d",
			identifier, idType, apolloPID, nodeID)
		return &models.NodeIdentifier{PID: apolloPID, ID: nodeID}, nil
	}

	return nil, fmt.Errorf("%s was not found", identifier)
}
