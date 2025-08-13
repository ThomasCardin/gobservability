package api

import (
	"github.com/ThomasCardin/gobservability/cmd/server/formatter"
	"github.com/ThomasCardin/gobservability/shared/types"
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
