package main

import (
	"flag"
	"log"
	"os"

	"github.com/uvalib/apollo/backend/internal/models"
)

type apollosvrConfig struct {
	dbConfig   models.DBConfig
	port       int
	devUser    string
	solrDir    string
	qdcDir     string
	iiifManURL string
	fedoraURL  string
	hostname   string
}

func getConfig() apollosvrConfig {
	log.Printf("Loading configuration...")
	cfg := apollosvrConfig{}
	defSolrDir := "/lib_content23/record_source_for_solr_cores/apollo/data/record_dropbox"
	flag.StringVar(&cfg.dbConfig.Host, "dbhost", os.Getenv("APOLLO_DB_HOST"), "DB Host (required)")
	flag.StringVar(&cfg.dbConfig.Database, "dbname", os.Getenv("APOLLO_DB_NAME"), "DB Name (required)")
	flag.StringVar(&cfg.dbConfig.User, "dbuser", os.Getenv("APOLLO_DB_USER"), "DB User (required)")
	flag.StringVar(&cfg.dbConfig.Pass, "dbpass", os.Getenv("APOLLO_DB_PASS"), "DB Password (required)")
	//
	flag.IntVar(&cfg.port, "port", 8080, "Port to offer service on (default 8080)")
	flag.StringVar(&cfg.devUser, "devuser", "", "Computing ID to use for fake authentication in dev mode")
	flag.StringVar(&cfg.iiifManURL, "iiif", "https://iiifman.lib.virginia.edu/pid", "IIIF Manifest service URL")
	flag.StringVar(&cfg.solrDir, "solr_dir", defSolrDir, "Dropoff dir for generated solr add docs")
	flag.StringVar(&cfg.qdcDir, "qdc_dir", "/digiserv-delivery/patron/dpla/qdc", "Delivery dir for generated QDC files for DPLA")
	flag.StringVar(&cfg.fedoraURL, "fedora", "http://fedora01.lib.virginia.edu", "Production Fedora instance")
	flag.StringVar(&cfg.hostname, "host", "apollo.lib.virginia.edu", "Apollo Hostname")

	flag.Parse()
	log.Printf("%#v", cfg)

	// if anything is still not set, die
	if len(cfg.dbConfig.Host) == 0 || len(cfg.dbConfig.User) == 0 ||
		len(cfg.dbConfig.Pass) == 0 || len(cfg.dbConfig.Database) == 0 {
		flag.Usage()
		log.Printf("FATAL: Missing DB configuration")
		os.Exit(1)
	}

	return cfg
}
