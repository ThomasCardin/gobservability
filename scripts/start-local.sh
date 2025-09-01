#!/bin/bash

# Script to start gobservability with docker-compose for local testing

echo "🚀 Starting gobservability local environment..."

# Check if .env file exists
if [ ! -f .env ]; then
    echo "⚠️  .env file not found. Creating from .env.example..."
    if [ -f .env.example ]; then
        cp .env.example .env
        echo "📝 Please edit .env file and add your DISCORD_WEBHOOK_URL"
        echo "   You can get a webhook URL from Discord Server Settings > Integrations > Webhooks"
        read -p "Press Enter to continue after editing .env file..."
    else
        echo "❌ .env.example file not found!"
        echo "Please create a .env file with:"
        echo "DISCORD_WEBHOOK_URL=https://discord.com/api/webhooks/YOUR_WEBHOOK_ID/YOUR_WEBHOOK_TOKEN"
        exit 1
    fi
fi

# Load environment variables
export $(grep -v '^#' .env | xargs)

# Stop any existing containers
echo "🛑 Stopping existing containers..."
docker-compose down

# Start PostgreSQL first and wait for it to be ready
echo "🗄️  Starting PostgreSQL..."
docker-compose up -d postgres

# Wait for PostgreSQL to be healthy
echo "⏳ Waiting for PostgreSQL to be ready..."
for i in {1..30}; do
    if docker-compose exec -T postgres pg_isready -U gobs -d gobservability &>/dev/null; then
        echo "✅ PostgreSQL is ready!"
        break
    fi
    if [ $i -eq 30 ]; then
        echo "❌ PostgreSQL failed to start"
        docker-compose logs postgres
        exit 1
    fi
    sleep 1
done

# Build and start the server
echo "🔨 Building server..."
docker-compose build gobservability

echo "🚀 Starting server..."
docker-compose up -d gobservability

# Optional: Start agent for testing
read -p "Do you want to start the agent for testing? (y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "🤖 Starting agent..."
    docker-compose up -d gobservability-agent
fi

echo ""
echo "✨ gobservability is running!"
echo ""
echo "📊 Server UI: http://localhost:8080"
echo "🔌 gRPC endpoint: localhost:9090"
echo "🗄️  PostgreSQL: localhost:5432 (user: gobs, password: gobs123, database: gobservability)"
echo ""
echo "📝 View logs: docker-compose logs -f"
echo "🛑 Stop all: docker-compose down"
echo ""

# Show logs
docker-compose logs -f