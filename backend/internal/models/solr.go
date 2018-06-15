package models

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

type solrAdd struct {
	XMLName xml.Name `xml:"add"`
	Doc     solrDoc
}

type solrDoc struct {
	XMLName xml.Name `xml:"doc"`
	Fields  *[]solrField
}

type solrField struct {
	XMLName xml.Name `xml:"field"`
	Name    string   `xml:"name,attr"`
	Value   string   `xml:",chardata"`
}

type breadcrumb struct {
	PID   string
	Title string
}

// GetSolrXML will return the Solr Add XML for the specified nodeID
func (db *DB) GetSolrXML(nodeID int64, iiifURL string) (string, error) {
	// First, get this item regardless of its level (collection or item)
	// log.Printf("Get SOLR XML for %d", nodeID)
	item, dbErr := db.GetChildren(nodeID)
	if dbErr != nil {
		return "", dbErr
	}

	// Now get collection info if the item is not it already
	var ancestry *Node
	if item.parentID.Valid {
		// log.Printf("ID %d is not a collection; getting ancestry", nodeID)
		ancestry, _ = db.GetAncestry(item)
	} else {
		// log.Printf("ID %d is a collection", nodeID)
	}
	var add solrAdd
	var fields []solrField

	// Generate the field mappings based on:
	//    https://confluence.lib.virginia.edu/display/DCMD/Indexing+Apollo+Content+in+Virgo+3
	// If the start node has a externalPID use it. If not, default to Apollo PID
	outPID := getValue(item, "externalPID", item.PID)
	fields = append(fields, solrField{Name: "id", Value: outPID})
	fields = append(fields, solrField{Name: "source_facet", Value: "UVA Library Digital Repository"})
	var breadcrumbXML string
	if ancestry == nil {
		// the passed PID is for the collection
		title := getValue(item, "title", "")
		fields = append(fields, solrField{Name: "shadowed_location_facet", Value: "VISIBLE"})
		fields = append(fields, solrField{Name: "collection_title_display", Value: title})
		fields = append(fields, solrField{Name: "digital_collection_facet", Value: title})
		fields = append(fields, solrField{Name: "collection_title_text", Value: title})
		fields = append(fields, solrField{Name: "breadcrumbs_display", Value: "<breadcrumbs></breadcrumbs>"})
		fields = append(fields, solrField{Name: "hierarchy_level_display", Value: "collection"})
	} else {
		// the passed pid is an ITEM
		title := getValue(ancestry, "title", "")
		fields = append(fields, solrField{Name: "shadowed_location_facet", Value: "UNDISCOVERABLE"})
		fields = append(fields, solrField{Name: "collection_title_display", Value: title})
		fields = append(fields, solrField{Name: "digital_collection_facet", Value: title})
		fields = append(fields, solrField{Name: "collection_title_text", Value: title})
		breadcrumbXML = getBreadcrumbXML(ancestry)
		fields = append(fields, solrField{Name: "breadcrumbs_display", Value: breadcrumbXML})
	}

	fields = append(fields, solrField{Name: "feature_facet", Value: "dl_metadata"})
	fields = append(fields, solrField{Name: "feature_facet", Value: "suppress_ris_export"})
	fields = append(fields, solrField{Name: "feature_facet", Value: "suppress_refworks_export"})
	fields = append(fields, solrField{Name: "feature_facet", Value: "suppress_endnote_export"})
	fields = append(fields, solrField{Name: "date_received_facet", Value: getNow()})

	fields = append(fields, solrField{Name: "feature_facet", Value: "has_hierarchy"})
	hierarchyXML := db.getHierarchyXML(nodeID, breadcrumbXML)
	fields = append(fields, solrField{Name: "hierarchy_display", Value: hierarchyXML})

	title := getValue(item, "title", "")
	fields = append(fields, solrField{Name: "main_title_display", Value: title})
	fields = append(fields, solrField{Name: "title_display", Value: title})
	fields = append(fields, solrField{Name: "title_text", Value: title})
	fields = append(fields, solrField{Name: "full_title_text", Value: title})

	if hasChild(item, "reel") {
		reel := fmt.Sprintf("From Reel %s", getValue(item, "reel", ""))
		fields = append(fields, solrField{Name: "abstract_display", Value: reel})
		fields = append(fields, solrField{Name: "abstract_text", Value: reel})
	} else if hasChild(item, "description") {
		desc := getValue(item, "description", "")
		if len(desc) > 0 {
			fields = append(fields, solrField{Name: "abstract_display", Value: desc})
			fields = append(fields, solrField{Name: "abstract_text", Value: desc})
		}
	}

	if hasChild(item, "digitalObject") {
		// log.Printf("This node has an associated digital object; getting IIIF manifest")
		addIIIFMetadata(item, &fields, iiifURL)
	}

	add.Doc.Fields = &fields
	xmlOut, err := xml.MarshalIndent(add, "", "  ")
	if err != nil {
		return "", err
	}
	return string(xmlOut), nil
}

