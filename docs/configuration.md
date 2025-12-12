# Configuration Guide

This guide covers all configuration options for Gobservability.

---

## Environment Variables

### Server Configuration

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `POSTGRES_URL` | PostgreSQL connection string | `postgres://gobs:gobs123@localhost:5432/gobservability` | Yes |
| `DISCORD_WEBHOOK_URL` | Discord webhook for alert notifications | _(empty)_ | No |
| `GIN_MODE` | Gin framework mode (`debug`, `release`) | `release` | No |

**Example:**
```bash
export POSTGRES_URL="postgres://user:password@postgres-host:5432/gobservability?sslmode=disable"
export DISCORD_WEBHOOK_URL="https://discord.com/api/webhooks/123456789/abcdefgh"
export GIN_MODE="release"
```

### Agent Configuration

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `NODE_NAME` | Node identifier (auto-set in K8s via `spec.nodeName`) | hostname | No |
| `ENABLE_FLAMEGRAPH` | Enable flamegraph feature | `true` | No |

**Example:**
```bash
export NODE_NAME="my-node-01"
export ENABLE_FLAMEGRAPH="true"
```

---

## Resource Requirements

### Production Recommendations

#### Server
**CPU:**
- Request: `250m` (0.25 cores)
- Limit: `500m` (0.5 cores)

**Memory:**
- Request: `64Mi`
- Limit: `128Mi`

**Disk:**
- N/A (stateless, uses in-memory cache with 10s TTL)

**Kubernetes Example:**
```yaml
resources:
  requests:
    cpu: 250m
    memory: 64Mi
  limits:
    cpu: 500m
    memory: 128Mi
```

---

#### Agent (per node)
**CPU:**
- Request: `100m` (0.1 cores)
- Limit: `200m` (0.2 cores)

**Memory:**
- Request: `32Mi`
- Limit: `64Mi`

**Disk:**
- N/A (reads `/proc`, no persistent storage)

**Kubernetes Example:**
```yaml
resources:
  requests:
    cpu: 100m
    memory: 32Mi
  limits:
    cpu: 200m
    memory: 64Mi
```

---

#### PostgreSQL
**CPU:**
- Request: `250m`
- Limit: `1000m` (adjust based on alert volume)

**Memory:**
- Request: `256Mi`
- Limit: `512Mi`

**Disk:**
- Minimum: `5Gi` (for alert history)
- Recommended: `10Gi+` for production

**Kubernetes Example (CloudNativePG):**
```yaml
apiVersion: postgresql.cnpg.io/v1
kind: Cluster
spec:
  instances: 1
  storage:
    size: 5Gi
    storageClass: standard
  resources:
    requests:
      cpu: 250m
      memory: 256Mi
    limits:
      cpu: 1000m
      memory: 512Mi
```

---

## Helm Chart Configuration

### Complete values.yaml Example

```yaml
# Server configuration
server:
  image:
    repository: ghcr.io/thomascardin/gobservability-server
    tag: latest
    pullPolicy: Always

  replicas: 1

  resources:
    requests:
      cpu: 250m
      memory: 64Mi
    limits:
      cpu: 500m
      memory: 128Mi

  service:
    type: ClusterIP
    httpPort: 8080
    grpcPort: 9090

# Agent configuration
agent:
  image:
    repository: ghcr.io/thomascardin/gobservability-agent
    tag: latest
    pullPolicy: Always

  resources:
    requests:
      cpu: 100m
      memory: 32Mi
    limits:
      cpu: 200m
      memory: 64Mi

  # Interval for metric collection
  interval: 5s

# PostgreSQL configuration
postgresql:
  enabled: true  # Set to false if using external DB

  # CloudNativePG settings
  instances: 1
  storageClass: "standard"  # Change to your storage class
  size: 5Gi

  # Or use external database
  # enabled: false
  # externalConnectionString: "postgres://user:pass@external-db:5432/gobservability"

# Ingress configuration
ingress:
  enabled: true
  className: nginx
  host: gobservability.example.com

  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/ssl-redirect: "true"

  tls:
    enabled: true
    secretName: gobservability-tls

# Secrets configuration
secrets:
  # Discord webhook URL for alerts
  discordWebhook: "https://discord.com/api/webhooks/YOUR_WEBHOOK_URL"

  # PostgreSQL credentials (if using CloudNativePG)
  postgres:
    username: gobs
    password: gobs123  # Change in production!

# Image pull secrets
imagePullSecrets:
  - name: ghcr-secret

# Namespace
namespace: gobservability
```

