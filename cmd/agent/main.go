package main

import (
	"flag"
	"log"
	"os"
	"time"

	grpcClient "github.com/ThomasCardin/peek/cmd/agent/grpc"
	"github.com/ThomasCardin/peek/cmd/agent/pkg/collector"
)

const (
	DEFAULT_NODE_NAME = "unknown"
	DEFAULT_GRPC_ADDR = "localhost:9090"

	ENV_NODE_NAME = "NODE_NAME"
	ENV_DEV_MODE  = "DEV_MODE"
)

var (
	grpcAddr        = flag.String("grpc-server", DEFAULT_GRPC_ADDR, "Server gRPC address")
	collectInterval = flag.Duration("interval", 5*time.Second, "Collect interval")
	hostname        = flag.String("hostname", DEFAULT_NODE_NAME, "Custom hostname (overrides NODE_NAME env var)")
	dev             = flag.Bool("dev", false, "Development mode (use / instead of /host)")
)

func main() {
	flag.Parse()

	if *dev {
		os.Setenv(ENV_DEV_MODE, "true")
		log.Printf("Development mode enabled - using / paths")
	} else {
		log.Printf("Production mode - using /host paths")
	}

	var nodeName string
	var err error

	// Priority: flag hostname > NODE_NAME env var > system hostname
	if *hostname != DEFAULT_NODE_NAME {
		nodeName = *hostname
	} else if envNodeName, found := os.LookupEnv(ENV_NODE_NAME); found {
		nodeName = envNodeName
	} else {
		log.Printf("Warning: %s not set, using system hostname", ENV_NODE_NAME)
		nodeName, err = os.Hostname()
		if err != nil {
			nodeName = DEFAULT_NODE_NAME
		}
	}

	log.Printf("Starting gobservability agent for node: %s", nodeName)
	log.Printf("Metrics collection interval: %s", *collectInterval)
	log.Printf("gRPC server address: %s", *grpcAddr)

	// Initialize gRPC connection to server
	log.Printf("Connecting to gRPC server...")
	grpcSender, err := grpcClient.NewGRPCClient(*grpcAddr, ENV_DEV_MODE)
	if err != nil {
		log.Fatalf("Failed to create gRPC client: %v", err)
	}
	defer grpcSender.Close()
	log.Printf("Successfully connected to gRPC server")

	// Initialize metrics collector with gRPC client
	metricsCollector := collector.NewCollector(ENV_DEV_MODE, grpcSender)

	// Start the main collection and sending loop
	log.Printf("Starting metrics collection and gRPC transmission loop...")
	log.Printf("Will collect and send metrics every %s via gRPC", *collectInterval)
	metricsCollector.Start(nodeName, *collectInterval)
}