// Get a timespamp in form: yyyMMdd
func getNow() string {
	t := time.Now()
	return fmt.Sprintf("%04d%02d%02d", t.Year(), t.Month(), t.Day())
}

// Walk the node children and see if the target node type
// exists; if it does, return the value. If not, return a default
func getValue(node *Node, typeName string, defaultVal string) string {
	for _, c := range node.Children {
		if c.Type.Name == typeName {
			return c.Value
		}
	}
	return defaultVal
}

// Check if the source node contains a child of the specified type
func hasChild(node *Node, typeName string) bool {
	for _, c := range node.Children {
		if c.Type.Name == typeName {
			return true
		}
	}
	return false
}

func countComponents(node *Node) int {
	count := 0
	if len(node.Children) > 0 {
		for _, c := range node.Children {
			if len(c.Children) > 0 {
				count++
			}
		}
	}
	return count
}

// Get the subtree rooted at the target node and convert it into an escaped
// XML hierarchy document
func (db *DB) getHierarchyXML(rootID int64, breadcrumbXML string) string {
	// log.Printf("Get hierarchy XML for node %d", rootID)
	tree, err := db.GetTree(rootID)
	if err != nil {
		log.Printf("ERROR: Unable to get tree from node %d: %s", rootID, err.Error())
		return ""
	}

	var buffer bytes.Buffer
	buffer.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>")
	var counter int
	walkHierarchy(tree, &buffer, &counter)

	//return escapeXML(buffer.String())
	out := buffer.String()
	if breadcrumbXML != "" {
		// replace the first instance of the component tag with
		// a combimnation of the component tag and the breadcrumbs
		newComp := fmt.Sprintf("<component>%s", breadcrumbXML)
		return strings.Replace(out, "<component>", newComp, 1)
	}
	return out
}

func walkHierarchy(node *Node, buffer *bytes.Buffer, itemCnt *int) {
	// log.Printf("Walk node %s:%s", node.PID, node.Type.Name)
	if node.parentID.Valid == false {
		buffer.WriteString("<collection>")
		title := getValue(node, "title", "")
		buffer.WriteString(fmt.Sprintf("<title>%s</title>", title))
		buffer.WriteString(fmt.Sprintf("<shorttitle>%s</shorttitle>", title))
		buffer.WriteString(fmt.Sprintf("<component_count>%d</component_count>", countComponents(node)))
		// buffer.WriteString("<digitized_component_count>0</digitized_component_count>")
	} else {
		componentCnt := countComponents(node)
		if componentCnt == 0 {
			// if this node has no components, it is an item. count it
			*itemCnt = *itemCnt + 1
			if *itemCnt > 3 && bytes.Contains(buffer.Bytes(), []byte("<collection>")) == true {
				return
			}
		}
		buffer.WriteString("<component>")
		buffer.WriteString(fmt.Sprintf("<id>%s</id>", getValue(node, "externalPID", node.PID)))
		buffer.WriteString(fmt.Sprintf("<type>%s</type>", node.Type.Name))
		title := getValue(node, "title", "")
		buffer.WriteString(fmt.Sprintf("<unittitle>%s</unittitle>", title))
		buffer.WriteString(fmt.Sprintf("<shortunittitle>%s</shortunittitle>", title))

		if componentCnt > 0 {
			buffer.WriteString(fmt.Sprintf("<component_count>%d</component_count>", componentCnt))
			*itemCnt = 0
		}
	}
	// log.Printf("BEFORE CHILDREN=%d", localCnt)

	for _, c := range node.Children {
		if len(c.Children) > 0 {
			walkHierarchy(c, buffer, itemCnt)
		}
	}

	if node.parentID.Valid == false {
		buffer.WriteString("</collection>")
	} else {
		buffer.WriteString("</component>")
	}
}

