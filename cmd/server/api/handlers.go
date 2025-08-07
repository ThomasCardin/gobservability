package api

import (
	"log"
	"net/http"
	"sort"

	"github.com/ThomasCardin/peek/cmd/server/storage"
	"github.com/ThomasCardin/peek/shared/types"
	"github.com/gin-gonic/gin"
)

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

func ReceiveStatsHandler(c *gin.Context) {
	var payload types.NodeStatsPayload

	if err := c.ShouldBindJSON(&payload); err != nil {
		log.Printf("Erreur parsing JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	storage.GlobalStore.StoreNodeStats(payload)
	c.JSON(http.StatusOK, gin.H{"status": "received"})
}

func getUINodes() []UINode {
	nodes := storage.GlobalStore.GetAllNodes()
	var uiNodes []UINode

	for name, stats := range nodes {
		cpu := stats.Metrics.CPU
		mem := stats.Metrics.Memory
		net := stats.Metrics.Network
		disk := stats.Metrics.Disk

		totalCPU := float64(cpu.Total)
		userPct := float64(cpu.User+cpu.Nice) * 100 / totalCPU
		systemPct := float64(cpu.IRQ+cpu.SoftIRQ) * 100 / totalCPU
		activePct := (userPct + systemPct + float64(cpu.System)*100/totalCPU)
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

		uiNodes = append(uiNodes, UINode{
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
		})
	}

	sort.Slice(uiNodes, func(i, j int) bool {
		return uiNodes[i].Name < uiNodes[j].Name
	})

	return uiNodes
}

func IndexHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", nil)
}

func NodesFragmentHandler(c *gin.Context) {
	nodes := getUINodes()
	c.HTML(http.StatusOK, "nodes-fragment.html", nodes)
}
