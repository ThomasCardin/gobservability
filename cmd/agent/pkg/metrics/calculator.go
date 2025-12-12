package metrics

import (
	"time"

	"github.com/ThomasCardin/gobservability/shared/types"
)

type Calculator struct{}

func NewCalculator() *Calculator {
	return &Calculator{}
}

// CalculateNodeCPUPercentage calculates CPU usage percentage based on deltas
// previous *types.CPUStats is from the cache
func (c *Calculator) CalculateNodeCPUPercentage(current, previous *types.CPUStats, timeDelta time.Duration) float64 {
	if previous == nil || timeDelta == 0 {
		return 0
	}

	// Calculate total CPU time delta
	totalDelta := float64(current.Total - previous.Total)
	if totalDelta <= 0 {
		return 0
	}

	// Calculate active CPU time delta (everything except idle)
	activeDelta := float64((current.User - previous.User) +
		(current.Nice - previous.Nice) +
		(current.System - previous.System) +
		(current.IOWait - previous.IOWait) +
		(current.IRQ - previous.IRQ) +
		(current.SoftIRQ - previous.SoftIRQ) +
		(current.Steal - previous.Steal))

	return (activeDelta / totalDelta) * 100.0
}

// CalculatePodCPUPercentage calculates pod CPU usage percentage
func (c *Calculator) CalculatePodCPUPercentage(current, previous *types.PodCPUStats, timeDelta time.Duration) float64 {
	if previous == nil || timeDelta == 0 {
		return 0
	}

	// CPU time is in jiffies (typically 1/100 second)
	const jiffiesPerSecond = 100.0

	// Calculate CPU time delta in jiffies
	cpuDelta := float64((current.UTime - previous.UTime) + (current.STime - previous.STime))

	// Convert to percentage: (jiffies / jiffies_per_second) / seconds * 100
	timeInSeconds := timeDelta.Seconds()
	if timeInSeconds <= 0 {
		return 0
	}

	cpuPercent := (cpuDelta / jiffiesPerSecond) / timeInSeconds * 100.0

	// Cap at 100% (single core)
	if cpuPercent > 100 {
		cpuPercent = 100
	}

	return cpuPercent
}

// CalculateMemoryPercentage calculates memory usage percentage
func (c *Calculator) CalculateMemoryPercentage(vmRSS uint64, totalSystemMemory uint64) float64 {
	if totalSystemMemory == 0 {
		return 0
	}

	// Both values should be in the same unit (KB)
	return float64(vmRSS) / float64(totalSystemMemory) * 100.0
}

// CalculateNetworkRate calculates network throughput in MB/s
func (c *Calculator) CalculateNetworkRate(currentBytes, previousBytes uint64, timeDelta time.Duration) float64 {
	if previousBytes == 0 || timeDelta == 0 || currentBytes < previousBytes {
		return 0
	}

	// Calculate bytes per second
	byteDelta := float64(currentBytes - previousBytes)
	bytesPerSecond := byteDelta / timeDelta.Seconds()

	// Convert to MB/s
	return bytesPerSecond / (1024 * 1024)
}

// CalculateDiskRate calculates disk I/O rate in MB/s
func (c *Calculator) CalculateDiskRate(currentBytes, previousBytes uint64, timeDelta time.Duration) float64 {
	if previousBytes == 0 || timeDelta == 0 || currentBytes < previousBytes {
		return 0
	}

	// Calculate bytes per second
	byteDelta := float64(currentBytes - previousBytes)
	bytesPerSecond := byteDelta / timeDelta.Seconds()

	// Convert to MB/s
	return bytesPerSecond / (1024 * 1024)
}

// CalculateDiskRateFromSectors calculates disk I/O rate from sectors
func (c *Calculator) CalculateDiskRateFromSectors(currentSectors, previousSectors uint64, timeDelta time.Duration) float64 {
	if previousSectors == 0 || timeDelta == 0 || currentSectors < previousSectors {
		return 0
	}

	// Assume 512 bytes per sector (standard)
	const bytesPerSector = 512

	// Calculate sectors per second
	sectorDelta := float64(currentSectors - previousSectors)
	sectorsPerSecond := sectorDelta / timeDelta.Seconds()

	// Convert to MB/s
	bytesPerSecond := sectorsPerSecond * bytesPerSector
	return bytesPerSecond / (1024 * 1024)
}
