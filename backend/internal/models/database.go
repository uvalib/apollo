package models

import (
	"fmt"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
)

// DBConfig contains configuration data from command-line params or ENV variables
type DBConfig struct {
	Host     string
	Database string
	User     string
	Pass     string
}

// DB connection
type DB struct {
	*sqlx.DB
}

// ConnectDB will initialize the database connection based on data from config
func ConnectDB(cfg *DBConfig) (*DB, error) {
	log.Printf("Init DB connection to %s...", cfg.Host)
	connectStr := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true", cfg.User, cfg.Pass, cfg.Host, cfg.Database)

	// NOTE: Connect opens a connection and pings it in one step
	db, err := sqlx.Connect("mysql", connectStr)
	if err != nil {
		return nil, fmt.Errorf("Database connection failed: %s", err.Error())
	}
	db.SetConnMaxLifetime(time.Minute * 5)
	db.SetMaxIdleConns(5)
	db.SetMaxOpenConns(5)

	log.Printf("DB Connection established")
	return &DB{db}, nil
}
