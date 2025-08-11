package grpc

import (
	"context"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/ThomasCardin/peek/cmd/agent/pkg/flamegraph"
	pb "github.com/ThomasCardin/peek/proto"
	sharedGrpc "github.com/ThomasCardin/peek/shared/grpc"
	"github.com/ThomasCardin/peek/shared/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// StreamingGRPCClient handles bidirectional streaming with the server
type StreamingGRPCClient struct {
	serverAddr    string
	nodeName      string
	conn          *grpc.ClientConn
	stream        pb.NodeService_AgentStreamClient
	flamegraphGen *flamegraph.Generator
	currentPods   []*types.Pod
	mu            sync.RWMutex
	ctx           context.Context
	cancel        context.CancelFunc
}

// NewStreamingGRPCClient creates a new streaming gRPC client
func NewStreamingGRPCClient(serverAddr, nodeName, devMode string) (*StreamingGRPCClient, error) {
	conn, err := grpc.NewClient(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	client := &StreamingGRPCClient{
		serverAddr:    serverAddr,
		nodeName:      nodeName,
		conn:          conn,
		flamegraphGen: flamegraph.NewGenerator(devMode),
		currentPods:   make([]*types.Pod, 0),
		ctx:           ctx,
		cancel:        cancel,
	}

	// Establish streaming connection
	if err := client.establishStream(); err != nil {
		conn.Close()
		return nil, err
	}

	// Start background goroutine to handle incoming messages
	go client.handleIncomingMessages()

	return client, nil
}

// establishStream creates the bidirectional stream and sends hello
func (c *StreamingGRPCClient) establishStream() error {
	client := pb.NewNodeServiceClient(c.conn)

	stream, err := client.AgentStream(c.ctx)
	if err != nil {
		return fmt.Errorf("failed to create stream: %v", err)
	}

	c.stream = stream

	// Send hello message
	hello := &pb.AgentMessage{
		Message: &pb.AgentMessage_Hello{
			Hello: &pb.AgentHello{
				NodeName:     c.nodeName,
				AgentVersion: "1.0.0",
			},
		},
	}

	if err := c.stream.Send(hello); err != nil {
		return fmt.Errorf("failed to send hello: %v", err)
	}

	log.Printf("Established streaming connection with server for node %s", c.nodeName)
	return nil
}

// handleIncomingMessages processes messages from the server
func (c *StreamingGRPCClient) handleIncomingMessages() {
	for {
		msg, err := c.stream.Recv()
		if err == io.EOF {
			log.Printf("Server closed the stream")
			c.reconnect()
			return
		}
		if err != nil {
			log.Printf("Error receiving message: %v", err)
			c.reconnect()
			return
		}

		switch m := msg.Message.(type) {
		case *pb.ServerMessage_Ack:
			log.Printf("Received ack from server: %s", m.Ack.Message)

		case *pb.ServerMessage_FlamegraphRequest:
			// Handle flamegraph request
			go c.handleFlamegraphRequest(m.FlamegraphRequest)
		}
	}
}

// handleFlamegraphRequest processes flamegraph generation requests
func (c *StreamingGRPCClient) handleFlamegraphRequest(req *pb.FlamegraphRequest) {
	log.Printf("Received flamegraph request for pod %s, duration %ds", req.PodName, req.Duration)

	// Get current pods
	c.mu.RLock()
	pods := c.currentPods
	c.mu.RUnlock()

	// Generate flamegraph
	pid := c.flamegraphGen.GetPIDForPod(req.PodName, pods)
	data, err := c.flamegraphGen.GenerateFlamegraph(req.NodeName, req.PodName, req.Duration, req.Format, pid)

	// Prepare response
	response := &pb.AgentMessage{
		Message: &pb.AgentMessage_FlamegraphResponse{
			FlamegraphResponse: &pb.FlamegraphResponse{
				FlamegraphData: data,
				Format:         req.Format,
				RequestId:      req.RequestId,
			},
		},
	}

	if err != nil {
		response.Message.(*pb.AgentMessage_FlamegraphResponse).FlamegraphResponse.Error = err.Error()
	}

	// Send response
	if err := c.stream.Send(response); err != nil {
		log.Printf("Failed to send flamegraph response: %v", err)
	}
}

// Send sends metrics via the streaming connection
func (c *StreamingGRPCClient) Send(payload *types.NodeStatsPayload) error {
	// Update cached pods
	c.mu.Lock()
	c.currentPods = payload.Metrics.Pods
	c.mu.Unlock()

	// Convert and send stats
	stats := &pb.AgentMessage{
		Message: &pb.AgentMessage_Stats{
			Stats: &pb.NodeStatsRequest{
				NodeName:  payload.NodeName,
				Timestamp: timestamppb.New(payload.Timestamp),
				Metrics:   sharedGrpc.ConvertToGRPCMetrics(payload.Metrics),
			},
		},
	}

	if err := c.stream.Send(stats); err != nil {
		return fmt.Errorf("failed to send stats: %v", err)
	}

	return nil
}

// reconnect attempts to re-establish the connection
func (c *StreamingGRPCClient) reconnect() {
	log.Printf("Attempting to reconnect...")

	// Cancel existing context
	c.cancel()

	// Create new context
	c.ctx, c.cancel = context.WithCancel(context.Background())

	// Retry connection with exponential backoff
	retryDelay := 1 * time.Second
	maxDelay := 30 * time.Second

	for {
		if err := c.establishStream(); err != nil {
			log.Printf("Reconnection failed: %v, retrying in %s", err, retryDelay)
			time.Sleep(retryDelay)

			// Exponential backoff
			retryDelay *= 2
			if retryDelay > maxDelay {
				retryDelay = maxDelay
			}
		} else {
			// Successfully reconnected
			go c.handleIncomingMessages()
			break
		}
	}
}

// Close closes the streaming connection
func (c *StreamingGRPCClient) Close() error {
	c.cancel()
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
