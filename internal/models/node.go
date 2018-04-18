package models

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
)

// Node is an element in a tree of metadata
type Node struct {
	ID        int64     `json:"-"`
	PID       string    `json:"pid"`
	Parent    *Node     `json:"-"`
	Name      *NodeName `json:"name"`
	Value     string    `json:"value,omitempty"`
	Children  []*Node   `json:"children,omitempty"`
	User      *User     `json:"-"`
	Deleted   bool      `json:"-"`
	Current   bool      `json:"-"`
	CreatedAt time.Time `db:"created_at" json:"-"`
	UpdatedAt time.Time `db:"updated_at" json:"-"`
}

// GetCollectionPIDs finds a node by PID
func (db *DB) GetCollectionPIDs() []string {
	pids := []string{}
	qs := "select pid from nodes where parent_id is null"
	db.Select(&pids, qs)
	return pids
}

// GetCollection returns an entire collection identified by PID
func (db *DB) GetCollection(pid string) (*Node, error) {
	// first, get the root node
	var nodes []*Node
	root := db.GetNode(pid)
	nodes = append(nodes, root)

	// now get all children with the root ID as the start of their ancestry
	qs := fmt.Sprintf(
		`SELECT n.id, n.parent_id, n.pid, n.value, n.deleted, n.current, n.created_at,
         nn.id, nn.pid, nn.value,
         u.id, u.computing_id, u.last_name, u.first_name, u.email
       FROM nodes n
         inner join node_names nn on nn.id = n.node_name_id
         inner join users u on u.id = n.user_id
       WHERE deleted=0 and current=1 and (ancestry like '%d/%%' or ancestry=?)`, root.ID)
	rows, err := db.Queryx(qs, root.ID)
	if err != nil {
		log.Printf("ERROR: unable to retrieve collection: %s", err.Error())
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var n Node
		var nn NodeName
		var u User
		var parentID sql.NullInt64
		rows.Scan(&n.ID, &parentID, &n.PID, &n.Value, &n.Deleted, &n.Current, &n.CreatedAt,
			&nn.ID, &nn.PID, &nn.Value,
			&u.ID, &u.ComputingID, &u.LastName, &u.FirstName, &u.Email)
		n.Name = &nn
		n.User = &u
		nodes = append(nodes, &n)
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
		qs := "insert into nodes (node_name_id, value, user_id, created_at) values (?,?,?,NOW())"
		res, insertErr := tx.Exec(qs, node.Name.ID, node.Value, node.User.ID)
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
