package api

import (
	"sort"

	"github.com/ThomasCardin/peek/cmd/server/compute"
	"github.com/ThomasCardin/peek/cmd/server/storage"
)

func getUINodes() []compute.UINode {
	nodes := storage.GlobalStore.GetAllNodes()
	var uiNodes []compute.UINode

	for name, stats := range nodes {
		// Always recalculate UINode to ensure fresh data (no cache)
		// This ensures nodes reflect real-time changes from pods
		uiNode := compute.CalculateUINode(name, stats)
		storage.GlobalStore.StoreUINode(name, uiNode)
		uiNodes = append(uiNodes, uiNode)
	}

	sort.Slice(uiNodes, func(i, j int) bool {
		return uiNodes[i].Name < uiNodes[j].Name
	})

	return uiNodes
}

func getUINode(nodeName string) (*compute.UINode, bool) {
	// Get raw stats and always recalculate UINode (no cache)
	// This ensures nodes reflect real-time changes from pods
	stats, found := storage.GlobalStore.GetNodeStats(nodeName)
	if !found {
		return nil, false
	}

	uiNode := compute.CalculateUINode(nodeName, stats)
	storage.GlobalStore.StoreUINode(nodeName, uiNode)
	return &uiNode, true
}
