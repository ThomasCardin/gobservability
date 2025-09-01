# Testing Alerts System Locally

## Quick Start

1. **Setup Discord Webhook** (for notifications):
   - Go to your Discord server settings
   - Navigate to Integrations > Webhooks
   - Create a new webhook and copy the URL

2. **Configure environment**:
   ```bash
   cp .env.example .env
   # Edit .env and add your Discord webhook URL
   ```

3. **Start the environment**:
   ```bash
   ./scripts/start-local.sh
   ```

This will:
- Start PostgreSQL database
- Build and start the gobservability server
- Optionally start an agent for testing

## Access Points

- **Web UI**: http://localhost:8080
- **Alerts Page**: http://localhost:8080/alerts/[node-name]
- **gRPC**: localhost:9090
- **PostgreSQL**: localhost:5432
  - Database: `gobservability`
  - User: `gobs`
  - Password: `gobs123`

## Testing Alerts

1. Navigate to a node's page: http://localhost:8080
2. Click on a node to see its details
3. Click on "Alerts" button to access the alerts page
4. Create alert rules with thresholds
5. Monitor when metrics exceed thresholds

## Docker Commands

```bash
# View logs
docker-compose logs -f

# Stop all services
docker-compose down

# Stop and remove all data
docker-compose down -v

# Rebuild after code changes
docker-compose build --no-cache
docker-compose up -d
```

## Troubleshooting

### PostgreSQL Connection Issues
- Ensure PostgreSQL is running: `docker-compose ps`
- Check logs: `docker-compose logs postgres`
- Verify connection: `docker-compose exec postgres psql -U gobs -d gobservability -c '\dt'`

### Discord Notifications Not Working
- Verify webhook URL is correct in `.env`
- Check server logs: `docker-compose logs gobservability | grep Discord`
- Test webhook manually:
  ```bash
  curl -X POST -H "Content-Type: application/json" \
    -d '{"content":"Test message"}' \
    YOUR_DISCORD_WEBHOOK_URL
  ```

### Alerts Not Triggering
- Ensure rules are enabled in the UI
- Check that metrics exceed thresholds for the configured duration
- View server logs for evaluation messages
- Check PostgreSQL for stored rules:
  ```bash
  docker-compose exec postgres psql -U gobs -d gobservability \
    -c "SELECT * FROM alert_rules;"
  ```