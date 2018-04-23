package models

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
)

// Node is an element in a tree of metadata
type Node struct {
	ID        int64      `json:"-"`
	PID       string     `json:"pid"`
	Parent    *Node      `json:"-"`
	Name      *NodeName  `json:"name"`
	Value     string     `json:"value,omitempty"`
	Children  []*Node    `json:"children,omitempty"`
	User      *User      `json:"-"`
	Deleted   bool       `json:"-"`
	Current   bool       `json:"-"`
	CreatedAt time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt *time.Time `db:"updated_at" json:"updatedAt,omitempty"`
}

// Collection holds key data about a collection; its PID and Title
type Collection struct {
	PID   string
	Title string
}

// MarshalJSON will encode the Node structure as JSON
func (n *Node) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		PID       string     `json:"pid"`
		Name      *NodeName  `json:"name"`
		Value     string     `json:"value,omitempty"`
		CreatedAt time.Time  `db:"created_at" json:"createdAt"`
		UpdatedAt *time.Time `db:"updated_at" json:"updatedAt,omitempty"`
		Children  []*Node    `json:"children,omitempty"`
	}{
		PID:       n.PID,
		Name:      n.Name,
		Value:     encodeValue(n.Value),
		CreatedAt: n.CreatedAt,
		UpdatedAt: n.UpdatedAt,
		Children:  n.Children,
	})
}

func encodeValue(val string) string {
	if strings.Contains(val, "http:") || strings.Contains(val, "https:") {
		out, _ := url.QueryUnescape(val)
		return out
	}
	return val
}

// GetCollections returns a list of all collections. Data is PID/Title
func (db *DB) GetCollections() []Collection {
	var pids []struct {
		ID  int64
		PID string
	}

	var out []Collection
	qs := "select id,pid from nodes where parent_id is null"
	tq := "select value from nodes where ancestry=? and node_name_id=? order by id asc limit 1"
	db.Select(&pids, qs)

	for _, val := range pids {
		var title string
		db.QueryRow(tq, val.ID, 2).Scan(&title)
		out = append(out, Collection{val.PID, title})
	}
	return out
}

// GetCollection returns an entire collection identified by PID
func (db *DB) GetCollection(pid string) (*Node, error) {
	log.Printf("Get collecton with PID %s", pid)

	// first, get the root node
	var rootID int64
	db.QueryRow("select id from nodes where pid=?", pid).Scan(&rootID)
	var nodes []*Node
	var root *Node
	log.Printf("Collecton for PID %s has ID %d", pid, rootID)

	// now get all children with the root ID as the start of their ancestry
	qs := fmt.Sprintf(
		`SELECT n.id, n.parent_id, n.pid, n.value, n.created_at, n.updated_at, nn.pid, nn.value
       FROM nodes n
         inner join node_names nn on nn.id = n.node_name_id
       WHERE deleted=0 and current=1 and (n.id=? or ancestry like '%d/%%' or ancestry=?) order by n.id asc`, rootID)
	rows, err := db.Queryx(qs, rootID, rootID)
	if err != nil {
		log.Printf("ERROR: unable to retrieve collection: %s", err.Error())
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var n Node
		var nn NodeName
		var parentID sql.NullInt64
		var updateAt mysql.NullTime
		rows.Scan(&n.ID, &parentID, &n.PID, &n.Value, &n.CreatedAt, &updateAt, &nn.PID, &nn.Value)
		if updateAt.Valid {
			n.UpdatedAt = &updateAt.Time
		}
		n.Name = &nn
		nodes = append(nodes, &n)
		if n.ID == rootID {
			root = &n
		}
		if parentID.Valid {
			// a parentID exists. That node should be found in the nodes array
			found := false
			for _, val := range nodes {
				if val.ID == parentID.Int64 {
					n.Parent = val
					val.Children = append(val.Children, &n)
					found = true
					break
				}
			}
			if !found {
				msg := fmt.Sprintf("Unable to to find parentID %d for collection: %d", parentID.Int64, root.ID)
				log.Printf("ERROR: %s", msg)
				return nil, errors.New(msg)
			}
		}
	}
	return root, nil
}

// GetNode finds a SINGLE node by PID. No parent nor children is returned
// Details about user and revision history are included (TODO)
func (db *DB) GetNode(pid string) *Node {
	qs := `SELECT n.id, n.pid, n.value, n.deleted, n.current, n.created_at,
            nn.id, nn.pid, nn.value,
            u.id, u.computing_id, u.last_name, u.first_name, u.email
          FROM nodes n
            inner join node_names nn on nn.id = n.node_name_id
            inner join users u on u.id = n.user_id
          WHERE n.pid=?`
	row := db.QueryRow(qs, pid)
	var n Node
	var nn NodeName
	var u User
	row.Scan(&n.ID, &n.PID, &n.Value, &n.Deleted, &n.Current, &n.CreatedAt,
		&nn.ID, &nn.PID, &nn.Value,
		&u.ID, &u.ComputingID, &u.LastName, &u.FirstName, &u.Email)
	n.Name = &nn
	n.User = &u
	return &n
}

// CreateNodes creates a list of nodes
//
func (db *DB) CreateNodes(nodes []*Node) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	for _, node := range nodes {
		t := time.Now().Unix()
		tmpPID := fmt.Sprintf("TMP-%d", t)
		qs := "insert into nodes (pid, node_name_id, value, user_id, created_at) values (?,?,?,?,NOW())"
		res, insertErr := tx.Exec(qs, tmpPID, node.Name.ID, node.Value, node.User.ID)
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
			ancestry := getAncestry(node)
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

func getAncestry(node *Node) string {
	// walk back up the parent chain to build a backwards
	// list of ancestor IDs for this node
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

// UpdateNode : update node value as specified. This creates a version history.
//
func (db *DB) UpdateNode(updatedNode *Node, user *User) {
	// find all prior versions: select * from nodes where pid like 'PID.%' order created_at desc;
	// create a new node with existing node data and set pid to pid.N where N is 1 more that last from above
	// update existing node with data in updateNode
	//
	// _, err := db.Exec("update nodes set value=? where id=?", value, node.ID)
	// if err != nil {
	// 	log.Printf("ERROR: node value update failed %s", err.Error())
	// }
}