---

## Security Configuration

### Agent Privileges

The agent requires elevated privileges to access `/proc` filesystem and perform profiling:

```yaml
securityContext:
  privileged: true
  capabilities:
    add:
    - SYS_ADMIN   # Required for /proc access
    - SYS_PTRACE  # Required for perf profiling
    - SYS_RAWIO   # Required for perf profiling
```

**Why these capabilities are needed:**
- `SYS_ADMIN`: Access to `/proc` and `/sys` filesystems across namespaces
- `SYS_PTRACE`: Attach to processes for flamegraph generation
- `SYS_RAWIO`: Raw access to kernel performance counters

**Security considerations:**
- Agents run with `hostPID: true` to see all node processes
- `/proc` and `/sys` are mounted as **read-only**
- Agents do not have write access to host filesystem
- Consider using PodSecurityPolicies or Pod Security Standards to restrict which nodes can run agents

---

### RBAC Permissions

The agent ServiceAccount requires the following permissions:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: gobservability-agent
rules:
- apiGroups: [""]
  resources: ["pods", "nodes"]
  verbs: ["get", "list", "watch"]
```

**Permissions explained:**
- `pods`: Discover running pods on the node
- `nodes`: Get node metadata
- `get`, `list`, `watch`: Read-only operations (no create/update/delete)

**Least privilege principle:**
- Agent only needs **read** access
- No access to secrets, configmaps, or other sensitive resources
- Limited to `pods` and `nodes` resources only

---

### Secrets Management

**Development (Docker Compose / Local):**
```bash
# Set environment variables directly
export DISCORD_WEBHOOK_URL="https://discord.com/api/webhooks/..."
export POSTGRES_URL="postgres://user:pass@localhost:5432/gobservability"
```

**Kubernetes (Basic):**
```bash
# Create secret from literal
kubectl create secret generic discord-webhook \
  --from-literal=webhook_url="https://discord.com/api/webhooks/..." \
  -n gobservability
```

**Kubernetes (Production - Sealed Secrets):**
```bash
# Encrypt secrets before committing to Git
echo -n "https://discord.com/api/webhooks/..." | \
  kubeseal --raw --from-file=/dev/stdin \
  --namespace gobservability \
  --name discord-webhook
```

**Kubernetes (Production - External Secrets Operator):**
```yaml
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: gobservability-secrets
  namespace: gobservability
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: vault-backend
    kind: SecretStore
  target:
    name: discord-webhook
  data:
  - secretKey: webhook_url
    remoteRef:
      key: gobservability/discord-webhook
```

**Best Practices:**
- **Never** commit secrets to Git
- Use Kubernetes Secrets or external secret management (Vault, AWS Secrets Manager)
- Rotate credentials regularly (especially PostgreSQL passwords)
- Use RBAC to limit who can access secrets

---

## Network Configuration

### Ports

**Server:**
- `8080/TCP`: HTTP API and web interface
- `9090/TCP`: gRPC server for agent communication

**Agent:**
- No inbound ports (agents connect to server)

**PostgreSQL:**
- `5432/TCP`: PostgreSQL database (internal only)

### Firewall Rules (Kubernetes)

If using NetworkPolicies:

```yaml
# Allow server to receive HTTP/gRPC traffic
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: gobservability-server
  namespace: gobservability
spec:
  podSelector:
    matchLabels:
      app: gobservability-server
  policyTypes:
  - Ingress
  ingress:
  - from:
    - podSelector:
        matchLabels:
          app: gobservability-agent
    ports:
    - protocol: TCP
      port: 9090
  - from:
    - namespaceSelector: {}  # Allow ingress controller
    ports:
    - protocol: TCP
      port: 8080
