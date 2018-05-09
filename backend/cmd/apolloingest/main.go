package main

import (
	"encoding/xml"
	"flag"
	"io"
	"log"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/uvalib/apollo/backend/internal/models"
)

// Version of the command
const Version = "1.0.0"

type context struct {
	db    *models.DB
	names []models.NodeName
	user  *models.User
}

/**
 * MAIN
 */
func main() {
	log.Printf("===> Starting apollo ingest")

	// Get configuration
	// note: need to define all cmd line flags before calling GetConfig
	// as this calls flag.Parse(), and it requires all flags to be pre-defined
	var srcFile string
	var userID string
	flag.StringVar(&srcFile, "src", "", "File to ingest")
	flag.StringVar(&userID, "user", "lf6f", "File to ingest")
	dbCfg, err := models.GetConfig()
	if err != nil {
		log.Printf("FATAL: %s", err.Error())
		os.Exit(1)
	}

	// Use cfg to connect DB
	db, err := models.ConnectDB(&dbCfg)
	if err != nil {
		log.Printf("FATAL: %s", err.Error())
		os.Exit(1)
	}

	// Ingest requested file
	if len(srcFile) == 0 {
		log.Printf("FATAL: missing required -src parameter")
		os.Exit(1)
	}
	user, err := db.FindUserBy("computing_id", userID)
	if err != nil {
		log.Printf("FATAL: couldn't find user %s: %s", userID, err.Error())
		os.Exit(1)
	}

	// build a context for the ingest containing common data
	ctx := context{db: db, user: user, names: db.AllNames()}

	doIngest(ctx, srcFile)
}

/**
 * Ingest the XML file contained in the config data
 */
func doIngest(ctx context, srcFile string) {
	log.Printf("Start ingest of %s...", srcFile)
	xmlFile, err := os.Open(srcFile)
	if err != nil {
		log.Printf("ERROR: Unable to read source file %s: %s", srcFile, err.Error())
		return
	}
	defer xmlFile.Close()

	// stream the xml through a decoder, catching all start, data and end events
	// use the data at these events to build a list of nodes to be created
	decoder := xml.NewDecoder(xmlFile)
	nodeStack := []*models.Node{}
	nodes := []*models.Node{}
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
			node, err := startNode(&ctx, tok.Name.Local, nodeStack)
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
				node := nodeStack[len(nodeStack)-1]
				node.Value = val
				log.Printf("   value: %s", val)
			}
		case xml.EndElement:
			// pop last node from stack
			nodeStack = nodeStack[:len(nodeStack)-1]
		}
	}

	// Create all nodes now
	log.Printf("Creating all nodes...")
	err = ctx.db.CreateNodes(nodes)
	if err != nil {
		log.Printf("ERROR: Unable to create nodes: %s", err.Error())
	}
	log.Printf("==> DONE <==")
}

func startNode(ctx *context, name string, ancestors []*models.Node) (*models.Node, error) {
	var nn *models.NodeName
	var err error

	// first, find or create node name
	found := false
	for _, nodeName := range ctx.names {
		if strings.Compare(nodeName.Value, name) == 0 {
			nn = &nodeName
			found = true
			break
		}
	}
	if found == false {
		log.Printf("NodeName %s not found; CREATING...", name)
		nn, err = ctx.db.CreateNodeName(name)
		if err != nil {
			log.Printf("ERROR: Unable to create NodeName %s: %s", name, err.Error())
			return nil, err
		}
		ctx.names = append(ctx.names, *nn)
	}

	var parent *models.Node
	if len(ancestors) == 0 {
		log.Printf("Create ROOT node %s", name)
	} else {
		// get parent and full ancestry path
		parent = ancestors[len(ancestors)-1]
		log.Printf("Create node %s, parent %s", name, parent.Name.Value)
	}

	return &models.Node{Parent: parent, Name: nn, User: ctx.user}, nil
}
