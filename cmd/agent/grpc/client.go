package grpc

import (
	"context"
	"fmt"
	"time"

	"github.com/ThomasCardin/gobservability/cmd/agent/pkg/flamegraph"
	pb "github.com/ThomasCardin/gobservability/proto"
	sharedGrpc "github.com/ThomasCardin/gobservability/shared/grpc"
	"github.com/ThomasCardin/gobservability/shared/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	GRPCTimeout = 5 * time.Second
)

type GRPCClient struct {
	serverAddr    string
	client        pb.NodeServiceClient
	conn          *grpc.ClientConn
	flamegraphGen *flamegraph.Generator
	devMode       string
	currentPods   []*types.Pod // Cache pods
}

func NewGRPCClient(serverAddr, devMode string) (*GRPCClient, error) {
	conn, err := grpc.NewClient(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("error: failed to connect to gRPC server: %v", err)
	}

	client := pb.NewNodeServiceClient(conn)

	return &GRPCClient{
		serverAddr:    serverAddr,
		client:        client,
		conn:          conn,
		flamegraphGen: flamegraph.NewGenerator(devMode),
		devMode:       devMode,
		currentPods:   make([]*types.Pod, 0),
	}, nil
}

func (c *GRPCClient) Send(payload *types.NodeStatsPayload) error {
	ctx, cancel := context.WithTimeout(context.Background(), GRPCTimeout)
	defer cancel()

	c.currentPods = payload.Metrics.Pods

	req := &pb.NodeStatsRequest{
		NodeName:  payload.NodeName,
		Timestamp: timestamppb.New(payload.Timestamp),
		Metrics:   sharedGrpc.ConvertToGRPCMetrics(payload.Metrics),
	}

	resp, err := c.client.SendStats(ctx, req)
	if err != nil {
		return fmt.Errorf("error: gRPC SendStats failed: %v", err)
	}

	if resp.Status != "received" {
		return fmt.Errorf("error: server rejected data, status: %s", resp.Status)
	}

	return nil
}

func (c *GRPCClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *GRPCClient) GenerateFlamegraph(nodeName, podName string, duration int32) ([]byte, error) {
	pid := c.flamegraphGen.GetPIDForPod(podName, c.currentPods)
	data, err := c.flamegraphGen.GenerateFlamegraph(nodeName, podName, duration, pid)
	if err != nil {
		return nil, fmt.Errorf("error: failed to generate flamegraph %s", err.Error())
	}

	return data, nil
}
