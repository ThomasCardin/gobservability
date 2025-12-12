# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
AGENT_BINARY=agent
SERVER_BINARY=server

.PHONY: all build agent agents stop help proto clean-proto install-proto-deps agents-logs

all: build

build:
	$(GOBUILD) -o $(AGENT_BINARY) -v ./cmd/agent
	$(GOBUILD) -o $(SERVER_BINARY) -v ./cmd/server

agent: build
	@echo "Starting PostgreSQL..."
	@if [ ! -f .env ]; then echo "❌ .env file not found! Copy .env.example and configure it."; exit 1; fi
	@docker-compose up -d postgres
	@echo "Waiting for PostgreSQL to be ready..."
	@for i in $$(seq 1 30); do \
		if docker-compose exec -T postgres pg_isready -U gobs -d gobservability >/dev/null 2>&1; then \
			echo "✅ PostgreSQL is ready!"; \
			break; \
		fi; \
		if [ $$i -eq 30 ]; then \
			echo "❌ PostgreSQL failed to start"; \
			docker-compose logs postgres; \
			exit 1; \
		fi; \
		sleep 1; \
	done
	@echo "Starting web server in background..."
	@export $$(grep -v '^#' .env | xargs) && ./$(SERVER_BINARY) &
	@sleep 2
	@echo "Starting single agent..."
	./$(AGENT_BINARY) -dev

agents: build
	@echo "Starting PostgreSQL..."
	@if [ ! -f .env ]; then echo "❌ .env file not found! Copy .env.example and configure it."; exit 1; fi
	@docker-compose up -d postgres
	@echo "Waiting for PostgreSQL to be ready..."
	@for i in $$(seq 1 30); do \
		if docker-compose exec -T postgres pg_isready -U gobs -d gobservability >/dev/null 2>&1; then \
			echo "✅ PostgreSQL is ready!"; \
			break; \
		fi; \
		if [ $$i -eq 30 ]; then \
			echo "❌ PostgreSQL failed to start"; \
			docker-compose logs postgres; \
			exit 1; \
		fi; \
		sleep 1; \
	done
	@echo "Cleaning up any existing processes..."
	@killall -q $(SERVER_BINARY) $(AGENT_BINARY) 2>/dev/null || true
	@sleep 1
	@echo "Starting web server in background..."
	@export $$(grep -v '^#' .env | xargs) && ./$(SERVER_BINARY) &
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
	@echo "🌐 Web UI: http://localhost:8080"
	@echo "📊 Alerts: http://localhost:8080/alerts/[node-name]"

stop:
	@echo "Stopping all gobservability processes..."
	@killall -9 $(SERVER_BINARY) $(AGENT_BINARY) 2>/dev/null || true
	@echo "Stopping PostgreSQL container..."
	@docker-compose stop postgres
	@sleep 1
	@echo "All processes stopped."

agents-logs:
	@echo "Showing PostgreSQL logs (Ctrl+C to exit)..."
	docker-compose logs -f postgres

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
	@echo ""
	@echo "Building:"
	@echo "  build              - Build agent and web-server"
	@echo ""
	@echo "Development (with PostgreSQL):"
	@echo "  agent              - Start PostgreSQL + single agent + server"
	@echo "  agents             - Start PostgreSQL + 7 test agents + server"
	@echo "  stop               - Stop all processes and PostgreSQL"
	@echo "  agents-logs        - View PostgreSQL logs"
	@echo ""
	@echo "Protobuf:"
	@echo "  install-proto-deps - Install protobuf Go generators"
	@echo "  proto              - Generate Go code from proto files"
	@echo "  clean-proto        - Remove generated protobuf files"
	@echo ""
	@echo "  help               - Show this help"
	@echo ""
	@echo "Note: Requires .env file with DISCORD_WEBHOOK_URL for alerts"

.DEFAULT_GOAL := help