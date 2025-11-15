# criprof

Container Runtime Interface (CRI) profiling and introspection library and CLI for Go.

**criprof** detects and identifies container runtime environments, orchestration platforms, and image formats using environmental hints. It provides structured information about the executing container environment for debugging, telemetry, and runtime optimization.

## Features

- **Multi-Runtime Detection**: Docker, Podman, containerd, CRI-O, rkt, LXD, OpenVZ, Firecracker, Kata Containers, gVisor, Sysbox, Singularity/Apptainer, WebAssembly
- **Orchestrator Detection**: Kubernetes, Docker Swarm, HashiCorp Nomad, Apache Mesos, AWS ECS, AWS Fargate, Google Cloud Run, AWS Lambda, Azure Container Instances
- **Image Format Detection**: Docker, OCI, CRI, ACI, Singularity/Apptainer
- **Confidence Scoring**: Each detection includes a confidence score (0.0-1.0) indicating detection certainty
- **Non-Invasive**: Read-only detection using filesystem markers, environment variables, and process information
- **Extensible**: Plugin architecture allows custom detectors
- **Fast**: Priority-based execution with optional caching (5-minute TTL)
- **CLI & Library**: Use as a Go library or standalone command-line tool

## Installation

### As a Library

```bash
go get github.com/christianvozar/criprof
```

### As a CLI Tool

```bash
go install github.com/christianvozar/criprof/cmd/criprof@latest
```

Or build from source:

```bash
git clone https://github.com/christianvozar/criprof.git
cd criprof
go build -o criprof ./cmd/criprof
```

## Usage

### Library Usage

#### Basic Detection

```go
package main

import (
    "fmt"
    "github.com/christianvozar/criprof"
)

func main() {
    // Create inventory with default detectors
    inventory := criprof.New()

    // Access individual fields
    fmt.Printf("Runtime: %s\n", inventory.Runtime)
    fmt.Printf("Scheduler: %s\n", inventory.Scheduler)
    fmt.Printf("Image Format: %s\n", inventory.ImageFormat)
    fmt.Printf("Container ID: %s\n", inventory.ID)
    fmt.Printf("Hostname: %s\n", inventory.Hostname)
    fmt.Printf("PID: %d\n", inventory.PID)

    // Export as JSON
    fmt.Println(inventory.JSON())
}
```

#### Advanced Usage with Context and Timeout

```go
package main

import (
    "context"
    "fmt"
    "time"
    "github.com/christianvozar/criprof"
)

func main() {
    // Create context with 2-second timeout
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()

    // Detect with timeout
    inventory := criprof.NewWithContext(ctx)
    fmt.Println(inventory.JSON())
}
```

#### Custom Detectors and Configuration

```go
package main

import (
    "context"
    "fmt"
    "time"
    "github.com/christianvozar/criprof"
)

func main() {
    // Create custom engine with specific detectors
    engine := criprof.NewEngine(criprof.EngineConfig{
        Detectors:     criprof.FastDetectors(), // No network I/O
        EnableCaching: true,
        CacheTTL:      10 * time.Minute,
    })

    // Use custom engine
    inventory, err := criprof.NewWithEngine(context.Background(), engine)
    if err != nil {
        panic(err)
    }

    fmt.Println(inventory.JSON())
}
```

#### Cache Management

```go
package main

import (
    "fmt"
    "github.com/christianvozar/criprof"
)

func main() {
    // First call - performs detection
    inventory1 := criprof.New()
    fmt.Println("First call:", inventory1.JSON())

    // Second call - uses cache (default 5-minute TTL)
    inventory2 := criprof.New()
    fmt.Println("Second call (cached):", inventory2.JSON())

    // Invalidate cache manually
    criprof.InvalidateCache()

    // Third call - re-detects
    inventory3 := criprof.New()
    fmt.Println("Third call (re-detected):", inventory3.JSON())
}
```

### Command-Line Usage

#### Display Runtime Information

```bash
criprof hints
```

**Example Output:**
```json
{
  "hostname": "web-server-5d7c8f9b-xj2k9",
  "id": "8f4b2d1a3c5e",
  "image_format": "docker",
  "pid": 1,
  "runtime": "docker",
  "scheduler": "kubernetes"
}
```

#### Version Information

```bash
criprof version
```

#### Use in Shell Scripts

```bash
#!/bin/bash

# Detect runtime and take action
RUNTIME=$(criprof hints | jq -r '.runtime')

if [ "$RUNTIME" = "kubernetes" ]; then
    echo "Running in Kubernetes"
    # Kubernetes-specific logic
elif [ "$RUNTIME" = "docker" ]; then
    echo "Running in Docker"
    # Docker-specific logic
fi
```

