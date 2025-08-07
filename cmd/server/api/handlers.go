package api

import (
	"net/http"

	"github.com/ThomasCardin/peek/cmd/server/compute"
	"github.com/ThomasCardin/peek/cmd/server/storage"
	"github.com/ThomasCardin/peek/shared/types"
	"github.com/gin-gonic/gin"
)

func ReceiveStatsHandler(c *gin.Context) {
	var payload types.NodeStatsPayload

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	storage.GlobalStore.StoreNodeStats(payload)
	c.JSON(http.StatusOK, gin.H{"status": "received"})
}

func IndexHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", nil)
}

func NodesFragmentHandler(c *gin.Context) {
	nodes := getUINodes()
	c.HTML(http.StatusOK, "nodes-fragment.html", nodes)
}

func PodsHandler(c *gin.Context) {
	nodeName := c.Param("nodename")
	if nodeName == "" {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "Node name required"})
		return
	}

	// Get node data for pods
	nodeStats, found := storage.GlobalStore.GetNodeStats(nodeName)
	if !found {
		c.HTML(http.StatusOK, "pods.html", gin.H{
			"NodeName": nodeName,
			"Pods":     []*types.Pod{},
		})
		return
	}

	// Get UINode from cache (consistent with dashboard)
	uiNode, uiFound := getUINode(nodeName)
	if !uiFound {
		c.HTML(http.StatusOK, "pods.html", gin.H{
			"NodeName": nodeName,
			"Pods":     calculateUIPods(nodeStats.Metrics.Pods),
		})
		return
	}

	// Calculate UIPods with metrics using real node context
	uiPods := calculateUIPodsWithNodeContext(nodeStats.Metrics.Pods, nodeStats)

	c.HTML(http.StatusOK, "pods.html", gin.H{
		"NodeName":     uiNode.Name,
		"Timestamp":    uiNode.Timestamp,
		"Pods":         uiPods,
		"CPU":          uiNode.CPU,
		"CPUTotal":     uiNode.CPUTotal,
		"CPUUser":      uiNode.CPUUser,
		"CPUSystem":    uiNode.CPUSystem,
		"CPUUserRaw":   uiNode.CPUUserRaw,
		"CPUNiceRaw":   uiNode.CPUNiceRaw,
		"CPUIRQRaw":    uiNode.CPUIRQRaw,
		"CPUSIRQRaw":   uiNode.CPUSIRQRaw,
		"CPUIdle":      uiNode.CPUIdle,
		"Memory":       uiNode.Memory,
		"MemoryUsed":   uiNode.MemoryUsed,
		"MemoryFree":   uiNode.MemoryFree,
		"MemoryTotal":  uiNode.MemoryTotal,
		"Network":      uiNode.Network,
		"NetworkTotal": uiNode.NetworkTotal,
		"NetworkRX":    uiNode.NetworkRX,
		"NetworkTX":    uiNode.NetworkTX,
		"Disk":         uiNode.Disk,
		"DiskTotal":    uiNode.DiskTotal,
		"DiskRead":     uiNode.DiskRead,
		"DiskWrite":    uiNode.DiskWrite,
	})
}

func PodsFragmentHandler(c *gin.Context) {
	nodeName := c.Param("nodename")
	if nodeName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Node name required"})
		return
	}

	// Get node data for pods
	nodeStats, found := storage.GlobalStore.GetNodeStats(nodeName)
	var uiPods []compute.UIPod
	if found {
		uiPods = calculateUIPodsWithNodeContext(nodeStats.Metrics.Pods, nodeStats)
	}

	c.HTML(http.StatusOK, "pods-fragment.html", gin.H{
		"NodeName": nodeName,
		"Pods":     uiPods,
	})
}

func PodsMetricsFragmentHandler(c *gin.Context) {
	nodeName := c.Param("nodename")
	if nodeName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Node name required"})
		return
	}

	// Get UINode from cache for consistent metrics
	uiNode, found := getUINode(nodeName)
	if !found {
		c.HTML(http.StatusOK, "pods-metrics-fragment.html", gin.H{
			"NodeName": nodeName,
		})
		return
	}

	c.HTML(http.StatusOK, "pods-metrics-fragment.html", gin.H{
		"NodeName":     uiNode.Name,
		"Timestamp":    uiNode.Timestamp,
		"CPU":          uiNode.CPU,
		"CPUTotal":     uiNode.CPUTotal,
		"CPUUser":      uiNode.CPUUser,
		"CPUSystem":    uiNode.CPUSystem,
		"CPUUserRaw":   uiNode.CPUUserRaw,
		"CPUNiceRaw":   uiNode.CPUNiceRaw,
		"CPUIRQRaw":    uiNode.CPUIRQRaw,
		"CPUSIRQRaw":   uiNode.CPUSIRQRaw,
		"CPUIdle":      uiNode.CPUIdle,
		"Memory":       uiNode.Memory,
		"MemoryUsed":   uiNode.MemoryUsed,
		"MemoryFree":   uiNode.MemoryFree,
		"MemoryTotal":  uiNode.MemoryTotal,
		"Network":      uiNode.Network,
		"NetworkTotal": uiNode.NetworkTotal,
		"NetworkRX":    uiNode.NetworkRX,
		"NetworkTX":    uiNode.NetworkTX,
		"Disk":         uiNode.Disk,
		"DiskTotal":    uiNode.DiskTotal,
		"DiskRead":     uiNode.DiskRead,
		"DiskWrite":    uiNode.DiskWrite,
	})
}

