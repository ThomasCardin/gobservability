# Troubleshooting Guide

Common issues and solutions for Gobservability.

---

## Agent Issues

### Agent Not Collecting Metrics

**Symptoms:**
- No nodes showing up in dashboard
- Dashboard shows empty cluster

**Possible Causes & Solutions:**

#### 1. Agent Not Running
```bash
# Check agent pods
kubectl get pods -n gobservability -l app=gobservability-agent

# Expected: One pod per node in Running state
# If missing or not Running, check logs:
kubectl logs -n gobservability -l app=gobservability-agent --tail=50
```

#### 2. RBAC Permissions Missing
```bash
# Verify agent can access Kubernetes API
kubectl auth can-i get pods --as=system:serviceaccount:gobservability:gobservability-agent

# Should return: yes
# If no, apply RBAC manifests:
kubectl apply -f k8s/helm/templates/rbac.yaml
```

#### 3. /proc Mount Issues
```bash
# Verify /proc mount is working
kubectl exec -it -n gobservability gobservability-agent-xxxxx -- ls /host/proc

# Should show: 1, 2, 3, ... (process IDs)
# If error, check DaemonSet hostPath mounts
```

#### 4. gRPC Connection Failed
```bash
# Check server is reachable from agent
kubectl exec -it -n gobservability gobservability-agent-xxxxx -- nc -zv gobservability-server 9090

# Should show: Connection to gobservability-server 9090 port [tcp/*] succeeded!

# Check server logs for connection errors:
kubectl logs -n gobservability -l app=gobservability-server | grep -i grpc
```

---

### Agent Crash Loop

**Symptoms:**
- Agent pods constantly restarting
- `CrashLoopBackOff` status

**Solutions:**

#### Check Logs
```bash
# Get recent logs from crashed pod
kubectl logs -n gobservability gobservability-agent-xxxxx --previous

# Common errors:
# - "permission denied" → Check securityContext capabilities
# - "connection refused" → Check server service name/port
# - "context deadline exceeded" → Network policy blocking traffic
```

#### Verify Security Context
```bash
# Check agent has required capabilities
kubectl get daemonset gobservability-agent -n gobservability -o yaml | grep -A10 securityContext

# Must have:
# privileged: true
# capabilities: SYS_ADMIN, SYS_PTRACE, SYS_RAWIO
```

#### Resource Limits
```bash
# Check if agent is OOMKilled
kubectl describe pod -n gobservability gobservability-agent-xxxxx | grep -i oom

# If yes, increase memory limits in values.yaml:
agent:
  resources:
    limits:
      memory: 128Mi  # Increase from 64Mi
```

---

### Agent Shows "No Pods Found"

**Symptoms:**
- Agent running but not detecting pods
- Empty pod list in dashboard

**Solutions:**

#### Verify Kubernetes API Access
```bash
# Check agent can list pods
kubectl exec -it -n gobservability gobservability-agent-xxxxx -- \
  wget -qO- --header="Authorization: Bearer $(cat /var/run/secrets/kubernetes.io/serviceaccount/token)" \
  https://kubernetes.default.svc/api/v1/pods

# Should return JSON with pod list
```

#### Check Node Name
```bash
# Verify NODE_NAME is correctly set
kubectl exec -it -n gobservability gobservability-agent-xxxxx -- env | grep NODE_NAME

# Should match actual node name:
kubectl get nodes
```

---

## Server Issues

### Server Not Starting

**Symptoms:**
- Server pod in `CrashLoopBackOff`
- Dashboard not accessible

**Solutions:**

#### Check PostgreSQL Connection
```bash
# Check server logs for database errors
kubectl logs -n gobservability -l app=gobservability-server | grep -i postgres

# Common errors:
# - "connection refused" → PostgreSQL not ready
# - "authentication failed" → Wrong credentials
# - "database does not exist" → Database not created
```

#### Verify PostgreSQL is Running
```bash
# Check PostgreSQL pod
kubectl get pods -n gobservability | grep postgres

# If using CloudNativePG:
kubectl get cluster -n gobservability

# Check PostgreSQL logs:
kubectl logs -n gobservability gobservability-postgres-1
```

#### Test Database Connection Manually
```bash
# Connect to PostgreSQL from server pod
kubectl exec -it -n gobservability gobservability-server-xxxxx -- \
  apk add postgresql-client && \
  psql "postgres://gobs:gobs123@gobservability-postgres-rw:5432/gobservability"

# If successful, issue is in server code
# If failed, check PostgreSQL service/credentials
```

---

### Dashboard Shows No Data

**Symptoms:**
- Dashboard loads but no metrics displayed
- Empty node list

**Solutions:**

#### Verify Agents Are Connected
```bash
# Check server logs for gRPC connections
kubectl logs -n gobservability -l app=gobservability-server | grep -i "agent connected"

# Should see: "Agent connected: node-name"
```

#### Check Cache Storage
```bash
# Server uses in-memory cache with 10s TTL
# If no agents have sent data in >10s, cache is empty
# Wait for next metric collection cycle (5s default)
```

