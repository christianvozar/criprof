// Copyright © 2022 Christian R. Vozar ⚜
// Licensed under the MIT License. All rights reserved.

package criprof

import (
	"context"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	schedulerKubernetes   = "kubernetes"
	schedulerNomad        = "nomad"
	schedulerMesos        = "mesos"
	schedulerSwarm        = "swarm"
	schedulerUndetermined = "undetermined"

	// Network operation timeout
	networkTimeout = 2 * time.Second
)

// getScheduler returns the identified scheduler, if detected.
func getScheduler() string {
	if isKubernetes() {
		return schedulerKubernetes
	}

	if isNomad() {
		return schedulerNomad
	}

	if isSwarm() {
		return schedulerSwarm
	}

	if isMesos() {
		return schedulerMesos
	}

	return schedulerUndetermined
}

// isSwarm returns true if running in Docker Swarm.
func isSwarm() bool {
	// Check Docker Swarm port is open to detect if Docker Swarm cluster.
	// Use DialTimeout to avoid hanging indefinitely
	conn, err := net.DialTimeout("tcp", "127.0.0.1:2377", networkTimeout)
	if err == nil {
		conn.Close()
		return true
	}

	return false
}

// isKubernetes returns true if running in Kubernetes cluster.
func isKubernetes() bool {
	// Check if /run/secrets/kubernetes.io/serviceaccount/token file exists.
	if _, err := os.Stat("/run/secrets/kubernetes.io/serviceaccount/token"); err == nil {
		return true
	}

	// Check if KUBERNETES_SERVICE_HOST environment variable is set.
	if _, ok := EnvironmentVariables["KUBERNETES_SERVICE_HOST"]; ok {
		return true
	}

	// Check if Kubernetes API server is accessible with timeout.
	ctx, cancel := context.WithTimeout(context.Background(), networkTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "http://kubernetes.default.svc", nil)
	if err != nil {
		return false
	}

	resp, err := http.DefaultClient.Do(req)
	if err == nil {
		resp.Body.Close()
		return true
	}

	return false
}

// isNomad returns true if running inside a HashiCorp Nomad.
func isNomad() bool {
	// Check if the NOMAD_TASK_DIR environment variable is set.
	// Use cached EnvironmentVariables for consistency
	if _, ok := EnvironmentVariables["NOMAD_TASK_DIR"]; ok {
		return true
	}

	// Check if the HOSTNAME environment variable starts with the prefix "nomad-task-".
	hostname, err := os.Hostname()
	if err == nil && strings.HasPrefix(hostname, "nomad-task-") {
		return true
	}

	return false
}

// isMesos returns true if running in a Mesos environment.
func isMesos() bool {
	// Check if  MESOS_TASK_ID environment variable is set.
	if _, ok := EnvironmentVariables["MESOS_TASK_ID"]; ok {
		return true
	}

	// Check if the MESOS_CONTAINER_NAME environment variable is set.
	if _, ok := EnvironmentVariables["MESOS_CONTAINER_NAME"]; ok {
		return true
	}

	// Check if the /proc/1/cgroup file contains the "mesos" string.
	cgroup, err := os.ReadFile("/proc/1/cgroup")
	if err == nil && strings.Contains(string(cgroup), "mesos") {
		return true
	}

	return false
}
