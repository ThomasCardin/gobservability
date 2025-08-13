package grpc

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "github.com/ThomasCardin/gobservability/proto"
	cache "github.com/patrickmn/go-cache"
)

// AgentConnection represents a connected agent with its stream
type AgentConnection struct {
	NodeName string
	Stream   pb.NodeService_AgentStreamServer
	Context  context.Context
	Cancel   context.CancelFunc
}

// AgentManager manages all connected agents using go-cache
type AgentManager struct {
	agents   *cache.Cache // nodeName -> *AgentConnection
	requests *cache.Cache // requestID -> chan *pb.FlamegraphResponse
}

// NewAgentManager creates a new agent manager with go-cache
func NewAgentManager() *AgentManager {
	return &AgentManager{
		// Agents expire after 5 minutes, cleanup every minute
		agents: cache.New(5*time.Minute, 1*time.Minute),
		// Requests expire after 5 minutes (to handle long flamegraph operations), cleanup every 30 seconds
		requests: cache.New(5*time.Minute, 30*time.Second),
	}
}

// RegisterAgent registers a new agent connection
func (am *AgentManager) RegisterAgent(nodeName string, stream pb.NodeService_AgentStreamServer, ctx context.Context, cancel context.CancelFunc) {
	// Check if there's an existing connection
	if existing, found := am.agents.Get(nodeName); found {
		if conn, ok := existing.(*AgentConnection); ok {
			log.Printf("Closing existing connection for node %s", nodeName)
			conn.Cancel()
		}
	}

	conn := &AgentConnection{
		NodeName: nodeName,
		Stream:   stream,
		Context:  ctx,
		Cancel:   cancel,
	}

	// Set with default expiration (5 minutes)
	am.agents.Set(nodeName, conn, cache.DefaultExpiration)

	// Set up cleanup on expiration
	am.agents.OnEvicted(func(key string, value interface{}) {
		if conn, ok := value.(*AgentConnection); ok {
			log.Printf("Agent connection expired for node %s", key)
			conn.Cancel()
		}
	})

	log.Printf("Registered agent for node %s", nodeName)
}

// UnregisterAgent removes an agent connection
func (am *AgentManager) UnregisterAgent(nodeName string) {
	if existing, found := am.agents.Get(nodeName); found {
		if conn, ok := existing.(*AgentConnection); ok {
			conn.Cancel()
		}
	}
	am.agents.Delete(nodeName)
	log.Printf("Unregistered agent for node %s", nodeName)
}

// GetAgent returns the connection for a specific node
func (am *AgentManager) GetAgent(nodeName string) (*AgentConnection, error) {
	if item, found := am.agents.Get(nodeName); found {
		if conn, ok := item.(*AgentConnection); ok {
			// Reset expiration on access to keep active connections alive
			am.agents.Set(nodeName, conn, cache.DefaultExpiration)
			return conn, nil
		}
	}
	return nil, fmt.Errorf("no agent connected for node %s", nodeName)
}

// UpdateLastSeen refreshes the TTL for an agent (called when receiving stats)
func (am *AgentManager) UpdateLastSeen(nodeName string) {
	if item, found := am.agents.Get(nodeName); found {
		// Reset expiration to keep connection alive
		am.agents.Set(nodeName, item, cache.DefaultExpiration)
	}
}

// RegisterRequest registers a flamegraph request with a response channel
func (am *AgentManager) RegisterRequest(requestID string) chan *pb.FlamegraphResponse {
	ch := make(chan *pb.FlamegraphResponse, 1)
	am.requests.Set(requestID, ch, cache.DefaultExpiration)
	return ch
}

// RegisterRequestWithTimeout registers a flamegraph request with a custom timeout
func (am *AgentManager) RegisterRequestWithTimeout(requestID string, timeout time.Duration) chan *pb.FlamegraphResponse {
	ch := make(chan *pb.FlamegraphResponse, 1)
	am.requests.Set(requestID, ch, timeout)
	return ch
}

// CompleteRequest completes a flamegraph request
func (am *AgentManager) CompleteRequest(requestID string, response *pb.FlamegraphResponse) {
	if item, found := am.requests.Get(requestID); found {
		if ch, ok := item.(chan *pb.FlamegraphResponse); ok {
			ch <- response
			close(ch)
		}
		am.requests.Delete(requestID)
	}
}

// GetConnectedAgents returns a list of currently connected agents
func (am *AgentManager) GetConnectedAgents() []string {
	items := am.agents.Items()
	agents := make([]string, 0, len(items))

	for nodeName := range items {
		agents = append(agents, nodeName)
	}

	return agents
}