#### Network Policy Blocking Traffic
```bash
# Check if NetworkPolicy is blocking agent→server traffic
kubectl get networkpolicy -n gobservability

# If exists, verify it allows:
# - Agent egress to server:9090
# - Server ingress from agents on port 9090
```

---

## Alert Issues

### Alerts Not Firing

**Symptoms:**
- Metrics exceed threshold but no Discord notification
- No alerts shown in alerts page

**Solutions:**

#### Verify Alert Rule is Enabled
```bash
# Navigate to alerts page: http://localhost:8080/alerts/{nodename}
# Check "Enabled" column shows "Yes"
```

#### Check Discord Webhook URL
```bash
# Verify secret exists
kubectl get secret discord-webhook -n gobservability -o yaml

# Decode webhook URL (base64)
kubectl get secret discord-webhook -n gobservability -o jsonpath='{.data.webhook_url}' | base64 -d

# Test webhook manually:
curl -X POST "YOUR_WEBHOOK_URL" \
  -H "Content-Type: application/json" \
  -d '{"content": "Test notification from Gobservability"}'
```

#### Check Server Logs for Alert Evaluation
```bash
# Check alert evaluation logs
kubectl logs -n gobservability -l app=gobservability-server | grep -i alert

# Look for:
# - "Alert fired: ..."
# - "Discord notification sent"
# - "Failed to send Discord notification: ..." (errors)
```

#### Verify PostgreSQL Alert Storage
```bash
# Connect to database and check alerts table
kubectl exec -it -n gobservability gobservability-postgres-1 -- \
  psql -U gobs -d gobservability -c "SELECT * FROM alerts ORDER BY created_at DESC LIMIT 10;"

# If empty, alert evaluation is not working
# Check server logs for database errors
```

---

### Duplicate Alert Notifications

**Symptoms:**
- Receiving multiple Discord notifications for same alert

**Solution:**

This is a bug - alerts should only notify on **state change**. Check server logs:
```bash
kubectl logs -n gobservability -l app=gobservability-server | grep "Discord notification"

# If seeing multiple "sent" messages for same alert, report issue
```

---

### Cannot Delete Alert Rule

**Symptoms:**
- "Cannot delete rule with active alert" error

**Solution:**

This is expected behavior. You must either:
1. **Dismiss the alert manually** via UI
2. **Wait for alert to auto-resolve** (when metric returns below threshold)
3. **Disable the rule** (keeps rule but stops evaluation)

```bash
# Option 1: Dismiss via API
curl -X PUT http://localhost:8080/api/alerts/dismiss/{alert-id}

# Option 2: Disable rule (not delete)
# Navigate to alerts page and click "Disable"
```

---

## Flamegraph Issues

### Flamegraph Generation Fails

**Symptoms:**
- "Flamegraph generation failed" error
- Task status shows "error"

**Solutions:**

#### Verify Agent Has perf Tools
```bash
# Check perf is installed in agent
kubectl exec -it -n gobservability gobservability-agent-xxxxx -- which perf

# Should return: /usr/bin/perf
# If not found, agent image is broken (should use Ubuntu base with perf)
```

#### Check Agent Capabilities
```bash
# Verify SYS_ADMIN, SYS_PTRACE, SYS_RAWIO capabilities
kubectl get daemonset gobservability-agent -n gobservability -o yaml | grep -A5 capabilities

# Must include all three capabilities
```

#### Verify Target Pod Has Valid PID
```bash
# Navigate to process details page
# Check that PID field shows a positive number (not -1)

# If PID is -1, pod has no running process (might be completed/failed)
```

#### Check Agent Logs
```bash
# Look for perf errors
kubectl logs -n gobservability gobservability-agent-xxxxx | grep -i perf

# Common errors:
# - "perf: permission denied" → Missing capabilities
# - "No such process" → PID no longer exists
# - "Cannot attach to process" → Process security context blocks ptrace
```

---

### Flamegraph Task Stuck in "Processing"

**Symptoms:**
- Task status never completes
- No error shown

**Solutions:**

#### Check Task Timeout
```bash
# Default timeout is 10 minutes
# If profiling duration is 600s (10 minutes), task may time out

# Wait for task completion or check agent logs for errors
kubectl logs -n gobservability gobservability-agent-xxxxx | grep flamegraph
```

#### Verify gRPC Connection
```bash
# Flamegraph uses gRPC server-to-agent communication
# Check server can send requests to agents

kubectl logs -n gobservability -l app=gobservability-server | grep -i flamegraph
```

---

## Performance Issues

### High Memory Usage (Server)

**Symptoms:**
- Server pod using >128Mi memory
- OOMKilled events

**Solutions:**

#### Increase Memory Limits
```yaml
# In values.yaml
server:
  resources:
    limits:
      memory: 256Mi  # Double the limit
```

#### Reduce Cache Size
The server caches all metrics in-memory. For large clusters:
- Reduce cache TTL (modify `storage/store.go`)
- Limit number of nodes monitored
- Add metric sampling (collect every 10s instead of 5s)

---

### High Memory Usage (Agent)

**Symptoms:**
- Agent pod using >64Mi memory
- OOMKilled on nodes with many pods

