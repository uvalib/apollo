package models

import (
	"encoding/xml"
	"log"
	"strings"
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
	outPID := getExternalPID(item)
	fields = append(fields, solrField{Name: "id", Value: outPID})
	if ancestry == nil {
		fields = append(fields, solrField{Name: "shadowed_location_facet", Value: "VISIBLE"})
	} else {
		fields = append(fields, solrField{Name: "shadowed_location_facet", Value: "UNDISCOVERABLE"})
	}
	fields = append(fields, solrField{Name: "feature_facet", Value: "dl_metadata"})
	fields = append(fields, solrField{Name: "feature_facet", Value: "has_hierarchy"})
	fields = append(fields, solrField{Name: "feature_facet", Value: "suppress_ris_export"})
	fields = append(fields, solrField{Name: "feature_facet", Value: "suppress_refworks_export"})
	fields = append(fields, solrField{Name: "feature_facet", Value: "suppress_endnote_export"})
	fields = append(fields, solrField{Name: "source_facet",
		Value: "UVA Library Digital Repository"})

	add.Doc.Fields = &fields
	xmlOut, err := xml.MarshalIndent(add, "", "  ")
	if err != nil {
		return "", err
	}
	return string(xmlOut), nil
}

// Walk the node children and see if an externalPID can be found.
// If nothing found, just return the PID of the node
func getExternalPID(node *Node) string {
	if strings.Compare(node.Type.Name, "externalPID") == 0 {
		return node.Value
	}
	for _, c := range node.Children {
		if strings.Compare(c.Type.Name, "externalPID") == 0 {
			return c.Value
		}
	}
	return node.PID
}
