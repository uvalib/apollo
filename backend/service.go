package main

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

// DB wraps the database connection object
type DB struct {
	*sqlx.DB
}

// Apollo is the applicatin object through which all requests are handled.
// It contains common config information and services, like the DB
type Apollo struct {
	Version         string
	ApolloURL       string
	WSLSURL         string
	DB              DB
	DevAuthUser     string
	AuthComputingID string
	IIIF            string
	QDCTemplate     *template.Template
}

// InitService will initialize the service context based on the config parameters
func InitService(version string, cfg *apolloConfig) (*Apollo, error) {
	svc := Apollo{Version: version,
		ApolloURL:   cfg.apolloURL,
		DevAuthUser: cfg.devUser,
		IIIF:        cfg.iiifManURL,
		WSLSURL:     cfg.wslsURL,
	}

	log.Printf("INFO: connecting to DB...")
	connectStr := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true&timeout=%ss&readTimeout=%ss&writeTimeout=%ss",
		cfg.dbConfig.User, cfg.dbConfig.Pass, cfg.dbConfig.Host, cfg.dbConfig.Database,
		cfg.dbConfig.Timeout, cfg.dbConfig.Timeout, cfg.dbConfig.Timeout)
	db, err := sqlx.Connect("mysql", connectStr)
	if err != nil {
		return nil, fmt.Errorf("database connection failed: %s", err.Error())
	}
	db.SetConnMaxLifetime(time.Minute * 5)
	db.SetMaxIdleConns(5)
	db.SetMaxOpenConns(5)
	svc.DB = DB{db}
	log.Printf("INFO: DB Connection established")

	log.Printf("INFO: Load QDC template")
	svc.QDCTemplate = template.Must(template.ParseFiles("./templates/wsls_qdc.xml"))

	return &svc, nil
}

// HealthCheck will report health of this and associated services
func (app *Apollo) HealthCheck(c *gin.Context) {
	err := app.DB.Ping()
	if err != nil {
		log.Printf("ERROR: healthcheck failure: %s", err)
		// gin.H is a shortcut for map[string]interface{}
		c.JSON(http.StatusInternalServerError, gin.H{"alive": "true", "mysql": "false"})
		return
	}
	log.Printf("INFO: healthcheck OK")
	c.JSON(http.StatusOK, gin.H{"alive": "true", "mysql": "true"})
}

// VersionInfo will report the version of the serivce
func (app *Apollo) VersionInfo(c *gin.Context) {
	build := "unknown"

	// cos our CWD is the bin directory
	files, _ := filepath.Glob("../buildtag.*")
	if len(files) == 1 {
		build = strings.Replace(files[0], "../buildtag.", "", 1)
	}

	vMap := make(map[string]string)
	vMap["version"] = app.Version
	vMap["build"] = build
	c.JSON(http.StatusOK, vMap)
}

// IgnoreFavicon is a dummy to handle browser favicon requests without warnings
func (app *Apollo) IgnoreFavicon(c *gin.Context) {
}
