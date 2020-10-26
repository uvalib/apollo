package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
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
	if tgtFormat != "json" && tgtFormat != "xml" && tgtFormat != "uvamap" {
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
		xml, err := generateXML(root, tgtFormat)
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

func generateXML(node *Node, xmlType string) (string, error) {
	log.Printf("Generate %s for collection %s", xmlType, node.PID)
	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)
	traverseTree(writer, node, xmlType)
	writer.Flush()
	return buf.String(), nil
}

type digitalObjectInfo struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type nodeMapping struct {
	OpenTag  string
	CloseTag string
	Sibling  string
}

func traverseTree(out *bufio.Writer, node *Node, xmlType string) {
	nm := mapNodeName(node.Type.Name, xmlType)
	if node.Type.Container {
		out.WriteString(fmt.Sprintf("%s\n", nm.OpenTag))
		if node.Type.Name == "collection" && xmlType == "uvamap" {
			out.WriteString("<metadataSource>Apollo</metadataSource>\n")
			out.WriteString(fmt.Sprintf("<sourceRecordIdentifier source=\"Apollo\">%s</sourceRecordIdentifier>\n", node.PID))
		}
		for _, child := range node.Children {
			if child.Type.Name == "dpla" {
				// skip the DPLA tag; it is no longer used
				continue
			}

			if child.Type.Name == "filmBoxLabel" && xmlType == "uvamap" {
				if child.Value != "no label" {
					out.WriteString(fmt.Sprintf("<alternativeTitle>%s</alternativeTitle>\n", cleanValue(child.Value)))
					out.WriteString(fmt.Sprintf("<orig_note>Container title: %s</orig_note>\n", cleanValue(child.Value)))
				}
				continue
			}
			if child.Type.Name == "hasScript" && xmlType == "uvamap" {
				if child.Value == "true" {
					out.WriteString("<orig_note>Script available</orig_note>\n")
				} else {
					out.WriteString("<orig_note>Script not available</orig_note>\n")
				}
				continue
			}
			if child.Type.Name == "hasVideo" && xmlType == "uvamap" {
				if child.Value == "true" {
					out.WriteString("<orig_note>Video available</orig_note>\n")
				} else {
					out.WriteString("<orig_note>Video not available</orig_note>\n")
				}
				continue
			}
			if child.Type.Name == "title" && xmlType == "uvamap" {
				t := cleanValue(child.Value)
				out.WriteString(fmt.Sprintf("<title>%s</title>\n", t))
				out.WriteString(fmt.Sprintf("<displayTitle>%s</displayTitle>\n", t))
				r := regexp.MustCompile("\\A(A\\s+)|(An\\s+)|(The\\s+)")
				st := r.ReplaceAllString(t, "")
				out.WriteString(fmt.Sprintf("<sortTitle>%s</sortTitle>\n", st))
				continue
			}
			if child.Type.Name == "wslsColor" && xmlType == "uvamap" {
				if strings.Index(child.Value, "black") >= 0 {
					out.WriteString("<colorContent>black and white</colorContent>\n<physDetails>negative</phyDetails>\n")
				} else {
					out.WriteString("<colorContent>color</colorContent>\n")
				}
				continue
			}
			if child.Type.Name == "wslsTag" && xmlType == "uvamap" {
				val := strings.Split(child.Value, " ")[0]
				out.WriteString(fmt.Sprintf("<soundContent>%s</soundContent>\n", val))
				continue
			}

			if child.Type.Name == "digitalObject" {
				// value looks like: {type: images|wsls id: external_id}
				var doInfo digitalObjectInfo
				doErr := json.Unmarshal([]byte(child.Value), &doInfo)
				if doErr != nil {
					log.Printf("ERROR: unable to read digital object info %s", doErr.Error())
				} else {
					embedURL := fmt.Sprintf("https://iiif-manifest.internal.lib.virginia.edu/pid/%s", doInfo.ID)
					val := fmt.Sprintf("https://curio.lib.virginia.edu/view/uv/uv.html#?manifest=%s", url.QueryEscape(embedURL))
					if xmlType == "xml" {
						out.WriteString(fmt.Sprintf("<%s>%s</%s>\n", child.Type.Name, val, child.Type.Name))
					} else {
						out.WriteString(fmt.Sprintf("<uri access=\"%s\" usage=\"primary\"></uri>\n", val))
						out.WriteString(fmt.Sprintf("<uri access=\"%s\" displayLabel=\"iiifManifest\"></uri>\n", embedURL))
					}
				}
			} else {
				cm := mapNodeName(child.Type.Name, xmlType)
				if child.Type.Container == false {
					if child.ValueURI != "" {
						t := cm.OpenTag
						r := regexp.MustCompile("</|<|>")
						st := r.ReplaceAllString(t, "")
						out.WriteString(fmt.Sprintf("<%s href=\"%s\">%s</%s>\n",
							st, child.ValueURI, child.Value, st))
					} else {
						out.WriteString(fmt.Sprintf("%s%s%s\n", cm.OpenTag, cleanValue(child.Value), cm.CloseTag))
						if cm.Sibling != "" {
							out.WriteString(fmt.Sprintf("<%s>%s</%s>\n", cm.Sibling, cleanValue(child.Value), cm.Sibling))
						}
					}
				} else {
					traverseTree(out, child, xmlType)
				}
			}
		}
		out.WriteString(fmt.Sprintf("%s\n", nm.CloseTag))
	} else {
		out.WriteString(fmt.Sprintf("%s%s%s\n", nm.OpenTag, cleanValue(node.Value), nm.CloseTag))
	}
}

