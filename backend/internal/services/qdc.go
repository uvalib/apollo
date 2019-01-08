package services

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/uvalib/apollo/backend/internal/models"
)

// qdcControlledValue is a controllev value and the source URI for
// a QDC entry value
type qdcControlledValue struct {
	Value    string
	ValueURI string
}

// wslsQdcData holds all of the data needed to populate the QDC XML template for an item
// in the collection.
// NOTE: much ow WSLS has hardoced values, so for now, this code is specific to that collection
// and simplified. Once new collections need this functionality, it will have to be generalized
type wslsQdcData struct {
	PID         string
	Title       string
	Description string
	DateCreated string
	Duration    string
	Color       string
	Tag         string
	Places      []qdcControlledValue
	Topics      []qdcControlledValue
	Preview     string
}

func (d *wslsQdcData) CleanXMLSting(val string) string {
	out := strings.Replace(val, "&", "&amp;", -1)
	out = strings.Replace(out, "<", "&lt;", -1)
	out = strings.Replace(out, ">", "&gt;", -1)
	return out
}

func (d *wslsQdcData) FixDate(origDate string) string {
	if strings.Contains(origDate, "/") == false {
		return origDate
	}
	log.Printf("NOTICE: Date with slashes %s", origDate)
	r := regexp.MustCompile("^0/0/")
	if r.MatchString(origDate) {
		yr := strings.Split(origDate, "/")[2]
		log.Printf("   Fixed: %s", yr)
		return yr
	}
	r = regexp.MustCompile("/0/")
	out := r.ReplaceAllString(origDate, "/uu/")
	bits := strings.Split(out, "/")
	if len(bits) == 2 {
		d := bits[1][0:2]
		y := bits[1][2:6]
		out = fmt.Sprintf("%s-%s-%s", y, bits[0], d)
	} else {
		m := bits[0]
		if len(m) < 2 {
			m = fmt.Sprintf("0%s", m)
		}
		out = fmt.Sprintf("%s-%s-%s", bits[2], m, bits[1])
	}
	log.Printf("   Fixed: %s", out)
	return out
}

