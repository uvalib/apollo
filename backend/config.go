package main

import (
	"flag"
	"log"
	"os"
)

type dbConfig struct {
	Host     string
	Database string
	User     string
	Pass     string
}

type apolloConfig struct {
	dbConfig   dbConfig
	port       int
	devUser    string
	iiifManURL string
	apolloURL  string
}

func getConfig() apolloConfig {
	log.Printf("Loading configuration...")
	cfg := apolloConfig{}
	flag.StringVar(&cfg.dbConfig.Host, "dbhost", os.Getenv("APOLLO_DB_HOST"), "DB Host (required)")
	flag.StringVar(&cfg.dbConfig.Database, "dbname", os.Getenv("APOLLO_DB_NAME"), "DB Name (required)")
	flag.StringVar(&cfg.dbConfig.User, "dbuser", os.Getenv("APOLLO_DB_USER"), "DB User (required)")
	flag.StringVar(&cfg.dbConfig.Pass, "dbpass", os.Getenv("APOLLO_DB_PASS"), "DB Password (required)")
	//
	flag.IntVar(&cfg.port, "port", 8080, "Port to offer service on (default 8080)")
	flag.StringVar(&cfg.devUser, "devuser", "", "Computing ID to use for fake authentication in dev mode")
	flag.StringVar(&cfg.iiifManURL, "iiif", "https://iiifman.lib.virginia.edu/pid", "IIIF Manifest service URL")
	flag.StringVar(&cfg.apolloURL, "apollo", "https://apollo.lib.virginia.edu", "Apollo URL")

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