func mapNodeName(nodeName string, xmlType string) nodeMapping {
	out := nodeMapping{OpenTag: fmt.Sprintf("<%s>", nodeName), CloseTag: fmt.Sprintf("</%s>", nodeName)}
	if xmlType == "xml" {
		// nothing to do... just return as-is
		return out
	}

	// NOTE: the elements that have href fo not get the angle brackets as they will be added by the traverse
	mapping := map[string]nodeMapping{"abstract": {OpenTag: "<abstractSummary>", CloseTag: "</abstractSummary>"},
		"barcode":     {OpenTag: "<itemID>", CloseTag: "</itemID>"},
		"catalogKey":  {OpenTag: "<sourceRecordIdentifier source=\"SIRSI\">", CloseTag: "</sourceRecordIdentifier>"},
		"collection":  {OpenTag: "<metadata type=\"collection\">", CloseTag: "</metadata>"},
		"description": {OpenTag: "<abstractSummary>", CloseTag: "</abstractSummary>"},
		"duration":    {OpenTag: "<playingTime>", CloseTag: "</playingTime>"},
		"entity":      {OpenTag: "<subject>", CloseTag: "</subject>", Sibling: "subjectName"},
		"externalPID": {OpenTag: "<localIdentifier displayLabel=\"UVA PID\">", CloseTag: "</localIdentifier>"},
		"issue":       {OpenTag: "<metadata type=\"issue\">", CloseTag: "</metadata>"},
		"item":        {OpenTag: "<metadata type=\"item\">", CloseTag: "</metadata>"},
		"month":       {OpenTag: "<metadata type=\"month\">", CloseTag: "</metadata>"},
		"reel":        {OpenTag: "<callNumber displayLabel=\"reel\">", CloseTag: "</callNumber>"},
		"useRights":   {OpenTag: "useRestrict", CloseTag: "useRestrict"},
		"volume":      {OpenTag: "<metadata type=\"volume\">", CloseTag: "</metadata>"},
		"wslsID":      {OpenTag: "<localIdentifier displayLabel=\"WSLS ID\">", CloseTag: "</localIdentifier>"},
		"wslsPlace":   {OpenTag: "subjectGeographic", CloseTag: "subjectGeographic"},
		"wslsRights":  {OpenTag: "<useRestrict>", CloseTag: "</useRestrict>"},
		"wslsTopic":   {OpenTag: "<subject>", CloseTag: "</subject>", Sibling: "subjectName"},
		"year":        {OpenTag: "<metadata type=\"year\">", CloseTag: "</metadata>"},
	}
	if val, ok := mapping[nodeName]; ok {
		return val
	}

	return out
}

func cleanValue(val string) string {
	clean := strings.TrimSpace(val)
	clean = strings.ReplaceAll(clean, "&", "&amp;")
	clean = strings.ReplaceAll(clean, "<", "&lt;")
	clean = strings.ReplaceAll(clean, ">", "&gt;")
	return clean
}
