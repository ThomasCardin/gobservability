package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/ThomasCardin/peek/cmd/agent/internal"
	"github.com/ThomasCardin/peek/cmd/agent/pkg"
	"github.com/ThomasCardin/peek/shared/types"
)

const (
	HTTP_API_STATS = "/api/stats"

	DEFAULT_NODE_NAME   = "unknown"
	DEFAULT_SERVER_ADDR = "http://localhost:8080"

	ENV_NODE_NAME = "NODE_NAME"
	ENV_DEV_MODE  = "DEV_MODE"
)

var (
	serverURL       = flag.String("server", DEFAULT_SERVER_ADDR, "Server URL")
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

	log.Printf("gobservability node started %s scrapping at %s interval", nodeName, collectInterval)

	ticker := time.NewTicker(*collectInterval)
	defer ticker.Stop()

	for {
		cpuStats, err := internal.ProcStat(ENV_DEV_MODE)
		if err != nil {
			log.Printf("%v", err.Error())
			continue
		}

		memStats, err := internal.ProcMeminfo(ENV_DEV_MODE)
		if err != nil {
			log.Printf("%v", err.Error())
			continue
		}

		netStats, err := internal.ProcNetDev(ENV_DEV_MODE)
		if err != nil {
			log.Printf("%v", err.Error())
			continue
		}

		diskStats, err := internal.ProcDiskstats(ENV_DEV_MODE)
		if err != nil {
			log.Printf("%v", err.Error())
			continue
		}

		pods, err := pkg.GetPodsPID(ENV_DEV_MODE, nodeName)
		if err != nil {
			log.Printf("%v", err.Error())
		}

		payload := types.NodeStatsPayload{
			NodeName:  nodeName,
			Timestamp: time.Now(),
			Metrics: types.NodeMetrics{
				CPU:     cpuStats,
				Memory:  memStats,
				Network: netStats,
				Disk:    diskStats,
				Pods:    pods,
			},
		}

		if err := sendStats(*serverURL, payload); err != nil {
			log.Printf("%s", err.Error())
		} else {
			log.Printf("%s sent for %s", HTTP_API_STATS, nodeName)
		}

		<-ticker.C
	}
}

func sendStats(serverURL string, payload types.NodeStatsPayload) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error: JSON: %v", err)
	}

	resp, err := http.Post(fmt.Sprintf("%s%s", serverURL, HTTP_API_STATS), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error: HTTP POST: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error: server returned: %d", resp.StatusCode)
	}

	return nil
}
