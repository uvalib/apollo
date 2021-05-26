package main

import (
	"bytes"
	"container/list"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
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
	Rights      string
}

func (d *wslsQdcData) CleanXMLSting(val string) string {
	out := strings.Replace(val, "&", "&amp;", -1)
	out = strings.Replace(out, "<", "&lt;", -1)
	out = strings.Replace(out, ">", "&gt;", -1)
	return out
}

func (d *wslsQdcData) FixDate(origDate string) string {
	if origDate == "" {
		return "[1951..1971]"
	}
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

// GetDPLAPIDs returns a list of PIDs for items that are published to the DPLA
func (app *Apollo) GetDPLAPIDs(c *gin.Context) {
	pid := "uva-an109873"
	log.Printf("INFO: get collection for Apollo PID %s", pid)
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
	log.Printf("INFO: collection tree retrieved from DB; find items with video")
	pidList := list.New()
	traverseTreeForDPLA(pidList, root)
	out := ""
	cnt := 0
	for e := pidList.Front(); e != nil; e = e.Next() {
		if out != "" {
			out += ","
		}
		cnt++
		out += fmt.Sprintf("%s", e.Value)
	}
	log.Printf("INFO: %d DPLA PIDS found", cnt)
	c.String(http.StatusOK, out)
}

// GetQDC returns QDC for a WSLS item
func (app *Apollo) GetQDC(c *gin.Context) {
	pid := c.Param("pid")
	log.Printf("INFO: Get QDC for %s", pid)
	itemIDs, dbErr := lookupIdentifier(&app.DB, pid)
	if dbErr != nil {
		log.Printf("ERROR: %s", dbErr.Error())
		c.String(http.StatusNotFound, dbErr.Error())
		return
	}

	item, dbErr := getNode(&app.DB, itemIDs.ID)
	if dbErr != nil {
		log.Printf("ERROR: %s", dbErr.Error())
		c.String(http.StatusNotFound, dbErr.Error())
		return
	}

	// note: if above was successful, this will be as well
	parent, _ := getNodeCollection(&app.DB, item)
	if parent.PID != "uva-an109873" {
		log.Printf("%s is not a QDC candidate", pid)
		c.String(http.StatusBadRequest, fmt.Sprintf("%s is not q QDC candidate", pid))
		return
	}

	var data wslsQdcData
	for _, child := range item.Children {
		switch name := child.Type.Name; name {
		case "externalPID":
			data.PID = child.Value
		case "wslsRights":
			if child.Value == "Local" {
				data.Rights = "https://creativecommons.org/licenses/by/4.0/"
			} else {
				data.Rights = "http://rightsstatements.org/vocab/NoC-US/1.0/"
			}
		case "wslsID":
			data.Preview = fmt.Sprintf("%s/%s/%s-thumbnail.jpg", app.WSLSURL, child.Value, child.Value)
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

	if data.PID == "" {
		log.Printf("ERROR: %s has not been published", pid)
		c.String(http.StatusNotFound, fmt.Sprintf("%s not published", pid))
		return
	}
	if data.Title == "" {
		log.Printf("ERROR: %s has no title", pid)
		c.String(http.StatusNotFound, fmt.Sprintf("%s has no title", pid))
		return
	}

	var buf bytes.Buffer
	if err := app.QDCTemplate.Execute(&buf, data); err != nil {
		log.Printf("ERROR: %s", err.Error())
		c.String(http.StatusInternalServerError, "unable to generate qdc")
		return
	}
	c.String(http.StatusOK, buf.String())
}

func traverseTreeForDPLA(pids *list.List, node *Node) {
	if node.Type.Container {
		externalPID := ""
		hasVideo := false
		for _, child := range node.Children {
			if child.Type.Name == "externalPID" {
				externalPID = child.Value
			}
			if child.Type.Name == "hasVideo" {
				if child.Value != "false" {
					hasVideo = true
				}
			}
			if child.Type.Container {
				traverseTreeForDPLA(pids, child)
			}
		}
		if externalPID != "" && node.Type.Name == "item" {
			if hasVideo {
				pids.PushBack(externalPID)
			} else {
				log.Printf("INFO: Skip %s with no video", externalPID)
			}
		}
	}
}