```

```yaml
# Allow agents to connect to server
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: gobservability-agent
  namespace: gobservability
spec:
  podSelector:
    matchLabels:
      app: gobservability-agent
  policyTypes:
  - Egress
  egress:
  - to:
    - podSelector:
        matchLabels:
          app: gobservability-server
    ports:
    - protocol: TCP
      port: 9090
  - to:
    - podSelector:
        matchLabels:
          app: gobservability-postgres
    ports:
    - protocol: TCP
      port: 5432
```

---

## Storage Configuration

### Server (Stateless)
- **In-memory cache**: Metrics stored for 10 seconds
- **No persistent storage required**
- Restart-safe (agents resend metrics)

### PostgreSQL (Stateful)
- **Storage class**: Use fast storage (SSD recommended)
- **Size**: Minimum 5Gi, scale based on alert volume
- **Backup**: Configure CloudNativePG backups

**CloudNativePG Backup Configuration:**
```yaml
apiVersion: postgresql.cnpg.io/v1
kind: Cluster
metadata:
  name: gobservability-postgres
spec:
  instances: 1
  storage:
    size: 5Gi
    storageClass: fast-ssd

  backup:
    barmanObjectStore:
      destinationPath: s3://my-bucket/gobservability-backups
      s3Credentials:
        accessKeyId:
          name: s3-credentials
          key: access-key-id
        secretAccessKey:
          name: s3-credentials
          key: secret-access-key
    retentionPolicy: "30d"
```

---

## Performance Tuning

### Metric Collection Interval

Default: **5 seconds** per agent

**To reduce load:**
```bash
# In agent args
./agent -grpc-server=server:9090 -interval=10s
```

**In Helm values.yaml:**
```yaml
agent:
  interval: 10s  # Change from default 5s
```

### Server Cache TTL

Default: **10 seconds**

Metrics are cached in-memory for 10s before being evicted. Adjust in code if needed:
```go
// cmd/server/storage/store.go
const DefaultTTL = 10 * time.Second
```

### PostgreSQL Connection Pool

Default GORM settings apply. For high alert volume:
```go
// cmd/server/main.go
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(10)
db.SetConnMaxLifetime(5 * time.Minute)
```

---

## Alert Configuration

### Alert Evaluation Interval

Alerts are evaluated **every time metrics are received** (~5 seconds).

### Discord Rate Limiting

To prevent notification spam:
- Notifications are sent **only on state change** (firing ↔ resolved)
- No duplicate notifications for already-firing alerts

### Alert Retention

Alerts are stored in PostgreSQL indefinitely. To clean up old alerts:
```sql
-- Delete alerts older than 90 days
DELETE FROM alerts WHERE created_at < NOW() - INTERVAL '90 days';
```

Or create a Kubernetes CronJob:
```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: cleanup-old-alerts
  namespace: gobservability
spec:
  schedule: "0 2 * * 0"  # Every Sunday at 2 AM
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: cleanup
            image: postgres:16-alpine
            env:
            - name: PGPASSWORD
              valueFrom:
                secretKeyRef:
                  name: postgres-user-secret
                  key: password
            command:
            - psql
            - -h
            - gobservability-postgres-rw
            - -U
            - gobs
            - -d
            - gobservability
            - -c
            - "DELETE FROM alerts WHERE created_at < NOW() - INTERVAL '90 days';"
          restartPolicy: OnFailure
```

---

## Development Configuration

### Local Development (without Kubernetes)

**Start PostgreSQL:**
```bash
docker-compose up -d postgres
```

**Start server:**
```bash
export POSTGRES_URL="postgres://gobs:gobs123@localhost:5432/gobservability?sslmode=disable"
export DISCORD_WEBHOOK_URL="https://discord.com/api/webhooks/..."
./server -port=8080 -grpc-port=9090 -mode=debug
```

**Start agent (dev mode):**
```bash
./agent -grpc-server=localhost:9090 -interval=5s -hostname=my-dev-node -dev
```

### Rebuilding Protocol Buffers

When modifying `proto/gobservability.proto`:
```bash
make install-proto-deps  # One-time setup
make proto               # Regenerate Go code
```
