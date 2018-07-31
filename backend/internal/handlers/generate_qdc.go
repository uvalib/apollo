package handlers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/julienschmidt/httprouter"
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

// GenerateQDC generates the QDC XML documents needed to publish to the DPLA
// NOTE: Test with this: curl --header "remote_user: lf6f" -X POST http://localhost:8085/api/qdc/[PID]
func (app *ApolloHandler) GenerateQDC(rw http.ResponseWriter, req *http.Request, params httprouter.Params) {
	pid := params.ByName("pid")

	// HACK for now, only WSLS is an option. Choke on all other pids
	if pid != "uva-an109873" {
		out := fmt.Sprintf("QDC generation is not supported for %s", pid)
		http.Error(rw, out, http.StatusBadRequest)
		return
	}

	// Convert the PID to a nodeID. Given the above, it is safe to ignore
	// errors as the pid is known to exist
	nodeID, _ := app.DB.GetNodeIDFromPID(pid)

	// Get a list of identifters for all items in this collection. This
	// is a struct containing both PID and DB ID. Items are the only thing
	// that goes to DPLA
	ids, err := app.DB.GetCollectionContainerIdentifiers(nodeID, "item")
	if err != nil {
		out := fmt.Sprintf("Unable to retrieve collection items %s", err.Error())
		http.Error(rw, out, http.StatusInternalServerError)
		return
	}

	// kick off the generation of QDC in a goroutine...
	go app.generateQDCForItems(nodeID, ids)

	fmt.Fprintf(rw, "QDC is being generated to %s...", app.QdcDir)
}

func (app *ApolloHandler) generateQDCForItems(collectionID int64, items []models.ItemIDs) {
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
		log.Printf("Item %s : %.2f%% complete.", item.PID, float32(cnt)/float32(len(items))*100.0)
		data := app.getItemQDCData(itemNode)
		if data.PID == "" {
			log.Printf("Item %d:%s has no external PID and hasn't been published to DL. SKIPPING", item.ID, item.PID)
			continue
		}

		// Generate the nested directory structure needed to store the files...
		pidSubdir := filepath.Join(app.QdcDir, generatePIDPath(data.PID))
		os.MkdirAll(pidSubdir, os.ModePerm)

		// open the destination file and truncate it to prepare for new content....
		qdcFilename := fmt.Sprintf("%s.xml", data.PID)
		outPath := filepath.Join(pidSubdir, qdcFilename)
		outFile, err := os.OpenFile(outPath, os.O_CREATE|os.O_RDWR, 0666)
		if err != nil {
			log.Printf("ERROR: Unable to open destination QDC file %s: %s", outPath, err.Error())
			continue
		}
		outFile.Truncate(0)
		outFile.Seek(0, 0)

		qdcTemplate.Execute(outFile, data)
		outFile.Close()
		gen++
	}
	log.Printf("QDC generation done; %d records generated from %d total items", gen, cnt)
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

func (app *ApolloHandler) getItemQDCData(itemNode *models.Node) wslsQdcData {
	// Walk the child attributes and pluck out the ones we want
	var data wslsQdcData
	for _, child := range itemNode.Children {
		switch name := child.Type.Name; name {
		case "externalPID":
			data.PID = child.Value
		case "wslsID":
			data.Preview = fmt.Sprintf("%s/wsls/%s-thumbnail.jpg", app.FedoraURL, child.Value)
		case "title":
			data.Title = child.Value
		case "abstract":
			data.Description = child.Value
		case "dateCreated":
			data.DateCreated = child.Value
		case "duration":
			data.Duration = child.Value
		case "wslsColor":
			data.Color = child.Value
		case "wslsTag":
			data.Tag = child.Value
		case "wslsTopic":
			cv := qdcControlledValue{Value: child.Value, ValueURI: child.ValueURI}
			data.Topics = append(data.Topics, cv)
		case "wslsPlace":
			cv := qdcControlledValue{Value: child.Value, ValueURI: child.ValueURI}
			data.Places = append(data.Places, cv)
		}
	}
	return data
}
