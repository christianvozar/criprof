// Copyright © 2022 Christian R. Vozar ⚜
// Licensed under the MIT License. All rights reserved.

package criprof

import (
	"fmt"
	"os"
	"regexp"
)

var (
	// Compiled regexes for container ID extraction
	dockerIDRegex  = regexp.MustCompile(`cpu:/docker/([0-9a-z]+)`)
	coreOSIDRegex  = regexp.MustCompile(`cpuset:/system\.slice/docker-([0-9a-z]+)`)
)

// getCgroupContent reads and caches the content of /proc/self/cgroup.
// Returns the content as a string, or an empty string if the file cannot be read.
func getCgroupContent() string {
	data, err := os.ReadFile("/proc/self/cgroup")
	if err != nil {
		return ""
	}
	return string(data)
}

// IsContainer determines whether the current application is running inside a
// container runtime or engine.
//
// This function uses multiple detection methods to identify container environments:
//   - Checks for the presence of /.dockerinit file (legacy Docker)
//   - Checks for the presence of /.dockerenv file (Docker)
//   - Inspects /proc/self/cgroup for container-specific cgroup entries
//
// The function returns true if any of these container indicators are found.
// It is safe to call in any environment and will return false when running on
// bare metal or in a virtual machine without container isolation.
//
// Example:
//
//	if criprof.IsContainer() {
//	    fmt.Println("Running in a container")
//	} else {
//	    fmt.Println("Running on bare metal or VM")
//	}
//
// Returns:
//   - true if running inside a container, false otherwise
func IsContainer() bool {
	if _, err := os.Stat("/.dockerinit"); err == nil {
		return true
	}

	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}

	if c := getContainerID(); c != "undetermined" {
		return true
	}

	return false
}

// getContainerID extracts the container identifier from cgroup information.
// This is an unexported function used internally by the Inventory type.
//
// The function attempts to parse /proc/self/cgroup to extract container IDs
// using regular expressions that match common container runtime patterns:
//   - Standard Docker format: cpu:/docker/[container-id]
//   - CoreOS format: cpuset:/system.slice/docker-[container-id]
//
// Returns "undetermined" if the container ID cannot be extracted.
func getContainerID() string {
	cgroupContent := getCgroupContent()
	if cgroupContent == "" {
		return "undetermined"
	}

	// Try standard Docker format using capture group
	if matches := dockerIDRegex.FindStringSubmatch(cgroupContent); matches != nil && len(matches) > 1 {
		return matches[1]
	}

	// Try CoreOS format using capture group
	if matches := coreOSIDRegex.FindStringSubmatch(cgroupContent); matches != nil && len(matches) > 1 {
		return matches[1]
	}

	return "undetermined"
}

// getHostname returns the DNS hostname of the system.
// This is an unexported function used internally by the Inventory type.
//
// The function wraps os.Hostname() with additional error context.
//
// Returns the system hostname or an error if it cannot be determined.
func getHostname() (string, error) {
	// Use the os package to get the hostname of the system.
	hostname, err := os.Hostname()
	if err != nil {
		return "", fmt.Errorf("failed to get hostname: %v", err)
	}

	return hostname, nil
}
