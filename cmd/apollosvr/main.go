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
	"github.com/uvalib/apollo/internal/handlers"
	"github.com/uvalib/apollo/internal/models"
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
	var key, crt string
	defPort, err := strconv.Atoi(os.Getenv("APOLLO_PORT"))
	if err != nil {
		defPort = 8080
	}
	defHTTPS, err := strconv.Atoi(os.Getenv("APOLLO_PORT"))
	if err != nil {
		defHTTPS = 0
	}
	flag.IntVar(&port, "port", defPort, "Port to offer service on (default 8080)")
	flag.IntVar(&https, "https", defHTTPS, "Use HTTPS? (default 0)")
	flag.StringVar(&key, "key", os.Getenv("APOLLO_KEY"), "Key for https connection")
	flag.StringVar(&crt, "crt", os.Getenv("APOLLO_CRT"), "Crt for https connection")
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
	app := handlers.ApolloHandler{Version: Version, DB: db}

	// Set routes and start server
	// use julienschmidt router for all things API/version/health
	// These handlers are accessed throu the ApolloHandler which provides some
	// shared configuration info, like DB, versions, etc...
	router := httprouter.New()
	router.GET("/version", app.VersionInfo)
	router.GET("/healthcheck", app.HealthCheck)
	router.GET("/api/collections", app.CollectionsIndex)
	router.GET("/api/collections/:pid", app.CollectionsShow)
	router.GET("/api/users", app.UsersIndex)
	router.GET("/api/users/:id", app.UsersShow)
	router.GET("/api/names", app.NamesIndex)

	// Create a standard go Mux to serve static files, and pass off
	// all other stuff the the router. this allows static files to be
	// served from /, and other stuff to be served under /api
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir("public")))
	mux.Handle("/version", router)
	mux.Handle("/healthcheck", router)
	mux.Handle("/api/", router)

	// Serve the mux with cors and logging enabled
	portStr := fmt.Sprintf(":%d", port)
	if https == 1 {
		log.Printf("Start HTTPS service on port %s with CORS support enabled", portStr)
		log.Fatal(http.ListenAndServeTLS(portStr, crt, key, cors.Default().Handler(loggingHandler(mux))))
	} else {
		log.Printf("Start HTTP service on port %s with CORS support enabled", portStr)
		log.Fatal(http.ListenAndServe(portStr, cors.Default().Handler(loggingHandler(mux))))
	}
}

func loggingHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("Started %s %s", r.Method, r.RequestURI)
		next.ServeHTTP(w, r)
		log.Printf("Completed %s %s in %s", r.Method, r.RequestURI, time.Since(start))
	})
}
