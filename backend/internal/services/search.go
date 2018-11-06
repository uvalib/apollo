package services

import (
	"fmt"
	"log"

	"github.com/uvalib/apollo/backend/internal/models"
)

// LookupIdentifier will accept any sort of known identifier and find a matching
// Apolloo ItemID which includes internal ID and PID
func LookupIdentifier(db *models.DB, identifier string) (*models.NodeIdentifier, error) {
	log.Printf("Lookup identifier %s", identifier)

	// First easy case; the identifier is an apollo PID
	var nodeID int64
	db.QueryRow("select id from nodes where pid=?", identifier).Scan(&nodeID)
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
	db.QueryRow(qs, identifier).Scan(&idType, &nodeID, &apolloPID)
	if apolloPID != "" {
		log.Printf("%s matches type %s. ApolloPID: %s ID: %d",
			identifier, idType, apolloPID, nodeID)
		return &models.NodeIdentifier{PID: apolloPID, ID: nodeID}, nil
	}

	return nil, fmt.Errorf("%s was not found", identifier)
}
