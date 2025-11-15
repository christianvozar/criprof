# Proposed Architecture: Plugin-Based Detector System

## Overview

Refactor the hinting system to use a **Strategy Pattern** with pluggable detectors, dependency injection, and lazy evaluation.

## Core Concepts

### 1. Detector Interface

```go
// Detector represents a single detection strategy
type Detector interface {
    // Name returns the detector's identifier
    Name() string

    // Detect runs the detection logic and returns a result with confidence
    Detect(ctx context.Context) (*Detection, error)

    // Priority returns the detector's priority (higher = runs first)
    Priority() int
}

// Detection represents a detection result
type Detection struct {
    Type       DetectionType  // Runtime, Scheduler, ImageFormat, etc.
    Value      string         // "docker", "kubernetes", etc.
    Confidence float64        // 0.0 to 1.0
    Source     string         // Which detector produced this
}

type DetectionType int

const (
    DetectionTypeRuntime DetectionType = iota
    DetectionTypeScheduler
    DetectionTypeImageFormat
)
```

### 2. Filesystem/Network Abstraction

```go
// FileSystem abstracts file operations for testability
type FileSystem interface {
    Stat(name string) (os.FileInfo, error)
    ReadFile(name string) ([]byte, error)
}

// Network abstracts network operations
type Network interface {
    DialTimeout(network, address string, timeout time.Duration) (net.Conn, error)
    HTTPGet(ctx context.Context, url string) (*http.Response, error)
}

// DefaultFS implements FileSystem using os package
type DefaultFS struct{}

func (DefaultFS) Stat(name string) (os.FileInfo, error) {
    return os.Stat(name)
}

func (DefaultFS) ReadFile(name string) ([]byte, error) {
    return os.ReadFile(name)
}
```

### 3. Individual Detectors

```go
// DockerFileDetector checks for Docker marker files
type DockerFileDetector struct {
    fs FileSystem
}

func (d *DockerFileDetector) Name() string {
    return "docker-file-marker"
}

func (d *DockerFileDetector) Priority() int {
    return 100 // File checks are fast, run early
}

func (d *DockerFileDetector) Detect(ctx context.Context) (*Detection, error) {
    // Check /.dockerenv
    if _, err := d.fs.Stat("/.dockerenv"); err == nil {
        return &Detection{
            Type:       DetectionTypeRuntime,
            Value:      "docker",
            Confidence: 0.95, // High confidence - file is definitive
            Source:     d.Name(),
        }, nil
    }

    // Check /.dockerinit (legacy)
    if _, err := d.fs.Stat("/.dockerinit"); err == nil {
        return &Detection{
            Type:       DetectionTypeRuntime,
            Value:      "docker",
            Confidence: 0.90, // Slightly lower - legacy marker
            Source:     d.Name(),
        }, nil
    }

    return nil, nil // No detection
}

// KubernetesServiceAccountDetector checks for K8s service account
type KubernetesServiceAccountDetector struct {
    fs FileSystem
}

func (d *KubernetesServiceAccountDetector) Name() string {
    return "kubernetes-service-account"
}

func (d *KubernetesServiceAccountDetector) Priority() int {
    return 90
}

func (d *KubernetesServiceAccountDetector) Detect(ctx context.Context) (*Detection, error) {
    if _, err := d.fs.Stat("/run/secrets/kubernetes.io/serviceaccount/token"); err == nil {
        return &Detection{
            Type:       DetectionTypeScheduler,
            Value:      "kubernetes",
            Confidence: 0.99, // Very high confidence
            Source:     d.Name(),
        }, nil
    }
    return nil, nil
}

// KubernetesAPIDetector makes network call to K8s API
type KubernetesAPIDetector struct {
    network Network
    timeout time.Duration
}

func (d *KubernetesAPIDetector) Name() string {
    return "kubernetes-api-probe"
}

func (d *KubernetesAPIDetector) Priority() int {
    return 10 // Network checks are slow, run last
}

func (d *KubernetesAPIDetector) Detect(ctx context.Context) (*Detection, error) {
    ctx, cancel := context.WithTimeout(ctx, d.timeout)
    defer cancel()

    resp, err := d.network.HTTPGet(ctx, "http://kubernetes.default.svc")
    if err != nil {
        return nil, nil // No detection, not an error
    }
    resp.Body.Close()

    return &Detection{
        Type:       DetectionTypeScheduler,
        Value:      "kubernetes",
        Confidence: 0.80, // Lower confidence - could be other service
        Source:     d.Name(),
    }, nil
}
```

### 4. Detection Engine

