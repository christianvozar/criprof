// Copyright © 2022 Christian R. Vozar ⚜
// Licensed under the MIT License. All rights reserved.

package criprof

import (
	"net"
	"net/http"
	"os"
	"strings"
)

const (
	schedulerKubernetes   = "kubernetes"
	schedulerNomad        = "nomad"
	scehdulerMesos        = "mesos"
	schedulerSwarm        = "swarm"
	schedulerUndetermined = "undetermined"
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

	return schedulerUndetermined
}

// isSwarm returns true if running in Docker Swarm.
func isSwarm() bool {
	// Check Docker Swarm port is open to detect if Docker Swarm cluster.
	conn, err := net.Dial("tcp", "127.0.0.1:2377")
	if err == nil {
		conn.Close()
		return true
	}

	return false
}

// isKubernetes returns true if running in Kubernetes cluster.
func isKubernetes() bool {
	// Check if  /run/secrets/kubernetes.io/serviceaccount/token file exists to
	// detect Kubernetes.
	if _, err := os.Stat("/run/secrets/kubernetes.io/serviceaccount/token"); err == nil {
		return true
	}

	// Check if KUBERNETES_SERVICE_HOST environment variable is set to detect
	// Kubernetes.
	if _, ok := EnvironmentVariables["KUBERNETES_SERVICE_HOST"]; ok {
		return true
	}

	// Check if Kubernetes API server is accessible to detect Kubernetes.
	resp, err := http.Get("http://kubernetes.default.svc")
	if err == nil {
		resp.Body.Close()
		return true
	}

	return false
}

// isNomad returns true if running inside a HashiCorp Nomad.
func isNomad() bool {
	// Check if the NOMAD_TASK_DIR environment variable is set to detect
	// Nomad task.
	if _, ok := os.LookupEnv("NOMAD_TASK_DIR"); ok {
		return true
	}

	// Check if the HOSTNAME environment variable starts with the prefix
	// "nomad-task-" to detect Nomad task.
	hostname, err := os.Hostname()
	if err == nil && strings.HasPrefix(hostname, "nomad-task-") {
		return true
	}

	return false
}
