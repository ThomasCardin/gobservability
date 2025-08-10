package api

import (
	"github.com/ThomasCardin/peek/cmd/server/formatter"
	"github.com/ThomasCardin/peek/shared/types"
)

// formatPodsForUI formats pod data for UI display
func formatPodsForUI(pods []*types.Pod) []formatter.UIPod {
	var uiPods []formatter.UIPod
	for _, pod := range pods {
		uiPod := formatter.FormatPodForUI(pod)
		uiPods = append(uiPods, uiPod)
	}
	return uiPods
}
