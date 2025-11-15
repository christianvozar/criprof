// Copyright © 2022-2023 Christian R. Vozar
// Licensed under the MIT License. All rights reserved.

// Package criprof provides Container Runtime Interface (CRI) profiling and introspection.
//
// This library detects and identifies the container runtime environment, scheduler,
// and image format in which an application is running. It uses environmental "hints"
// to introspect the container runtime and provides structured information about the
// executing environment.
//
// # Supported Container Runtimes
//
// The package can detect the following container runtimes:
//   - Docker
//   - rkt (CoreOS)
//   - containerd
//   - CRI-O
//   - LXD
//   - OpenVZ
//   - WASM (WebAssembly)
//   - RunC (Open Container Initiative)
//
// # Supported Orchestrators
//
// The package can detect the following orchestration platforms:
//   - Kubernetes
//   - Docker Swarm
//   - HashiCorp Nomad
//   - Apache Mesos
//
// # Supported Image Formats
//
// The package can detect the following container image formats:
//   - Docker format
//   - CRI (Container Runtime Interface) format
//   - ACI (App Container Image) format
//   - OCF (Open Container Format)
//
// # Usage
//
// Basic usage to detect container runtime information:
//
//	import "github.com/christianvozar/criprof"
//
//	func main() {
//	    // Create a new inventory of the container environment
//	    inventory := criprof.New()
//
//	    // Access individual fields
//	    fmt.Printf("Runtime: %s\n", inventory.Runtime)
//	    fmt.Printf("Scheduler: %s\n", inventory.Scheduler)
//
//	    // Export as JSON
//	    jsonData := inventory.JSON()
//	    fmt.Println(jsonData)
//
//	    // Check if running in a container
//	    if criprof.IsContainer() {
//	        fmt.Println("Running in a container")
//	    }
//	}
//
// # Detection Methods
//
// The library uses multiple detection methods including:
//   - File system markers (e.g., /.dockerenv, /run/.containerenv)
//   - Environment variables (e.g., KUBERNETES_SERVICE_HOST, NOMAD_TASK_DIR)
//   - Process information (e.g., /proc/self/cgroup)
//   - Network probes (e.g., Kubernetes API checks)
//
// All detection is non-invasive, read-only, and safe for production use.
package criprof

import (
	"encoding/json"
	"fmt"
	"os"
)

// EnvironmentVariables is a cached map of all environment variables available at
// package initialization time. This cache is populated once when the package is
// imported to improve performance during runtime and scheduler detection.
//
// The map keys are environment variable names and values are their corresponding values.
// This variable is exported to allow advanced users to inspect the cached environment
// if needed for debugging purposes.
var EnvironmentVariables map[string]string

func init() {
	EnvironmentVariables = environMap()
}

// Inventory holds an application's container and runtime information.
// It provides a comprehensive snapshot of the container environment including
// runtime type, orchestration platform, and system identifiers.
//
// All fields are automatically populated when creating a new Inventory using New().
// If a particular aspect cannot be determined, the corresponding field will contain
// the value "undetermined" (for string fields) or system defaults (for other types).
type Inventory struct {
	// Hostname is the DNS hostname of the system running the container.
	Hostname string `json:"hostname"`

	// ID is the unique container identifier extracted from cgroup information.
	// Returns "undetermined" if the container ID cannot be detected.
	ID string `json:"id"`

	// ImageFormat identifies the container image format (e.g., "docker", "cri", "aci").
	// Returns "undetermined" if the format cannot be detected.
	ImageFormat string `json:"image_format"`

	// PID is the process ID of the current application.
	PID int `json:"pid"`

	// Runtime identifies the container runtime engine (e.g., "docker", "containerd", "rkt").
	// Returns "undetermined" if the runtime cannot be detected.
	Runtime string `json:"runtime"`

	// Scheduler identifies the orchestration platform (e.g., "kubernetes", "nomad", "swarm").
	// Returns "undetermined" if no orchestrator is detected.
	Scheduler string `json:"scheduler"`
}

// New creates and returns a new Inventory with all fields automatically populated
// based on the current container environment.
//
// This function performs comprehensive detection of the container runtime, orchestration
// platform, image format, and system identifiers. Detection is non-invasive and uses
// multiple hints including file system markers, environment variables, process information,
// and network probes.
//
// The function will gracefully handle detection failures by setting undetermined values
// rather than returning errors, making it safe to use in any environment.
//
// Example:
//
//	inventory := criprof.New()
//	fmt.Printf("Runtime: %s\n", inventory.Runtime)
//	fmt.Printf("Scheduler: %s\n", inventory.Scheduler)
//	fmt.Printf("Container ID: %s\n", inventory.ID)
//
// Returns:
//   - A pointer to a fully populated Inventory struct
func New() *Inventory {
	f, _ := getImageFormat()
	h, _ := getHostname()

	return &Inventory{
		Hostname:    h,
		ID:          getContainerID(),
		ImageFormat: f,
		PID:         os.Getpid(),
		Runtime:     getRuntime(),
		Scheduler:   getScheduler(),
	}
}

// JSON serializes the Inventory to a JSON-formatted string.
//
// This method converts all inventory fields to JSON format, making it easy to
// export container environment information for logging, diagnostics, or integration
// with other tools.
//
// If JSON marshaling fails (which is extremely rare for this struct), an error
// message is printed to stdout and an empty string is returned.
//
// Example:
//
//	inventory := criprof.New()
//	jsonStr := inventory.JSON()
//	fmt.Println(jsonStr)
//	// Output: {"hostname":"web-server","id":"abc123","image_format":"docker","pid":1234,"runtime":"docker","scheduler":"kubernetes"}
//
// Returns:
//   - A JSON-formatted string representation of the Inventory, or an empty string on error
func (i Inventory) JSON() string {
	j, err := json.Marshal(i)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	return string(j)
}
