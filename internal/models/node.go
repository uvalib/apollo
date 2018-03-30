package models

import (
	"fmt"
	"time"
)

// Node is an element in a tree of metadata
type Node struct {
	ID           int64
	PID          string
	Parent       *Node
	Name         *NodeName
	Value        string
	User         *User
	Deleted      bool
	Current      bool
	PriorVersion *Node
	CreatedAt    time.Time
}

// GetNode finds a node by PID
func (db *DB) GetNode(pid string) *Node {
	qs := `SELECT n.id, n.pid, n.value, n.deleted, n.current, n.created_at,
            nn.id, nn.pid, nn.value,
            u.id, u.computing_id, u.last_name, u.first_name, u.email
          FROM nodes n
            inner join node_names nn on nn.id = n.node_name_id
            inner join users u on u.id = n.user_id
          WHERE n.pid=?`
	row := db.QueryRow(qs, pid)
	n := Node{}
	nn := NodeName{}
	u := User{}
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
		var parentPID string
		if node.Parent != nil {
			parentPID = node.Parent.PID
		}
		qs := "insert into nodes (parent_pid,node_name_id,value,user_id,created_at) values (?,?,?,?,NOW())"
		res, insertErr := tx.Exec(qs, parentPID, node.Name.ID, node.Value, node.User.ID)
		if insertErr != nil {
			tx.Rollback()
			return insertErr
		}

		id, _ := res.LastInsertId()
		pid := fmt.Sprintf("uva-an%d", id)
		node.PID = pid
		qs = "update nodes set pid=? where id=?"
		res, insertErr = tx.Exec(qs, pid, id)
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

// UpdateNode : update node value as specified. This creates a version history.
//
func (db *DB) UpdateNode(node *Node, user *User) {
	// _, err := db.Exec("update nodes set value=? where id=?", value, node.ID)
	// if err != nil {
	// 	log.Printf("ERROR: node value update failed %s", err.Error())
	// }
}
