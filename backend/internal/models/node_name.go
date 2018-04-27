package models

import (
	"fmt"
	"log"
	"time"
)

// NodeName is a controlled vocabulary for node names
type NodeName struct {
	ID    int64  `json:"-"`
	PID   string `json:"pid"`
	Value string `json:"value"`
}

// AllNames returns a list of all available names
func (db *DB) AllNames() []NodeName {
	names := []NodeName{}
	db.Select(&names, "select * from node_names order by value asc")
	return names
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
	t := time.Now().Unix()
	tmpPID := fmt.Sprintf("TMP-%d", t)
	qs := "insert into node_names (pid,value) values (?,?)"
	res, err := tx.Exec(qs, tmpPID, name)
	if err != nil {
		log.Printf("ERROR: Unable to create node_name: %s", err.Error())
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
