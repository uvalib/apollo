package main

import (
	"encoding/xml"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/uvalib/apollo/backend/internal/models"
)

// Ingest the XML UPDATE file contained in the config data
func doUpdate(ctx *context, srcFile string) {
	log.Printf("Start ingest of %s...", srcFile)
	xmlFile, err := os.Open(srcFile)
	if err != nil {
		log.Printf("ERROR: Unable to read source file %s: %s", srcFile, err.Error())
		return
	}
	defer xmlFile.Close()

	decoder := xml.NewDecoder(xmlFile)
	nodeStack := []*models.Node{}
	nodes := []*models.Node{}
	var valueNode string
	var parentID int64
	var sequence int
	var updateMode string
	var wslsCache map[string]int64
	specialNodes := []string{"update", "insert", "append", "parent", "sequence", "wslsParent"}
	for {
		token, terr := decoder.Token()
		if terr == io.EOF {
			break
		} else if terr != nil {
			log.Printf("ERROR: unable to parse file: %s", terr.Error())
			return
		}

		switch tok := token.(type) {
		case xml.StartElement:
			valueNode = ""
			if tok.Name.Local == "update" {
				log.Printf("Start of update data...")
				continue
			}
			if tok.Name.Local == "insert" || tok.Name.Local == "append" {
				updateMode = tok.Name.Local
				log.Printf("Update mode=%s", updateMode)
				continue
			}
			if tok.Name.Local == "wslsParent" || tok.Name.Local == "parent" || tok.Name.Local == "sequence" {
				valueNode = tok.Name.Local
				log.Printf("Start of value for %s", valueNode)
				continue
			}
			node, err := startNode(ctx, tok.Name.Local, nodeStack)
			if err != nil {
				log.Printf("FATAL: unable to start node for %s: %s", tok.Name.Local, err.Error())
				os.Exit(1)
			} else {
				nodeStack = append(nodeStack, node)
				nodes = append(nodes, node) //  add to sequental list of all nodes
			}
		case xml.CharData:
			val := strings.TrimSpace(string(tok))
			if len(val) > 0 {
				if valueNode == "parent" {
					parentID, err = strconv.ParseInt(val, 10, 64)
					if err != nil {
						log.Printf("FATAL: unable to parse parentID %s: %s", val, err.Error())
						os.Exit(1)
					}
				} else if valueNode == "wslsParent" {
					log.Printf("Lookup ID for WSLS ID: %s", val)
					if len(wslsCache) == 0 {
						wslsCache = cacheWSLSItemIDs(ctx)
					}
					parentID = wslsCache[val]
					log.Printf("    Found ID: %d", parentID)
				} else if valueNode == "sequence" {
					sequence, err = strconv.Atoi(val)
					if err != nil {
						log.Printf("FATAL: unable to parse sequence %s: %s", val, err.Error())
						os.Exit(1)
					}
				} else {
					node := nodeStack[len(nodeStack)-1]
					setNodeValue(ctx, node, val)
				}
			}
		case xml.EndElement:
			if includes(specialNodes, tok.Name.Local) == false {
				nodeStack = nodeStack[:len(nodeStack)-1]
			} else if tok.Name.Local == updateMode {
				// this is the end of the insert/append data. Do the DB update NOW
				log.Printf("===> %s mode. Parent: %d, Sequence: %d", updateMode, parentID, sequence)
				if updateMode != "append" {
					sequenceNodes(nodes[0])
					nodes[0].Sequence = sequence
				}
				log.Printf("Add nodes....")
				// log.Printf("MODE: %s, parent: %d data %#v", updateMode, parentID, *nodes[0])
				err := ctx.db.AddNodes(updateMode, nodes, parentID)
				if err != nil {
					log.Printf("FATAL: Unable to insert nodes: %s", err.Error())
					os.Exit(1)
				}
				log.Printf("Update completed")
				nodes = nil // reset to start again
			}
		}
	}
}

func cacheWSLSItemIDs(ctx *context) map[string]int64 {
	cache := make(map[string]int64)
	qs := "select value,parent_id from nodes where node_type_id=23"
	rows, _ := ctx.db.Query(qs)
	defer rows.Close()
	for rows.Next() {
		var wslsID string
		var parentID int64
		rows.Scan(&wslsID, &parentID)
		cache[wslsID] = parentID
	}
	return cache
}
