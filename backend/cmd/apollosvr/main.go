package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/uvalib/apollo/backend/internal/handlers"
	"github.com/uvalib/apollo/backend/internal/models"
)

// Version of the service
const Version = "1.0.0"

/**
 * MAIN
 */
func main() {
	log.Printf("===> apollo staring up <===")
	var port int
	var https int
	var key, crt, devUser, iiifServer, solrDir, qdcDir, fedoraURL string
	defPort, err := strconv.Atoi(os.Getenv("APOLLO_PORT"))
	if err != nil {
		defPort = 8080
	}
	defHTTPS, err := strconv.Atoi(os.Getenv("APOLLO_HTTPS"))
	if err != nil {
		defHTTPS = 0
	}
	flag.IntVar(&port, "port", defPort, "Port to offer service on (default 8080)")
	flag.IntVar(&https, "https", defHTTPS, "Use HTTPS? (default 0)")
	flag.StringVar(&key, "key", os.Getenv("APOLLO_KEY"), "Key for https connection")
	flag.StringVar(&crt, "crt", os.Getenv("APOLLO_CRT"), "Crt for https connection")
	flag.StringVar(&devUser, "devuser", "", "Computing ID to use for fake authentication in dev mode")
	flag.StringVar(&iiifServer, "iiif", "https://tracksys.lib.virginia.edu:8080", "IIIF Manifest service URL")
	flag.StringVar(&solrDir, "solr_dir", "./tmp", "Dropoff dir for generated solr add docs")
	flag.StringVar(&qdcDir, "qdc_dir", "/digiserv-delivery/patron/dpla/qdc", "Delivery dir for generated QDC files for DPLA")
	flag.StringVar(&fedoraURL, "fedora", "http://fedora01.lib.virginia.edu", "Production Fedora instance")

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

	// Create the main handler object which has access to common
	app := handlers.ApolloHandler{Version: Version, DB: db, DevAuthUser: devUser,
		IIIF: iiifServer, FedoraURL: fedoraURL, SolrDir: solrDir, QdcDir: qdcDir}
	log.Printf("Config: %#v", app)

	// Set routes and start server
	router := gin.Default()
	router.Use(static.Serve("/", static.LocalFile("./public", true)))
	router.GET("/version", app.VersionInfo)
	router.GET("/healthcheck", app.HealthCheck)
	router.GET("/authenticate", app.Authenticate)

	// create an api routing group and gzip all of its responses
	api := router.Group("/api")
	// TODO add cors upport with gin middleware
	api.Use(gzip.Gzip(gzip.DefaultCompression))
	api.Use(cors.Default())
	{
		api.GET("/aries/:id", app.AriesLookup)
		api.GET("/collections", app.CollectionsIndex)
		api.GET("/collections/:pid", app.CollectionsShow)
		api.GET("/external/:pid", app.ExternalPIDLookup)
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

	portStr := fmt.Sprintf(":%d", port)
	if https == 1 {
		log.Printf("Start HTTPS service on port %s with CORS support enabled", portStr)
		log.Fatal(router.RunTLS(portStr, crt, key))
	} else {
		log.Printf("Start HTTP service on port %s with CORS support enabled", portStr)
		log.Fatal(router.Run(portStr))
	}
}
