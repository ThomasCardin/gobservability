# Gobservability

A real-time monitoring system for Kubernetes that collects and displays performance metrics for nodes and pods through a modern web interface.

## What does Gobservability do?

Gobservability consists of two main components. The agent runs as a collector that gathers system metrics every 5 seconds from Kubernetes nodes, monitors pod metrics running on each node, and sends this data to the central server via REST API while operating in containerized mode within the Kubernetes cluster. The web server provides a real-time interface that automatically updates every 2 seconds, featuring a node dashboard with cluster overview, detailed node views showing all pods, in-depth process analysis with complete per-pod metrics, and visual animations to indicate real-time value changes.

## Metrics Sources

All metrics are collected directly from the Linux `/proc` filesystem. Here are the exact sources:

### Node Metrics (System)

#### CPU - `/proc/stat`

- **User time**: CPU time spent in user mode
- **System time**: CPU time spent in kernel mode  
- **Nice time**: CPU time for processes with modified priority
- **Idle time**: CPU idle time
- **IRQ time**: CPU time handling hardware interrupts
- **SoftIRQ time**: CPU time handling software interrupts

#### Memory - `/proc/meminfo`

- **MemTotal**: Total physical memory
- **MemFree**: Available free memory
- **MemAvailable**: Memory available for new applications
- **Buffers**: Memory used for buffers
- **Cached**: Memory used for cache
- **SwapTotal/SwapFree**: Total/free swap space

#### Network - `/proc/net/dev`

- **Bytes received/transmitted** per network interface
- **Packets received/transmitted** per network interface
- **Network errors** received/transmitted per interface
- **Network drops** received/transmitted per interface

#### Disk - `/proc/diskstats`

- **Sectors read/written** per storage device
- **Read/write operations** per device
- **Time spent** reading/writing per device

### Pod Metrics (Process)

For each pod/process identified via the Kubernetes API, the following metrics are collected:

#### Process CPU - `/proc/{PID}/stat`

- **utime** (field 14): Process user CPU time
- **stime** (field 15): Process system CPU time
- **cutime** (field 16): Children processes user CPU time
- **cstime** (field 17): Children processes system CPU time
- **priority** (field 18): Process priority
- **nice** (field 19): Process nice value
- **threads** (field 20): Number of threads
- **starttime** (field 22): Process start time
- **processor** (field 39): CPU the process is scheduled on

#### Process Memory - `/proc/{PID}/status`

- **VmSize**: Total virtual memory size
- **VmRSS**: Resident memory size (physical memory used)
- **VmPeak**: Peak virtual memory used
- **VmLck**: Locked memory
- **VmPin**: Pinned memory
- **voluntary_ctxt_switches**: Voluntary context switches
- **nonvoluntary_ctxt_switches**: Forced context switches

#### Process Disk I/O - `/proc/{PID}/io`

- **read_bytes**: Bytes read from storage
- **write_bytes**: Bytes written to storage
- **cancelled_write_bytes**: Cancelled write bytes

#### Process Network - `/proc/{PID}/net/dev`

- **bytes**: Bytes received/transmitted by the process
- **packets**: Packets received/transmitted by the process
- **errs**: Network errors received/transmitted
- **drop**: Dropped packets received/transmitted

#### Process System Information - `/proc/{PID}/status`

- **Seccomp**: System call filtering mode
- **Speculation_Store_Bypass**: Speculative vulnerability protection
- **Cpus_allowed_list**: CPUs allowed for this process
- **Mems_allowed_list**: Memory nodes allowed

## Architecture

```
┌─────────────────┐    HTTP POST     ┌─────────────────┐
│  Agent (Node)   │ ────────────────▶ │  Web Server     │
│                 │   /api/stats      │                 │
│ • Collect /proc │                   │ • Web Interface │
│ • K8s API       │                   │ • Real-time     │
│ • Every 5s      │                   │ • HTMX Updates  │
└─────────────────┘                   └─────────────────┘
```

## Interface Features

The interface provides a responsive dashboard with adaptive grid layout, real-time updates every 2 seconds via HTMX, and visual animations for value changes. Users can access detailed process views with complete metrics dumps, browse organized metrics by categories including CPU, Memory, Network, and Disk, all presented through a modern GitHub-style dark theme interface with smooth navigation between nodes, pods, and processes.

## Metrics Calculations

