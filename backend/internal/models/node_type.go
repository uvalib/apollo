package models

import (
	"fmt"
	"log"
	"time"
)

// NodeType is a controlled vocabulary for node names
type NodeType struct {
	ID              int64  `json:"-"`
	PID             string `json:"pid"`
	Name            string `json:"name"`
	ControlledVocab bool   `db:"controlled_vocab" json:"controlledVocab"`
	Validation      string `json:"validation,omitempty"`
	Container       bool   `json:"container"`
}

// AllTypes returns a list of all available names
func (db *DB) AllTypes() []NodeType {
	names := []NodeType{}
	db.Select(&names, "select * from node_types order by name asc")
	return names
}

// GetNodeType finds a node name record by name
func (db *DB) GetNodeType(name string) *NodeType {
	nn := NodeType{}
	err := db.Get(&nn, "SELECT * FROM node_types WHERE name=?", name)
	if err != nil {
		log.Printf("Unable to find node_type %s: %s", name, err.Error())
		return nil
	}
	return &nn
}

// CreateNodeType creates a nnew node name record and returns it
func (db *DB) CreateNodeType(name string) (*NodeType, error) {
	tx, err := db.Begin()
	if err != nil {
		log.Printf("ERROR: create transaction for node_type create: %s", err.Error())
		return nil, err
	}
	t := time.Now().Unix()
	tmpPID := fmt.Sprintf("TMP-%d", t)
	qs := "insert into node_types (pid,name) values (?,?)"
	res, err := tx.Exec(qs, tmpPID, name)
	if err != nil {
		log.Printf("ERROR: Unable to create node_type: %s", err.Error())
		tx.Rollback()
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		log.Printf("ERROR: unable to get lastInsertID: %s", err.Error())
		tx.Rollback()
		return nil, err
	}
	pid := fmt.Sprintf("uva-ann%d", id)
	qs = "update node_types set pid=? where id=?"
	res, err = tx.Exec(qs, pid, id)
	if err != nil {
		log.Printf("ERROR: unable to set node_type PID: %s", err.Error())
		tx.Rollback()
		return nil, err
	}
	err = tx.Commit()
	if err != nil {
		log.Printf("ERROR: node_type transaction commit failed %s", err.Error())
		tx.Rollback()
		return nil, err
	}
	return &NodeType{ID: id, PID: pid, Name: name}, nil
}
