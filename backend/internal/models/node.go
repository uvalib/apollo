package models

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
)

// NodeIdentifier holds the primary apollo IDs for an item; PID and ID
type NodeIdentifier struct {
	ID  int64  `db:"id" json:"-"`
	PID string `db:"pid" json:"pid"`
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
	Parent      *Node          `json:"-"`
	Sequence    int            `json:"sequence"`
	Type        *NodeType      `json:"type"`
	Value       string         `json:"value,omitempty"`
	ValueURI    string         `json:"valueURI,omitempty"`
	Children    []*Node        `json:"children,omitempty"`
	User        *User          `json:"-"`
	Deleted     bool           `json:"-"`
	Current     bool           `json:"-"`
	CreatedAt   time.Time      `db:"created_at" json:"createdAt"`
	UpdatedAt   *time.Time     `db:"updated_at" json:"updatedAt,omitempty"`
	PublishedAt *time.Time     `json:"publishedAt,omitempty"`
	Ancestry    sql.NullString `json:"-"`
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
		PID         string     `json:"pid"`
		Sequence    int        `json:"sequence"`
		Type        *NodeType  `json:"type"`
		Value       string     `json:"value,omitempty"`
		ValueURI    string     `json:"valueURI,omitempty"`
		CreatedAt   time.Time  `db:"created_at" json:"createdAt"`
		UpdatedAt   *time.Time `db:"updated_at" json:"updatedAt,omitempty"`
		PublishedAt *time.Time `json:"publishedAt,omitempty"`
		Children    []*Node    `json:"children,omitempty"`
	}{
		PID:         n.PID,
		Sequence:    n.Sequence,
		Type:        n.Type,
		Value:       n.encodeValue(n.Value),
		ValueURI:    n.ValueURI,
		CreatedAt:   n.CreatedAt,
		UpdatedAt:   n.UpdatedAt,
		PublishedAt: n.PublishedAt,
		Children:    n.Children,
	})
}

// GetAncestry returns all ancestors of the source node
func (db *DB) GetAncestry(node *Node) (*Node, error) {
	log.Printf("Get ancestors for %s with ancestry [%s]", node.PID, node.Ancestry.String)

	// Get the ancestry string. If there is none, there are no ancestors
	ancestry := node.Ancestry.String
	if ancestry == "" {
		return nil, nil
	}

	// split ancestry into a list of itemID
	var root, prior *Node
	ancestryArray := strings.Split(ancestry, "/")
	for _, stringID := range ancestryArray {
		// Pull details for the item and assemble into ancestry structure
		id, _ := strconv.ParseInt(stringID, 10, 64)
		ancestor, err := db.GetItem(id)
		if err != nil {
			return nil, err
		}
		if root == nil {
			root = ancestor
		} else {
			prior.Children = append(prior.Children, ancestor)
		}
		prior = ancestor
	}
	return root, nil
}

// GetParentCollection returns details about the collection that contains the source node
func (db *DB) GetParentCollection(node *Node) (*Node, error) {
	log.Printf("Get Parent for %s with ancestry [%s]", node.PID, node.Ancestry.String)

	// Get the ancestry string. If there is none, this node is the collection
	ancestry := node.Ancestry.String
	if ancestry == "" {
		return node, nil
	}

	// The collection node is the one with  ID matching the first ancestry substring
	rootID, _ := strconv.ParseInt(strings.Split(ancestry, "/")[0], 10, 64)
	log.Printf("Ancestry rootID: %d", rootID)

	// Dont want deleted or non-current nodes. Non-root nodes without values are the start of
	// child containers of the collection; skip them. Only take the parent node itself (id match)
	// or all nodes that have ONLY that node as their ancestor.
	qs := fmt.Sprintf(`%s WHERE deleted=0 and current=1 and (n.id=? or ancestry REGEXP '^%d$' and n.value <> '')`,
		getNodeSelect(), rootID)
	return db.queryNodes(qs, rootID)
}

