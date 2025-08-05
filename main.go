package main

import (
	"fmt"
	"log"

	"github.com/ThomasCardin/peek/api"
	"github.com/gin-gonic/gin"
)

const (
	PORT = "8080"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// Load HTML templates
	r.LoadHTMLGlob("templates/*.html")

	// Serve static files (CSS, JS, etc.)
	r.Static("/static", "./templates/static")

	// API endpoints (keep for compatibility)
	r.GET("/api/nodes", api.NodesHandler)

	// HTMX endpoints
	r.GET("/nodes", api.NodesTemplateHandler)
	r.GET("/node/:nodeName/stats", api.NodeStatsHandler)

	// Serve main page
	r.GET("/", api.IndexHandler)

	fmt.Printf("Starting Gin web server on http://localhost:%s\n", PORT)
	fmt.Printf("HTMX interface available at: http://localhost:%s\n", PORT)

	log.Fatal(r.Run(":" + PORT))
}
