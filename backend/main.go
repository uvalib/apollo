package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
)

// Version of the service
const version = "2.1.3"

/**
 * MAIN
 */
func main() {
	log.Printf("===> Apollo staring up <===")

	log.Printf("INFO: load configuration....")
	cfg := getConfig()

	log.Printf("INFO: initialize service....")
	app, err := InitService(version, &cfg)
	if err != nil {
		log.Printf("FATAL: %s", err.Error())
		os.Exit(1)
	}

	gin.SetMode(gin.ReleaseMode)
	gin.DisableConsoleColor()
	router := gin.Default()
	router.Use(cors.Default())

	router.GET("/version", app.VersionInfo)
	router.GET("/favicon.ico", app.IgnoreFavicon)
	router.GET("/healthcheck", app.HealthCheck)

	// create an api routing group and gzip all of its responses
	api := router.Group("/api")
	api.Use(gzip.Gzip(gzip.DefaultCompression))
	{
		api.GET("/aries", app.AriesPing)
		api.GET("/aries/:id", app.AriesLookup)
		api.GET("/collections", app.ListCollections)
		api.GET("/collections/:pid", app.GetCollection)
		api.GET("/items/:pid", app.GetItemDetails)
		api.GET("/search", app.SearchHandler)
		api.GET("/types", app.GetNodeTypes)
		api.GET("/values/:name", app.GeControlledValues)
		api.GET("/published/dpla", app.GetDPLAPIDs)
		api.GET("/dpla/:pid", app.GetQDC)
	}

	// Note: in dev mode, this is never actually used. The front end is served
	// by yarn and it proxies all requests to the API to the routes above
	router.Use(static.Serve("/", static.LocalFile("./public", true)))

	// add a catchall route that renders the index page.
	// based on no-history config setup info here:
	//    https://router.vuejs.org/guide/essentials/history-mode.html#example-server-configurations
	router.NoRoute(func(c *gin.Context) {
		c.File("./public/index.html")
	})

	log.Printf("INFO: start Apollo on port %d", cfg.port)
	log.Fatal(router.Run(fmt.Sprintf(":%d", cfg.port)))
}