// GetTree returns the node tree rooted at the specified node ID
func (db *DB) GetTree(rootID int64) (*Node, error) {
	// Get all children with the root ID as the start of their ancestry -- OR
	// any nodes that contain the ID as part of their ancestry (this is necessary for subtrees)
	log.Printf("Get tree rooted at ID %d", rootID)
	qs := fmt.Sprintf(`
		%s WHERE deleted=0 and current=1 AND (n.id=? or ancestry REGEXP '(^.+/|^)%d($|/.+)')
		ORDER BY n.id ASC`,
		getNodeSelect(), rootID)
	return db.queryNodes(qs, rootID)
}

// GetItem returns the items with the specified ID. An item is a node and all of its immediate children
func (db *DB) GetItem(nodeID int64) (*Node, error) {
	// now get all children with the above PID as the end of their ancestry
	qs := fmt.Sprintf(`
		%s WHERE deleted=0 and current=1 and (n.id=? or ancestry REGEXP '(^.+/|^)%d$' and n.value <> "")
		ORDER BY n.id ASC`,
		getNodeSelect(), nodeID)
	return db.queryNodes(qs, nodeID)
}

func getNodeSelect() string {
	return `SELECT n.id, n.parent_id, n.ancestry, n.sequence, n.pid, n.value, n.created_at, n.updated_at,
			  		nt.pid, nt.name, nt.controlled_vocab, nt.container
	 		  FROM nodes n
			  		INNER JOIN node_types nt ON nt.id = n.node_type_id`
}

func (db *DB) queryNodes(query string, rootID int64) (*Node, error) {
	// log.Printf("%s, %d", query, rootID)
	nodes := make(map[int64]*Node)
	nodeParents := make(map[int64]int64)
	var root *Node
	controlledValues := make(map[int64]*ControlledValue)
	rows, err := db.Query(query, rootID)
	if err != nil {
		log.Printf("ERROR: unable to retrieve nodes: %s", err.Error())
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var n Node
		var nt NodeType
		var updateAt mysql.NullTime
		var parentID sql.NullInt64

		rows.Scan(&n.ID, &parentID, &n.Ancestry, &n.Sequence, &n.PID, &n.Value, &n.CreatedAt, &updateAt, &nt.PID,
			&nt.Name, &nt.ControlledVocab, &nt.Container)

		if updateAt.Valid {
			n.UpdatedAt = &updateAt.Time
		}
		if parentID.Valid {
			nodeParents[n.ID] = parentID.Int64
		}

		n.Type = &nt
		if nt.ControlledVocab {
			id, _ := strconv.ParseInt(n.Value, 10, 64)
			if cv, ok := controlledValues[id]; ok {
				n.Value = cv.Value
				n.ValueURI = cv.ValueURI.String
			} else {
				cv := db.GetControlledValueByID(id)
				if cv == nil {
					log.Printf("ERROR: no controlled value match for %d", id)
				} else {
					n.Value = cv.Value
					n.ValueURI = cv.ValueURI.String
					controlledValues[id] = cv
				}
			}
		}

		// Save a map of ID -> Node. This will be used to assemble this list of rw nodes
		// into a heirarchy below
		nodes[n.ID] = &n
		if n.ID == rootID {
			root = &n
		}
	}

	// Build te tree: hook all nodes that have a parent with the parent
	for _, node := range nodes {
		if parentID, hasParent := nodeParents[node.ID]; hasParent {
			if parent, ok := nodes[parentID]; ok {
				parent.Children = append(parent.Children, node)
				node.Parent = parent
			} else {
				msg := fmt.Sprintf("Unable to to find parentID %d for node %d", parentID, node.ID)
				log.Printf("ERROR: %s", msg)
				return nil, errors.New(msg)
			}
		}
	}
	sortNodes(root)

	return root, nil
}

func sortNodes(node *Node) {
	if len(node.Children) > 0 {
		sort.Slice(node.Children, func(i, j int) bool {
			return node.Children[i].Sequence < node.Children[j].Sequence
		})
		for _, c := range node.Children {
			sortNodes(c)
		}
	}
}

