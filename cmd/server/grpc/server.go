package grpc

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/ThomasCardin/peek/cmd/server/storage"
	pb "github.com/ThomasCardin/peek/proto"
	sharedGrpc "github.com/ThomasCardin/peek/shared/grpc"
	"github.com/ThomasCardin/peek/shared/types"
	"google.golang.org/grpc"
)

// Server implémente le service gRPC NodeService
type Server struct {
	pb.UnimplementedNodeServiceServer
	agentManager *AgentManager
}

// NewServer creates a new gRPC server with agent management
func NewServer() *Server {
	return &Server{
		agentManager: NewAgentManager(),
	}
}

// SendStats remplace le handler HTTP /api/stats
func (s *Server) SendStats(ctx context.Context, req *pb.NodeStatsRequest) (*pb.StatsResponse, error) {
	// Convertir la requête gRPC vers les types Go existants
	payload := types.NodeStatsPayload{
		NodeName:  req.NodeName,
		Timestamp: req.Timestamp.AsTime(),
		Metrics:   sharedGrpc.ConvertNodeMetrics(req.Metrics),
	}

	// Utiliser le storage existant (même logique que l'ancien /api/stats)
	storage.GlobalStore.StoreNodeStats(payload)

	// Update agent last seen time when receiving stats
	s.agentManager.UpdateLastSeen(req.NodeName)

	return &pb.StatsResponse{
		Status: "received",
	}, nil
}

// GenerateFlamegraph handles flamegraph generation requests
func (s *Server) GenerateFlamegraph(ctx context.Context, req *pb.FlamegraphRequest) (*pb.FlamegraphResponse, error) {
	log.Printf("Received flamegraph request for node: %s, pod: %s, duration: %ds, format: %s",
		req.NodeName, req.PodName, req.Duration, req.Format)

	// Get the agent connection for the requested node
	agent, err := s.agentManager.GetAgent(req.NodeName)
	if err != nil {
		return &pb.FlamegraphResponse{
			Error: fmt.Sprintf("Agent not connected: %v", err),
		}, nil
	}

	// Generate a unique request ID
	requestID := fmt.Sprintf("%s-%s-%d", req.NodeName, req.PodName, time.Now().UnixNano())

	// Register the request and get response channel
	responseChan := s.agentManager.RegisterRequest(requestID)

	// Add request ID to the request
	reqWithID := &pb.FlamegraphRequest{
		NodeName:  req.NodeName,
		PodName:   req.PodName,
		Duration:  req.Duration,
		Format:    req.Format,
		RequestId: requestID,
	}

	// Send the flamegraph request to the agent
	err = agent.Stream.Send(&pb.ServerMessage{
		Message: &pb.ServerMessage_FlamegraphRequest{
			FlamegraphRequest: reqWithID,
		},
	})
	if err != nil {
		return &pb.FlamegraphResponse{
			Error: fmt.Sprintf("Failed to send request to agent: %v", err),
		}, nil
	}

	// Wait for response or timeout
	select {
	case response := <-responseChan:
		return response, nil
	case <-ctx.Done():
		return &pb.FlamegraphResponse{
			Error: "Request cancelled",
		}, nil
	}
}

// AgentStream implements bidirectional streaming for agents
func (s *Server) AgentStream(stream pb.NodeService_AgentStreamServer) error {
	ctx, cancel := context.WithCancel(stream.Context())
	defer cancel()

	var nodeName string

	// Handle incoming messages from agent
	for {
		msg, err := stream.Recv()
		if err != nil {
			log.Printf("Stream error for node %s: %v", nodeName, err)
			if nodeName != "" {
				s.agentManager.UnregisterAgent(nodeName)
			}
			return err
		}

		switch m := msg.Message.(type) {
		case *pb.AgentMessage_Hello:
			// Agent registration
			nodeName = m.Hello.NodeName
			log.Printf("Agent hello from node %s, version %s", nodeName, m.Hello.AgentVersion)

			// Register the agent
			s.agentManager.RegisterAgent(nodeName, stream, ctx, cancel)

			// Send acknowledgment
			err = stream.Send(&pb.ServerMessage{
				Message: &pb.ServerMessage_Ack{
					Ack: &pb.ServerAck{
						Message: fmt.Sprintf("Welcome agent %s", nodeName),
					},
				},
			})
			if err != nil {
				log.Printf("Failed to send ack to %s: %v", nodeName, err)
				return err
			}

		case *pb.AgentMessage_Stats:
			// Handle stats submission via streaming
			payload := types.NodeStatsPayload{
				NodeName:  m.Stats.NodeName,
				Timestamp: m.Stats.Timestamp.AsTime(),
				Metrics:   sharedGrpc.ConvertNodeMetrics(m.Stats.Metrics),
			}
			storage.GlobalStore.StoreNodeStats(payload)
			s.agentManager.UpdateLastSeen(m.Stats.NodeName)

		case *pb.AgentMessage_FlamegraphResponse:
			// Handle flamegraph response
			resp := m.FlamegraphResponse
			if resp.RequestId != "" {
				log.Printf("Received flamegraph response from agent %s for request %s", nodeName, resp.RequestId)
				s.agentManager.CompleteRequest(resp.RequestId, resp)
			} else {
				log.Printf("Warning: Received flamegraph response without request ID from agent %s", nodeName)
			}
		}
	}
}

// StartGRPCServer démarre le serveur gRPC sur le port spécifié
func StartGRPCServer(port string) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer()
	server := NewServer()
	pb.RegisterNodeServiceServer(grpcServer, server)

	log.Printf("gRPC server starting on port %s", port)
	return grpcServer.Serve(lis)
}
