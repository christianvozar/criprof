// Copyright © 2022-2023 Christian R. Vozar
// Licensed under the MIT License. All rights reserved.

package criprof

import (
	"context"
	"os"
	"strings"
	"time"
)

// KubernetesServiceAccountDetector detects Kubernetes via service account token
type KubernetesServiceAccountDetector struct {
	fs FileSystem
}

func (d *KubernetesServiceAccountDetector) Name() string {
	return "kubernetes-service-account"
}

func (d *KubernetesServiceAccountDetector) Priority() int {
	return 95
}

func (d *KubernetesServiceAccountDetector) Detect(ctx context.Context) (*Detection, error) {
	if _, err := d.fs.Stat("/run/secrets/kubernetes.io/serviceaccount/token"); err == nil {
		return &Detection{
			Type:       DetectionTypeScheduler,
			Value:      schedulerKubernetes,
			Confidence: 0.99, // Very high - definitive marker
			Source:     d.Name(),
		}, nil
	}

	return nil, nil
}

// KubernetesEnvDetector detects Kubernetes via environment variables
type KubernetesEnvDetector struct{}

func (d *KubernetesEnvDetector) Name() string {
	return "kubernetes-env"
}

func (d *KubernetesEnvDetector) Priority() int {
	return 85
}

func (d *KubernetesEnvDetector) Detect(ctx context.Context) (*Detection, error) {
	if _, ok := EnvironmentVariables["KUBERNETES_SERVICE_HOST"]; ok {
		return &Detection{
			Type:       DetectionTypeScheduler,
			Value:      schedulerKubernetes,
			Confidence: 0.95,
			Source:     d.Name(),
		}, nil
	}

	return nil, nil
}

// KubernetesAPIDetector detects Kubernetes via API server probe
type KubernetesAPIDetector struct {
	network Network
	timeout time.Duration
}

func (d *KubernetesAPIDetector) Name() string {
	return "kubernetes-api-probe"
}

func (d *KubernetesAPIDetector) Priority() int {
	return 10 // Network check is slow, run last
}

func (d *KubernetesAPIDetector) Detect(ctx context.Context) (*Detection, error) {
	// Create timeout context
	ctx, cancel := context.WithTimeout(ctx, d.timeout)
	defer cancel()

	resp, err := d.network.HTTPGet(ctx, "http://kubernetes.default.svc")
	if err != nil {
		return nil, nil // No detection, not an error
	}
	resp.Body.Close()

	return &Detection{
		Type:       DetectionTypeScheduler,
		Value:      schedulerKubernetes,
		Confidence: 0.80, // Lower - could be other service
		Source:     d.Name(),
	}, nil
}

// NomadEnvDetector detects HashiCorp Nomad via environment variables
type NomadEnvDetector struct{}

func (d *NomadEnvDetector) Name() string {
	return "nomad-env"
}

func (d *NomadEnvDetector) Priority() int {
	return 85
}

func (d *NomadEnvDetector) Detect(ctx context.Context) (*Detection, error) {
	if _, ok := EnvironmentVariables["NOMAD_TASK_DIR"]; ok {
		return &Detection{
			Type:       DetectionTypeScheduler,
			Value:      schedulerNomad,
			Confidence: 0.95,
			Source:     d.Name(),
		}, nil
	}

	return nil, nil
}

// NomadHostnameDetector detects Nomad via hostname prefix
type NomadHostnameDetector struct{}

func (d *NomadHostnameDetector) Name() string {
	return "nomad-hostname"
}

func (d *NomadHostnameDetector) Priority() int {
	return 80
}

func (d *NomadHostnameDetector) Detect(ctx context.Context) (*Detection, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, nil
	}

	if strings.HasPrefix(hostname, "nomad-task-") {
		return &Detection{
			Type:       DetectionTypeScheduler,
			Value:      schedulerNomad,
			Confidence: 0.85,
			Source:     d.Name(),
		}, nil
	}

	return nil, nil
}

// MesosEnvDetector detects Apache Mesos via environment variables
type MesosEnvDetector struct{}

func (d *MesosEnvDetector) Name() string {
	return "mesos-env"
}

func (d *MesosEnvDetector) Priority() int {
	return 85
}

func (d *MesosEnvDetector) Detect(ctx context.Context) (*Detection, error) {
	if _, ok := EnvironmentVariables["MESOS_TASK_ID"]; ok {
		return &Detection{
			Type:       DetectionTypeScheduler,
			Value:      schedulerMesos,
			Confidence: 0.95,
			Source:     d.Name(),
		}, nil
	}

	if _, ok := EnvironmentVariables["MESOS_CONTAINER_NAME"]; ok {
		return &Detection{
			Type:       DetectionTypeScheduler,
			Value:      schedulerMesos,
			Confidence: 0.95,
			Source:     d.Name(),
		}, nil
	}

	return nil, nil
}

// MesosCgroupDetector detects Mesos via cgroup inspection
type MesosCgroupDetector struct {
	fs FileSystem
}

func (d *MesosCgroupDetector) Name() string {
	return "mesos-cgroup"
}

func (d *MesosCgroupDetector) Priority() int {
	return 80
}

func (d *MesosCgroupDetector) Detect(ctx context.Context) (*Detection, error) {
	data, err := d.fs.ReadFile("/proc/1/cgroup")
	if err != nil {
		return nil, nil
	}

	if strings.Contains(string(data), "mesos") {
		return &Detection{
			Type:       DetectionTypeScheduler,
			Value:      schedulerMesos,
			Confidence: 0.90,
			Source:     d.Name(),
		}, nil
	}

	return nil, nil
}

// SwarmPortDetector detects Docker Swarm via port probe
type SwarmPortDetector struct {
	network Network
	timeout time.Duration
}

func (d *SwarmPortDetector) Name() string {
	return "swarm-port-probe"
}

func (d *SwarmPortDetector) Priority() int {
	return 20 // Network check is slower
}

func (d *SwarmPortDetector) Detect(ctx context.Context) (*Detection, error) {
	conn, err := d.network.DialTimeout("tcp", "127.0.0.1:2377", d.timeout)
	if err != nil {
		return nil, nil // No detection, not an error
	}
	conn.Close()

	return &Detection{
		Type:       DetectionTypeScheduler,
		Value:      schedulerSwarm,
		Confidence: 0.80, // Port could be used by other services
		Source:     d.Name(),
	}, nil
}