```go
// Engine coordinates multiple detectors
type Engine struct {
    detectors []Detector
    cache     *DetectionCache
    mu        sync.RWMutex
}

type EngineConfig struct {
    Detectors      []Detector
    EnableCaching  bool
    CacheTTL       time.Duration
    SkipNetwork    bool  // Skip expensive network detectors
}

func NewEngine(config EngineConfig) *Engine {
    engine := &Engine{
        detectors: config.Detectors,
    }

    if config.EnableCaching {
        engine.cache = NewDetectionCache(config.CacheTTL)
    }

    // Sort by priority (highest first)
    sort.Slice(engine.detectors, func(i, j int) bool {
        return engine.detectors[i].Priority() > engine.detectors[j].Priority()
    })

    return engine
}

// DetectAll runs all detectors and aggregates results
func (e *Engine) DetectAll(ctx context.Context) (*Inventory, error) {
    // Check cache first
    if e.cache != nil {
        if cached := e.cache.Get(); cached != nil {
            return cached, nil
        }
    }

    results := make(map[DetectionType]*Detection)

    for _, detector := range e.detectors {
        select {
        case <-ctx.Done():
            return nil, ctx.Err()
        default:
        }

        detection, err := detector.Detect(ctx)
        if err != nil {
            // Log error but continue with other detectors
            continue
        }

        if detection == nil {
            continue
        }

        // Keep detection with highest confidence
        existing := results[detection.Type]
        if existing == nil || detection.Confidence > existing.Confidence {
            results[detection.Type] = detection
        }
    }

    inventory := e.buildInventory(results)

    // Cache result
    if e.cache != nil {
        e.cache.Set(inventory)
    }

    return inventory, nil
}

func (e *Engine) buildInventory(results map[DetectionType]*Detection) *Inventory {
    inv := &Inventory{
        PID:         os.Getpid(),
        Runtime:     "undetermined",
        Scheduler:   "undetermined",
        ImageFormat: "undetermined",
    }

    if det := results[DetectionTypeRuntime]; det != nil {
        inv.Runtime = det.Value
    }
    if det := results[DetectionTypeScheduler]; det != nil {
        inv.Scheduler = det.Value
    }
    if det := results[DetectionTypeImageFormat]; det != nil {
        inv.ImageFormat = det.Value
    }

    // Get hostname
    if h, err := os.Hostname(); err == nil {
        inv.Hostname = h
    } else {
        inv.Hostname = "unknown"
    }

    return inv
}
```

### 5. Detection Cache

```go
type DetectionCache struct {
    inventory *Inventory
    timestamp time.Time
    ttl       time.Duration
    mu        sync.RWMutex
}

func NewDetectionCache(ttl time.Duration) *DetectionCache {
    return &DetectionCache{ttl: ttl}
}

func (c *DetectionCache) Get() *Inventory {
    c.mu.RLock()
    defer c.mu.RUnlock()

    if c.inventory == nil {
        return nil
    }

    if time.Since(c.timestamp) > c.ttl {
        return nil // Cache expired
    }

    return c.inventory
}

func (c *DetectionCache) Set(inv *Inventory) {
    c.mu.Lock()
    defer c.mu.Unlock()

    c.inventory = inv
    c.timestamp = time.Now()
}

func (c *DetectionCache) Invalidate() {
    c.mu.Lock()
    defer c.mu.Unlock()

    c.inventory = nil
}
```

### 6. Public API

```go
var (
    defaultEngine *Engine
    once          sync.Once
)

// New creates an Inventory using the default engine
func New() *Inventory {
    return NewWithContext(context.Background())
}

// NewWithContext creates an Inventory with timeout control
func NewWithContext(ctx context.Context) *Inventory {
    once.Do(initDefaultEngine)

    inv, err := defaultEngine.DetectAll(ctx)
    if err != nil {
        // Fallback to minimal inventory
        return &Inventory{
            PID:         os.Getpid(),
            Runtime:     "undetermined",
            Scheduler:   "undetermined",
            ImageFormat: "undetermined",
            Hostname:    "unknown",
        }
    }

    return inv
}

// NewWithEngine creates an Inventory using a custom engine
func NewWithEngine(ctx context.Context, engine *Engine) (*Inventory, error) {
    return engine.DetectAll(ctx)
}

func initDefaultEngine() {
    fs := DefaultFS{}
    net := DefaultNetwork{}

    detectors := []Detector{
        // Runtime detectors (priority 100-90)
        &DockerFileDetector{fs: fs},
        &DockerCgroupDetector{fs: fs},
        &RktEnvDetector{},
        &ContainerdFileDetector{fs: fs},
        &LXDSocketDetector{fs: fs},
        &OpenVZDetector{fs: fs},
        &WASMDetector{},

        // Scheduler detectors (priority 90-80)
        &KubernetesServiceAccountDetector{fs: fs},
        &KubernetesEnvDetector{},
        &NomadEnvDetector{},
        &MesosEnvDetector{},
        &MesosCgroupDetector{fs: fs},

        // Network detectors (priority 10-1)
        &KubernetesAPIDetector{network: net, timeout: 2 * time.Second},
        &SwarmPortDetector{network: net, timeout: 2 * time.Second},

        // Image format detectors
        &DockerImageDetector{fs: fs},
        &CRIImageDetector{fs: fs},
        &ACIEnvDetector{},
    }

    defaultEngine = NewEngine(EngineConfig{
        Detectors:     detectors,
        EnableCaching: true,
        CacheTTL:      5 * time.Minute,
    })
}
```

