package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/uvalib/apollo/backend/internal/handlers"
)

func initRoutes(app *handlers.Apollo) *gin.Engine {
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
		api.GET("/search", app.SearchHandler)
		api.GET("/collections", app.CollectionsIndex)
		api.GET("/collections/:pid", app.CollectionsShow)
		api.GET("/items/:pid", app.ItemShow)
		api.POST("/qdc/:pid", app.GenerateQDC)
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
	return router
}
