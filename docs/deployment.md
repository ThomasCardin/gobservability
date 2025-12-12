# Deployment Guide

This guide covers all available deployment options for Gobservability.

---

## Option 1: Docker Compose (Local Development - Full Stack)

**Best for:** Testing the complete system locally without Kubernetes

**Requirements:**
- Docker and Docker Compose installed

**Quick Start:**
```bash
# Start PostgreSQL + Server + Agent
docker-compose up -d

# View logs
docker-compose logs -f

# Access web interface
open http://localhost:8080

# Stop all services
docker-compose down
```

**What's Included:**
- PostgreSQL 16 (persistent database for alerts)
- Gobservability Server (HTTP :8080, gRPC :9090)
- Gobservability Agent (monitors the Docker host)

**Configuration:**
- Edit `docker-compose.yml` to customize environment variables
- Discord webhook URL in `DISCORD_WEBHOOK_URL` environment variable
- PostgreSQL credentials in `POSTGRES_URL`

---

## Option 2: Makefile Development Mode (Quick Testing)

**Best for:** Rapid development and testing without Docker

**Requirements:**
- Go 1.24+ installed
- PostgreSQL running locally (or use docker-compose for DB only)

**Single Agent Mode:**
```bash
# Build binaries and start server + 1 agent
make agent

# Access web interface at http://localhost:8080
# Uses fake pod data for demonstration
```

**Multi-Agent Simulation (7 fake nodes):**
```bash
# Simulate a multi-node cluster
make agents

# Creates 7 agents with different hostnames:
# node-01, agent-02, worker-03, controlplane, gpunode, aiworkloadsonly, node-07
```

**Build Only:**
```bash
make build  # Compiles binaries without running
make stop   # Stop all running gobservability processes
```

**Development Features:**
- Agents run with `-dev` flag (uses fake pod data)
- Real system metrics from local `/proc` filesystem
- No Kubernetes cluster required
- Perfect for UI/UX development

---

## Option 3: Kubernetes Helm Deployment (Production)

**Best for:** Production clusters with persistent monitoring

**Requirements:**
- Kubernetes cluster (1.19+) with `kubectl` configured
- Helm 3.x installed
- CloudNativePG operator (for PostgreSQL) **OR** external PostgreSQL database
- (Optional) Nginx Ingress Controller + cert-manager for external access

### Prerequisites: Install CloudNativePG Operator

```bash
# Install CloudNativePG operator (if not already installed)
kubectl apply -f https://raw.githubusercontent.com/cloudnative-pg/cloudnative-pg/release-1.24/releases/cnpg-1.24.0.yaml
```

**Or use an external PostgreSQL database** (skip CloudNativePG):
```yaml
# In values.yaml
postgresql:
  enabled: false  # Disable CloudNativePG
  externalConnectionString: "postgres://user:pass@external-db:5432/gobservability"
```

### Installation Steps

**1. Create Image Pull Secret (if using private registry):**
```bash
kubectl create secret docker-registry ghcr-secret \
  --docker-server=ghcr.io \
  --docker-username=YOUR_GITHUB_USERNAME \
  --docker-password=YOUR_GITHUB_PAT \
  -n gobservability
```

**2. Configure `values.yaml`:**

Create a custom `values.yaml` file (see [k8s/helm/values.yaml](../k8s/helm/values.yaml)):

```yaml
# Example values.yaml
server:
  image:
    repository: ghcr.io/thomascardin/gobservability-server
    tag: latest
  resources:
    requests:
      cpu: 250m
      memory: 64Mi
    limits:
      cpu: 500m
      memory: 128Mi

agent:
  image:
    repository: ghcr.io/thomascardin/gobservability-agent
    tag: latest
  resources:
    requests:
      cpu: 100m
      memory: 32Mi
    limits:
      cpu: 200m
      memory: 64Mi

postgresql:
  enabled: true  # Set to false if using external DB
  storageClass: "standard"  # Change to your storage class
  size: 5Gi

ingress:
  enabled: true
  className: nginx
  host: gobservability.example.com  # Change to your domain
  tls:
    enabled: true
    secretName: gobservability-tls

secrets:
  discordWebhook: "https://discord.com/api/webhooks/YOUR_WEBHOOK_URL"
```

