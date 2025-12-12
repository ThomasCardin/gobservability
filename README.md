# Gobservability

**A lightweight, real-time Kubernetes monitoring platform built in Go**

Gobservability is a cloud-native observability system designed specifically for Kubernetes clusters. It provides deep visibility into node and pod performance through direct `/proc` filesystem monitoring, offering granular metrics collection without the overhead of traditional monitoring solutions.

## 📖 Project Overview

Gobservability is a two-tier monitoring system consisting of:

- **Agent (DaemonSet)**: Runs on every Kubernetes node, collecting system and process-level metrics every 5 seconds from the `/proc` filesystem and Kubernetes API
- **Server (Deployment)**: Central aggregator that receives metrics via gRPC streaming, stores alert configurations in PostgreSQL, and provides a real-time web interface powered by HTMX

Unlike heavyweight monitoring solutions, Gobservability is purpose-built for Kubernetes with minimal resource footprint, using native Linux kernel interfaces for accurate, low-overhead metrics collection.

### Key Design Principles

- **Direct `/proc` access**: No kernel modules or eBPF required - pure userspace monitoring
- **gRPC streaming**: Efficient bidirectional communication between agents and server
- **Stateless agents**: Agents are ephemeral and discover pods dynamically via Kubernetes API
- **Real-time UI**: Auto-refreshing dashboard with HTMX (no frontend framework bloat)
- **Flamegraph integration**: On-demand CPU profiling using `perf` tools

## ✨ Features

### 1. Real-Time Metrics Collection

- **Node-Level Metrics** (from `/proc/stat`, `/proc/meminfo`, `/proc/net/dev`, `/proc/diskstats`)
  - CPU usage breakdown (user, system, nice, idle, IRQ, SoftIRQ)
  - Memory utilization (total, free, available, buffers, cached, swap)
  - Network throughput (bytes, packets, errors, drops per interface)
  - Disk I/O (read/write sectors, operations, latency per device)

- **Pod/Process-Level Metrics** (from `/proc/{PID}/...`)
  - Per-pod CPU time (user, system, children, priority, nice value)
  - Per-pod memory (VmSize, VmRSS, VmPeak, context switches)
  - Per-pod disk I/O (read/write bytes, cancelled writes)
  - Per-pod network statistics (bytes, packets, errors, drops)
  - Process system info (Seccomp, CPU affinity, memory nodes)

- **5-second collection interval** with configurable retention

### 2. Dynamic Alert System

- **Flexible Rule Configuration**
  - Create alerts for **nodes** or individual **pods**
  - Monitor any metric: CPU, Memory, Network, Disk
  - Configurable thresholds with **greater than (>)** or **less than (<)** conditions
  - Enable/disable rules without deletion

- **Alert Lifecycle Management**
  - Automatic alert firing when thresholds are exceeded
  - Automatic resolution when metrics return to normal
  - Manual alert dismissal via UI
  - Cannot modify/delete rules with active alerts (prevents accidental data loss)

- **Discord Notifications**
  - Real-time webhook notifications for alert events
  - Alert firing notifications (includes metric value, threshold, timestamp)
  - Alert resolved notifications (automatic or manual dismissal)
  - Rate limiting to prevent notification spam

- **Alert History**
  - PostgreSQL-backed persistent storage
  - Query historical alerts by node (configurable time window)
  - Track alert status: `firing`, `resolved`, timestamps

### 3. CPU Flamegraph Profiling

- **On-Demand Profiling**
  - Generate CPU flamegraphs for any running pod
  - Uses Linux `perf` tool for accurate call stack sampling
  - Configurable profiling duration (30-600 seconds)
  - JSON output format for interactive visualization

- **Asynchronous Task Management**
  - Non-blocking flamegraph generation (returns task ID immediately)
  - Poll task status endpoint for completion
  - Download completed flamegraphs via REST API

- **Privileged Container Support**
  - Agent runs with `SYS_ADMIN`, `SYS_PTRACE`, `SYS_RAWIO` capabilities
  - Required for `perf` profiling across process boundaries

