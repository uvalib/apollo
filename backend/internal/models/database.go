package models

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
)

// DBConfig : Configuration data from command-line params or ENV variables
type DBConfig struct {
	host     string
	database string
	user     string
	pass     string
}

// DB connection
type DB struct {
	*sqlx.DB
}

// GetConfig : get DB configuration data from ENV or command-line params
func GetConfig() (DBConfig, error) {
	// FIRST, try command line flags. Fallback is ENV variables
	var cfg DBConfig
	flag.StringVar(&cfg.host, "dbhost", os.Getenv("APOLLO_DB_HOST"), "DB Host (required)")
	flag.StringVar(&cfg.database, "dbname", os.Getenv("APOLLO_DB_NAME"), "DB Name (required)")
	flag.StringVar(&cfg.user, "dbuser", os.Getenv("APOLLO_DB_USER"), "DB User (required)")
	flag.StringVar(&cfg.pass, "dbpass", os.Getenv("APOLLO_DB_PASS"), "DB Password (required)")
	flag.Parse()

	// if anything is still not set, die
	if len(cfg.host) == 0 || len(cfg.user) == 0 ||
		len(cfg.pass) == 0 || len(cfg.database) == 0 {
		flag.Usage()
		return cfg, errors.New("Missing configuration")
	}
	return cfg, nil
}

// ConnectDB : Initialize the database connection based on data from config
func ConnectDB(cfg *DBConfig) (*DB, error) {
	log.Printf("Init DB connection to %s...", cfg.host)
	connectStr := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true", cfg.user, cfg.pass, cfg.host, cfg.database)

	// NOTE: Connect opens a connection and pings it in one step
	db, err := sqlx.Connect("mysql", connectStr)
	if err != nil {
		return nil, fmt.Errorf("Database connection failed: %s", err.Error())
	}
	log.Printf("DB Connection established")
	return &DB{db}, nil
}
