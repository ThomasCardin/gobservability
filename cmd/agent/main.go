package main

import (
	"flag"
	"fmt"
	"log/slog"
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
		slog.Info("development mode enabled - using / paths", "component", "env")
	} else {
		slog.Info("production mode - using /host paths", "component", "env")
	}

	// Priority: flag hostname > NODE_NAME env var > system hostname
	if *hostname != DEFAULT_NODE_NAME {
		nodeName = *hostname
	} else if envNodeName, found := os.LookupEnv(ENV_NODE_NAME); found {
		nodeName = envNodeName
	} else {
		slog.Warn("env var not set, using system hostname", "component", "env", "env_var", ENV_NODE_NAME)
		nodeName, err = os.Hostname()
		if err != nil {
			nodeName = DEFAULT_NODE_NAME
		}
	}

	slog.Info("starting gobservability agent", "component", "env", "node", nodeName, "interval", *collectInterval, "grpc_addr", *grpcAddr)

	// Initialize streaming gRPC connection to server
	devModeValue := fmt.Sprintf("%t", *dev)
	grpcSender, err := grpcClient.NewStreamingGRPCClient(*grpcAddr, nodeName, devModeValue)
	if err != nil {
		slog.Error("failed to create streaming gRPC client", "error", err)
		os.Exit(1)
	}
	defer grpcSender.Close()

	// Initialize metrics collector with gRPC client
	metricsCollector := collector.NewCollector(ENV_DEV_MODE, grpcSender)
	metricsCollector.Start(nodeName, *collectInterval)
}
