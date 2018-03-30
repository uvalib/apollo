package models

import (
	"fmt"
	"log"
)

// NodeName is a controlled vocabulary for node names
type NodeName struct {
	ID    int64
	PID   string
	Value string
}

// GetNodeName finds a node name record by name
func (db *DB) GetNodeName(name string) *NodeName {
	nn := NodeName{}
	err := db.Get(&nn, "SELECT * FROM node_names WHERE value=?", name)
	if err != nil {
		log.Printf("Unable to find node_name %s: %s", name, err.Error())
		return nil
	}
	return &nn
}

// CreateNodeName creates a nnew node name record and returns it
func (db *DB) CreateNodeName(name string) (*NodeName, error) {
	tx, err := db.Begin()
	if err != nil {
		log.Printf("ERROR: create transaction for node_name create: %s", err.Error())
		return nil, err
	}
	qs := "insert into node_names (value) values (?)"
	res, err := tx.Exec(qs, name)
	if err != nil {
		log.Printf("ERROR: Unable to creat node_name: %s", err.Error())
		tx.Rollback()
		return nil, err
	}
	id, _ := res.LastInsertId()
	pid := fmt.Sprintf("uva-ann%d", id)
	qs = "update node_names set pid=? where id=?"
	res, err = tx.Exec(qs, pid, id)
	if err != nil {
		log.Printf("ERROR: unable to set node_name PID: %s", err.Error())
		tx.Rollback()
		return nil, err
	}
	err = tx.Commit()
	if err != nil {
		log.Printf("ERROR: node_name transaction commit failed %s", err.Error())
		tx.Rollback()
		return nil, err
	}
	return &NodeName{ID: id, PID: pid, Value: name}, nil
}
