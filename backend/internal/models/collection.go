package models

import "fmt"

// Collection holds key data about a collection; its PID and Title
type Collection struct {
	ID    int64  `json:"-"`
	PID   string `json:"pid"`
	Title string `json:"title"`
}

// GetCollections returns a list of all collections. Data is PID/Title
func (db *DB) GetCollections() []Collection {
	var IDs []NodeIdentifier
	var out []Collection
	qs := "select id,pid from nodes where parent_id is null"
	tq := "select value from nodes where ancestry=? and node_type_id=? order by id asc limit 1"
	db.Select(&IDs, qs)

	for _, val := range IDs {
		var title string
		db.QueryRow(tq, val.ID, 2).Scan(&title)
		out = append(out, Collection{ID: val.ID, PID: val.PID, Title: title})
	}
	return out
}

// GetCollectionItemIdentifiers returns items (nodes that contain other nodes) owned by the collection. The itemType parameter
// controls which types of containers are returned. If you want everything, cet containerType to 'all'
func (db *DB) GetCollectionItemIdentifiers(collectionID int64, itemType string) ([]NodeIdentifier, error) {
	var ids []NodeIdentifier
	var err error
	if itemType == "all" {
		// Return ALL containers owned by the collection
		qs := fmt.Sprintf(`select n.id,n.pid from nodes n inner join node_types nt on nt.id = n.node_type_id
				 where nt.container = 1 and ancestry regexp '^%d($|/.*)' order by n.id asc;`, collectionID)
		err = db.Select(&ids, qs)
	} else {
		// Return only a specific type of container
		qs := fmt.Sprintf(`select n.id,n.pid from nodes n inner join node_types nt on nt.id = n.node_type_id
				 where nt.container = 1 and nt.name=? and ancestry regexp '^%d($|/.*)' order by n.id asc;`, collectionID)
		err = db.Select(&ids, qs, itemType)
	}

	if err != nil {
		return ids, err
	}

	return ids, nil
}