Percentages and displayed values are calculated in real-time using various methods. CPU percentages are based on time deltas between collections, memory percentages use the used-to-total system ratio, while network and disk throughput (MB/s) are calculated based on byte deltas and sector read/write deltas respectively.

The complete system provides a comprehensive and detailed view of your Kubernetes cluster performance in real-time!

## Local Testing & Development

### Prerequisites

- Go 1.19+ installed
- Docker (for Kubernetes deployment)
- kubectl configured (for Kubernetes deployment)
- Skaffold (optional, for automated Kubernetes deployment)

### Local Development Mode

For local testing and development, use the provided Makefile targets:

#### 1. Single Agent Testing
```bash
make agent
```
This command will:
- Build both agent and server binaries
- Start the web server on port 8080 in the background
- Start a single agent in development mode (`-dev` flag)
- Access the interface at http://localhost:8080

#### 2. Multi-Agent Testing (Recommended)
```bash
make agents
```
This command simulates a cluster with:
- 7 different agents with various hostnames (`node-01`, `agent-02`, `worker-03`, `controlplane`, `gpunode`, `aiworkloadsonly`, `node-07`)
- Each agent collects metrics every 5 seconds
- All agents run in development mode (uses fake pod data for demonstration)
- Web interface available at http://localhost:8080

#### 3. Stop All Processes
```bash
make stop
```
Cleanly stops all running gobservability processes.

#### 4. Build Only
```bash
make build
```
Builds the agent and server binaries without running them.

### Development Mode Features

When running with `-dev` flag, the agent uses:
- Fake Kubernetes pod data for demonstration
- Local `/proc` filesystem metrics (real system metrics)
- No actual Kubernetes API calls
- Perfect for testing the interface and metrics collection

### Troubleshooting Local Setup

- **Port conflicts**: Ensure port 8080 is available
- **Permission issues**: The agent needs to read `/proc` filesystem
- **Process cleanup**: Use `make stop` before restarting if processes hang

## Kubernetes Cluster Deployment

### Prerequisites

- Kubernetes cluster access with `kubectl` configured
- Appropriate RBAC permissions for the agent to access Kubernetes API
- Container registry access (for custom images)

### Option 1: Manual Kubernetes Deployment

#### 1. Apply Kubernetes Manifests
```bash
# Create namespace and RBAC
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/rbac.yaml

# Deploy server
kubectl apply -f k8s/server-deployment.yaml

# Deploy agent as DaemonSet (runs on every node)
kubectl apply -f k8s/agent-daemonset.yaml
```

#### 2. Access the Interface
```bash
# Port forward to access the web interface
kubectl port-forward -n gobservability service/gobservability-server 8080:8080
```
Then access http://localhost:8080

#### 3. Monitor Deployment
```bash
# Check pods status
kubectl get pods -n gobservability

# Check agent logs
kubectl logs -n gobservability -l app=gobservability-agent

# Check server logs
kubectl logs -n gobservability -l app=gobservability-server
```

### Option 2: Automated Deployment with Skaffold

#### 1. Configure Skaffold
Edit `skaffold.yaml` and replace `<your-user>` with your container registry username:
```yaml
artifacts:
- image: ghcr.io/yourusername/gobservability-server
- image: ghcr.io/yourusername/gobservability-agent
```

#### 2. Deploy with Skaffold
```bash
# Deploy and build images
make skaffold-run

# Or directly with skaffold
skaffold run
```

#### 3. Clean Up
```bash
# Remove deployment
make skaffold-delete

# Or directly with skaffold
skaffold delete
```

### Kubernetes Features

In a real Kubernetes cluster:
- **DaemonSet**: Agent runs on every node automatically
- **Real pod detection**: Agents discover and monitor actual pods via Kubernetes API
- **RBAC**: Proper permissions for agents to access Kubernetes resources
- **Service discovery**: Server is accessible via Kubernetes service
- **Multi-node monitoring**: Complete cluster visibility
- **Real-time updates**: Live pod lifecycle tracking (creation, deletion, restarts)

### Monitoring Production Clusters

For production use:
- Configure resource limits in Kubernetes manifests
- Set up proper RBAC with minimal required permissions  
- Consider persistent storage for historical metrics
- Configure ingress for external access
- Set up alerts for agent failures
- Monitor server resource usage

The system provides complete real-time visibility into your Kubernetes cluster performance!