### 4. Modern Web Interface

- **Dashboard Features**
  - Cluster overview with all nodes
  - Real-time metric updates every 2 seconds (HTMX polling)
  - Visual animations for value changes
  - Adaptive grid layout (responsive design)

- **Navigation Hierarchy**
  - **Nodes page**: Cluster-wide overview
  - **Pods page**: All pods running on a specific node
  - **Process details page**: Deep dive into individual pod metrics
  - **Alerts page**: Configure and view alert rules per node
  - **Flamegraph page**: Interactive CPU profiling visualization

- **GitHub-Style Dark Theme**
  - Monospace fonts for technical data
  - Clean, minimal interface
  - Smooth transitions and animations

### 5. gRPC Bidirectional Streaming

- **Efficient Communication Protocol**
  - Agents stream metrics to server via gRPC (port 9090)
  - Server sends commands to agents (e.g., flamegraph generation requests)
  - Protocol Buffers for compact serialization
  - Connection pooling and automatic reconnection

- **Agent Discovery**
  - Agents identify themselves by node name (from Kubernetes `spec.nodeName`)
  - Server maintains active agent registry
  - Supports dynamic agent scaling (DaemonSet auto-scaling)

### 6. Kubernetes-Native Architecture

- **Agent Deployment (DaemonSet)**
  - Runs on every cluster node automatically
  - Host PID namespace access (`hostPID: true`) for `/proc` visibility
  - Read-only mounts for `/proc` and `/sys` filesystems
  - ServiceAccount with RBAC for Kubernetes API access (pod discovery)

- **Server Deployment**
  - Stateless server (metrics cached in-memory, 10s TTL)
  - Horizontal scaling ready (share PostgreSQL for alerts)
  - ClusterIP service for internal communication
  - Optional Ingress for external web access

- **PostgreSQL Database (CloudNativePG)**
  - Stores alert rules and history
  - GORM ORM with automatic migrations
  - UUID extension support
  - Configurable storage class

### 7. Multiple Deployment Options

- **Production Kubernetes**
  - Helm chart in `k8s/helm/` (customizable via `values.yaml`)
  - Nginx Ingress with Let's Encrypt TLS
  - Resource limits and requests pre-configured
  - Multi-platform images (amd64, arm64)

- **Local Development**
  - `docker-compose.yml` for full stack (PostgreSQL + Server + Agent)
  - `make agent` for single-agent testing with fake data
  - `make agents` for multi-agent simulation (7 fake nodes)
  - No Kubernetes cluster required for development

- **CI/CD Integration**
  - GitHub Actions workflow for image builds
  - Skaffold configuration for automated deployments
  - Multi-arch image support via Docker Buildx

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
┌──────────────────────────────────────────────────────────────────┐
│                       Kubernetes Cluster                         │
│                                                                  │
│  ┌────────────────┐         gRPC Stream         ┌─────────────┐ │
│  │ Agent (Node 1) │ ◀──────────────────────────▶ │             │ │
│  │                │     Bidirectional            │   Server    │ │
│  │ • /proc read   │                              │             │ │
│  │ • K8s API      │                              │ • gRPC :9090│ │
│  │ • Flamegraph   │                              │ • HTTP :8080│ │
│  └────────────────┘                              │ • HTMX UI   │ │
│                                                  │ • Alerts    │ │
│  ┌────────────────┐                              └──────┬──────┘ │
│  │ Agent (Node 2) │ ◀──────────────────────────────────┤        │
│  └────────────────┘                                     │        │
│                                                         │        │
│  ┌────────────────┐                              ┌─────▼──────┐ │
│  │ Agent (Node N) │ ◀────────────────────────────│ PostgreSQL │ │
│  └────────────────┘                              │            │ │
│                                                  │ • Alerts   │ │
│  DaemonSet (runs on every node)                  │ • History  │ │
│                                                  └────────────┘ │
└──────────────────────────────────────────────────────────────────┘
```

**Data Flow:**
1. Agents collect metrics from `/proc` and Kubernetes API every 5 seconds
2. Metrics streamed to server via gRPC bidirectional connection
3. Server stores data in-memory cache (10s TTL) and persists alerts to PostgreSQL
4. Web UI polls server every 2 seconds via HTMX for real-time updates
5. Server can send commands to agents (e.g., flamegraph generation)

## Metrics Calculations

Percentages and displayed values are calculated in real-time using various methods. CPU percentages are based on time deltas between collections, memory percentages use the used-to-total system ratio, while network and disk throughput (MB/s) are calculated based on byte deltas and sector read/write deltas respectively.

---

## 🚀 Quick Start

### Docker Compose (Recommended for Local Testing)

```bash
# Start full stack (PostgreSQL + Server + Agent)
docker-compose up -d

