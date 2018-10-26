package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/uvalib/apollo/backend/internal/handlers"
	"github.com/uvalib/apollo/backend/internal/models"
)

// Version of the service
const Version = "1.2.0"

/**
 * MAIN
 */
func main() {
	log.Printf("===> Apollo staring up <===")
	var port int
	var devUser, iiifServer, solrDir, qdcDir, fedoraURL, apolloHost string
	var dbCfg models.DBConfig
	defSolrDir := "/lib_content23/record_source_for_solr_cores/apollo/data/record_dropbox"
	flag.StringVar(&dbCfg.Host, "dbhost", os.Getenv("APOLLO_DB_HOST"), "DB Host (required)")
	flag.StringVar(&dbCfg.Database, "dbname", os.Getenv("APOLLO_DB_NAME"), "DB Name (required)")
	flag.StringVar(&dbCfg.User, "dbuser", os.Getenv("APOLLO_DB_USER"), "DB User (required)")
	flag.StringVar(&dbCfg.Pass, "dbpass", os.Getenv("APOLLO_DB_PASS"), "DB Password (required)")

	flag.IntVar(&port, "port", 8080, "Port to offer service on (default 8080)")
	flag.StringVar(&devUser, "devuser", "", "Computing ID to use for fake authentication in dev mode")
	flag.StringVar(&iiifServer, "iiif", "https://iiifman.lib.virginia.edu/pid", "IIIF Manifest service URL")
	flag.StringVar(&solrDir, "solr_dir", defSolrDir, "Dropoff dir for generated solr add docs")
	flag.StringVar(&qdcDir, "qdc_dir", "/digiserv-delivery/patron/dpla/qdc", "Delivery dir for generated QDC files for DPLA")
	flag.StringVar(&fedoraURL, "fedora", "http://fedora01.lib.virginia.edu", "Production Fedora instance")
	flag.StringVar(&apolloHost, "host", "apollo.lib.virginia.edu", "Apollo Hostname")

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
		log.Printf("FATAL: Unable to connect DB: %s", err.Error())
		os.Exit(1)
	}

	if devUser != "" {
		log.Printf("Running in dev mode with devUser=%s", devUser)
	}

	// Create the main handler object which has access to common
	app := handlers.ApolloHandler{Version: Version, DB: db, DevAuthUser: devUser,
		IIIF: iiifServer, FedoraURL: fedoraURL, SolrDir: solrDir, QdcDir: qdcDir, ApolloHost: apolloHost}
	log.Printf("Config: %#v", app)

	// Set routes and start server
	gin.SetMode(gin.ReleaseMode)
	gin.DisableConsoleColor()
	router := gin.Default()
	router.Use(static.Serve("/", static.LocalFile("./public", true)))
	router.GET("/version", app.VersionInfo)
	router.GET("/healthcheck", app.HealthCheck)
	router.GET("/authenticate", app.Authenticate)

	// create an api routing group and gzip all of its responses
	api := router.Group("/api")
	api.Use(gzip.Gzip(gzip.DefaultCompression))
	api.Use(cors.Default())
	{
		api.GET("/aries", app.AriesPing)
		api.GET("/aries/:id", app.AriesLookup)
		api.GET("/collections", app.CollectionsIndex)
		api.GET("/collections/:pid", app.CollectionsShow)
		api.GET("/items/:pid", app.ItemShow)
		api.GET("/qdc/:pid", app.GenerateQDC)
		api.GET("/solr/:pid", app.GenerateSolr)
		api.GET("/types", app.TypesIndex)
		api.GET("/users", app.UsersIndex)
		api.GET("/users/:id", app.UsersShow)
		api.GET("/values/:name", app.ValuesForName)

		// require the user auth info in headers for these
		api.POST("/publish/:pid", app.AuthMiddleware, app.PublishCollection)
	}

	// add a catchall route that renders the index page.
	// based on no-history config setup info here:
	//    https://router.vuejs.org/guide/essentials/history-mode.html#example-server-configurations
	router.NoRoute(func(c *gin.Context) {
		c.File("./public/index.html")
	})

	portStr := fmt.Sprintf(":%d", port)
	log.Printf("Start Apollo on port %s", portStr)
	log.Fatal(router.Run(portStr))
}
