package collector

import (
	"fmt"
	"time"

	"github.com/ThomasCardin/peek/cmd/agent/pkg/kubernetes"
	"github.com/ThomasCardin/peek/cmd/agent/pkg/metrics"
	"github.com/ThomasCardin/peek/shared/types"
)

type Collector struct {
	nodeCollector *NodeCollector
	podCollector  *PodCollector
	k8sClient     *kubernetes.Client
	cache         *metrics.Cache
	calculator    *metrics.Calculator
	devMode       string
	grpcClient    GRPCSender
}

// GRPCSender interface pour envoyer les métriques
type GRPCSender interface {
	Send(*types.NodeStatsPayload) error
}

// NewCollector creates a new collector instance
func NewCollector(devMode string, grpcClient GRPCSender) *Collector {
	cache := metrics.NewCache()
	calculator := metrics.NewCalculator()

	return &Collector{
		nodeCollector: NewNodeCollector(cache, calculator, devMode),
		podCollector:  NewPodCollector(cache, calculator, devMode),
		k8sClient:     kubernetes.NewClient(devMode),
		cache:         cache,
		calculator:    calculator,
		devMode:       devMode,
		grpcClient:    grpcClient,
	}
}

// CollectAll collects all metrics (node + pods) with calculations
func (c *Collector) CollectAll(nodeName string) (*types.NodeStatsPayload, error) {
	nodeMetrics, err := c.nodeCollector.CollectNodeMetrics(nodeName)
	if err != nil {
		return nil, fmt.Errorf("error: failed to collect node metrics %s", err.Error())
	}

	pods, err := c.k8sClient.GetPodsForNode(nodeName)
	if err != nil {
		return nil, fmt.Errorf("error: failed to get pods %s", err.Error())
	}

	// Get total system memory for pod percentage calculations
	totalSystemMemoryKB := uint64(nodeMetrics.Memory.MemTotal)

	if err := c.podCollector.CollectAllPodMetrics(pods, totalSystemMemoryKB); err != nil {
		// Log error but continue - some pods may have succeeded
		fmt.Printf("error: failed to collect pod metrics %s\n", err.Error())
	}

	nodeMetrics.Pods = pods

	payload := &types.NodeStatsPayload{
		NodeName:  nodeName,
		Timestamp: time.Now(),
		Metrics:   *nodeMetrics,
	}

	return payload, nil
}

func (c *Collector) Start(nodeName string, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Initial collection
	c.collectAndSend(nodeName)

	// Collection loop
	for range ticker.C {
		c.collectAndSend(nodeName)
	}
}

func (c *Collector) collectAndSend(nodeName string) {
	payload, err := c.CollectAll(nodeName)
	if err != nil {
		fmt.Printf("error: %s failed to collect metrics %s\n", nodeName, err.Error())
		return
	}

	if err := c.grpcClient.Send(payload); err != nil {
		fmt.Printf("error: %s failed to send metrics via gRPC %s\n", nodeName, err.Error())
	}

	fmt.Printf("OK: %s sent metrics via gRPC\n", nodeName)
}
