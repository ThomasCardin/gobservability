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
	"github.com/ThomasCardin/peek/shared/types"
)

const (
	API_STATS           = "/api/stats"
	DEFAULT_SERVER_ADDR = "http://localhost:8080"
	DEFAULT_NODE_NAME   = "unknown (no name)"
)

var (
	serverURL       = flag.String("server", DEFAULT_SERVER_ADDR, "Server URL")
	collectInterval = flag.Duration("interval", 5*time.Second, "Collect interval")
	hostname        = flag.String("hostname", "", "Custom hostname (defaults to system hostname)")
)

func main() {
	flag.Parse()

	var nodeID string
	if *hostname != "" {
		nodeID = *hostname
	} else {
		var err error
		nodeID, err = os.Hostname()
		if err != nil {
			nodeID = DEFAULT_NODE_NAME
		}
	}

	log.Printf("gobservability node started %s scrapping at %s interval", nodeID, collectInterval)

	ticker := time.NewTicker(*collectInterval)
	defer ticker.Stop()

	for {
		cpuStats, err := internal.ProcStat()
		if err != nil {
			log.Printf("%v", err.Error())
			continue
		}

		memStats, err := internal.ProcMeminfo()
		if err != nil {
			log.Printf("%v", err.Error())
			continue
		}

		netStats, err := internal.ProcNetDev()
		if err != nil {
			log.Printf("%v", err.Error())
			continue
		}

		diskStats, err := internal.ProcDiskstats()
		if err != nil {
			log.Printf("%v", err.Error())
			continue
		}

		payload := types.NodeStatsPayload{
			NodeName:  nodeID,
			Timestamp: time.Now(),
			Metrics: types.NodeMetrics{
				CPU:     *cpuStats,
				Memory:  *memStats,
				Network: *netStats,
				Disk:    *diskStats,
			},
		}

		if err := sendStats(*serverURL, payload); err != nil {
			log.Printf("%s", err.Error())
		} else {
			log.Printf("%s sent for %s", API_STATS, nodeID)
		}

		<-ticker.C
	}
}

func sendStats(serverURL string, payload types.NodeStatsPayload) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error: JSON: %v", err)
	}

	resp, err := http.Post(fmt.Sprintf("%s%s", serverURL, API_STATS), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error: HTTP POST: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error: server returned: %d", resp.StatusCode)
	}

	return nil
}
