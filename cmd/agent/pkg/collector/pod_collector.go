package collector

import (
	"fmt"
	"os"
	"time"

	"github.com/ThomasCardin/peek/cmd/agent/internal"
	"github.com/ThomasCardin/peek/cmd/agent/pkg/metrics"
	"github.com/ThomasCardin/peek/shared/types"
)

type PodCollector struct {
	cache      *metrics.Cache
	calculator *metrics.Calculator
	devMode    string
}

func NewPodCollector(cache *metrics.Cache, calculator *metrics.Calculator, devMode string) *PodCollector {
	return &PodCollector{
		cache:      cache,
		calculator: calculator,
		devMode:    devMode,
	}
}

// CollectPodMetrics collects metrics for a single pod and calculates percentages
func (pc *PodCollector) CollectPodMetrics(pod *types.Pod, totalSystemMemoryKB uint64) error {
	if pod.PID <= 0 {
		return fmt.Errorf("error: invalid PID: %d", pod.PID)
	}

	if isDev := os.Getenv(pc.devMode); isDev == "true" {
		return nil
	}

	// Collect raw metrics
	podMetrics, pidDetails, err := internal.CollectPodMetrics(pc.devMode, pod.PID)
	if err != nil {
		return fmt.Errorf("error: failed to collect pod metrics: %v", err)
	}

	// Get previous metrics from cache
	prev, hasPrev := pc.cache.UpdatePodMetrics(
		pod.PID,
		&podMetrics.CPU,
		&podMetrics.Network,
		&podMetrics.Disk,
	)

	// Calculate CPU percentage if we have previous data
	if hasPrev && prev != nil {
		timeDelta := time.Since(prev.Timestamp)

		// Calculate CPU percentage
		cpuPercent := pc.calculator.CalculatePodCPUPercentage(&podMetrics.CPU, prev.CPU, timeDelta)
		podMetrics.CPU.CPUPercent = cpuPercent

		// Calculate memory percentage
		memPercent := pc.calculator.CalculateMemoryPercentage(podMetrics.Memory.VmRSS, totalSystemMemoryKB)
		podMetrics.Memory.MemPercent = memPercent

		// Note: Network and disk rates could be calculated here too if needed
		// but they're typically shown as absolute values for pods
	} else {
		// First collection - set percentages to 0
		podMetrics.CPU.CPUPercent = 0
		podMetrics.Memory.MemPercent = pc.calculator.CalculateMemoryPercentage(podMetrics.Memory.VmRSS, totalSystemMemoryKB)
	}

	// Update pod with calculated metrics
	pod.PodMetrics = *podMetrics
	pod.PidDetails = *pidDetails

	return nil
}

func (pc *PodCollector) CollectAllPodMetrics(pods []*types.Pod, totalSystemMemoryKB uint64) error {
	activePIDs := make([]int, 0, len(pods))

	for _, pod := range pods {
		if pod.PID > 0 {
			activePIDs = append(activePIDs, pod.PID)

			if err := pc.CollectPodMetrics(pod, totalSystemMemoryKB); err != nil {
				fmt.Printf("error: failed to collect metrics for pod %s (PID %d) %s\n", pod.Name, pod.PID, err.Error())
				pod.PID = -1 // Mark as failed
			}
		}
	}

	// Clean up stale cache entries
	pc.cache.CleanupStaleEntries(activePIDs)

	return nil
}
