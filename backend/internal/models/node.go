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
	Name      *NodeName  `json:"name"`
	Value     string     `json:"value,omitempty"`
	ValueURI  string     `json:"valueURI,omitempty"`
	Children  []*Node    `json:"children,omitempty"`
	User      *User      `json:"-"`
	Deleted   bool       `json:"-"`
	Current   bool       `json:"-"`
	CreatedAt time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt *time.Time `db:"updated_at" json:"updatedAt,omitempty"`
}

// Collection holds key data about a collection; its PID and Title
type Collection struct {
	PID   string `json:"pid"`
	Title string `json:"title"`
}

// MarshalJSON will encode the Node structure as JSON
func (n *Node) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		PID       string     `json:"pid"`
		Name      *NodeName  `json:"name"`
		Value     string     `json:"value,omitempty"`
		ValueURI  string     `json:"valueURI,omitempty"`
		CreatedAt time.Time  `db:"created_at" json:"createdAt"`
		UpdatedAt *time.Time `db:"updated_at" json:"updatedAt,omitempty"`
		Children  []*Node    `json:"children,omitempty"`
	}{
		PID:       n.PID,
		Name:      n.Name,
		Value:     encodeValue(n.Value),
		ValueURI:  n.ValueURI,
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

// LegacyLookup will find an ApolloPID for a legacy TrackSys PID
func (db *DB) LegacyLookup(componentPID string) (string, error) {
	var apolloPID string
	qs := `SELECT np.pid FROM nodes ns
		INNER JOIN nodes np ON np.id = ns.parent_id
 		WHERE ns.value=?`
	log.Printf("Q: %s, PID: %s", qs, componentPID)
	db.QueryRow(qs, componentPID).Scan(&apolloPID)
	if len(apolloPID) == 0 {
		return "", fmt.Errorf("Unable to find match for legacy PID %s", componentPID)
	}
	return apolloPID, nil
}

// GetParentCollection returns details about the collection that contains the PID
func (db *DB) GetParentCollection(pid string) (*Node, error) {
	var ancestry string
	log.Printf("Get Parent Collection for PID %s", pid)
	db.QueryRow("select ancestry from nodes where pid=?", pid).Scan(&ancestry)
	rootID, _ := strconv.ParseInt(strings.Split(ancestry, "/")[0], 10, 64)

	qs := fmt.Sprintf(`
		%s WHERE deleted=0 and current=1 and (n.id=? or ancestry REGEXP '(^.*/|^)%d$')
	 	ORDER BY n.id ASC`, getNodeSelect(), rootID)
	return db.queryNodes(qs, rootID, true)
}

// GetTree returns the node tree rooted at the specified PID
func (db *DB) GetTree(pid string) (*Node, error) {
	log.Printf("Get tree rooted at PID %s", pid)
	var rootID int64
	db.QueryRow("select id from nodes where pid=?", pid).Scan(&rootID)

	// now get all children with the root ID as the start of their ancestry
	qs := fmt.Sprintf(`
		%s WHERE deleted=0 and current=1 AND (n.id=? or ancestry REGEXP '(^.*/|^)%d($|/.*)')
		ORDER BY n.id ASC`, getNodeSelect(), rootID)
	return db.queryNodes(qs, rootID, false)
}

// GetChildren returns this node and all of its immediate children
func (db *DB) GetChildren(pid string) (*Node, error) {
	var itemID int64
	log.Printf("Get Children of PID %s", pid)
	db.QueryRow("select id from nodes where pid=?", pid).Scan(&itemID)

	// now get all children with the above PID as the end of their ancestry
	qs := fmt.Sprintf(`
		%s WHERE deleted=0 and current=1 and (n.id=? or ancestry REGEXP '(^.*/|^)%d$')
	 	ORDER BY n.id ASC`, getNodeSelect(), itemID)
	return db.queryNodes(qs, itemID, true)
}

func getNodeSelect() string {
	return `SELECT n.id, n.parent_id, n.pid, n.value, n.created_at, n.updated_at,
			  		nn.pid, nn.value, nn.controlled_vocab
	 		  FROM nodes n
			  		INNER JOIN node_names nn ON nn.id = n.node_name_id`
}

func (db *DB) queryNodes(query string, rootID int64, stripNoValue bool) (*Node, error) {
	var nodes []*Node
	var root *Node
	rows, err := db.Queryx(query, rootID)
	if err != nil {
		log.Printf("ERROR: unable to retrieve nodes: %s", err.Error())
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var n Node
		var nn NodeName
		var parentID sql.NullInt64
		var updateAt mysql.NullTime
		rows.Scan(&n.ID, &parentID, &n.PID, &n.Value, &n.CreatedAt, &updateAt, &nn.PID, &nn.Value, &nn.ControlledVocab)
		if updateAt.Valid {
			n.UpdatedAt = &updateAt.Time
		}
		n.Name = &nn
		if nn.ControlledVocab {
			id, _ := strconv.ParseInt(n.Value, 10, 64)
			cv := db.GetControlledValueByID(id)
			if cv == nil {
				log.Printf("ERROR: no controlled value match for %d", id)
			} else {
				n.Value = cv.Value
				n.ValueURI = cv.ValueURI
			}
		}

		if n.ID == rootID {
			root = &n
			nodes = append(nodes, &n)
		} else if parentID.Valid {
			if stripNoValue && len(n.Value) == 0 {
				// skip no-value nodes mode and one encountered. Skip it!
				continue
			}
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