# Access web interface
open http://localhost:8080

# Stop all services
docker-compose down
```

### Development Mode

```bash
# Simulate multi-node cluster (7 fake nodes)
make agents

# Access interface at http://localhost:8080
```

### Kubernetes (Production)

```bash
# Install with Helm
helm install gobservability ./k8s/helm \
  --namespace gobservability \
  --create-namespace \
  --values values.yaml
```

**For detailed installation instructions, see:** **[📖 Deployment Guide](docs/deployment.md)**

---

## 📚 Documentation

### Core Guides

- **[📖 Deployment Guide](docs/deployment.md)** - Complete deployment instructions
  - Docker Compose for local development
  - Makefile development mode
  - Kubernetes Helm deployment (production)
  - Skaffold automated workflow
  - Building custom images

- **[⚙️ Configuration Reference](docs/configuration.md)** - All configuration options
  - Environment variables
  - Resource requirements
  - Security settings (RBAC, capabilities, secrets)
  - Network configuration
  - Performance tuning

- **[🐛 Troubleshooting Guide](docs/troubleshooting.md)** - Common issues and solutions
  - Agent not collecting metrics
  - Server connection issues
  - Alerts not firing
  - Flamegraph generation failures
  - Performance problems

### Additional Resources

- [Protocol Buffers Definition](proto/gobservability.proto) - gRPC API schema
- [Kubernetes Manifests](k8s/helm/templates/) - Helm chart templates
- [CloudNativePG Documentation](https://cloudnative-pg.io/) - PostgreSQL operator
- [HTMX Documentation](https://htmx.org/) - Web interface framework

---

## 🛠️ Development

### Prerequisites

- Go 1.24+ installed
- Docker (for building images)
- Protocol Buffers compiler (for proto files)

### Building from Source

```bash
# Clone repository
git clone https://github.com/ThomasCardin/gobservability.git
cd gobservability

# Build binaries
make build

# Run local development environment
make agents
```

### Compiling Protocol Buffers

```bash
# Install protoc dependencies (one-time)
make install-proto-deps

# Generate Go code from proto files
make proto
```

### Contributing

Contributions welcome! Please:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests (if applicable)
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

**Development Setup:**
```bash
# Start PostgreSQL for local development
docker-compose up -d postgres

# Start server in debug mode
export GIN_MODE=debug
export POSTGRES_URL="postgres://gobs:gobs123@localhost:5432/gobservability?sslmode=disable"
./server -port=8080 -grpc-port=9090

# Start agent in dev mode
./agent -grpc-server=localhost:9090 -dev -hostname=dev-node
```

---

## 📝 License

This project is licensed under the **GNU General Public License v3.0** - see the [LICENSE](LICENSE) file for details.

**What this means:**
- ✅ You can use, modify, and distribute this software
- ✅ You must disclose the source code of any modifications
- ✅ You must license derivative works under GPL-3.0
- ✅ Commercial use is allowed

---

## 🤝 Support

**Found a bug or have a feature request?**
- Open an issue: https://github.com/ThomasCardin/gobservability/issues

**Need help deploying?**
- Check the [Deployment Guide](docs/deployment.md)
- Check the [Troubleshooting Guide](docs/troubleshooting.md)

---

**Built with ❤️ using Go, gRPC, HTMX, and Kubernetes**
