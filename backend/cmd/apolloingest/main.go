package main

import (
	"encoding/xml"
	"flag"
	"fmt"
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
	db     *models.DB
	types  []models.NodeType
	values []models.ControlledValue
	user   *models.User
}

/**
 * MAIN
 */
func main() {
	log.Printf("===> Starting apollo ingest")

	// Get configuration
	// note: need to define all cmd line flags before calling GetConfig
	// as this calls flag.Parse(), and it requires all flags to be pre-defined
	var srcFile, userID, mode string
	var dbCfg models.DBConfig
	flag.StringVar(&srcFile, "src", "", "File to ingest")
	flag.StringVar(&userID, "user", "lf6f", "File to ingest")
	flag.StringVar(&mode, "mode", "create", "Ingest mode [create|update] Default: create")
	flag.StringVar(&dbCfg.Host, "dbhost", os.Getenv("APOLLO_DB_HOST"), "DB Host (required)")
	flag.StringVar(&dbCfg.Database, "dbname", os.Getenv("APOLLO_DB_NAME"), "DB Name (required)")
	flag.StringVar(&dbCfg.User, "dbuser", os.Getenv("APOLLO_DB_USER"), "DB User (required)")
	flag.StringVar(&dbCfg.Pass, "dbpass", os.Getenv("APOLLO_DB_PASS"), "DB Password (required)")

	flag.Parse()

	// if anything is still not set, die
	if len(dbCfg.Host) == 0 || len(dbCfg.User) == 0 ||
		len(dbCfg.Pass) == 0 || len(dbCfg.Database) == 0 {
		flag.Usage()
		log.Printf("FATAL: Missing DB configuration")
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
	ctx := context{db: db, user: user, types: db.ListNodeTypes(), values: db.ListAllControlledValues()}

	if mode == "create" {
		// doIngest(&ctx, srcFile)
	} else {
		doUpdate(&ctx, srcFile)
	}
}

func includes(data []string, tgt string) bool {
	for _, val := range data {
		if val == tgt {
			return true
		}
	}
	return false
}

/**
 * Ingest the XML file contained in the config data
 */
func doIngest(ctx *context, srcFile string) {
	log.Printf("Start ingest of %s...", srcFile)
	xmlFile, xmlErr := os.Open(srcFile)
	if xmlErr != nil {
		log.Printf("ERROR: Unable to read source file %s: %s", srcFile, xmlErr.Error())
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
			node, err := startNode(ctx, tok.Name.Local, nodeStack)
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
				setNodeValue(ctx, node, val)
			}
		case xml.EndElement:
			// pop last node from stack
			nodeStack = nodeStack[:len(nodeStack)-1]
		}
	}

	log.Printf("Initialize sequence...")
	sequenceNodes(nodes[0])

	// Create all nodes now
	log.Printf("Creating all nodes...")
	createErr := ctx.db.CreateNodes(nodes)
	if createErr != nil {
		log.Printf("ERROR: Unable to create nodes: %s", createErr.Error())
	}
	log.Printf("==> DONE <==")
}

func sequenceNodes(node *models.Node) {
	if len(node.Children) > 0 {
		seq := 0
		for _, c := range node.Children {
			c.Sequence = seq
			seq++
			sequenceNodes(c)
		}
	}
}

func startNode(ctx *context, name string, ancestors []*models.Node) (*models.Node, error) {
	var nt *models.NodeType
	var err error

	// first, find or create node name
	found := false
	for _, nodeType := range ctx.types {
		if strings.Compare(nodeType.Name, name) == 0 {
			nt = &nodeType
			found = true
			break
		}
	}
	if found == false {
		log.Printf("NodeType %s not found; CREATING...", name)
		nt, err = ctx.db.CreateNodeType(name)
		if err != nil {
			log.Printf("ERROR: Unable to create NodeType %s: %s", name, err.Error())
			return nil, err
		}
		ctx.types = append(ctx.types, *nt)
	}

	var parent *models.Node
	if len(ancestors) == 0 {
		log.Printf("Create ROOT node %s", name)
	} else {
		// get parent and full ancestry path
		parent = ancestors[len(ancestors)-1]
		log.Printf("Create node %s, parent %s", name, parent.Type.Name)
	}

	newNode := &models.Node{Parent: parent, Type: nt, User: ctx.user}
	if parent != nil {
		parent.Children = append(parent.Children, newNode)
	}

	return newNode, nil
}

func setNodeValue(ctx *context, node *models.Node, val string) {
	if node.Type.ControlledVocab == false {
		node.Value = val
		log.Printf("   value: %s", val)
		return
	}
	log.Printf("Look up controlled value [%s]", val)
	cv := findControlledValue(ctx.values, val)
	if cv == nil {
		log.Printf("WARN: no controlled value match found for %s. Just setting value directly.", val)
		node.Value = val
	}
	if cv.TypeID != node.Type.ID {
		log.Printf("WARN: controlled value / node name mismatch (%d vs %d) for %s. Just setting value directly.",
			node.Type.ID, cv.TypeID, val)
		node.Value = val
	}
	log.Printf("Controlled value %s replaced with ID %d", val, cv.ID)
	node.Value = fmt.Sprintf("%d", cv.ID)
}

func findControlledValue(values []models.ControlledValue, tgtVal string) *models.ControlledValue {
	for _, val := range values {
		if val.Value == tgtVal {
			return &val
		}
	}
	return nil
}