// CreateNodes creates all nodes contained in the source list
func (db *DB) CreateNodes(nodes []*Node) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	for _, node := range nodes {
		addErr := addNode(tx, node)
		if addErr != nil {
			return addErr
		}

		// add parent and ancestry if needed
		if node.Parent != nil {
			ancestry := generateAncestryString(node)
			qs := "update nodes set parent_id=?, ancestry=? where id=?"
			_, insertErr := tx.Exec(qs, node.Parent.ID, ancestry, node.ID)
			if insertErr != nil {
				tx.Rollback()
				return err
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

func addNode(tx *sql.Tx, node *Node) error {
	t := time.Now().Unix()
	tmpPID := fmt.Sprintf("TMP-%d", t)
	qs := "insert into nodes (pid, node_type_id, sequence, value, user_id, created_at) values (?,?,?,?,?,NOW())"
	res, insertErr := tx.Exec(qs, tmpPID, node.Type.ID, node.Sequence, node.Value, node.User.ID)
	if insertErr != nil {
		tx.Rollback()
		return insertErr
	}

	// update the PID using last insert ID
	node.ID, _ = res.LastInsertId()
	node.PID = fmt.Sprintf("uva-an%d", node.ID)
	qs = "update nodes set pid=? where id=?"
	_, insertErr = tx.Exec(qs, node.PID, node.ID)
	if insertErr != nil {
		tx.Rollback()
		return insertErr
	}
	return nil
}

// AddNodes adds nodes to the specified parent starting with the specified sequence.
// If in insert mode, all nodes after have their sequence increased by 1
func (db *DB) AddNodes(mode string, nodes []*Node, parentID int64) error {
	var parentAncestry string
	db.QueryRow("select ancestry from nodes where id=?", parentID).Scan(&parentAncestry)
	rootAncestry := parentAncestry
	if parentAncestry != "" {
		rootAncestry = fmt.Sprintf("%s/%d", parentAncestry, parentID)
	}

	currSequence := nodes[0].Sequence
	if mode == "append" && currSequence == 0 {
		// in append mode, but no sequence speified, just pick a big sequence to start from
		currSequence = 25
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// NOTE: Inserting a bunch of individual attribute nodes does NOT work
	//
	if mode == "insert" {
		insertedTypeID := nodes[0].Type.ID
		sequenceStart := nodes[0].Sequence
		qs := fmt.Sprintf(`update nodes set sequence=sequence+1
			where node_type_id=%d and sequence >= %d and (id=%d or parent_id=%d or ancestry like "%%/%d"  or ancestry like "%%/%d/%%")`,
			insertedTypeID, sequenceStart, parentID, parentID, parentID, parentID)
		_, updateErr := tx.Exec(qs)
		if updateErr != nil {
			tx.Rollback()
			return updateErr
		}
	}

	for idx, node := range nodes {
		if mode == "append" && node.Sequence == 0 {
			node.Sequence = currSequence
			currSequence++
		}
		err = addNode(tx, node)
		if err != nil {
			return err
		}

		ancestry := rootAncestry
		nodeParentID := parentID
		if idx != 0 && node.Parent != nil {
			nodeParentID = node.Parent.ID
			ancestry = fmt.Sprintf("%s/%s", rootAncestry, generateAncestryString(node))
		}

		qs := "update nodes set parent_id=?, ancestry=? where id=?"
		_, insertErr := tx.Exec(qs, nodeParentID, ancestry, node.ID)
		if insertErr != nil {
			tx.Rollback()
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

// generateAncestryString will walk back up the parent chain to build a backwards
// list of ancestor IDs for this node
func generateAncestryString(node *Node) string {
	var parentIDs []string
	curr := node
	for {
		parent := curr.Parent
		if parent == nil {
			break
		} else {
			parentIDs = append(parentIDs, fmt.Sprintf("%d", parent.ID))
			curr = parent
		}
	}
	// reverse to get proper ancestry ordering
	for i, j := 0, len(parentIDs)-1; i < j; i, j = i+1, j-1 {
		parentIDs[i], parentIDs[j] = parentIDs[j], parentIDs[i]
	}
	return strings.Join(parentIDs, "/")
}
