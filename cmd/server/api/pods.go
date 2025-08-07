package api

import (
	"github.com/ThomasCardin/peek/cmd/server/compute"
	"github.com/ThomasCardin/peek/shared/types"
)

// calculateUIPods converts raw Pod data to UIPod with computed metrics
func calculateUIPods(pods []*types.Pod) []compute.UIPod {
	return calculateUIPodsWithNodeContext(pods, nil)
}

// calculateUIPodsWithNodeContext converts raw Pod data to UIPod with node context for real calculations
func calculateUIPodsWithNodeContext(pods []*types.Pod, nodeStats *types.NodeStatsPayload) []compute.UIPod {
	var uiPods []compute.UIPod
	for _, pod := range pods {
		uiPod := compute.CalculateUIPodWithNodeContext(pod, nodeStats)
		uiPods = append(uiPods, uiPod)
	}
	return uiPods
}