// GenerateQDC will generate a single QDC record and return it as a string. NOTE: at
// present, there is only one collection (WSLS) that supports this. Fail fi another is requested
func (svc *ApolloSvc) GenerateQDC(tgtID *models.NodeIdentifier) (string, error) {
	log.Printf("Generating QDC for Item %s", tgtID.PID)
	qdcTemplate := template.Must(template.ParseFiles("./templates/wsls_qdc.xml"))
	itemNode, err := svc.DB.GetItem(tgtID.ID)
	if err != nil {
		return "", err
	}
	coll, _ := svc.DB.GetParentCollection(itemNode)
	if coll.PID != "uva-an109873" {
		return "", errors.New("Target PID not a QDC candidate")
	}
	data := svc.getItemQDCData(itemNode)
	if data.PID == "" {
		return "", errors.New("Target PID has not been published")
	}
	if data.Title == "" {
		return "", errors.New("Target PID has no title")
	}

	var buf bytes.Buffer
	if err := qdcTemplate.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// PublishQDCForItems generates QDC for all items in the list
func (svc *ApolloSvc) PublishQDCForItems(outDir string, collectionID int64, items []models.NodeIdentifier, limit int) {
	log.Printf("Generating QDC for %d items in collection", len(items))
	qdcTemplate := template.Must(template.ParseFiles("./templates/wsls_qdc.xml"))

	// Rather than doing thousands of queries for items, grab the whole tree
	collection, _ := svc.DB.GetTree(collectionID)

	cnt := 0
	gen := 0
	for _, item := range items {
		cnt++

		// Pull item data from in-memory tree rather than multiple queries
		itemNode := findItemByID(item.ID, collection)
		if itemNode == nil {
			log.Printf("ERROR: Unable to find nodeID %d. SKIPPING", item.ID)
			continue
		}

		// Grab details for this item. Some items have not been published to the DL
		// and will not have an externalPID defined. Do not generate QDC for them.
		if cnt%1000 == 0 {
			log.Printf("Item %s : %.2f%% complete.", item.PID, float32(cnt)/float32(len(items))*100.0)
		}
		data := svc.getItemQDCData(itemNode)
		if data.PID == "" || data.Title == "" {
			continue
		}

		fileErr := writeQDCFile(outDir, data, qdcTemplate)
		if fileErr != nil {
			log.Printf("%s", fileErr.Error())
			continue
		}
		gen++
		if limit >= 0 && gen >= limit {
			log.Printf("Stopping after requested limit %d", limit)
			break
		}
	}

	if limit <= 0 {
		log.Printf("QDC generation done; %d records generated from %d total items", gen, cnt)
		svc.DB.NodePublished("dpla", collectionID, svc.AuthComputingID)
	}
}

func (svc *ApolloSvc) getItemQDCData(itemNode *models.Node) wslsQdcData {
	// Walk the child attributes and pluck out the ones we want
	var data wslsQdcData
	for _, child := range itemNode.Children {
		switch name := child.Type.Name; name {
		case "externalPID":
			data.PID = child.Value
		case "wslsID":
			data.Preview = fmt.Sprintf("%s/wsls/%s/%s-thumbnail.jpg", svc.FedoraURL, child.Value, child.Value)
		case "title":
			data.Title = data.CleanXMLSting(child.Value)
		case "abstract":
			data.Description = data.CleanXMLSting(child.Value)
		case "dateCreated":
			data.DateCreated = data.FixDate(child.Value)
		case "duration":
			if child.Value != "mag" {
				data.Duration = child.Value
			}
		case "wslsColor":
			data.Color = child.Value
		case "wslsTag":
			data.Tag = child.Value
		case "wslsTopic":
			cv := qdcControlledValue{Value: data.CleanXMLSting(child.Value), ValueURI: child.ValueURI}
			data.Topics = append(data.Topics, cv)
		case "wslsPlace":
			cv := qdcControlledValue{Value: data.CleanXMLSting(child.Value), ValueURI: child.ValueURI}
			data.Places = append(data.Places, cv)
		}
	}
	return data
}

// findItemByID will walk the cached collection tree and find the item level
// node that has the same ID as the target
func findItemByID(tgtID int64, currNode *models.Node) *models.Node {
	if currNode.ID == tgtID {
		return currNode
	}

	for _, child := range currNode.Children {
		hit := findItemByID(tgtID, child)
		if hit != nil {
			return hit
		}
	}
	return nil
}

func writeQDCFile(baseDir string, data wslsQdcData, qdcTemplate *template.Template) error {
	// Generate the nested directory structure needed to store the files...
	pidSubdir := filepath.Join(baseDir, generatePIDPath(data.PID))
	os.MkdirAll(pidSubdir, os.ModePerm)

	// open the destination file and truncate it to prepare for new content....
	qdcFilename := fmt.Sprintf("%s.xml", data.PID)
	outPath := filepath.Join(pidSubdir, qdcFilename)
	outFile, err := os.OpenFile(outPath, os.O_CREATE|os.O_RDWR, 0666)

	if err != nil {
		return fmt.Errorf("ERROR: Unable to open destination QDC file %s: %s", outPath, err.Error())
	}
	outFile.Truncate(0)
	outFile.Seek(0, 0)

	// log.Printf("Rendering QDC file %s", outPath)
	qdcTemplate.Execute(outFile, data)
	outFile.Close()
	os.Chmod(outPath, 0666)
	return nil
}

// generatePIDPath will break a PID up into a set of directories using 2-digit segments
// of the numeric portion of the PID
func generatePIDPath(pid string) string {
	parts := strings.Split(pid, ":")
	out := parts[0]
	numbers := parts[1]
	var subdir string
	for _, char := range numbers {
		subdir += string(char)
		if len(subdir) == 2 {
			out = filepath.Join(out, subdir)
			subdir = ""
		}
	}
	if subdir != "" {
		out = filepath.Join(out, subdir)
	}
	return out
}
