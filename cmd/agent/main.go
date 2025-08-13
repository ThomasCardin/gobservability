package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	grpcClient "github.com/ThomasCardin/gobservability/cmd/agent/grpc"
	"github.com/ThomasCardin/gobservability/cmd/agent/pkg/collector"
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

	var nodeName string
	var err error

	if *dev {
		os.Setenv(ENV_DEV_MODE, "true")
		log.Printf("[ENV] Development mode enabled - using / paths")
	} else {
		log.Printf("[ENV] Production mode - using /host paths")
	}

	// Priority: flag hostname > NODE_NAME env var > system hostname
	if *hostname != DEFAULT_NODE_NAME {
		nodeName = *hostname
	} else if envNodeName, found := os.LookupEnv(ENV_NODE_NAME); found {
		nodeName = envNodeName
	} else {
		log.Printf("[ENV] Warning: %s not set, using system hostname", ENV_NODE_NAME)
		nodeName, err = os.Hostname()
		if err != nil {
			nodeName = DEFAULT_NODE_NAME
		}
	}

	log.Printf("[ENV] Starting gobservability agent for node: %s", nodeName)
	log.Printf("[ENV] Metrics collection interval: %s", *collectInterval)
	log.Printf("[ENV] gRPC server address: %s", *grpcAddr)

	// Initialize streaming gRPC connection to server
	devModeValue := fmt.Sprintf("%t", *dev)
	grpcSender, err := grpcClient.NewStreamingGRPCClient(*grpcAddr, nodeName, devModeValue)
	if err != nil {
		log.Fatalf("error: failed to create streaming gRPC client: %v", err)
	}
	defer grpcSender.Close()

	// Initialize metrics collector with gRPC client
	metricsCollector := collector.NewCollector(ENV_DEV_MODE, grpcSender)
	metricsCollector.Start(nodeName, *collectInterval)
}
