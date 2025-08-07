package compute

import (
	"fmt"

	"github.com/ThomasCardin/peek/shared/types"
)

// UINode represents a computed node for the UI with all calculated metrics
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

// UIPod represents a computed pod for the UI with all calculated metrics
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

// CalculateUINode computes all UI metrics from raw node stats
func CalculateUINode(name string, stats *types.NodeStatsPayload) UINode {
	cpu := stats.Metrics.CPU
	mem := stats.Metrics.Memory
	net := stats.Metrics.Network
	disk := stats.Metrics.Disk

	totalCPU := float64(cpu.Total)
	userTotal := float64(cpu.User + cpu.Nice)
	systemTotal := float64(cpu.IRQ + cpu.SoftIRQ)
	activePct := (userTotal + systemTotal + float64(cpu.System)) * 100 / totalCPU
	userPct := userTotal * 100 / totalCPU
	systemPct := systemTotal * 100 / totalCPU
	nicePct := float64(cpu.Nice) * 100 / totalCPU
	irqPct := float64(cpu.IRQ) * 100 / totalCPU
	sirqPct := float64(cpu.SoftIRQ) * 100 / totalCPU
	idlePct := float64(cpu.Idle) * 100 / totalCPU

	totalMem := float64(mem.MemTotal) / 1024 / 1024
	availMem := float64(mem.MemAvailable) / 1024 / 1024
	usedMem := totalMem - availMem

	netRX := float64(net.BytesReceived) / 1024 / 1024
	netTX := float64(net.BytesTransmitted) / 1024 / 1024
	netTotal := netRX + netTX

	diskRead := float64(disk.SectorsRead) * 512 / 1024 / 1024
	diskWrite := float64(disk.SectorsWritten) * 512 / 1024 / 1024
	diskTotal := diskRead + diskWrite

	return UINode{
		Name:         name,
		Timestamp:    stats.Timestamp.Format("15:04:05"),
		CPU:          cpu.FormatCPU(),
		CPUTotal:     activePct,
		CPUUser:      userPct,
		CPUSystem:    systemPct,
		CPUUserRaw:   float64(cpu.User) * 100 / totalCPU,
		CPUNiceRaw:   nicePct,
		CPUIRQRaw:    irqPct,
		CPUSIRQRaw:   sirqPct,
		CPUIdle:      idlePct,
		Memory:       mem.FormatMemory(),
		MemoryUsed:   usedMem,
		MemoryFree:   availMem,
		MemoryTotal:  totalMem,
		Network:      net.FormatNetwork(),
		NetworkTotal: netTotal,
		NetworkRX:    netRX,
		NetworkTX:    netTX,
		Disk:         disk.FormatDisk(),
		DiskTotal:    diskTotal,
		DiskRead:     diskRead,
		DiskWrite:    diskWrite,
	}
}

// CalculateUIPod computes all UI metrics for a pod from raw pod data
func CalculateUIPod(pod *types.Pod) UIPod {
	return CalculateUIPodWithNodeContext(pod, nil)
}

// CalculateUIPodWithNodeContext computes all UI metrics for a pod with node context for real calculations
func CalculateUIPodWithNodeContext(pod *types.Pod, nodeStats *types.NodeStatsPayload) UIPod {
	if pod.PID == -1 {
		// Failed pod - return basic info only
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

	// Calculate CPU metrics
	totalCPUTime := float64(pod.PodMetrics.CPU.UTime + pod.PodMetrics.CPU.STime)
	cpuPercent := pod.PodMetrics.CPU.CPUPercent

	// If CPUPercent is 0 and we have node context, calculate real percentage
	if cpuPercent == 0 && nodeStats != nil && totalCPUTime > 0 {
		// Real calculation: CPU percentage relative to total system CPU
		// Get total CPU time from node stats (all cores combined)
		nodeCPU := nodeStats.Metrics.CPU
		totalNodeCPUTime := float64(nodeCPU.User + nodeCPU.Nice + nodeCPU.System + nodeCPU.Idle + nodeCPU.IOWait + nodeCPU.IRQ + nodeCPU.SoftIRQ + nodeCPU.Steal)

		if totalNodeCPUTime > 0 {
			// Calculate process CPU as percentage of total system CPU capacity
			// This gives the real CPU percentage used by this process
			cpuPercent = (totalCPUTime / totalNodeCPUTime) * 100.0
			if cpuPercent > 100 {
				cpuPercent = 100
			}
		}
	}

	var userPercent, systemPercent float64
	if totalCPUTime > 0 {
		userPercent = float64(pod.PodMetrics.CPU.UTime) / totalCPUTime * cpuPercent
		systemPercent = float64(pod.PodMetrics.CPU.STime) / totalCPUTime * cpuPercent
	}

	// Calculate Memory metrics (convert KB to MB)
	memoryUsedMB := float64(pod.PodMetrics.Memory.VmRSS) / 1024
	memoryVirtualMB := float64(pod.PodMetrics.Memory.VmSize) / 1024
	memoryPercent := pod.PodMetrics.Memory.MemPercent

	// If MemPercent is 0 and we have node context, calculate real percentage
	if memoryPercent == 0 && nodeStats != nil && pod.PodMetrics.Memory.VmRSS > 0 {
		// Real calculation: Memory percentage relative to total system memory
		nodeMemory := nodeStats.Metrics.Memory
		totalSystemMemoryKB := float64(nodeMemory.MemTotal) // Real total memory from node

		if totalSystemMemoryKB > 0 {
			// Calculate process memory as percentage of total system memory
			memoryPercent = float64(pod.PodMetrics.Memory.VmRSS) / totalSystemMemoryKB * 100.0
			if memoryPercent > 100 {
				memoryPercent = 100
			}
		}
	}

	// Calculate Network metrics (convert bytes to MB)
	networkRX := float64(pod.PodMetrics.Network.BytesReceived) / 1024 / 1024
	networkTX := float64(pod.PodMetrics.Network.BytesTransmitted) / 1024 / 1024
	networkTotal := networkRX + networkTX

	// Calculate Disk metrics (convert bytes to MB)
	diskRead := float64(pod.PodMetrics.Disk.ReadBytes) / 1024 / 1024
	diskWrite := float64(pod.PodMetrics.Disk.WriteBytes) / 1024 / 1024
	diskTotal := diskRead + diskWrite

	return UIPod{
		Name:        pod.Name,
		ContainerID: pod.ContainerID,
		PID:         pod.PID,
		Status:      "RUNNING",

		CPU:        formatPercentage(cpuPercent),
		CPUPercent: cpuPercent,
		CPUUser:    userPercent,
		CPUSystem:  systemPercent,

		Memory:        formatPercentage(memoryPercent),
		MemoryUsed:    memoryUsedMB,
		MemoryVirtual: memoryVirtualMB,
		MemoryPercent: memoryPercent,

		Network:      formatMegabytes(networkTotal),
		NetworkTotal: networkTotal,
		NetworkRX:    networkRX,
		NetworkTX:    networkTX,

		Disk:      formatMegabytes(diskTotal),
		DiskTotal: diskTotal,
		DiskRead:  diskRead,
		DiskWrite: diskWrite,

		ProcessName: pod.PidDetails.Name,
		State:       pod.PidDetails.State,
		Threads:     pod.PidDetails.Threads,
	}
}

// Helper functions for consistent formatting with nodes
func formatPercentage(value float64) string {
	return fmt.Sprintf("%.1f%%", value)
}

func formatMegabytes(value float64) string {
	return fmt.Sprintf("%.1fM", value)
}
