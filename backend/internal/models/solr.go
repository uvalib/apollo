package models

import (
	"encoding/xml"
	"fmt"
	"log"
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

// GetSolrXML will return the Solr Add XML for the specified PID
func (db *DB) GetSolrXML(pid string) (string, error) {
	// First, get this item regardless of its level (collection or item)
	item, dbErr := db.GetChildren(pid)
	if dbErr != nil {
		return "", dbErr
	}

	// Now get collection info if the item is not it already
	var ancestry *Node
	if item.parentID.Valid {
		log.Printf("PID %s is not a collection; getting ancestry", pid)
		ancestry, _ = db.GetParentCollection(pid)
	} else {
		log.Printf("PID %s is a collection", pid)
	}
	var add solrAdd
	var fields []solrField

	// Generate the field mappings based on:
	//    https://confluence.lib.virginia.edu/display/DCMD/Indexing+Apollo+Content+in+Virgo+3
	outPID := getValue(item, "externalPID", item.PID)
	fields = append(fields, solrField{Name: "id", Value: outPID})
	fields = append(fields, solrField{Name: "source_facet", Value: "UVA Library Digital Repository"})
	if ancestry == nil {
		// the passed PID is for the collection
		title := getValue(item, "title", "")
		fields = append(fields, solrField{Name: "shadowed_location_facet", Value: "VISIBLE"})
		fields = append(fields, solrField{Name: "collection_title_display", Value: title})
		fields = append(fields, solrField{Name: "digital_collection_facet", Value: title})
		fields = append(fields, solrField{Name: "collection_title_text", Value: title})
	} else {
		// the passed pid is an ITEM
		title := getValue(ancestry, "title", "")
		fields = append(fields, solrField{Name: "shadowed_location_facet", Value: "UNDISCOVERABLE"})
		fields = append(fields, solrField{Name: "collection_title_display", Value: title})
		fields = append(fields, solrField{Name: "digital_collection_facet", Value: title})
		fields = append(fields, solrField{Name: "collection_title_text", Value: title})
		// TODO breadcrumbs_display
	}

	fields = append(fields, solrField{Name: "feature_facet", Value: "dl_metadata"})
	fields = append(fields, solrField{Name: "feature_facet", Value: "has_hierarchy"})
	fields = append(fields, solrField{Name: "feature_facet", Value: "suppress_ris_export"})
	fields = append(fields, solrField{Name: "feature_facet", Value: "suppress_refworks_export"})
	fields = append(fields, solrField{Name: "feature_facet", Value: "suppress_endnote_export"})
	fields = append(fields, solrField{Name: "date_received_facet", Value: getNow()})

	// TODO hierarchy_display

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
		log.Printf("This node has an associated digital object; getting IIIF manifest")
		// TODO
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
	if strings.Compare(node.Type.Name, typeName) == 0 {
		return node.Value
	}
	for _, c := range node.Children {
		if strings.Compare(c.Type.Name, typeName) == 0 {
			return c.Value
		}
	}
	return defaultVal
}

// Check if the source node contains a child of the specified type
func hasChild(node *Node, typeName string) bool {
	if strings.Compare(node.Type.Name, typeName) == 0 {
		return true
	}
	for _, c := range node.Children {
		if strings.Compare(c.Type.Name, typeName) == 0 {
			return true
		}
	}
	return false
}
