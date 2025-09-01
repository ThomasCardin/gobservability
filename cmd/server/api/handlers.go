package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/ThomasCardin/gobservability/cmd/server/formatter"
	grpcServer "github.com/ThomasCardin/gobservability/cmd/server/grpc"
	"github.com/ThomasCardin/gobservability/cmd/server/storage"
	pb "github.com/ThomasCardin/gobservability/proto"
	"github.com/ThomasCardin/gobservability/shared/types"
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
			"Pods":     formatPodsForUI(nodeStats.Metrics.Pods),
		})
		return
	}

	// Calculate UIPods with metrics using real node context
	uiPods := formatPodsForUI(nodeStats.Metrics.Pods)

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
	var uiPods []formatter.UIPod
	if found {
		uiPods = formatPodsForUI(nodeStats.Metrics.Pods)
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

	// Format pod for UI display
	uiPod := formatter.FormatPodForUI(targetPod)

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

	// Format pod for UI display
	uiPod := formatter.FormatPodForUI(targetPod)

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
		"PodName":            podName,
		"PID":                targetPod.PID,
		"ProcessDetails":     &targetPod.PidDetails,
		"ResourceLimits":     &targetPod.ResourceLimits,
		"ResourceRequests":   &targetPod.ResourceRequests,
	})
}

// GenerateFlamegraphHandler starts flamegraph generation and returns a task ID
func GenerateFlamegraphHandler(c *gin.Context) {
	nodeName := c.Param("nodename")
	podName := c.Param("podname")

	log.Printf("Received flamegraph HTTP request for node: %s, pod: %s", nodeName, podName)

	if nodeName == "" || podName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Node name and pod name required"})
		return
	}

	// Get optional query parameters
	duration, err := strconv.Atoi(c.DefaultQuery("duration", "60"))
	if err != nil || duration < 30 || duration > 600 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid duration (30-600 seconds)"})
		return
	}

	format := "json"

	// Verify the pod exists
	nodeStats, found := storage.GlobalStore.GetNodeStats(nodeName)
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "Node not found"})
		return
	}

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

	if targetPod.PID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Pod has no valid PID"})
		return
	}

	// Get the gRPC server instance
	grpcServer := grpcServer.GetServerInstance()
	if grpcServer == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gRPC server not initialized"})
		return
	}

	// Generate a unique task ID
	taskID := fmt.Sprintf("%s-%s-%d", nodeName, podName, time.Now().UnixNano())

	// Create task in pending state
	storage.GlobalStore.CreateFlamegraphTask(taskID, nodeName, podName, format)

	// Start flamegraph generation asynchronously
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()

		// Create the flamegraph request
		req := &pb.FlamegraphRequest{
			NodeName: nodeName,
			PodName:  podName,
			Duration: int32(duration),
			Format:   format,
		}

		// Call the gRPC method
		resp, err := grpcServer.GenerateFlamegraph(ctx, req)

		// Store the result (you'll need to implement a storage mechanism)
		storage.GlobalStore.StoreFlamegraphResult(taskID, resp, err, format, nodeName, podName)
	}()

	// Return task ID immediately
	c.JSON(http.StatusAccepted, gin.H{
		"task_id": taskID,
		"status":  "started",
		"message": "Flamegraph generation started. Use task_id to check status.",
	})
}

// FlamegraphStatusHandler checks the status of a flamegraph generation task
func FlamegraphStatusHandler(c *gin.Context) {
	taskID := c.Param("taskid")

	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Task ID required"})
		return
	}

	// Check task status (you'll need to implement this)
	result := storage.GlobalStore.GetFlamegraphResult(taskID)
	if result == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	if !result.Completed {
		c.JSON(http.StatusOK, gin.H{
			"status":  "processing",
			"message": "Flamegraph is still being generated",
		})
		return
	}

	if result.Error != "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  result.Error,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "completed",
		"ready":  true,
		"size":   len(result.Data),
	})
}

// DownloadFlamegraphHandler downloads the completed flamegraph
func DownloadFlamegraphHandler(c *gin.Context) {
	taskID := c.Param("taskid")

	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Task ID required"})
		return
	}

	// Get completed task result
	result := storage.GlobalStore.GetFlamegraphResult(taskID)
	if result == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	if !result.Completed {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Task not completed yet"})
		return
	}

	if result.Error != "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error})
		return
	}

	c.Header("Content-Type", "application/json")

	c.Data(http.StatusOK, "", result.Data)
}

// FlamegraphPageHandler renders the dedicated flamegraph page
func FlamegraphPageHandler(c *gin.Context) {
	nodeName := c.Param("nodename")
	podName := c.Param("podname")
	taskID := c.Query("task_id")

	if nodeName == "" || podName == "" {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "Node and pod name required"})
		return
	}

	c.HTML(http.StatusOK, "flamegraph-simple.html", gin.H{
		"NodeName": nodeName,
		"PodName":  podName,
		"TaskID":   taskID,
	})
}
