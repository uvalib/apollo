package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
)

// nodeSelect is the bas query used to get a variety of node dat from the DB
const nodeSelect = `SELECT n.id, n.parent_id, n.ancestry, n.sequence, n.pid, n.value, n.created_at, n.updated_at,
 nt.pid, nt.name, nt.controlled_vocab, nt.container
 FROM nodes n
 INNER JOIN node_types nt ON nt.id = n.node_type_id`

// GetItemDetails will return a block of JSON metadata for the specified ITEM PID. This includes
// details of the specific item as well as some basic data amout the colection it
// belongs to.
func (app *Apollo) GetItemDetails(c *gin.Context) {
	pid := c.Param("pid")
	itemIDs, dbErr := lookupIdentifier(&app.DB, pid)
	if dbErr != nil {
		log.Printf("ERROR: %s", dbErr.Error())
		c.String(http.StatusNotFound, dbErr.Error())
		return
	}

	item, dbErr := getNode(&app.DB, itemIDs.ID)
	if dbErr != nil {
		log.Printf("ERROR: %s", dbErr.Error())
		c.String(http.StatusNotFound, dbErr.Error())
		return
	}

	// note: if above was successful, this will be as well
	parent, _ := getNodeCollection(&app.DB, item)

	jsonItem, _ := json.MarshalIndent(item, "", "  ")
	jsonParent, _ := json.MarshalIndent(parent, "", "  ")
	out := fmt.Sprintf("{\n\"collection\": %s,\n\"item\": %s}", jsonParent, jsonItem)
	c.Header("Content-Type", "application/json")
	c.String(http.StatusOK, out)
}

// getNode returns the node specified by nodeID and all of its immediate children
func getNode(db *DB, nodeID int64) (*Node, error) {
	// Get all children with the above nodeID PID as the end of their ancestry
	qs := fmt.Sprintf(`
		%s WHERE deleted=0 and current=1 and (n.id=? or ancestry REGEXP '(^.+/|^)%d$' and n.value <> "")
		ORDER BY n.id ASC`,
		nodeSelect, nodeID)
	return queryNodes(db, qs, nodeID)
}

// GetTree returns the node tree rooted at the specified node ID
func getTree(db *DB, rootID int64) (*Node, error) {
	// Get all children with the root ID as the start of their ancestry -- OR
	// any nodes that contain the ID as part of their ancestry (this is necessary for subtrees)
	log.Printf("Get tree rooted at ID %d", rootID)
	qs := fmt.Sprintf(`
		%s WHERE deleted=0 and current=1 AND (n.id=? or ancestry REGEXP '(^.+/|^)%d($|/.+)')
		ORDER BY n.id ASC`,
		nodeSelect, rootID)
	return queryNodes(db, qs, rootID)
}

// getNodeCollection returns details about the collection that contains the source node
func getNodeCollection(db *DB, node *Node) (*Node, error) {
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
		nodeSelect, rootID)
	return queryNodes(db, qs, rootID)
}

func queryNodes(db *DB, query string, rootID int64) (*Node, error) {
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
				cv, err := getControlledValueByID(db, id)
				if err != nil {
					log.Printf("ERROR: no controlled value match for %d: %s", id, err.Error())
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
		// In the case when we are requesting a sub-tree, the parent of the
		// start of the tree will not exist. Don't try to find it!
		if node.ID == rootID {
			continue
		}
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
