package api

import (
	"sort"

	"github.com/ThomasCardin/gobservability/cmd/server/formatter"
	"github.com/ThomasCardin/gobservability/cmd/server/storage"
)

func getUINodes() []formatter.UINode {
	nodes := storage.GlobalStore.GetAllNodes()
	var uiNodes []formatter.UINode

	for name, stats := range nodes {
		// Format node for UI display
		uiNode := formatter.FormatNodeForUI(name, stats)
		uiNodes = append(uiNodes, uiNode)
	}

	sort.Slice(uiNodes, func(i, j int) bool {
		return uiNodes[i].Name < uiNodes[j].Name
	})

	return uiNodes
}

func getUINode(nodeName string) (*formatter.UINode, bool) {
	// Get raw stats and format for UI display
	stats, found := storage.GlobalStore.GetNodeStats(nodeName)
	if !found {
		return nil, false
	}

	uiNode := formatter.FormatNodeForUI(nodeName, stats)
	return &uiNode, true
}
