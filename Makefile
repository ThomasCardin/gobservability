# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
AGENT_BINARY=agent
SERVER_BINARY=server

.PHONY: all build agent agents stop help proto clean-proto install-proto-deps

all: build

build:
	$(GOBUILD) -o $(AGENT_BINARY) -v ./cmd/agent
	$(GOBUILD) -o $(SERVER_BINARY) -v ./cmd/server

agent: build
	@echo "Starting web server in background..."
	@./$(SERVER_BINARY) &
	@sleep 2
	@echo "Starting single agent..."
	./$(AGENT_BINARY) -dev

agents: build
	@echo "Cleaning up any existing processes..."
	@killall -q $(SERVER_BINARY) $(AGENT_BINARY) 2>/dev/null || true
	@sleep 1
	@echo "Starting web server in background..."
	@./$(SERVER_BINARY) &
	@sleep 3
	@echo "Starting 7 agents..."
	@./$(AGENT_BINARY) -hostname="node-01" -interval=5s -dev &
	@./$(AGENT_BINARY) -hostname="agent-02" -interval=5s -dev &
	@./$(AGENT_BINARY) -hostname="worker-03" -interval=5s -dev &
	@./$(AGENT_BINARY) -hostname="controlplnae" -interval=5s -dev &
	@./$(AGENT_BINARY) -hostname="gpunode" -interval=5s -dev &
	@./$(AGENT_BINARY) -hostname="aiworkloadsonly" -interval=5s -dev &
	@./$(AGENT_BINARY) -hostname="node-07" -interval=5s -dev &
	@sleep 2
	@echo "All 7 agents + web server started!"
	@echo "Check http://localhost:8080"

stop:
	@echo "Stopping all gobservability processes..."
	@killall -9 $(SERVER_BINARY) $(AGENT_BINARY) 2>/dev/null || true
	@sleep 1
	@echo "All processes stopped."

# Protobuf/gRPC targets
install-proto-deps:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

proto:
	@echo "Generating Go code from protobuf files..."
	export PATH=$$PATH:$(shell go env GOPATH)/bin && protoc --go_out=. --go-grpc_out=. proto/gobservability.proto

clean-proto:
	@echo "Cleaning generated protobuf files..."
	rm -f proto/*.pb.go

help:
	@echo "Available targets:"
	@echo "  build              - Build agent and web-server"
	@echo "  agent              - Build and run single agent + server"
	@echo "  agents             - Build and run 7 test agents + server"
	@echo "  stop               - Stop all gobservability processes"
	@echo "  install-proto-deps - Install protobuf Go generators"
	@echo "  proto              - Generate Go code from proto files"
	@echo "  clean-proto        - Remove generated protobuf files"
	@echo "  help               - Show this help"

.DEFAULT_GOAL := help