## Compilation and Deployment

### Build for Current Platform

```bash
go build -o criprof ./cmd/criprof
```

### Cross-Compilation

```bash
# Linux AMD64
GOOS=linux GOARCH=amd64 go build -o criprof-linux-amd64 ./cmd/criprof

# Linux ARM64
GOOS=linux GOARCH=arm64 go build -o criprof-linux-arm64 ./cmd/criprof

# Darwin (macOS) AMD64
GOOS=darwin GOARCH=amd64 go build -o criprof-darwin-amd64 ./cmd/criprof

# Darwin (macOS) ARM64 (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o criprof-darwin-arm64 ./cmd/criprof

# Windows AMD64
GOOS=windows GOARCH=amd64 go build -o criprof-windows-amd64.exe ./cmd/criprof
```

### Static Binary (for Alpine/Scratch containers)

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags '-extldflags "-static"' -o criprof ./cmd/criprof
```

## Sidecar Pattern

criprof is ideal for use as a sidecar container to provide runtime introspection to your application pods.

### Kubernetes Sidecar Example

**1. Create Dockerfile for sidecar:**

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /build
COPY . .
RUN CGO_ENABLED=0 go build -o criprof ./cmd/criprof

FROM scratch
COPY --from=builder /build/criprof /criprof
ENTRYPOINT ["/criprof"]
CMD ["hints"]
```

**2. Build and push the image:**

```bash
docker build -t your-registry/criprof:latest .
docker push your-registry/criprof:latest
```

**3. Add sidecar to your Kubernetes deployment:**

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
spec:
  replicas: 3
  selector:
    matchLabels:
      app: myapp
  template:
    metadata:
      labels:
        app: myapp
    spec:
      containers:
      # Main application container
      - name: app
        image: your-registry/myapp:latest
        ports:
        - containerPort: 8080

      # criprof sidecar
      - name: criprof
        image: your-registry/criprof:latest
        command: ["/bin/sh", "-c"]
        args:
          - |
            while true; do
              /criprof hints > /shared/runtime-info.json
              sleep 300
            done
        volumeMounts:
        - name: shared-data
          mountPath: /shared

      volumes:
      - name: shared-data
        emptyDir: {}
```

**4. Access runtime info from main container:**

Your main application can read `/shared/runtime-info.json` to get runtime information updated every 5 minutes.

### Docker Compose Sidecar Example

```yaml
version: '3.8'

services:
  app:
    image: your-app:latest
    volumes:
      - shared-data:/shared

  criprof:
    image: your-registry/criprof:latest
    command: >
      sh -c "while true; do
        /criprof hints > /shared/runtime-info.json;
        sleep 300;
      done"
    volumes:
      - shared-data:/shared

volumes:
  shared-data:
```

## Detection Methods

criprof uses multiple detection strategies with confidence scoring:

### Runtime Detection

| Runtime | Detection Method | Confidence |
|---------|-----------------|------------|
| Docker | `/.dockerenv` file | 0.95 |
| Docker | `/proc/self/cgroup` contains "docker" | 0.90 |
| Podman | `/run/.containerenv` contains "podman" | 0.95 |
| Podman | Generic `/run/.containerenv` | 0.70 |
| CRI-O | `/var/run/crio/crio.sock` exists | 0.95 |
| containerd | `/run/containerd/containerd.sock` | 0.90 |
| Firecracker | DMI product name = "Firecracker" | 0.98 |
| Kata | `/run/kata-containers` exists | 0.95 |
| gVisor | `/proc/self/cgroup` contains "gvisor" | 0.95 |
| Sysbox | `SYSBOX_CONTAINER` env var | 0.98 |
| Singularity | `SINGULARITY_CONTAINER` env var | 0.99 |
| WASM | `GOOS=js GOARCH=wasm` | 1.00 |

### Scheduler Detection

| Scheduler | Detection Method | Confidence |
|-----------|-----------------|------------|
| Kubernetes | Service account token exists | 0.99 |
| Kubernetes | `KUBERNETES_SERVICE_HOST` env var | 0.95 |
| AWS ECS | `ECS_CONTAINER_METADATA_URI` env var | 0.98 |
| AWS Fargate | `AWS_EXECUTION_ENV` contains "fargate" | 0.99 |
| Cloud Run | `K_SERVICE` env var | 0.98 |
| Lambda | `AWS_LAMBDA_FUNCTION_NAME` env var | 0.99 |
| Azure ACI | `ACI_RESOURCE_GROUP` env var | 0.98 |
| Nomad | `NOMAD_TASK_DIR` env var | 0.95 |
| Mesos | `MESOS_TASK_ID` env var | 0.95 |
| Swarm | Docker Swarm port reachable | 0.80 |

### Image Format Detection

| Format | Detection Method | Confidence |
|--------|-----------------|------------|
| Docker | Docker runtime detected | 0.95 |
| OCI | `/var/lib/containers` exists | 0.80 |
| CRI | CRI-O runtime detected | 0.90 |
| ACI | `AC_METADATA_URL` env var | 0.95 |
| Singularity | `SINGULARITY_CONTAINER` env var | 0.95 |

## Architecture

criprof uses a **Strategy Pattern** for detection:

```
┌──────────────┐
│   Detector   │  (Interface)
│  Interface   │
└──────┬───────┘
       │
       ├── DockerFileDetector (priority: 100)
       ├── PodmanDetector (priority: 95)
       ├── KubernetesServiceAccountDetector (priority: 95)
       ├── FirecrackerDetector (priority: 90)
       ├── ...
       └── KubernetesAPIDetector (priority: 10, network I/O)

