package formatter

import (
	"fmt"

	"github.com/ThomasCardin/peek/shared/types"
)

// UINode represents a formatted node for the UI display
type UINode struct {
	Name         string  `json:"name"`
	Timestamp    string  `json:"timestamp"`
	CPU          string  `json:"cpu"`
	CPUTotal     float64 `json:"cpu_total"`
	CPUUser      float64 `json:"cpu_user"`
	CPUSystem    float64 `json:"cpu_system"`
	CPUUserRaw   float64 `json:"cpu_user_raw"`
	CPUNiceRaw   float64 `json:"cpu_nice_raw"`
	CPUIRQRaw    float64 `json:"cpu_irq_raw"`
	CPUSIRQRaw   float64 `json:"cpu_sirq_raw"`
	CPUIdle      float64 `json:"cpu_idle"`
	Memory       string  `json:"memory"`
	MemoryUsed   float64 `json:"memory_used"`
	MemoryFree   float64 `json:"memory_free"`
	MemoryTotal  float64 `json:"memory_total"`
	Network      string  `json:"network"`
	NetworkTotal float64 `json:"network_total"`
	NetworkRX    float64 `json:"network_rx"`
	NetworkTX    float64 `json:"network_tx"`
	Disk         string  `json:"disk"`
	DiskTotal    float64 `json:"disk_total"`
	DiskRead     float64 `json:"disk_read"`
	DiskWrite    float64 `json:"disk_write"`
}

// UIPod represents a formatted pod for the UI display
type UIPod struct {
	Name        string `json:"name"`
	ContainerID string `json:"container_id"`
	PID         int    `json:"pid"`
	Status      string `json:"status"` // RUNNING or ERROR based on PID

	// CPU metrics (same format as nodes)
	CPU        string  `json:"cpu"`         // Formatted CPU percentage
	CPUPercent float64 `json:"cpu_percent"` // Raw CPU percentage
	CPUUser    float64 `json:"cpu_user"`    // User time percentage
	CPUSystem  float64 `json:"cpu_system"`  // System time percentage

	// Memory metrics (same format as nodes)
	Memory        string  `json:"memory"`         // Formatted memory percentage
	MemoryUsed    float64 `json:"memory_used"`    // Used memory in MB
	MemoryVirtual float64 `json:"memory_virtual"` // Virtual memory in MB
	MemoryPercent float64 `json:"memory_percent"` // Percentage of node memory

	// Network metrics (same format as nodes)
	Network      string  `json:"network"`       // Formatted network total
	NetworkTotal float64 `json:"network_total"` // Total network in MB
	NetworkRX    float64 `json:"network_rx"`    // Received in MB
	NetworkTX    float64 `json:"network_tx"`    // Transmitted in MB

	// Disk metrics (same format as nodes)
	Disk      string  `json:"disk"`       // Formatted disk total
	DiskTotal float64 `json:"disk_total"` // Total disk I/O in MB
	DiskRead  float64 `json:"disk_read"`  // Read in MB
	DiskWrite float64 `json:"disk_write"` // Write in MB

	// Process details
	ProcessName string `json:"process_name"` // Name of the process
	State       string `json:"state"`        // Process state
	Threads     int    `json:"threads"`      // Number of threads
}

// FormatNodeForUI formats raw node stats for UI display
func FormatNodeForUI(name string, stats *types.NodeStatsPayload) UINode {
	cpu := stats.Metrics.CPU
	mem := stats.Metrics.Memory
	net := stats.Metrics.Network
	disk := stats.Metrics.Disk

	// Calculate component percentages for detailed view (only for display breakdown)
	totalCPU := float64(cpu.Total)

	return UINode{
		Name:         name,
		Timestamp:    stats.Timestamp.Format("15:04:05"),
		CPU:          cpu.FormatCPU(),
		CPUTotal:     cpu.CPUPercent, // From agent calculation
		CPUUser:      float64(cpu.User+cpu.Nice) * 100 / totalCPU,
		CPUSystem:    float64(cpu.IRQ+cpu.SoftIRQ) * 100 / totalCPU,
		CPUUserRaw:   float64(cpu.User) * 100 / totalCPU,
		CPUNiceRaw:   float64(cpu.Nice) * 100 / totalCPU,
		CPUIRQRaw:    float64(cpu.IRQ) * 100 / totalCPU,
		CPUSIRQRaw:   float64(cpu.SoftIRQ) * 100 / totalCPU,
		CPUIdle:      float64(cpu.Idle) * 100 / totalCPU,
		Memory:       mem.FormatMemory(),
		MemoryUsed:   float64(mem.MemTotal-mem.MemAvailable) / 1024 / 1024,
		MemoryFree:   float64(mem.MemAvailable) / 1024 / 1024,
		MemoryTotal:  float64(mem.MemTotal) / 1024 / 1024,
		Network:      net.FormatNetwork(),
		NetworkTotal: net.TotalRate, // From agent calculation
		NetworkRX:    net.RxRate,    // From agent calculation
		NetworkTX:    net.TxRate,    // From agent calculation
		Disk:         disk.FormatDisk(),
		DiskTotal:    disk.TotalRate, // From agent calculation
		DiskRead:     disk.ReadRate,  // From agent calculation
		DiskWrite:    disk.WriteRate, // From agent calculation
	}
}

