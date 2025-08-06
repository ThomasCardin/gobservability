package types

import (
	"fmt"
	"time"
)

type CPUStats struct {
	User    int `json:"user"`
	Nice    int `json:"nice"`
	System  int `json:"system"`
	Idle    int `json:"idle"`
	IOWait  int `json:"iowait"`
	IRQ     int `json:"irq"`
	SoftIRQ int `json:"softirq"`
	Total   int `json:"total"`
}

type MemoryStats struct {
	MemTotal     int `json:"mem_total"`
	MemFree      int `json:"mem_free"`
	MemAvailable int `json:"mem_available"`
	Buffers      int `json:"buffers"`
	Cached       int `json:"cached"`
	SwapCached   int `json:"swap_cached"`
	SwapTotal    int `json:"swap_total"`
	SwapFree     int `json:"swap_free"`
}

type NetworkStats struct {
	BytesReceived      uint64 `json:"bytes_received"`
	BytesTransmitted   uint64 `json:"bytes_transmitted"`
	PacketsReceived    uint64 `json:"packets_received"`
	PacketsTransmitted uint64 `json:"packets_transmitted"`
	ErrorsReceived     uint64 `json:"errors_received"`
	ErrorsTransmitted  uint64 `json:"errors_transmitted"`
}

type DiskStats struct {
	ReadsCompleted  uint64 `json:"reads_completed"`
	ReadsMerged     uint64 `json:"reads_merged"`
	SectorsRead     uint64 `json:"sectors_read"`
	TimeReading     uint64 `json:"time_reading"`
	WritesCompleted uint64 `json:"writes_completed"`
	WritesMerged    uint64 `json:"writes_merged"`
	SectorsWritten  uint64 `json:"sectors_written"`
	TimeWriting     uint64 `json:"time_writing"`
}

type NodeMetrics struct {
	CPU     CPUStats     `json:"cpu"`
	Memory  MemoryStats  `json:"memory"`
	Network NetworkStats `json:"network"`
	Disk    DiskStats    `json:"disk"`
}

type NodeStatsPayload struct {
	NodeName  string      `json:"node_name"`
	Timestamp time.Time   `json:"timestamp"`
	Metrics   NodeMetrics `json:"metrics"`
}

func (c *CPUStats) FormatCPU() string {
	if c.Total == 0 {
		return "0%"
	}
	total := float64(c.Total)
	userTotal := float64(c.User + c.Nice)
	systemTotal := float64(c.IRQ + c.SoftIRQ)
	activePct := (userTotal + systemTotal + float64(c.System)) * 100 / total
	return fmt.Sprintf("%.1f%%", activePct)
}

func (m *MemoryStats) FormatMemory() string {
	if m.MemTotal == 0 {
		return "0%"
	}
	totalGB := float64(m.MemTotal) / 1024 / 1024
	availableGB := float64(m.MemAvailable) / 1024 / 1024
	usedGB := totalGB - availableGB
	usedPct := (usedGB / totalGB) * 100
	return fmt.Sprintf("%.1f%%", usedPct)
}

func (n *NetworkStats) FormatNetwork() string {
	totalMB := float64(n.BytesReceived+n.BytesTransmitted) / 1024 / 1024
	return fmt.Sprintf("%.1fM", totalMB)
}

func (d *DiskStats) FormatDisk() string {
	totalMB := float64(d.SectorsRead+d.SectorsWritten) * 512 / 1024 / 1024
	return fmt.Sprintf("%.1fM", totalMB)
}
