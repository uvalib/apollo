package services

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/uvalib/apollo/backend/internal/models"
)

// Hit is one match found in the search
type Hit struct {
	CollectionPID string `json:"collection_pid"`
	Title         string `json:"title"`
	CollectionURL string `json:"collection_url"`
	PID           string `json:"pid"`
	Type          string `json:"item_type"`
	Match         string `json:"match"`
	ItemURL       string `json:"item_url"`
}

// SearchResults contains all of the results for a search operation
type SearchResults struct {
	Hits           int   `json:"hits"`
	ResponseTimeMS int64 `json:"response_time_ms"`
	Results        []Hit `json:"results,omitempty"`
}

type hitRow struct {
	ID              int64  `db:"id"`
	PID             string `db:"pid"`
	Type            string `db:"type"`
	Ancestry        string `db:"ancestry"`
	Value           string `db:"value"`
	ControlledValue string `db:"controlled_value"`
}

// Search will search node values for the query string and return a struct containing match results
func (svc *Apollo) Search(query string) *SearchResults {
	query = strings.ToLower(query)
	start := time.Now()
	searchQ := `select n.id,n.pid,n.ancestry,nt.name as type,n.value,cv.value as controlled_value from nodes n 
		inner join node_types nt on nt.id=n.node_type_id
		left join controlled_values cv on cv.id=n.value
		where n.node_type_id != 6 and (n.value REGEXP ? or cv.value REGEXP ?)`
	rows, err := svc.DB.Queryx(searchQ, query, query)
	if err != nil {
		log.Printf("Query failed: %s", err.Error())
		elapsedNanoSec := time.Since(start)
		elapsedMS := int64(elapsedNanoSec / time.Millisecond)
		return &SearchResults{Hits: 0, ResponseTimeMS: elapsedMS}
	}

	// get minimal info on all collections; OID, PID and TItle. Only a few exist right
	// now, so this brute force grab is OK
	collections := svc.DB.GetCollections()

	apollorURL := fmt.Sprintf("https://%s", svc.Hostname)
	if svc.HTTPS == false {
		apollorURL = fmt.Sprintf("http://%s", svc.Hostname)
	}

	// Walk the rows from the search query and generate hit list for response
	out := SearchResults{}
	hits := 0
	for rows.Next() {
		hits++
		var hitRow hitRow
		rows.StructScan(&hitRow)

		hit := Hit{PID: hitRow.PID, Type: hitRow.Type}

		collID, _ := strconv.ParseInt(strings.Split(hitRow.Ancestry, "/")[0], 10, 64)
		for _, coll := range collections {
			if coll.ID == collID {
				hit.CollectionPID = coll.PID
				hit.Title = coll.Title
				hit.CollectionURL = fmt.Sprintf("%s/collections/%s", apollorURL, hit.CollectionPID)
				break
			}
		}

		// FIXME this is wrong; can only link direct to ITEMS, not the nodes that make them up
		hit.ItemURL = fmt.Sprintf("%s/collections/%s?item=%s", apollorURL, hit.CollectionPID, hitRow.PID)

		// see if the hit was in value or controlled value...
		if strings.Contains(strings.ToLower(hitRow.Value), query) {
			hit.Match = hitRow.Value
		} else {
			hit.Match = hitRow.ControlledValue
		}

		out.Results = append(out.Results, hit)
	}

	elapsedNanoSec := time.Since(start)
	elapsedMS := int64(elapsedNanoSec / time.Millisecond)
	out.ResponseTimeMS = elapsedMS
	out.Hits = hits
	return &out
}

// LookupIdentifier will accept any sort of known identifier and find a matching
// Apollo ItemID which includes internal ID and PID
func (svc *Apollo) LookupIdentifier(identifier string) (*models.NodeIdentifier, error) {
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
