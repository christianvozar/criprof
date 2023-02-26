// Copyright © 2022-2023 Christian R. Vozar
// Licensed under the MIT License. All rights reserved.

package criprof

import (
	"io/ioutil"
	"os"
	"strings"
)

// Detectable container runtimes.
const (
	runtimeDocker       = "docker"       // Docker
	runtimeRkt          = "rkt"          // CoreOS rkt
	runtimeRunC         = "runc"         // Open Container Initiative runc
	runtimeContainerD   = "containerd"   // containerd
	runtimeLXC          = "lxc"          // LXC (Linux Containers)
	runtimeLXD          = "lxd"          // LXD (containerd + LXC)
	runtimeOpenVZ       = "openvz"       // OpenVZ
	runtimeUndetermined = "undetermined" // Undetermined
)

// getRuntime returns the name of the container runtime that is currently running.
func getRuntime() string {
	// Check if the /.dockerinit file exists to detect a Docker runtime.
	if _, err := os.Stat("/.dockerinit"); err == nil {
		return runtimeDocker
	}

	// Check if the /.dockerenv file exists to detect a Docker runtime.
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return runtimeDocker
	}

	// Check if /run/.containerenv file exists to detect a CRI-O or containerd
	// runtime.
	if _, err := os.Stat("/run/.containerenv"); err == nil {
		return runtimeContainerD
	}

	// Check the cgroup to detect a Docker runtime.
	cgroup, _ := ioutil.ReadFile("/proc/self/cgroup")
	if strings.Contains(string(cgroup), "docker") {
		return runtimeDocker
	}

	// Check if the AC_METADATA_URL environment variable is set to detect an rkt runtime.
	if _, ok := EnvironmentVariables["AC_METADATA_URL"]; ok {
		return runtimeRkt
	}

	// Check if the AC_APP_NAME environment variable is set to detect an rkt runtime.
	if _, ok := EnvironmentVariables["AC_APP_NAME"]; ok {
		return runtimeRkt
	}

	// Check if the /dev/lxd/sock file exists to detect an LXD runtime.
	if _, err := os.Stat("/dev/lxd/sock"); err == nil {
		return runtimeLXD
	}

	// If none of the above checks pass, return an undetermined runtime.
	return runtimeUndetermined
}
