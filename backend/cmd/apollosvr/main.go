package main

import (
	"fmt"
	"log"
	"os"

	"github.com/uvalib/apollo/backend/internal/handlers"
	"github.com/uvalib/apollo/backend/internal/models"
)

// Version of the service
const version = "1.5.5"

/**
 * MAIN
 */
func main() {
	log.Printf("===> Apollo staring up <===")

	cfg := getConfig()
	db, err := models.ConnectDB(&cfg.dbConfig)
	if err != nil {
		log.Printf("FATAL: Unable to connect DB: %s", err.Error())
		os.Exit(1)
	}

	// Create the main Application object which has access to common config
	log.Printf("Setup routes...")
	app := handlers.Apollo{Version: version, DB: db, DevAuthUser: cfg.devUser,
		IIIF: cfg.iiifManURL, FedoraURL: cfg.fedoraURL, SolrDir: cfg.solrDir, QdcDir: cfg.qdcDir, ApolloHost: cfg.hostname}
	router := initRoutes(&app)

	log.Printf("Start Apollo on port %d", cfg.port)
	log.Fatal(router.Run(fmt.Sprintf(":%d", cfg.port)))
}
