package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/julienschmidt/httprouter"
	"github.com/rs/cors"
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
	var key, crt, devUser, iiifServer, solrDir, qdcDir string
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
	flag.StringVar(&qdcDir, "qdc_dir", "./tmp/qdc", "Delivery dir for generated QDC files for DPLA")

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
	// config information, like the database
	app := handlers.ApolloHandler{Version: Version, DB: db, DevAuthUser: devUser, IIIF: iiifServer, SolrDir: solrDir, QdcDir: qdcDir}

	// Set routes and start server
	// use julienschmidt router for all things API/version/health
	// These handlers are accessed thru the ApolloHandler which provides some
	// shared configuration info; DB, versions, authUser
	router := httprouter.New()
	router.GET("/version", app.VersionInfo)
	router.GET("/healthcheck", app.HealthCheck)
	router.GET("/authenticate", app.Authenticate)
	router.GET("/api/collections", handlers.GzipMiddleware(app.CollectionsIndex))
	router.GET("/api/collections/:pid", handlers.GzipMiddleware(app.CollectionsShow))
	router.GET("/api/items/:pid", handlers.GzipMiddleware(app.ItemShow))
	router.GET("/api/users", handlers.GzipMiddleware(app.UsersIndex))
	router.GET("/api/users/:id", handlers.GzipMiddleware(app.UsersShow))
	router.GET("/api/types", handlers.GzipMiddleware(app.TypesIndex))
	router.GET("/api/values/:name", handlers.GzipMiddleware(app.ValuesForName))
	router.GET("/api/external/:pid", handlers.GzipMiddleware(app.ExternalPIDLookup))
	router.GET("/api/solr/:pid", handlers.GzipMiddleware(app.GenerateSolr))

	// require the user auth info in headers for these
	router.POST("/api/publish/:pid", app.AuthMiddleware(app.PublishCollection))
	router.POST("/api/qdc/:pid", app.AuthMiddleware(app.GenerateQDC))

	// Create a standard go Mux to serve static files, and pass off
	// all other stuff the the router. this allows static files to be
	// served from /, and other stuff to be served under /api
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir("public")))
	mux.Handle("/authenticate", router)
	mux.Handle("/version", router)
	mux.Handle("/healthcheck", router)
	mux.Handle("/api/", cors.Default().Handler(router))

	// Serve the mux with cors and logging enabled
	portStr := fmt.Sprintf(":%d", port)
	if https == 1 {
		log.Printf("Start HTTPS service on port %s with CORS support enabled", portStr)
		log.Fatal(http.ListenAndServeTLS(portStr, crt, key, loggingHandler(mux)))
	} else {
		log.Printf("Start HTTP service on port %s with CORS support enabled", portStr)
		log.Fatal(http.ListenAndServe(portStr, loggingHandler(mux)))
	}
}

func loggingHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("Started %s %s", r.Method, r.RequestURI)
		next.ServeHTTP(w, r)
		log.Printf("COMPLETED %s %s in %s", r.Method, r.RequestURI, time.Since(start))
	})
}