**Solutions:**

#### Increase Memory Limits
```yaml
# In values.yaml
agent:
  resources:
    limits:
      memory: 128Mi
```

#### Reduce Collection Frequency
```yaml
# In values.yaml
agent:
  interval: 10s  # Instead of default 5s
```

---

### High CPU Usage (Agent)

**Symptoms:**
- Agent using >200m CPU constantly

**Solutions:**

#### Normal Behavior
- Agents parse `/proc` files which is CPU-intensive
- Expected on nodes with 50+ pods

#### Optimization Options
- Reduce collection interval (10s instead of 5s)
- Increase CPU limits if acceptable
- Optimize `/proc` parsing code (e.g., skip unused fields)

---

## Network Issues

### Ingress Not Working

**Symptoms:**
- Cannot access dashboard via external domain
- "404 Not Found" or "Service Unavailable"

**Solutions:**

#### Verify Ingress Controller is Installed
```bash
kubectl get pods -n ingress-nginx

# If not installed:
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/cloud/deploy.yaml
```

#### Check Ingress Resource
```bash
kubectl get ingress -n gobservability

# Verify:
# - Host matches your domain
# - Backend service is gobservability-server:8080
```

#### Verify DNS
```bash
# Check domain resolves to cluster
nslookup gobservability.example.com

# Should point to ingress controller LoadBalancer IP
```

#### Check TLS Certificate
```bash
# If using cert-manager
kubectl get certificate -n gobservability

# Should show "Ready: True"
# If not, check cert-manager logs:
kubectl logs -n cert-manager -l app=cert-manager
```

---

### Port Forward Not Working

**Symptoms:**
- `kubectl port-forward` hangs or times out

**Solutions:**

#### Check Service Exists
```bash
kubectl get svc -n gobservability gobservability-server

# Should show ClusterIP service with ports 8080, 9090
```

#### Verify Pod is Running
```bash
kubectl get pods -n gobservability -l app=gobservability-server

# Should show 1/1 Running
```

#### Use Correct Syntax
```bash
# Correct:
kubectl port-forward -n gobservability svc/gobservability-server 8080:8080

# Incorrect:
kubectl port-forward -n gobservability pod/gobservability-server-xxxxx 8080:8080
```

---

## Database Issues

### PostgreSQL Not Starting

**Symptoms:**
- PostgreSQL pod in `Pending` or `CrashLoopBackOff`

**Solutions:**

#### Check PVC is Bound
```bash
kubectl get pvc -n gobservability

# Should show "Bound" status
# If "Pending", check storage class exists:
kubectl get storageclass
```

#### Verify CloudNativePG Operator is Installed
```bash
kubectl get pods -n cnpg-system

# If not installed:
kubectl apply -f https://raw.githubusercontent.com/cloudnative-pg/cloudnative-pg/release-1.24/releases/cnpg-1.24.0.yaml
```

#### Check PostgreSQL Logs
```bash
kubectl logs -n gobservability gobservability-postgres-1

# Look for:
# - "database system is ready to accept connections" (success)
# - "FATAL: ..." (errors)
```

---

### Database Migration Fails

**Symptoms:**
- Server logs show "failed to migrate database"

**Solutions:**

#### Check Database Exists
```bash
# Connect to PostgreSQL
kubectl exec -it -n gobservability gobservability-postgres-1 -- \
  psql -U gobs -c "\l"

# If "gobservability" database missing, create it:
kubectl exec -it -n gobservability gobservability-postgres-1 -- \
  psql -U gobs -c "CREATE DATABASE gobservability;"
```

#### Run Migration Manually
```bash
# Server uses GORM auto-migration
# If fails, check PostgreSQL logs for permission errors
```

---

## Development Issues

### `make agents` Not Working

**Symptoms:**
- Script errors or no agents starting

**Solutions:**

#### Check PostgreSQL is Running
```bash
# Start PostgreSQL first
docker-compose up -d postgres

# Or use local PostgreSQL:
export POSTGRES_URL="postgres://user:pass@localhost:5432/gobservability"
```

#### Check Port 8080 is Free
```bash
# Kill existing server
lsof -ti:8080 | xargs kill -9

# Then retry
make agents
```

---

### Docker Compose Build Fails

**Symptoms:**
- "failed to solve" or build errors

**Solutions:**

#### Check Docker Daemon is Running
```bash
docker ps

# If error, start Docker daemon
```

#### Clean Build Cache
```bash
docker-compose build --no-cache
```

#### Check Go Version
```bash
# Requires Go 1.24+
go version
```

---

## Getting Help

If your issue is not listed here:

1. **Check Logs:**
   ```bash
   kubectl logs -n gobservability -l app=gobservability-server
   kubectl logs -n gobservability -l app=gobservability-agent
   ```

2. **Enable Debug Mode:**
   ```yaml
   # In server deployment
   env:
   - name: GIN_MODE
     value: "debug"
   ```

3. **Report Issue:**
   - Open issue on GitHub: https://github.com/ThomasCardin/gobservability/issues
   - Include logs, Kubernetes version, and deployment method