┌──────────────┐
│    Engine    │  Runs detectors in priority order,
│              │  keeps highest confidence per type
└──────────────┘

┌──────────────┐
│  Detection   │  Contains:
│    Cache     │  - Type (runtime/scheduler/image)
│              │  - Value (e.g., "docker", "kubernetes")
└──────────────┘  - Confidence (0.0-1.0)
                  - Source (detector name)
```

**Priority Levels:**
- **90-100**: Fast filesystem/environment checks
- **70-89**: Cgroup parsing, moderate I/O
- **10-40**: Network operations (with 2s timeout)

**Confidence Scoring:**
- **0.95-1.00**: Definitive markers (e.g., specific env vars, unique files)
- **0.80-0.94**: Strong indicators (e.g., socket files, API responses)
- **0.60-0.79**: Weak indicators (e.g., generic files, pattern matches)

## Contributing

Contributions are welcome! Here's how you can help:

### Adding New Detectors

1. **Identify detection markers** for the runtime/scheduler/image format
2. **Determine confidence level** based on marker uniqueness
3. **Implement the Detector interface:**

```go
type MyRuntimeDetector struct {
    fs FileSystem
}

func (d *MyRuntimeDetector) Name() string {
    return "my-runtime-detector"
}

func (d *MyRuntimeDetector) Priority() int {
    return 90 // Adjust based on operation cost
}

func (d *MyRuntimeDetector) Detect(ctx context.Context) (*Detection, error) {
    // Check for runtime markers
    if _, err := d.fs.Stat("/path/to/marker"); err == nil {
        return &Detection{
            Type:       DetectionTypeRuntime,
            Value:      "my-runtime",
            Confidence: 0.95,
            Source:     d.Name(),
        }, nil
    }
    return nil, nil
}
```

4. **Add tests** in `*_test.go` with mock filesystem
5. **Update `DefaultDetectors()`** in `api.go`
6. **Submit a pull request**

### Reporting Issues

If you discover a detection issue or missing runtime support:

1. Open an issue with:
   - Runtime/scheduler/platform name
   - Detection markers (env vars, files, etc.)
   - Example environment where it runs
   - Expected vs. actual detection results

2. Include output from `criprof hints` if possible

### Development Setup

```bash
# Clone repository
git clone https://github.com/christianvozar/criprof.git
cd criprof

# Install dependencies
go mod download

# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Build CLI
go build ./cmd/criprof

# Run CLI
./criprof hints
```

### Code Guidelines

- Follow Go best practices and idioms
- Add godoc comments for public APIs
- Include tests for all new detectors
- Use mock filesystem/network for tests
- Keep confidence scores realistic
- Prioritize non-invasive detection methods
- Avoid external dependencies when possible

## Use Cases

- **Multi-Runtime Applications**: Detect which container runtime is executing your code and adjust behavior accordingly
- **Debugging**: Identify container environment issues by inspecting runtime information
- **Telemetry**: Include runtime metadata in logs and metrics for better observability
- **Security**: Detect unexpected runtime environments that may indicate compromise
- **Testing**: Verify applications run correctly across different container platforms
- **CI/CD**: Conditional logic based on deployment target (Kubernetes vs. ECS vs. Cloud Run)
- **Sidecar Monitoring**: Continuously monitor and report container runtime information

## License

MIT License - see [LICENSE](LICENSE) for details

## Author

Christian R. Vozar

## Acknowledgments

Inspired by the [Ohai](https://github.com/chef/ohai) project's system profiling approach.
