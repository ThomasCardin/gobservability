package collector

import (
	"fmt"
	"time"

	"github.com/ThomasCardin/gobservability/cmd/agent/internal"
	"github.com/ThomasCardin/gobservability/cmd/agent/pkg/metrics"
	"github.com/ThomasCardin/gobservability/shared/types"
)

type NodeCollector struct {
	cache      *metrics.Cache
	calculator *metrics.Calculator
	devMode    string
}

func NewNodeCollector(cache *metrics.Cache, calculator *metrics.Calculator, devMode string) *NodeCollector {
	return &NodeCollector{
		cache:      cache,
		calculator: calculator,
		devMode:    devMode,
	}
}

func (nc *NodeCollector) CollectNodeMetrics(nodeName string) (*types.NodeMetrics, error) {
	cpu, err := internal.ProcStat(nc.devMode)
	if err != nil {
		return nil, fmt.Errorf("error: failed to read CPU stats: %v", err)
	}

	memory, err := internal.ProcMeminfo(nc.devMode)
	if err != nil {
		return nil, fmt.Errorf("error: failed to read memory stats: %v", err)
	}

	network, err := internal.ProcNetDev(nc.devMode)
	if err != nil {
		return nil, fmt.Errorf("error: failed to read network stats: %v", err)
	}

	disk, err := internal.ProcDiskstats(nc.devMode)
	if err != nil {
		return nil, fmt.Errorf("error: failed to read disk stats: %v", err)
	}

	// Get previous metrics from cache
	prev, hasPrev := nc.cache.UpdateNodeMetrics(nodeName, cpu, network, disk)

	// Calculate rates and percentages if we have previous data
	if hasPrev && prev != nil {
		timeDelta := time.Since(prev.Timestamp)

		cpu.CPUPercent = nc.calculator.CalculateNodeCPUPercentage(cpu, prev.CPU, timeDelta)

		network.RxRate = nc.calculator.CalculateNetworkRate(network.BytesReceived, prev.Network.BytesReceived, timeDelta)
		network.TxRate = nc.calculator.CalculateNetworkRate(network.BytesTransmitted, prev.Network.BytesTransmitted, timeDelta)
		network.TotalRate = network.RxRate + network.TxRate

		disk.ReadRate = nc.calculator.CalculateDiskRateFromSectors(disk.SectorsRead, prev.Disk.SectorsRead, timeDelta)
		disk.WriteRate = nc.calculator.CalculateDiskRateFromSectors(disk.SectorsWritten, prev.Disk.SectorsWritten, timeDelta)
		disk.TotalRate = disk.ReadRate + disk.WriteRate

		memory.MemoryPercent = float64(memory.MemTotal-memory.MemAvailable) / float64(memory.MemTotal) * 100.0
	} else {
		// First collection - set calculated values to 0
		cpu.CPUPercent = 0
		network.RxRate = 0
		network.TxRate = 0
		network.TotalRate = 0
		disk.ReadRate = 0
		disk.WriteRate = 0
		disk.TotalRate = 0
		memory.MemoryPercent = float64(memory.MemTotal-memory.MemAvailable) / float64(memory.MemTotal) * 100.0
	}

	return &types.NodeMetrics{
		CPU:     cpu,
		Memory:  memory,
		Network: network,
		Disk:    disk,
		Pods:    nil, // Pods will be collected separately
	}, nil
}
