package main

import (
	"database/sql"
	"encoding/json"
	"net/url"
	"strings"
	"time"
)

// Collection holds key data about a collection; its PID and Title
type Collection struct {
	ID    int64  `json:"id"`
	PID   string `json:"pid"`
	Title string `json:"title"`
}

// NodeType is a controlled vocabulary for node names
type NodeType struct {
	ID              int64  `json:"-"`
	PID             string `json:"pid"`
	Name            string `json:"name"`
	ControlledVocab bool   `db:"controlled_vocab" json:"controlledVocab"`
	Validation      string `json:"validation,omitempty"`
	Container       bool   `json:"container"`
}

// NodeIdentifier holds the primary apollo IDs for an item; PID and ID
type NodeIdentifier struct {
	ID  int64  `db:"id" json:"id"`
	PID string `db:"pid" json:"pid"`
}

// ControlledValue is a controlled vocabulary for node values
type ControlledValue struct {
	ID       int64          `json:"-"`
	PID      string         `json:"pid"`
	TypeID   int64          `db:"node_type_id" json:"-"`
	Value    string         `json:"value"`
	ValueURI sql.NullString `db:"value_uri" json:"valueURI"`
}

// Node is a single element in a tree of metadata. This is the smallest unit
// of data in the system; essentially a name/value pair.
//
// An Item is a collection of nodes with
// the same parent.
//
// A Collection is the hierarchical representation of all nodes stemming from a
// single PID.
type Node struct {
	NodeIdentifier
	Parent    *Node          `json:"-"`
	Sequence  int            `json:"sequence"`
	Type      *NodeType      `json:"type"`
	Value     string         `json:"value,omitempty"`
	ValueURI  string         `json:"valueURI,omitempty"`
	Children  []*Node        `json:"children,omitempty"`
	Deleted   bool           `json:"-"`
	Current   bool           `json:"-"`
	CreatedAt time.Time      `db:"created_at" json:"createdAt"`
	UpdatedAt *time.Time     `db:"updated_at" json:"updatedAt,omitempty"`
	Ancestry  sql.NullString `json:"-"`
}

func (n *Node) encodeValue(val string) string {
	if strings.Contains(val, "http:") || strings.Contains(val, "https:") {
		out, _ := url.QueryUnescape(val)
		return out
	}
	return val
}

// MarshalJSON will encode the Node structure as JSON
func (n *Node) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		PID       string     `json:"pid"`
		Sequence  int        `json:"sequence"`
		Type      *NodeType  `json:"type"`
		Value     string     `json:"value,omitempty"`
		ValueURI  string     `json:"valueURI,omitempty"`
		CreatedAt time.Time  `db:"created_at" json:"createdAt"`
		UpdatedAt *time.Time `db:"updated_at" json:"updatedAt,omitempty"`
		Children  []*Node    `json:"children,omitempty"`
	}{
		PID:       n.PID,
		Sequence:  n.Sequence,
		Type:      n.Type,
		Value:     n.encodeValue(n.Value),
		ValueURI:  n.ValueURI,
		CreatedAt: n.CreatedAt,
		UpdatedAt: n.UpdatedAt,
		Children:  n.Children,
	})
}
