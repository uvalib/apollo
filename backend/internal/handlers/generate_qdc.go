package handlers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"github.com/gin-gonic/gin"
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

func (d wslsQdcData) CleanXMLSting(val string) string {
	out := strings.Replace(val, "&", "&amp;", -1)
	out = strings.Replace(out, "<", "&lt;", -1)
	out = strings.Replace(out, ">", "&gt;", -1)
	return out
}

func (d wslsQdcData) FixDate(origDate string) string {
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

// GenerateQDC generates the QDC XML documents needed to publish to the DPLA
// NOTE: Test with this: curl -X POST http://localhost:8085/api/qdc/[PID]
func (app *Apollo) GenerateQDC(c *gin.Context) {
	pid := c.Param("pid")
	tgtPID := c.Query("item")
	limit, err := strconv.Atoi(c.Query("limit"))
	if err != nil {
		limit = -1
	}

	// HACK for now, only WSLS is an option. Choke on all other pids
	if pid != "uva-an109873" {
		out := fmt.Sprintf("QDC generation is not supported for %s", pid)
		c.String(http.StatusBadRequest, out)
		return
	}

	// lookup identifiers for the passed PID
	ids, err := app.DB.Lookup(pid)
	if err != nil {
		c.String(http.StatusNotFound, "%s not found", pid)
		return
	}

	// Get a list of identifters for all items in this collection. This
	// is a struct containing both PID and DB ID. Items are the only thing
	// that goes to DPLA
	itemIDs, err := app.DB.GetCollectionContainerIdentifiers(ids.ID, "item")
	if err != nil {
		out := fmt.Sprintf("Unable to retrieve collection items %s", err.Error())
		c.String(http.StatusInternalServerError, out)
		return
	}

	if tgtPID != "" {
		app.generateSingleQDCRecord(ids.ID, tgtPID)
	} else {
		// kick off the generation of QDC in a goroutine...
		go app.generateQDCForItems(ids.ID, itemIDs, limit)
	}

	c.String(http.StatusOK, "QDC is being generated to %s...", app.QdcDir)
}

func (app *Apollo) generateSingleQDCRecord(collectionID int64, tgtPID string) {
	log.Printf("Generating QDC for Item %s", tgtPID)
	tgtIDs, err := app.DB.Lookup(tgtPID)
	if err != nil {
		log.Printf("ERROR: Unable to find target PID %s.", tgtPID)
		return
	}
	qdcTemplate := template.Must(template.ParseFiles("./templates/wsls_qdc.xml"))
	collection, _ := app.DB.GetTree(collectionID)
	itemNode := findItemByID(tgtIDs.ID, collection)
	if itemNode == nil {
		log.Printf("ERROR: Unable to find nodeID %d. SKIPPING", tgtIDs.ID)
		return
	}
	data := app.getItemQDCData(itemNode)
	if data.PID == "" {
		log.Printf("Item %d:%s has no external PID and hasn't been published to DL. SKIPPING", tgtIDs.ID, tgtPID)
		return
	}
	if data.Title == "" {
		log.Printf("Item %d:%s has no Title. SKIPPING", tgtIDs.ID, tgtPID)
		return
	}

	fileErr := app.writeQDCFile(data, qdcTemplate)
	if fileErr != nil {
		log.Printf("%s", fileErr.Error())
	}
}

func (app *Apollo) generateQDCForItems(collectionID int64, items []models.ItemIDs, limit int) {
	log.Printf("Generating QDC for %d items in collection", len(items))
	qdcTemplate := template.Must(template.ParseFiles("./templates/wsls_qdc.xml"))
	collection, _ := app.DB.GetTree(collectionID)

	cnt := 0
	gen := 0
	for _, item := range items {
		cnt++
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
		data := app.getItemQDCData(itemNode)
		if data.PID == "" {
			// log.Printf("Item %d:%s has no external PID and hasn't been published to DL. SKIPPING", item.ID, item.PID)
			continue
		}
		if data.Title == "" {
			// log.Printf("Item %d:%s has no Title. SKIPPING", item.ID, item.PID)
			continue
		}

		fileErr := app.writeQDCFile(data, qdcTemplate)
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
	}
}

func (app *Apollo) writeQDCFile(data wslsQdcData, qdcTemplate *template.Template) error {
	// Generate the nested directory structure needed to store the files...
	pidSubdir := filepath.Join(app.QdcDir, generatePIDPath(data.PID))
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
	os.Chown(outPath, 118698, 10708) // libsnlocal:	libr-snlocal
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

func (app *Apollo) getItemQDCData(itemNode *models.Node) wslsQdcData {
	// Walk the child attributes and pluck out the ones we want
	var data wslsQdcData
	for _, child := range itemNode.Children {
		switch name := child.Type.Name; name {
		case "externalPID":
			data.PID = child.Value
		case "wslsID":
			data.Preview = fmt.Sprintf("%s/wsls/%s/%s-thumbnail.jpg", app.FedoraURL, child.Value, child.Value)
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