// Get an escaped xml snippet detailing the ancestors in the passed node tree
func getBreadcrumbXML(ancestry *Node) string {
	var breadcrumbs []breadcrumb
	getBreadcrumbs(ancestry, &breadcrumbs)
	out := "<breadcrumbs>"
	for _, bc := range breadcrumbs {
		out += fmt.Sprintf("<ancestor><id>%s</id><title>%s</title></ancestor>", bc.PID, bc.Title)
	}
	out += "</breadcrumbs>"
	// return escapeXML(out)
	return out
}

// recursively walk down the ancestry tree rooted at the node param. Return
// an array of breadcrumb structs
func getBreadcrumbs(node *Node, breadcrumbs *[]breadcrumb) {
	if len(node.Children) > 0 {
		bc := breadcrumb{
			PID:   getValue(node, "externalPID", node.PID),
			Title: getValue(node, "title", "")}
		*breadcrumbs = append(*breadcrumbs, bc)
		for _, c := range node.Children {
			if len(c.Children) > 0 {
				getBreadcrumbs(c, breadcrumbs)
			}
		}
	}
}

// Get IIIF manifest for the target node and add data to the solr fields array
func addIIIFMetadata(node *Node, fields *[]solrField, iiifURL string) {
	pid := getValue(node, "externalPID", node.PID)
	iiifManifest, err := getAPIResponse(fmt.Sprintf("%s/%s", iiifURL, pid))
	if err != nil {
		log.Printf("ERROR: Unable to retrieve IIIF Manifest: %s", err.Error())
		return
	}
	*fields = append(*fields, solrField{Name: "format_facet", Value: "Online"})
	*fields = append(*fields, solrField{Name: "feature_facet", Value: "iiif"})
	*fields = append(*fields, solrField{Name: "iiif_presentation_metadata_display", Value: iiifManifest})
	*fields = append(*fields, solrField{Name: "feature_facet", Value: "pdf_service"})
	*fields = append(*fields, solrField{Name: "pdf_url_display", Value: "http://pdfws.lib.virginia.edu:8088"})
	*fields = append(*fields, solrField{Name: "thumbnail_url_display", Value: parseExemplar(iiifManifest)})
}

func parseExemplar(iiifManifest string) string {
	// Looking for first line like this:
	//  "thumbnail":"https://iiif.lib.virginia.edu/iiif/tsm:2601265/full/!200,200/0/default.jpg",
	// Parse out the url from between the quotes
	idxThumb := strings.Index(iiifManifest, "thumbnail")
	log.Printf("thumb idx %d", idxThumb)
	fn := "default.jpg"
	idxJPG := strings.Index(iiifManifest[idxThumb:len(iiifManifest)], fn)
	idxJPG += idxThumb
	log.Printf("thumb JPG %d", idxJPG)
	if idxThumb == -1 || idxJPG == -1 || idxThumb > idxJPG {
		log.Printf("ERROR: Couldn't find thumbnail in IIIF")
		return ""
	}
	line := iiifManifest[idxThumb : idxJPG+len(fn)+1]
	colonIdx := strings.Index(line, ":")
	quotedURL := line[colonIdx+1 : len(line)]
	out := strings.Replace(quotedURL, "\"", "", -1)
	log.Printf("Thumb %s", out)
	return out
}

func escapeXML(src string) string {
	var outBuf bytes.Buffer
	xml.EscapeText(&outBuf, []byte(src))
	return outBuf.String()
}

func getAPIResponse(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	respString := string(bodyBytes)
	if resp.StatusCode != 200 {
		return "", errors.New(respString)
	}
	return respString, nil
}