// PodProcessDetailsHandler returns the PidDetails for a specific pod
func PodProcessDetailsHandler(c *gin.Context) {
	nodeName := c.Param("nodename")
	podName := c.Param("podname")

	if nodeName == "" || podName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Node name and pod name required"})
		return
	}

	// Get node data
	nodeStats, found := storage.GlobalStore.GetNodeStats(nodeName)
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "Node not found"})
		return
	}

	// Find the specific pod
	var targetPod *types.Pod
	for _, pod := range nodeStats.Metrics.Pods {
		if pod.Name == podName {
			targetPod = pod
			break
		}
	}

	if targetPod == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pod not found"})
		return
	}

	if targetPod.PID == -1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Pod has no valid PID"})
		return
	}

	// Return the PidDetails directly
	c.JSON(http.StatusOK, targetPod.PidDetails)
}

// ProcessDetailsPageHandler returns the complete process details page
func ProcessDetailsPageHandler(c *gin.Context) {
	nodeName := c.Param("nodename")
	podName := c.Param("podname")

	if nodeName == "" || podName == "" {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "Node name and pod name required"})
		return
	}

	// Get node data
	nodeStats, found := storage.GlobalStore.GetNodeStats(nodeName)
	if !found {
		c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "Node not found"})
		return
	}

	// Find the specific pod
	var targetPod *types.Pod
	for _, pod := range nodeStats.Metrics.Pods {
		if pod.Name == podName {
			targetPod = pod
			break
		}
	}

	if targetPod == nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "Pod not found"})
		return
	}

	if targetPod.PID == -1 {
		c.HTML(http.StatusOK, "process-details.html", gin.H{
			"NodeName":       nodeName,
			"PodName":        podName,
			"Pod":            nil,
			"ProcessDetails": nil,
		})
		return
	}

	// Get UINode from cache for node metrics
	uiNode, uiFound := getUINode(nodeName)

	// Calculate UIPod for this specific pod with node context
	uiPod := compute.CalculateUIPodWithNodeContext(targetPod, nodeStats)

	if !uiFound {
		c.HTML(http.StatusOK, "process-details.html", gin.H{
			"NodeName":       nodeName,
			"PodName":        podName,
			"Pod":            &uiPod,
			"ProcessDetails": &targetPod.PidDetails,
		})
		return
	}

	c.HTML(http.StatusOK, "process-details.html", gin.H{
		"NodeName":       nodeName,
		"PodName":        podName,
		"Pod":            &uiPod,
		"ProcessDetails": &targetPod.PidDetails,
		// Node metrics
		"CPU":          uiNode.CPU,
		"CPUTotal":     uiNode.CPUTotal,
		"CPUUser":      uiNode.CPUUser,
		"CPUSystem":    uiNode.CPUSystem,
		"CPUUserRaw":   uiNode.CPUUserRaw,
		"CPUNiceRaw":   uiNode.CPUNiceRaw,
		"CPUIRQRaw":    uiNode.CPUIRQRaw,
		"CPUSIRQRaw":   uiNode.CPUSIRQRaw,
		"CPUIdle":      uiNode.CPUIdle,
		"Memory":       uiNode.Memory,
		"MemoryUsed":   uiNode.MemoryUsed,
		"MemoryFree":   uiNode.MemoryFree,
		"MemoryTotal":  uiNode.MemoryTotal,
		"Network":      uiNode.Network,
		"NetworkTotal": uiNode.NetworkTotal,
		"NetworkRX":    uiNode.NetworkRX,
		"NetworkTX":    uiNode.NetworkTX,
		"Disk":         uiNode.Disk,
		"DiskTotal":    uiNode.DiskTotal,
		"DiskRead":     uiNode.DiskRead,
		"DiskWrite":    uiNode.DiskWrite,
		"Timestamp":    uiNode.Timestamp,
	})
}

// PodInfoHandler returns just the pod information for HTMX updates
func PodInfoHandler(c *gin.Context) {
	nodeName := c.Param("nodename")
	podName := c.Param("podname")

	if nodeName == "" || podName == "" {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "Node name and pod name required"})
		return
	}

	// Get node data
	nodeStats, found := storage.GlobalStore.GetNodeStats(nodeName)
	if !found {
		c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "Node not found"})
		return
	}

	// Find the specific pod
	var targetPod *types.Pod
	for _, pod := range nodeStats.Metrics.Pods {
		if pod.Name == podName {
			targetPod = pod
			break
		}
	}

	if targetPod == nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "Pod not found"})
		return
	}

	// Calculate UIPod for this specific pod with node context
	uiPod := compute.CalculateUIPodWithNodeContext(targetPod, nodeStats)

	c.HTML(http.StatusOK, "pod-info-fragment.html", gin.H{
		"Pod": &uiPod,
	})
}

// ProcessDetailsFragmentHandler returns just the process details for HTMX updates
func ProcessDetailsFragmentHandler(c *gin.Context) {
	nodeName := c.Param("nodename")
	podName := c.Param("podname")

	if nodeName == "" || podName == "" {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "Node name and pod name required"})
		return
	}

	// Get node data
	nodeStats, found := storage.GlobalStore.GetNodeStats(nodeName)
	if !found {
		c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "Node not found"})
		return
	}

	// Find the specific pod
	var targetPod *types.Pod
	for _, pod := range nodeStats.Metrics.Pods {
		if pod.Name == podName {
			targetPod = pod
			break
		}
	}

	if targetPod == nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "Pod not found"})
		return
	}

	if targetPod.PID == -1 {
		c.HTML(http.StatusOK, "process-details-fragment.html", gin.H{
			"PodName":        podName,
			"ProcessDetails": nil,
		})
		return
	}

	c.HTML(http.StatusOK, "process-details-fragment.html", gin.H{
		"PodName":        podName,
		"ProcessDetails": &targetPod.PidDetails,
	})
}
