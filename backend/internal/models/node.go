package models

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
)

// Collection holds key data about a collection; its PID and Title
type Collection struct {
	PID   string `json:"pid"`
	Title string `json:"title"`
}

// Node is a single element in a tree of metadata. This is the smallest unit
// of data in the system; essentially a name/value pair.
// An Item is a collection of nodes with
// the same parent.
// A Collection is the hierarchical representation of all nodes stemming from a
// single PID.
type Node struct {
	ID        int64      `json:"-"`
	PID       string     `json:"pid"`
	Parent    *Node      `json:"-"`
	Sequence  int        `json:"sequence"`
	Type      *NodeType  `json:"type"`
	Value     string     `json:"value,omitempty"`
	ValueURI  string     `json:"valueURI,omitempty"`
	Children  []*Node    `json:"children,omitempty"`
	User      *User      `json:"-"`
	Deleted   bool       `json:"-"`
	Current   bool       `json:"-"`
	CreatedAt time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt *time.Time `db:"updated_at" json:"updatedAt,omitempty"`
	parentID  sql.NullInt64
	ancestry  sql.NullString
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

// ExternalPIDLookup will find an ApolloPID for an external PID
func (db *DB) ExternalPIDLookup(externalPID string) (string, error) {
	var apolloPID string
	qs := `SELECT np.pid FROM nodes ns
		INNER JOIN nodes np ON np.id = ns.parent_id
 		WHERE ns.value=?`
	db.QueryRow(qs, externalPID).Scan(&apolloPID)
	if len(apolloPID) == 0 {
		return "", fmt.Errorf("Unable to find match for legacy PID %s", externalPID)
	}
	return apolloPID, nil
}

// GetNodeIDFromPID takes a PID and returns the corresponding node ID
func (db *DB) GetNodeIDFromPID(pid string) (int64, error) {
	log.Printf("Get node ID for PID %s", pid)
	var nodeID int64
	db.QueryRow("select id from nodes where pid=?", pid).Scan(&nodeID)
	if nodeID > 0 {
		return nodeID, nil
	}
	return 0, fmt.Errorf("PID %s was not found", pid)
}

// GetCollections returns a list of all collections. Data is PID/Title
func (db *DB) GetCollections() []Collection {
	var pids []struct {
		ID  int64
		PID string
	}

	var out []Collection
	qs := "select id,pid from nodes where parent_id is null"
	tq := "select value from nodes where ancestry=? and node_type_id=? order by id asc limit 1"
	db.Select(&pids, qs)

	for _, val := range pids {
		var title string
		db.QueryRow(tq, val.ID, 2).Scan(&title)
		out = append(out, Collection{val.PID, title})
	}
	return out
}

// GetAncestry returns all ancestors of the source node
func (db *DB) GetAncestry(node *Node) (*Node, error) {
	log.Printf("Get ancestors for %s with ancestry [%s]", node.PID, node.ancestry.String)

	// Get the ancestry string. If there is none, there are no ancestors
	ancestry := node.ancestry.String
	if ancestry == "" {
		return nil, nil
	}

	// Pull rootID off of ancestry string
	var root, prior *Node
	ancestryArray := strings.Split(ancestry, "/")
	for _, stringID := range ancestryArray {
		id, _ := strconv.ParseInt(stringID, 10, 64)
		ancestor, err := db.GetChildren(id)
		if err != nil {
			return nil, err
		}
		log.Printf("Ancestor: %+v", ancestor)
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
	log.Printf("Get Parent for %s with ancestry [%s]", node.PID, node.ancestry.String)

	// Get the ancestry string. If there is none, this node is the collection
	ancestry := node.ancestry.String
	if ancestry == "" {
		return node, nil
	}

	// The collection node is the one with  ID matching the first ancestry substring
	rootID, _ := strconv.ParseInt(strings.Split(ancestry, "/")[0], 10, 64)

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
		%s WHERE deleted=0 and current=1 AND (n.id=? or ancestry REGEXP '(^.*/|^)%d($|/.*)')
		ORDER BY n.id ASC`,
		getNodeSelect(), rootID)
	return db.queryNodes(qs, rootID)
}

// GetChildren returns this node and all of its immediate children
func (db *DB) GetChildren(nodeID int64) (*Node, error) {
	// now get all children with the above PID as the end of their ancestry
	qs := fmt.Sprintf(`
		%s WHERE deleted=0 and current=1 and (n.id=? or ancestry REGEXP '(^.*/|^)%d$' and n.value <> "")
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
	nodes := make(map[int64]*Node)
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

		rows.Scan(&n.ID, &n.parentID, &n.ancestry, &n.Sequence, &n.PID, &n.Value, &n.CreatedAt, &updateAt, &nt.PID,
			&nt.Name, &nt.ControlledVocab, &nt.Container)

		if updateAt.Valid {
			n.UpdatedAt = &updateAt.Time
		}

		n.Type = &nt
		if nt.ControlledVocab {
			id, _ := strconv.ParseInt(n.Value, 10, 64)
			if cv, ok := controlledValues[id]; ok {
				n.Value = cv.Value
				n.ValueURI = cv.ValueURI
				log.Printf("Use cached controlled value %d", id)
			} else {
				cv := db.GetControlledValueByID(id)
				if cv == nil {
					log.Printf("ERROR: no controlled value match for %d", id)
				} else {
					n.Value = cv.Value
					n.ValueURI = cv.ValueURI
					controlledValues[id] = cv
					log.Printf("Cache controlled value %d", id)
				}
			}
		}

		nodes[n.ID] = &n
		if n.ID == rootID {
			log.Printf("Set root node to %d", n.ID)
			root = &n
		}
	}

	// hook all nodes that have a parent with the parent
	for _, node := range nodes {
		// In the case when we are requesting a sub-tree, the parent of the
		// start of the tree will not exist. Don't try to find it!
		if node.ID == rootID {
			continue
		}
		if node.parentID.Valid == false {
			continue
		}
		if parent, ok := nodes[node.parentID.Int64]; ok {
			parent.Children = append(parent.Children, node)
		} else {
			msg := fmt.Sprintf("Unable to to find parentID %d for node %d", node.parentID.Int64, node.ID)
			log.Printf("ERROR: %s", msg)
			return nil, errors.New(msg)
		}
	}

	return root, nil
}

// CreateNodes creates all nodes contained in the source list
func (db *DB) CreateNodes(nodes []*Node) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	for _, node := range nodes {
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
		res, insertErr = tx.Exec(qs, node.PID, node.ID)
		if insertErr != nil {
			tx.Rollback()
			return err
		}

		// add parent and ancestry if needed
		if node.Parent != nil {
			ancestry := generateAncestryString(node)
			qs = "update nodes set parent_id=?, ancestry=? where id=?"
			res, insertErr = tx.Exec(qs, node.Parent.ID, ancestry, node.ID)
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

// TODO Implement these methods
// // UpdateNode : update node value as specified. This creates a version history.
// //
// func (db *DB) UpdateNode(updatedNode *Node, user *User) {
// 	// find all prior versions: select * from nodes where pid like 'PID.%' order created_at desc;
// 	// create a new node with existing node data and set pid to pid.N where N is 1 more that last from above
// 	// update existing node with data in updateNode
// 	//
// 	// _, err := db.Exec("update nodes set value=? where id=?", value, node.ID)
// 	// if err != nil {
// 	// 	log.Printf("ERROR: node value update failed %s", err.Error())
// 	// }
// }

// // GetNode finds a SINGLE node by PID. No parent nor children is returned
// // Details about user and revision history are included
// func (db *DB) GetNode(pid string) *Node {
// 	qs := `SELECT n.id, n.pid, n.value, n.deleted, n.current, n.created_at,
//             nt.id, nt.pid, nt.value,
//             u.id, u.computing_id, u.last_name, u.first_name, u.email
//           FROM nodes n
//             inner join node_types nt on nt.id = n.node_type_id
//             inner join users u on u.id = n.user_id
//           WHERE n.pid=?`
// 	row := db.QueryRow(qs, pid)
// 	var n Node
// 	var nt NodeType
// 	var u User
// 	row.Scan(&n.ID, &n.PID, &n.Value, &n.Deleted, &n.Current, &n.CreatedAt,
// 		&nt.ID, &nt.PID, &nt.Name,
// 		&u.ID, &u.ComputingID, &u.LastName, &u.FirstName, &u.Email)
// 	n.Type = &nt
// 	n.User = &u
// 	return &n
// }