**3. Install with Helm:**
```bash
# Install the chart
helm install gobservability ./k8s/helm \
  --namespace gobservability \
  --create-namespace \
  --values values.yaml

# Or upgrade existing installation
helm upgrade gobservability ./k8s/helm \
  --namespace gobservability \
  --values values.yaml
```

**4. Verify Deployment:**
```bash
# Check all pods are running
kubectl get pods -n gobservability

# Expected output:
# NAME                                      READY   STATUS    RESTARTS   AGE
# gobservability-server-xxxxx               1/1     Running   0          1m
# gobservability-agent-xxxxx                1/1     Running   0          1m
# gobservability-agent-yyyyy                1/1     Running   0          1m
# gobservability-postgres-1                 1/1     Running   0          2m

# Check agent logs
kubectl logs -n gobservability -l app=gobservability-agent --tail=20

# Check server logs
kubectl logs -n gobservability -l app=gobservability-server --tail=20
```

**5. Access the Web Interface:**

**Option A: Port Forward (local access)**
```bash
kubectl port-forward -n gobservability svc/gobservability-server 8080:8080
# Open http://localhost:8080
```

**Option B: Ingress (external access)**
```bash
# If ingress is enabled, access via your configured domain
open https://gobservability.example.com
```

### Uninstall

```bash
# Remove Helm release
helm uninstall gobservability -n gobservability

# Delete namespace (including PVCs)
kubectl delete namespace gobservability
```

---

## Option 4: Skaffold (Automated Development Workflow)

**Best for:** Developers who want continuous deployment during development

**Requirements:**
- Skaffold installed
- Kubernetes cluster configured
- Docker for multi-arch builds

**Quick Start:**
```bash
# Deploy and watch for file changes
skaffold dev

# Or build + deploy once
skaffold run

# Delete deployment
skaffold delete
```

**Configuration:**
- Edit `skaffold.yaml` to change image registry
- Automatically rebuilds images on code changes (in dev mode)
- Port-forwards server to localhost:8080

---

## Building Custom Images

### Build Multi-Arch Images

```bash
# Build server image (amd64 + arm64)
docker buildx build --platform linux/amd64,linux/arm64 \
  -t ghcr.io/yourusername/gobservability-server:latest \
  -f cmd/server/Dockerfile .

# Build agent image (amd64 + arm64)
docker buildx build --platform linux/amd64,linux/arm64 \
  -t ghcr.io/yourusername/gobservability-agent:latest \
  -f cmd/agent/Dockerfile .
```

### GitHub Actions CI/CD

The repository includes a GitHub Actions workflow (`.github/workflows/build-push-ghcr.yaml`) that automatically:
- Builds multi-arch images on version tags (e.g., `v1.0.0`)
- Pushes to GitHub Container Registry (ghcr.io)
- Tags images with semantic version

**Trigger a release:**
```bash
git tag v1.0.0
git push origin v1.0.0
```

---

## Usage Examples

### Creating Alerts via UI

1. Navigate to a node's alerts page: `http://localhost:8080/alerts/{nodename}`
2. Click "Create Alert Rule"
3. Configure rule:
   - **Target**: `node` (entire node) or `podname` (specific pod)
   - **Metric**: `cpu`, `memory`, `network`, `disk`
   - **Condition**: `>` (greater than) or `<` (less than)
   - **Threshold**: Numeric value (e.g., `80` for 80% CPU)
4. Save rule - Discord notifications will fire when threshold is exceeded

### Generating Flamegraphs

1. Navigate to process details: `http://localhost:8080/nodes/{nodename}/pods/{podname}`
2. Click "Generate Flamegraph"
3. Configure duration (30-600 seconds)
4. View interactive flamegraph visualization (JSON format)

### Monitoring Best Practices

- **Set resource limits** on all Kubernetes deployments to prevent runaway processes
- **Create baseline alerts** (e.g., CPU > 80%, Memory > 90%)
- **Monitor agent health** - agents should always be running on every node
- **Review alert history** regularly to identify patterns
- **Use flamegraphs** for CPU-intensive pods to optimize performance
