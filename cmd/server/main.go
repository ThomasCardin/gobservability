package main

import (
	"flag"
	"log"

	"github.com/ThomasCardin/peek/cmd/server/api"
	grpcServer "github.com/ThomasCardin/peek/cmd/server/grpc"
	"github.com/gin-gonic/gin"
)

var (
	port     = flag.String("port", "8080", "Port d'écoute du serveur HTTP")
	grpcPort = flag.String("grpc-port", "9090", "Port d'écoute du serveur gRPC")
	ginMode  = flag.String("mode", "release", "Mode Gin (debug|release)")
)

func main() {
	flag.Parse()

	gin.SetMode(*ginMode)
	r := gin.Default()

	r.LoadHTMLGlob("cmd/server/templates/*.html")
	r.Static("/static", "./cmd/server/templates/static")

	// AGENT endpoints (deprecated - now using gRPC)
	// r.POST("/api/stats", api.ReceiveStatsHandler)

	// UI endpoints
	r.GET("/", api.IndexHandler)                                                              // Page principale
	r.GET("/nodes", api.NodesFragmentHandler)                                                 // HTMX fragment
	r.GET("/pods/:nodename", api.PodsHandler)                                                 // Page pods pour un nœud
	r.GET("/pods/:nodename/fragment", api.PodsFragmentHandler)                                // Fragment HTMX pour pods
	r.GET("/pods/:nodename/metrics", api.PodsMetricsFragmentHandler)                          // Fragment HTMX pour métriques du nœud
	r.GET("/process/:nodename/:podname", api.ProcessDetailsPageHandler)                       // Page détails processus
	r.GET("/api/pods/:nodename/:podname/details", api.PodProcessDetailsHandler)               // API pour détails processus JSON
	r.GET("/api/pods/:nodename/:podname/info", api.PodInfoHandler)                            // Fragment HTMX pour infos pod
	r.GET("/api/pods/:nodename/:podname/details-fragment", api.ProcessDetailsFragmentHandler) // Fragment HTMX pour process details

	// Démarrer le serveur gRPC en goroutine
	go func() {
		log.Printf("Starting gRPC server on port %s", *grpcPort)
		if err := grpcServer.StartGRPCServer(*grpcPort); err != nil {
			log.Fatalf("error: starting gRPC server: %v", err)
		}
	}()

	log.Printf("HTTP server started on port %s", *port)
	log.Printf("Mode: %s, Cache TTL: 10s", *ginMode)

	if err := r.Run(":" + *port); err != nil {
		log.Fatalf("error: starting HTTP server: %v", err)
	}
}
