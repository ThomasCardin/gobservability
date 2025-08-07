package main

import (
	"flag"
	"log"

	"github.com/ThomasCardin/peek/cmd/server/api"
	"github.com/gin-gonic/gin"
)

var (
	port    = flag.String("port", "8080", "Port d'écoute du serveur")
	ginMode = flag.String("mode", "release", "Mode Gin (debug|release)")
)

func main() {
	flag.Parse()

	gin.SetMode(*ginMode)
	r := gin.Default()

	r.LoadHTMLGlob("cmd/server/templates/*.html")
	r.Static("/static", "./cmd/server/templates/static")

	// AGENT endpoints
	r.POST("/api/stats", api.ReceiveStatsHandler)

	// UI endpoints
	r.GET("/", api.IndexHandler)              // Page principale
	r.GET("/nodes", api.NodesFragmentHandler) // HTMX fragment

	log.Printf("Gobservability Server démarré sur le port %s", *port)
	log.Printf("Interface disponible: http://localhost:%s", *port)
	log.Printf("Mode: %s, Cache TTL: 10s", *ginMode)

	if err := r.Run(":" + *port); err != nil {
		log.Fatalf("Erreur démarrage serveur: %v", err)
	}
}
