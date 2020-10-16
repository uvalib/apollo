package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// ListCollections returns a json array containg all collection in tghe system
func (app *Apollo) ListCollections(c *gin.Context) {
	log.Printf("Get all collections")
	collections := getCollections(&app.DB)
	c.JSON(http.StatusOK, collections)
}

// GetCollection finds a collection by PID and returns details as json
func (app *Apollo) GetCollection(c *gin.Context) {
	pid := c.Param("pid")
	tgtFormat := c.Query("format")
	if tgtFormat == "" {
		tgtFormat = "json"
	}
	if tgtFormat != "json" && tgtFormat != "xml" {
		log.Printf("ERROR: Unsupported format for %s requested %s", tgtFormat, pid)
		c.String(http.StatusBadRequest, fmt.Sprintf("unsupported format %s", tgtFormat))
		return
	}
	log.Printf("Get collection for PID %s as %s", pid, tgtFormat)
	rootID, dbErr := lookupIdentifier(&app.DB, pid)
	if dbErr != nil {
		log.Printf("ERROR: %s", dbErr.Error())
		c.String(http.StatusNotFound, dbErr.Error())
		return
	}

	root, dbErr := getTree(&app.DB, rootID.ID)
	if dbErr != nil {
		log.Printf("ERROR: %s", dbErr.Error())
		c.String(http.StatusInternalServerError, dbErr.Error())
		return
	}
	log.Printf("Collection tree retrieved from DB; sending to client")
	if tgtFormat == "json" {
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.json", pid))
		c.JSON(http.StatusOK, root)
	} else {
		xml, err := generateXML(root)
		if err != nil {
			log.Printf("ERROR: unable to generate XML for %s: %s", pid, err.Error())
			c.String(http.StatusInternalServerError, "unable to generate XML content")
			return
		}
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.xml", pid))
		c.Header("Content-Type", "application/xml")
		c.String(http.StatusOK, xml)
	}
}

// getCollections returns a list of all collections. Data is ID/PID/Title
func getCollections(db *DB) []Collection {
	var IDs []NodeIdentifier
	var out []Collection
	qs := "select id,pid from nodes where parent_id is null"
	db.Select(&IDs, qs)

	tq := "select value from nodes where ancestry=? and node_type_id=? order by id asc limit 1"
	for _, val := range IDs {
		var title string
		db.QueryRow(tq, val.ID, 2).Scan(&title)
		out = append(out, Collection{ID: val.ID, PID: val.PID, Title: title})
	}
	return out
}

func generateXML(node *Node) (string, error) {
	log.Printf("Generate XML for collection %s", node.PID)
	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)
	traverseTree(writer, node)
	writer.Flush()
	return buf.String(), nil
}

func traverseTree(out *bufio.Writer, node *Node) {
	if node.Type.Container {
		// log.Printf("<%s>", node.Type.Name)
		out.WriteString(fmt.Sprintf("<%s>", node.Type.Name))
		for _, child := range node.Children {
			if child.Type.Name != "dpla" && child.Type.Name != "digitalObject" {
				if child.Type.Container == false {
					// log.Printf("<%s>%s</%s>", child.Type.Name, child.Value, child.Type.Name)
					out.WriteString(fmt.Sprintf("<%s>%s</%s>\n", child.Type.Name, cleanValue(child.Value), child.Type.Name))
				} else {
					traverseTree(out, child)
				}
			}
		}
		// log.Printf("</%s>", node.Type.Name)
		out.WriteString(fmt.Sprintf("</%s>\n", node.Type.Name))
	} else {
		// log.Printf("<%s>%s</%s>", node.Type.Name, node.Value, node.Type.Name)
		out.WriteString(fmt.Sprintf("<%s>%s</%s>\n", node.Type.Name, cleanValue(node.Value), node.Type.Name))
	}
}

func cleanValue(val string) string {
	clean := strings.TrimSpace(val)
	clean = strings.ReplaceAll(clean, "&", "&amp;")
	return clean
}
