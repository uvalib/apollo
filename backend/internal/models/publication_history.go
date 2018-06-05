package models

import (
	"fmt"
	"log"
	"time"
)

// PublicationHistory contains publication details for a node
type PublicationHistory struct {
	ID     int64 `json:"-"`
	NodeID int64 `db:"node_id" json:"node_id"`
	User
	PublishedAt time.Time `db:"published_at" json:"published"`
}

// NodePublished adds a publication history record to the specified node (a collection)
func (db *DB) NodePublished(nodeID int64, computingID string) error {
	var userID int64
	db.QueryRow("select id from users where computing_id=?", computingID).Scan(&userID)
	if userID == 0 {
		return fmt.Errorf("User %s not found", computingID)
	}

	qs := "insert into publication_history (node_id, user_id, published_at) values (?,?,NOW())"
	_, err := db.Exec(qs, nodeID, userID)
	return err
}

// GetPublicationHistory returns all of the publication history for a node
func (db *DB) GetPublicationHistory(nodeID int64) ([]PublicationHistory, error) {
	log.Printf("Get publication history for %d", nodeID)
	var history []PublicationHistory
	qs := `select node_id, computing_id, last_name, first_name, email, published_at
      from publication_history p inner join users u on u.id = user_id
       where node_id=? order by published_at desc`
	err := db.Select(&history, qs, nodeID)
	return history, err
}

// GetLatestPublication will return the latest publication date for a node
func (db *DB) GetLatestPublication(nodeID int64) *time.Time {
	log.Printf("Get latest publication info for %d", nodeID)
	q := "select published_at from publication_history where node_id=? order by published_at desc limit 1"
	var ts *time.Time
	db.QueryRow(q, nodeID).Scan(&ts)
	return ts
}
