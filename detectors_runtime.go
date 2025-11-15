// Copyright © 2022-2023 Christian R. Vozar
// Licensed under the MIT License. All rights reserved.

package criprof

import (
	"context"
	"runtime"
	"strings"
)

// DockerFileDetector detects Docker via filesystem markers
type DockerFileDetector struct {
	fs FileSystem
}

func (d *DockerFileDetector) Name() string {
	return "docker-file-marker"
}

func (d *DockerFileDetector) Priority() int {
	return 100 // File checks are very fast
}

func (d *DockerFileDetector) Detect(ctx context.Context) (*Detection, error) {
	// Check /.dockerenv (most common)
	if _, err := d.fs.Stat("/.dockerenv"); err == nil {
		return &Detection{
			Type:       DetectionTypeRuntime,
			Value:      runtimeDocker,
			Confidence: 0.95,
			Source:     d.Name(),
		}, nil
	}

	// Check /.dockerinit (legacy)
	if _, err := d.fs.Stat("/.dockerinit"); err == nil {
		return &Detection{
			Type:       DetectionTypeRuntime,
			Value:      runtimeDocker,
			Confidence: 0.90,
			Source:     d.Name(),
		}, nil
	}

	return nil, nil
}

// DockerCgroupDetector detects Docker via cgroup inspection
type DockerCgroupDetector struct {
	fs FileSystem
}

func (d *DockerCgroupDetector) Name() string {
	return "docker-cgroup"
}

func (d *DockerCgroupDetector) Priority() int {
	return 90
}

func (d *DockerCgroupDetector) Detect(ctx context.Context) (*Detection, error) {
	data, err := d.fs.ReadFile("/proc/self/cgroup")
	if err != nil {
		return nil, nil
	}

	if strings.Contains(string(data), "docker") {
		return &Detection{
			Type:       DetectionTypeRuntime,
			Value:      runtimeDocker,
			Confidence: 0.85,
			Source:     d.Name(),
		}, nil
	}

	return nil, nil
}

// ContainerdFileDetector detects containerd/CRI-O via filesystem markers
type ContainerdFileDetector struct {
	fs FileSystem
}

func (d *ContainerdFileDetector) Name() string {
	return "containerd-file-marker"
}

func (d *ContainerdFileDetector) Priority() int {
	return 95
}

func (d *ContainerdFileDetector) Detect(ctx context.Context) (*Detection, error) {
	if _, err := d.fs.Stat("/run/.containerenv"); err == nil {
		return &Detection{
			Type:       DetectionTypeRuntime,
			Value:      runtimeContainerD,
			Confidence: 0.90,
			Source:     d.Name(),
		}, nil
	}

	return nil, nil
}

// RktEnvDetector detects rkt via environment variables
type RktEnvDetector struct{}

func (d *RktEnvDetector) Name() string {
	return "rkt-env"
}

func (d *RktEnvDetector) Priority() int {
	return 80
}

func (d *RktEnvDetector) Detect(ctx context.Context) (*Detection, error) {
	if _, ok := EnvironmentVariables["AC_METADATA_URL"]; ok {
		return &Detection{
			Type:       DetectionTypeRuntime,
			Value:      runtimeRkt,
			Confidence: 0.95,
			Source:     d.Name(),
		}, nil
	}

	if _, ok := EnvironmentVariables["AC_APP_NAME"]; ok {
		return &Detection{
			Type:       DetectionTypeRuntime,
			Value:      runtimeRkt,
			Confidence: 0.95,
			Source:     d.Name(),
		}, nil
	}

	return nil, nil
}

// LXDSocketDetector detects LXD via socket presence
type LXDSocketDetector struct {
	fs FileSystem
}

func (d *LXDSocketDetector) Name() string {
	return "lxd-socket"
}

func (d *LXDSocketDetector) Priority() int {
	return 95
}

func (d *LXDSocketDetector) Detect(ctx context.Context) (*Detection, error) {
	if _, err := d.fs.Stat("/dev/lxd/sock"); err == nil {
		return &Detection{
			Type:       DetectionTypeRuntime,
			Value:      runtimeLXD,
			Confidence: 0.95,
			Source:     d.Name(),
		}, nil
	}

	return nil, nil
}

// OpenVZDetector detects OpenVZ via /proc/vz
type OpenVZDetector struct {
	fs FileSystem
}

func (d *OpenVZDetector) Name() string {
	return "openvz-proc"
}

func (d *OpenVZDetector) Priority() int {
	return 90
}

func (d *OpenVZDetector) Detect(ctx context.Context) (*Detection, error) {
	if _, err := d.fs.Stat("/proc/vz"); err == nil {
		return &Detection{
			Type:       DetectionTypeRuntime,
			Value:      runtimeOpenVZ,
			Confidence: 0.90,
			Source:     d.Name(),
		}, nil
	}

	return nil, nil
}

// WASMDetector detects WebAssembly runtime
type WASMDetector struct{}

func (d *WASMDetector) Name() string {
	return "wasm-build-target"
}

func (d *WASMDetector) Priority() int {
	return 100 // Compile-time constant check is instant
}

func (d *WASMDetector) Detect(ctx context.Context) (*Detection, error) {
	if runtime.GOOS == "js" && runtime.GOARCH == "wasm" {
		return &Detection{
			Type:       DetectionTypeRuntime,
			Value:      runtimeWASM,
			Confidence: 1.0, // Definitive - compile-time constant
			Source:     d.Name(),
		}, nil
	}

	return nil, nil
}
