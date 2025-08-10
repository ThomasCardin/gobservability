package grpc

import (
	"context"
	"log"
	"net"

	"github.com/ThomasCardin/peek/cmd/server/storage"
	pb "github.com/ThomasCardin/peek/proto"
	sharedGrpc "github.com/ThomasCardin/peek/shared/grpc"
	"github.com/ThomasCardin/peek/shared/types"
	"google.golang.org/grpc"
)

// Server implémente le service gRPC NodeService
type Server struct {
	pb.UnimplementedNodeServiceServer
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

	return &pb.StatsResponse{
		Status: "received",
	}, nil
}

// GenerateFlamegraph handles flamegraph generation requests
func (s *Server) GenerateFlamegraph(ctx context.Context, req *pb.FlamegraphRequest) (*pb.FlamegraphResponse, error) {
	log.Printf("Received flamegraph request for node: %s, pod: %s, duration: %ds, format: %s", 
		req.NodeName, req.PodName, req.Duration, req.Format)

	// For now, return a placeholder response
	// In a real implementation, this would:
	// 1. Find the appropriate agent connection for the requested node
	// 2. Forward the request to that agent via bidirectional streaming
	// 3. Wait for the response and return it
	
	return &pb.FlamegraphResponse{
		FlamegraphData: []byte("Flamegraph generation not yet implemented"),
		Format:         req.Format,
		Error:          "",
	}, nil
}


// StartGRPCServer démarre le serveur gRPC sur le port spécifié
func StartGRPCServer(port string) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer()
	pb.RegisterNodeServiceServer(grpcServer, &Server{})

	log.Printf("gRPC server starting on port %s", port)
	return grpcServer.Serve(lis)
}
