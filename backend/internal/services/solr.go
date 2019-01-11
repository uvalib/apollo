package services

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/uvalib/apollo/backend/internal/models"
)

type solrAdd struct {
	XMLName xml.Name `xml:"add"`
	Doc     solrDoc
}

type solrDoc struct {
	XMLName xml.Name `xml:"doc"`
	Fields  []*solrField
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

type pbcoreData struct {
	WSLSID   string
	Abstract string
	Duration string
	Color    string
	Sound    string
}

// PublishSolrForItems will generate the Solr XML for all passed items, and publish it to the specified dir
func (svc *ApolloSvc) PublishSolrForItems(tgtDir string, IDs []models.NodeIdentifier, rootID int64) {
	// chop up id list into blocks chunks that can be executed concurrenty
	// limit the maximum number of concurrrent generation threads to 50
	// to avoid choking the DB, tracksys or IIIF manifest service
	var chunks [][]models.NodeIdentifier
	var maxConcurrent = 10
	var chunkSize = int(math.Round(float64(len(IDs)) / float64(maxConcurrent)))
	if chunkSize == 0 {
		chunkSize = 1
	}
	for i := 0; i < len(IDs); i += chunkSize {
		endIdx := i + chunkSize
		if endIdx > len(IDs) {
			endIdx = len(IDs)
		}
		chunks = append(chunks, IDs[i:endIdx])
	}

	// set up a wait group as big as the number of IDs to process
	var wg sync.WaitGroup
	wg.Add(len(chunks))

	// Kick off  generation of each block of IDs in a goroutine
	for _, chunk := range chunks {
		go svc.processIDs(tgtDir, chunk, &wg)
	}

	wg.Wait()
	log.Printf("All goroutines done; flagging publication complete by [%s]", svc.AuthComputingID)
	svc.DB.NodePublished("virgo", rootID, svc.AuthComputingID)
	log.Printf("Publication COMPLETE")
}

func (svc *ApolloSvc) processIDs(tgtDir string, IDs []models.NodeIdentifier, wg *sync.WaitGroup) {
	for _, ID := range IDs {
		xml, err := svc.GetSolrXML(ID.ID)
		if err != nil {
			log.Printf("ERROR: Unable to generate solr xml for %s: %s", ID.PID, err.Error())
		} else {
			filename := fmt.Sprintf("%s/%s.xml", tgtDir, ID.PID)
			log.Printf("Write file %s", filename)
			ioutil.WriteFile(filename, []byte(xml), 0777)
		}
	}
	wg.Done()
}

// GetSolrXML will return the Solr Add XML for the specified nodeID
// The general format is: <add><doc><field name="name"></field>, <field/>, ... </doc></add>
// If a field has multiple values, just add multiple field elements with
// the same name attribute
func (svc *ApolloSvc) GetSolrXML(nodeID int64) (string, error) {
	// First, get this item regardless of its level (collection or item)
	// log.Printf("Get SOLR XML for %d", nodeID)
	item, dbErr := svc.DB.GetItem(nodeID)
	if dbErr != nil {
		return "", dbErr
	}

	// Now get collection info if the item is not it already
	var ancestry *models.Node
	if item.Ancestry.Valid {
		ancestry, _ = svc.DB.GetAncestry(item)
	} else {
		log.Printf("ID %d is a collection", nodeID)
	}
	var add solrAdd
	var fields []*solrField

	// Generate the field mappings based on:
	//    https://confluence.lib.virginia.edu/display/DCMD/Indexing+Apollo+Content+in+Virgo+3
	// If the start node has a externalPID use it. If not, default to Apollo PID
	outPID := getValue(item, "externalPID", item.PID)
	fields = append(fields, &solrField{Name: "id", Value: outPID})
	fields = append(fields, &solrField{Name: "source_facet", Value: "UVA Library Digital Repository"})
	var breadcrumbXML string
	if ancestry == nil {
		// the passed PID is for the collection
		title := getValue(item, "title", "")
		fields = append(fields, &solrField{Name: "shadowed_location_facet", Value: "VISIBLE"})
		fields = append(fields, &solrField{Name: "collection_title_display", Value: title})
		fields = append(fields, &solrField{Name: "digital_collection_facet", Value: title})
		fields = append(fields, &solrField{Name: "collection_title_text", Value: title})
		fields = append(fields, &solrField{Name: "breadcrumbs_display", Value: "<breadcrumbs></breadcrumbs>"})
		fields = append(fields, &solrField{Name: "hierarchy_level_display", Value: "collection"})
	} else {
		// the passed pid is an ITEM
		title := getValue(ancestry, "title", "")
		fields = append(fields, &solrField{Name: "shadowed_location_facet", Value: "UNDISCOVERABLE"})
		fields = append(fields, &solrField{Name: "collection_title_display", Value: title})
		fields = append(fields, &solrField{Name: "digital_collection_facet", Value: title})
		fields = append(fields, &solrField{Name: "collection_title_text", Value: title})
		breadcrumbXML = getBreadcrumbXML(ancestry)
		fields = append(fields, &solrField{Name: "breadcrumbs_display", Value: breadcrumbXML})
	}

	fields = append(fields, &solrField{Name: "feature_facet", Value: "dl_metadata"})
	fields = append(fields, &solrField{Name: "feature_facet", Value: "suppress_ris_export"})
	fields = append(fields, &solrField{Name: "feature_facet", Value: "suppress_refworks_export"})
	fields = append(fields, &solrField{Name: "feature_facet", Value: "suppress_endnote_export"})
	fields = append(fields, &solrField{Name: "date_received_facet", Value: getNow()})

	fields = append(fields, &solrField{Name: "feature_facet", Value: "has_hierarchy"})
	hierarchyXML := svc.getHierarchyXML(nodeID, breadcrumbXML)
	fields = append(fields, &solrField{Name: "hierarchy_display", Value: hierarchyXML})

	title := getValue(item, "title", "")
	fields = append(fields, &solrField{Name: "main_title_display", Value: title})
	fields = append(fields, &solrField{Name: "title_display", Value: title})
	fields = append(fields, &solrField{Name: "title_text", Value: title})
	fields = append(fields, &solrField{Name: "full_title_text", Value: title})

	if hasChild(item, "reel") {
		reel := fmt.Sprintf("From Reel %s", getValue(item, "reel", ""))
		fields = append(fields, &solrField{Name: "abstract_display", Value: reel})
		fields = append(fields, &solrField{Name: "abstract_text", Value: reel})
	} else if hasChild(item, "description") {
		desc := getValue(item, "description", "")
		if len(desc) > 0 {
			fields = append(fields, &solrField{Name: "abstract_display", Value: desc})
			fields = append(fields, &solrField{Name: "abstract_text", Value: desc})
		}
	}

	dobj := getValue(item, "digitalObject", "")
	if strings.Index(dobj, "images") > -1 {
		svc.addIIIFMetadata(item, &fields)
	} else if strings.Index(dobj, "wsls") > -1 {
		svc.addWSLSMetadata(item, &fields)
	}

	add.Doc.Fields = fields
	xmlOut, err := xml.MarshalIndent(add, "", "   ")
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
func getValue(node *models.Node, typeName string, defaultVal string) string {
	for _, c := range node.Children {
		if c.Type.Name == typeName {
			return cleanXMLSting(c.Value)
		}
	}
	return defaultVal
}

func cleanXMLSting(val string) string {
	out := strings.Replace(val, "&", "&amp;", -1)
	out = strings.Replace(out, "<", "&lt;", -1)
	out = strings.Replace(out, ">", "&gt;", -1)
	out = strings.Replace(out, "'", "&apos;", -1)
	out = strings.Replace(out, "\"", "&quot;", -1)
	return out
}

// Check if the source node contains a child of the specified type
func hasChild(node *models.Node, typeName string) bool {
	for _, c := range node.Children {
		if c.Type.Name == typeName {
			return true
		}
	}
	return false
}

func countComponents(node *models.Node) int {
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
func (svc *ApolloSvc) getHierarchyXML(rootID int64, breadcrumbXML string) string {
	// log.Printf("Get hierarchy XML for node %d", rootID)
	tree, err := svc.DB.GetTree(rootID)
	if err != nil {
		log.Printf("ERROR: Unable to get tree from node %d: %s", rootID, err.Error())
		return ""
	}

	var buffer bytes.Buffer
	buffer.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>")
	var counter int
	walkHierarchy(tree, &buffer, &counter)

	out := buffer.String()
	if breadcrumbXML != "" {
		// replace the first instance of the component tag with
		// a combimnation of the component tag and the breadcrumbs
		newComp := fmt.Sprintf("<component>%s", breadcrumbXML)
		return strings.Replace(out, "<component>", newComp, 1)
	}
	return out
}

func walkHierarchy(node *models.Node, buffer *bytes.Buffer, itemCnt *int) {
	// log.Printf("Walk node %s:%s", node.PID, node.Type.Name)
	if node.Ancestry.Valid == false {
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

	if node.Ancestry.Valid == false {
		buffer.WriteString("</collection>")
	} else {
		buffer.WriteString("</component>")
	}
}

// Get an escaped xml snippet detailing the ancestors in the passed node tree
func getBreadcrumbXML(ancestry *models.Node) string {
	var breadcrumbs []breadcrumb
	getBreadcrumbs(ancestry, &breadcrumbs)
	out := "<breadcrumbs>"
	for _, bc := range breadcrumbs {
		out += fmt.Sprintf("<ancestor><id>%s</id><title>%s</title></ancestor>", bc.PID, bc.Title)
	}
	out += "</breadcrumbs>"
	return out
}

// recursively walk down the ancestry tree rooted at the node param. Return
// an array of breadcrumb structs
func getBreadcrumbs(node *models.Node, breadcrumbs *[]breadcrumb) {
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

// Add WSLS metadata to the solrAdd record
func (svc *ApolloSvc) addWSLSMetadata(node *models.Node, fields *[]*solrField) {
	wslsID := getValue(node, "wslsID", "")
	if wslsID == "" {
		log.Printf("ERROR: WSLS ID NOT FOUND")
		return
	}
	updateField(fields, "shadowed_location_facet", "VISIBLE")
	*fields = append(*fields, &solrField{Name: "format_facet", Value: "Online"})
	*fields = append(*fields, &solrField{Name: "content_model_facet", Value: "uva-lib:pbcore2CModel"})

	// set hierarchy_level_display; collection for top level, item for digitalObjects and series for others
	if node.Ancestry.String == "" {
		*fields = append(*fields, &solrField{Name: "hierarchy_level_display", Value: "collection"})
	} else if hasChild(node, "digitalObject") {
		*fields = append(*fields, &solrField{Name: "hierarchy_level_display", Value: "item"})
	} else {
		*fields = append(*fields, &solrField{Name: "hierarchy_level_display", Value: "series"})
	}

	if strings.Compare(getValue(node, "hasVideo", "false"), "true") == 0 {
		*fields = append(*fields, &solrField{Name: "format_facet", Value: "Video"})
		*fields = append(*fields, &solrField{Name: "format_facet", Value: "Streaming Video"})
		url := fmt.Sprintf("%s/wsls/%s/%s.webm", svc.FedoraURL, wslsID, wslsID)
		*fields = append(*fields, &solrField{Name: "url_display", Value: url})
	}
	if strings.Compare(getValue(node, "hasScript", "false"), "true") == 0 {
		*fields = append(*fields, &solrField{Name: "format_facet", Value: "Anchor Script"})
		url := fmt.Sprintf("%s/wsls/%s/%s.pdf", svc.FedoraURL, wslsID, wslsID)
		*fields = append(*fields, &solrField{Name: "anchor_script_pdf_url_display", Value: url})
		url = fmt.Sprintf("%s/wsls/%s/%s-script-thumbnail.jpg", svc.FedoraURL, wslsID, wslsID)
		*fields = append(*fields, &solrField{Name: "anchor_script_thumbnail_url_display", Value: url})
		url = fmt.Sprintf("%s/wsls/%s/%s.txt", svc.FedoraURL, wslsID, wslsID)
		*fields = append(*fields, &solrField{Name: "anchor_script_text_url_display", Value: url})
		scriptTxt, err := getAPIResponse(url)
		if err != nil {
			log.Printf("Unable to get anchor script text: %s", err.Error())
		} else {
			*fields = append(*fields, &solrField{Name: "anchor_script_text", Value: scriptTxt})
			*fields = append(*fields, &solrField{Name: "anchor_script_display", Value: scriptTxt})
		}
	}

	// Date stuff; date is of format YYYY-MM-DD or just YYYY
	dateCreated := getValue(node, "dateCreated", "")
	if dateCreated != "" {
		*fields = append(*fields, &solrField{Name: "date_coverage_display", Value: dateCreated})
		*fields = append(*fields, &solrField{Name: "date_coverage_text", Value: dateCreated})
		*fields = append(*fields, &solrField{Name: "published_date_display", Value: dateCreated})
		year := strings.Split(dateCreated, "-")[0]
		*fields = append(*fields, &solrField{Name: "year_multisort_i", Value: year})
		*fields = append(*fields, &solrField{Name: "published_date_facet", Value: getPublishedDateFacet(dateCreated)})
	}

	// add subjects (subject_facet and subject_text)
	for _, c := range node.Children {
		if c.Type.Name == "wslsTopic" {
			*fields = append(*fields, &solrField{Name: "subject_facet", Value: c.Value})
			*fields = append(*fields, &solrField{Name: "subject_text", Value: c.Value})
		}
	}

	// add regions (region_facet and region_text)
	for _, c := range node.Children {
		if c.Type.Name == "wslsPlace" {
			*fields = append(*fields, &solrField{Name: "region_facet", Value: c.Value})
			*fields = append(*fields, &solrField{Name: "region_text", Value: c.Value})
		}
	}

	// Add PBCore XML
	var data pbcoreData
	data.WSLSID = wslsID
	data.Abstract = getValue(node, "abstract", "")
	data.Duration = getValue(node, "duration", "")
	color := getValue(node, "wslsColor", "")
	data.Color = "BW"
	if strings.Contains(color, "color") {
		data.Color = "COLOR"
	} else if strings.Contains(color, "netagtives") {
		data.Color = "BW neg"
	}
	data.Sound = "Sound"
	if strings.Contains(getValue(node, "wslsTag", ""), "silent") {
		data.Sound = "Silent"
	}
	pbcTemplate := template.Must(template.ParseFiles("./templates/pb_core.xml"))
	var pbcBuf bytes.Buffer
	if err := pbcTemplate.Execute(&pbcBuf, data); err != nil {
		log.Printf("ERROR: Unable to render pbcCore template %s", err.Error())
	}
	*fields = append(*fields, &solrField{Name: "pbcore_display", Value: pbcBuf.String()})
}

// Get IIIF manifest for the target node and add data to the solr fields array
func (svc *ApolloSvc) addIIIFMetadata(node *models.Node, fields *[]*solrField) {
	pid := getValue(node, "externalPID", node.PID)
	iiifManURL := fmt.Sprintf("%s/%s", svc.IIIFManifestURL, pid)
	iiifManifest, err := getAPIResponse(iiifManURL)
	if err != nil {
		log.Printf("ERROR: Unable to retrieve IIIF Manifest: %s", err.Error())
		return
	}
	*fields = append(*fields, &solrField{Name: "format_facet", Value: "Online"})
	*fields = append(*fields, &solrField{Name: "feature_facet", Value: "iiif"})
	*fields = append(*fields, &solrField{Name: "iiif_presentation_metadata_display", Value: iiifManifest})
	*fields = append(*fields, &solrField{Name: "feature_facet", Value: "pdf_service"})
	*fields = append(*fields, &solrField{Name: "pdf_url_display", Value: "http://pdfws.lib.virginia.edu:8088"})
	exemplar := parseExemplar(iiifManifest)
	if exemplar == "" {
		log.Printf("WARN: No thumbnail for %s", node.PID)
	} else {
		*fields = append(*fields, &solrField{Name: "thumbnail_url_display", Value: exemplar})
	}
}

func parseExemplar(iiifManifest string) string {
	// Looking for first line like this:
	//  "thumbnail":"https://iiif.lib.virginia.edu/iiif/tsm:2601265/full/!200,200/0/default.jpg",
	// Parse out the url from between the quotes
	idxThumb := strings.Index(iiifManifest, "thumbnail")
	if idxThumb == -1 {
		return ""
	}
	fn := "default.jpg"
	idxJPG := strings.Index(iiifManifest[idxThumb:len(iiifManifest)], fn)
	idxJPG += idxThumb
	log.Printf("thumb JPG %d", idxJPG)
	if idxJPG == -1 || idxThumb > idxJPG {
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

func updateField(fields *[]*solrField, fieldName string, newValue string) {
	for _, field := range *fields {
		if strings.Compare(field.Name, fieldName) == 0 {
			field.Value = newValue
			break
		}
	}
}

func getAPIResponse(url string) (string, error) {
	log.Printf("Get API response from: %s", url)
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

// getPublishedDateFacet accept a date in the form YYYY-MM-DD and returns a string to be
// used in the  published_date_facet field. Details here:
// https://confluence.lib.virginia.edu/pages/viewpage.action?spaceKey=SSE&title=Format+of+Solr+add+documents+for+the+new+Virgo+Index
func getPublishedDateFacet(dateStr string) string {
	// possible values: This year, Last 12 months, Last 3 years, Last 10 years,
	// Last 50 years, More than 50 years ago
	const RFC3339FullDate = "2006-01-02"
	parsed, err := time.Parse(RFC3339FullDate, dateStr)
	if err != nil {
		log.Printf("Unable to parse date %s: %s", dateStr, err.Error())
		return ""
	}
	thisYear := time.Now().Local().Year()
	year := parsed.Year()
	switch deltaYears := thisYear - year; deltaYears {
	case 0:
		return "This year"
	case 1:
		return "Last 12 Months"
	case 3:
		return "Last 3 Years"
	case 10:
		return "Last 10 Years"
	case 50:
		return "Last 50 Years"
	default:
		return "More than 50 years ago"
	}
}
