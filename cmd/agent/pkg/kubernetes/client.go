package kubernetes

import (
	"github.com/ThomasCardin/peek/shared/types"
)

// Client provides Kubernetes functionality
type Client struct {
	devMode string
}

// NewClient creates a new Kubernetes client
func NewClient(devMode string) *Client {
	return &Client{
		devMode: devMode,
	}
}

// GetPodsForNode returns pods for a specific node
func (c *Client) GetPodsForNode(nodeName string) ([]*types.Pod, error) {
	return GetPodsPID(c.devMode, nodeName)
}