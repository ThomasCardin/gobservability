package api

import (
	"fmt"
	"log"
	"net/http"
	"sort"

	"github.com/ThomasCardin/peek/pkg/kubernetes"
	"github.com/gin-gonic/gin"
)

type ResourceStats struct {
	User   int `json:"user"`
	Kernel int `json:"kernel"`
	Total  int `json:"total"`
}

type NetworkStats struct {
	In    int `json:"in"`
	Out   int `json:"out"`
	Total int `json:"total"`
}

type DiskStats struct {
	Read  int `json:"read"`
	Write int `json:"write"`
	Total int `json:"total"`
}

type Stats struct {
	CPU     ResourceStats `json:"cpu"`
	Memory  ResourceStats `json:"memory"`
	Network NetworkStats  `json:"network"`
	Disk    DiskStats     `json:"disk"`
}

type Pod struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Stats     *Stats `json:"stats,omitempty"`
}

type Node struct {
	Name  string `json:"name"`
	Pods  []Pod  `json:"pods"`
	Stats *Stats `json:"stats,omitempty"`
}

type NodesResponse map[string]Node

func getNodesAndPods() (NodesResponse, error) {
	client, err := kubernetes.Client()
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %v", err)
	}

	nodes, err := kubernetes.GetNodes(client)
	if err != nil {
		return nil, fmt.Errorf("failed to get nodes: %v", err)
	}

	response := make(NodesResponse)

	for _, n := range nodes.Items {
		pods, err := kubernetes.GetPods(client, n.Name)
		if err != nil {
			log.Printf("Error getting pods for node %s: %v", n.Name, err)
			continue
		}

		var podList []Pod
		for _, pod := range pods.Items {
			podList = append(podList, Pod{
				Name:      pod.Name,
				Namespace: pod.Namespace,
			})
		}

		response[n.Name] = Node{
			Name: n.Name,
			Pods: podList,
		}
	}

	return response, nil
}

func NodesHandler(c *gin.Context) {
	data, err := getNodesAndPods()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error getting nodes and pods: %v", err)})
		return
	}

	c.JSON(http.StatusOK, data)
}

func NodesTemplateHandler(c *gin.Context) {
	data, err := getNodesAndPods()
	if err != nil {
		c.String(http.StatusInternalServerError, "Error loading nodes")
		return
	}

	var nodes []Node
	for _, node := range data {
		nodes = append(nodes, node)
	}

	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Name < nodes[j].Name
	})

	c.HTML(http.StatusOK, "nodes.html", nodes)
}

func NodeStatsHandler(c *gin.Context) {
	nodeName := c.Param("nodeName")
	_ = nodeName

	c.HTML(http.StatusOK, "node-stats.html", nil)
}

func IndexHandler(c *gin.Context) {
	data, err := getNodesAndPods()
	if err != nil {
		c.String(http.StatusInternalServerError, "Error loading nodes")
		return
	}

	var nodes []Node
	for _, node := range data {
		nodes = append(nodes, node)
	}

	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Name < nodes[j].Name
	})

	c.HTML(http.StatusOK, "index.html", nodes)
}