## Benefits of This Architecture

### 1. **Testability**
```go
func TestDockerFileDetector(t *testing.T) {
    // Mock filesystem
    mockFS := &MockFileSystem{
        files: map[string]bool{
            "/.dockerenv": true,
        },
    }

    detector := &DockerFileDetector{fs: mockFS}
    detection, err := detector.Detect(context.Background())

    assert.NoError(t, err)
    assert.Equal(t, "docker", detection.Value)
    assert.Equal(t, 0.95, detection.Confidence)
}
```

### 2. **Extensibility**
```go
// Users can add custom detectors
type CustomRuntimeDetector struct {}

func (d *CustomRuntimeDetector) Name() string {
    return "my-custom-runtime"
}

func (d *CustomRuntimeDetector) Detect(ctx context.Context) (*Detection, error) {
    // Custom detection logic
    return &Detection{
        Type:       DetectionTypeRuntime,
        Value:      "custom-runtime",
        Confidence: 0.85,
        Source:     d.Name(),
    }, nil
}

// Use custom engine
engine := NewEngine(EngineConfig{
    Detectors: []Detector{
        &CustomRuntimeDetector{},
        // ... other detectors
    },
})

inventory, _ := engine.DetectAll(context.Background())
```

### 3. **Performance Control**
```go
// Fast mode: skip network checks
engineFast := NewEngine(EngineConfig{
    Detectors: filterByPriority(allDetectors, 50), // Only fast detectors
    EnableCaching: true,
    CacheTTL: 10 * time.Minute,
})

// With timeout
ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
defer cancel()
inventory, err := engineFast.DetectAll(ctx)
```

### 4. **Debugging**
```go
// Enable debug mode to see which detectors ran
type DebugEngine struct {
    *Engine
    logger *log.Logger
}

func (e *DebugEngine) DetectAll(ctx context.Context) (*Inventory, error) {
    for _, detector := range e.detectors {
        e.logger.Printf("Running detector: %s (priority: %d)",
            detector.Name(), detector.Priority())

        detection, err := detector.Detect(ctx)
        if err != nil {
            e.logger.Printf("  Error: %v", err)
        } else if detection != nil {
            e.logger.Printf("  Found: %s (confidence: %.2f)",
                detection.Value, detection.Confidence)
        }
    }

    return e.Engine.DetectAll(ctx)
}
```

## Migration Path

### Phase 1: Add New Architecture Alongside Existing
- Keep current `New()` function working
- Implement new detector system in parallel
- Add feature flag to switch between old/new

### Phase 2: Deprecation
- Mark old functions as deprecated
- Document migration guide
- Provide compatibility shims

### Phase 3: Remove Old Code
- Remove deprecated functions in next major version
- Clean up codebase

## Example Usage Comparison

### Current (Simple but Inflexible)
```go
inventory := criprof.New()
fmt.Println(inventory.Runtime)
```

### Proposed (More Power, Same Simplicity)
```go
// Simple usage (same as before)
inventory := criprof.New()
fmt.Println(inventory.Runtime)

// Advanced usage
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

engine := criprof.NewEngine(criprof.EngineConfig{
    Detectors: criprof.DefaultDetectors(),
    EnableCaching: true,
    CacheTTL: 10 * time.Minute,
})

inventory, err := engine.DetectAll(ctx)
if err != nil {
    log.Fatal(err)
}
fmt.Println(inventory.Runtime)
```

## Conclusion

The proposed architecture:
- ✅ Maintains backward compatibility
- ✅ Improves testability dramatically
- ✅ Enables extensibility via plugins
- ✅ Adds performance controls (caching, timeouts, priorities)
- ✅ Provides confidence scoring for ambiguous cases
- ✅ Follows Go best practices (interfaces, dependency injection)
- ✅ Supports context-based cancellation

This would make criprof a **best-in-class** detection library.