// FormatPodForUI formats pod data for UI display (simplified - agent already calculated percentages)
func FormatPodForUI(pod *types.Pod) UIPod {
	if pod.PID == -1 {
		// Failed pod
		return UIPod{
			Name:        pod.Name,
			ContainerID: pod.ContainerID,
			PID:         pod.PID,
			Status:      "ERROR",
			CPU:         "0%",
			Memory:      "0%",
			Network:     "0M",
			Disk:        "0M",
		}
	}

	// Calculate user/system split from total CPU percentage
	totalCPUTime := float64(pod.PodMetrics.CPU.UTime + pod.PodMetrics.CPU.STime)
	var userPercent, systemPercent float64
	if totalCPUTime > 0 {
		userPercent = float64(pod.PodMetrics.CPU.UTime) / totalCPUTime * pod.PodMetrics.CPU.CPUPercent
		systemPercent = float64(pod.PodMetrics.CPU.STime) / totalCPUTime * pod.PodMetrics.CPU.CPUPercent
	}

	return UIPod{
		Name:        pod.Name,
		ContainerID: pod.ContainerID,
		PID:         pod.PID,
		Status:      "RUNNING",

		CPU:        formatPercentage(pod.PodMetrics.CPU.CPUPercent),
		CPUPercent: pod.PodMetrics.CPU.CPUPercent, // From agent calculation
		CPUUser:    userPercent,
		CPUSystem:  systemPercent,

		Memory:        formatPercentage(pod.PodMetrics.Memory.MemPercent),
		MemoryUsed:    float64(pod.PodMetrics.Memory.VmRSS) / 1024,
		MemoryVirtual: float64(pod.PodMetrics.Memory.VmSize) / 1024,
		MemoryPercent: pod.PodMetrics.Memory.MemPercent, // From agent calculation

		Network:      formatMegabytes(float64(pod.PodMetrics.Network.BytesReceived+pod.PodMetrics.Network.BytesTransmitted) / 1024 / 1024),
		NetworkTotal: float64(pod.PodMetrics.Network.BytesReceived+pod.PodMetrics.Network.BytesTransmitted) / 1024 / 1024,
		NetworkRX:    float64(pod.PodMetrics.Network.BytesReceived) / 1024 / 1024,
		NetworkTX:    float64(pod.PodMetrics.Network.BytesTransmitted) / 1024 / 1024,

		Disk:      formatMegabytes(float64(pod.PodMetrics.Disk.ReadBytes+pod.PodMetrics.Disk.WriteBytes) / 1024 / 1024),
		DiskTotal: float64(pod.PodMetrics.Disk.ReadBytes+pod.PodMetrics.Disk.WriteBytes) / 1024 / 1024,
		DiskRead:  float64(pod.PodMetrics.Disk.ReadBytes) / 1024 / 1024,
		DiskWrite: float64(pod.PodMetrics.Disk.WriteBytes) / 1024 / 1024,

		ProcessName: pod.PidDetails.Name,
		State:       pod.PidDetails.State,
		Threads:     pod.PidDetails.Threads,
	}
}

// Helper functions for consistent formatting
func formatPercentage(value float64) string {
	return fmt.Sprintf("%.1f%%", value)
}

func formatMegabytes(value float64) string {
	return fmt.Sprintf("%.1fM", value)
}
