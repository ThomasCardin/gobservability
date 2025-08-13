package grpc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"time"

	"github.com/ThomasCardin/gobservability/cmd/agent/pkg/flamegraph"
	pb "github.com/ThomasCardin/gobservability/proto"
	sharedGrpc "github.com/ThomasCardin/gobservability/shared/grpc"
	"github.com/ThomasCardin/gobservability/shared/types"
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
		return nil, errors.New("failed to connect to gRPC server")
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
		return errors.New("failed to create stream")
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
		return errors.New("failed to send hello")
	}

	slog.Info("node connected to server", "component", "grpc", "node", c.nodeName)
	return nil
}

// handleIncomingMessages processes messages from the server
func (c *StreamingGRPCClient) handleIncomingMessages() {
	for {
		msg, err := c.stream.Recv()
		if err == io.EOF {
			slog.Info("server closed the stream", "component", "grpc")
			c.reconnect()
			return
		}
		if err != nil {
			slog.Error("receiving message failed", "component", "grpc", "error", err)
			c.reconnect()
			return
		}

		switch m := msg.Message.(type) {
		case *pb.ServerMessage_Ack:
			slog.Info("received ack from server", "component", "grpc", "message", m.Ack.Message)
		case *pb.ServerMessage_FlamegraphRequest:
			go c.handleFlamegraphRequest(m.FlamegraphRequest)
		}
	}
}

// handleFlamegraphRequest processes flamegraph generation requests
func (c *StreamingGRPCClient) handleFlamegraphRequest(req *pb.FlamegraphRequest) {
	slog.Info("received flamegraph request", "component", "flamegraph", "pod", req.PodName, "duration", req.Duration)

	// Get current pods
	c.mu.RLock()
	pods := c.currentPods
	c.mu.RUnlock()

	// Generate flamegraph
	pid := c.flamegraphGen.GetPIDForPod(req.PodName, pods)
	slog.Info("found PID for pod", "component", "flamegraph", "pid", pid, "pod", req.PodName)

	if pid <= 0 {
		slog.Error("no valid PID found for pod", "component", "flamegraph", "pod", req.PodName)
		response := &pb.AgentMessage{
			Message: &pb.AgentMessage_FlamegraphResponse{
				FlamegraphResponse: &pb.FlamegraphResponse{
					Error:     fmt.Sprintf("[FLAMEGRAPH] error: no valid PID found for pod %s", req.PodName),
					RequestId: req.RequestId,
				},
			},
		}
		if err := c.stream.Send(response); err != nil {
			slog.Error("failed to send error response", "component", "flamegraph", "error", err)
		}
		return
	}

	slog.Info("generating flamegraph", "component", "flamegraph", "pid", pid, "duration", req.Duration)
	data, err := c.flamegraphGen.GenerateFlamegraph(req.NodeName, req.PodName, req.Duration, pid)

	if err != nil {
		slog.Error("error generating flamegraph", "component", "flamegraph", "error", err)
	} else {
		slog.Info("successfully generated flamegraph", "component", "flamegraph", "data_size_bytes", len(data))
	}

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
	slog.Info("sending flamegraph response", "component", "flamegraph", "request_id", req.RequestId)
	if err := c.stream.Send(response); err != nil {
		slog.Error("failed to send flamegraph response", "component", "flamegraph", "error", err)
	} else {
		slog.Info("successfully sent flamegraph response", "component", "flamegraph", "request_id", req.RequestId)
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
		return errors.New("failed to send stats")
	}

	return nil
}

// reconnect attempts to re-establish the connection
func (c *StreamingGRPCClient) reconnect() {
	slog.Info("attempting to reconnect", "component", "grpc")

	// Cancel existing context
	c.cancel()

	// Create new context
	c.ctx, c.cancel = context.WithCancel(context.Background())

	// Retry connection with exponential backoff
	retryDelay := 1 * time.Second
	maxDelay := 30 * time.Second

	for {
		if err := c.establishStream(); err != nil {
			slog.Error("reconnection failed, retrying", "component", "grpc", "error", err, "retry_delay", retryDelay)
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
