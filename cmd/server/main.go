package main

import (
	"flag"
	"log"

	"github.com/ThomasCardin/gobservability/cmd/server/alerts"
	"github.com/ThomasCardin/gobservability/cmd/server/api"
	grpcServer "github.com/ThomasCardin/gobservability/cmd/server/grpc"
	"github.com/ThomasCardin/gobservability/cmd/server/storage"
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
	r.POST("/api/pods/:nodename/:podname/flamegraph", api.GenerateFlamegraphHandler)          // API pour démarrer génération flamegraph
	r.GET("/api/flamegraph/:taskid/status", api.FlamegraphStatusHandler)                      // API pour vérifier statut flamegraph
	r.GET("/api/flamegraph/:taskid/download", api.DownloadFlamegraphHandler)                  // API pour télécharger flamegraph
	r.GET("/flamegraph/:nodename/:podname", api.FlamegraphPageHandler)                        // Page dédiée pour afficher flamegraph

	// Initialiser le système d'alertes
	alertsManager, err := alerts.NewAlertsManager()
	if err != nil {
		log.Printf("Warning: Failed to initialize alerts system: %v", err)
		log.Printf("Alerts will be disabled. Make sure POSTGRES_URL and DISCORD_WEBHOOK_URL are set.")
	} else {
		log.Printf("Alerts system initialized successfully")
		
		// Configurer le storage pour l'API et le GlobalStore pour l'évaluation
		api.SetAlertsStorage(alertsManager.GetStorage())
		api.SetDiscordNotifier(alertsManager.GetDiscord())
		storage.GlobalStore.SetAlertsManager(alertsManager)
		
		// Routes pour les alertes
		r.GET("/alerts/:nodename", api.AlertsPageHandler)                    // Page principale des alertes
		r.GET("/api/alerts/:nodename", api.GetAlertsHandler)                 // API JSON pour les alertes
		r.GET("/api/alerts/:nodename/fragment", api.GetAlertsFragmentHandler) // Fragment HTMX
		r.POST("/api/alerts/:nodename", api.CreateAlertRuleHandler)          // Créer une règle
		r.PUT("/api/alerts/:nodename/:ruleid", api.UpdateAlertRuleHandler)   // Modifier une règle
		r.DELETE("/api/alerts/:nodename/:ruleid", api.DeleteAlertRuleHandler) // Supprimer une règle
		r.PUT("/api/alerts/dismiss/:alertid", api.DismissAlertHandler)       // Dismiss une alerte active
		r.GET("/api/alerts/:nodename/history", api.GetAlertHistoryHandler)   // Historique des alertes
		
		// Cleanup à l'arrêt du serveur
		defer func() {
			if err := alertsManager.Close(); err != nil {
				log.Printf("Error closing alerts manager: %v", err)
			}
		}()
	}

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
