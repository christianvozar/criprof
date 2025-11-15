// Copyright © 2022-2023 Christian R. Vozar
// Licensed under the MIT License. All rights reserved.

package criprof

import (
	"context"
	"strings"
)

// Modern container runtimes and environments
const (
	runtimePodman     = "podman"     // Podman (daemonless container engine)
	runtimeCRIO       = "cri-o"      // CRI-O (Kubernetes CRI implementation)
	runtimeFirecracker = "firecracker" // Firecracker microVM
	runtimeKata       = "kata"       // Kata Containers (secure runtime)
	runtimeGVisor     = "gvisor"     // gVisor (application kernel)
	runtimeSysbox     = "sysbox"     // Sysbox (system containers)

	// Cloud-specific schedulers
	schedulerECS      = "ecs"        // AWS ECS
	schedulerFargate  = "fargate"    // AWS Fargate
	schedulerGKE      = "gke"        // Google Kubernetes Engine
	schedulerAKS      = "aks"        // Azure Kubernetes Service
	schedulerEKS      = "eks"        // Amazon EKS
	schedulerCloudRun = "cloud-run"  // Google Cloud Run
	schedulerLambda   = "lambda"     // AWS Lambda (container image support)
	schedulerACI      = "aci"        // Azure Container Instances

	// Modern image formats
	formatOCI         = "oci"        // OCI (Open Container Initiative)
	formatSingularity = "singularity" // Singularity/Apptainer (HPC)
)

// PodmanDetector detects Podman container runtime
type PodmanDetector struct {
	fs FileSystem
}

func (d *PodmanDetector) Name() string {
	return "podman-marker"
}

func (d *PodmanDetector) Priority() int {
	return 95
}

func (d *PodmanDetector) Detect(ctx context.Context) (*Detection, error) {
	// Check for Podman-specific marker
	if _, err := d.fs.Stat("/run/.containerenv"); err == nil {
		// Read the file to check for Podman-specific content
		data, err := d.fs.ReadFile("/run/.containerenv")
		if err == nil && strings.Contains(string(data), "podman") {
			return &Detection{
				Type:       DetectionTypeRuntime,
				Value:      runtimePodman,
				Confidence: 0.95,
				Source:     d.Name(),
			}, nil
		}
		// Generic containerenv could be Podman or CRI-O
		return &Detection{
			Type:       DetectionTypeRuntime,
			Value:      runtimePodman,
			Confidence: 0.70,
			Source:     d.Name(),
		}, nil
	}

	// Check for Podman environment variable
	if _, ok := EnvironmentVariables["PODMAN_SYSTEMD_UNIT"]; ok {
		return &Detection{
			Type:       DetectionTypeRuntime,
			Value:      runtimePodman,
			Confidence: 0.90,
			Source:     d.Name(),
		}, nil
	}

	return nil, nil
}

// CRIODetector detects CRI-O container runtime
type CRIODetector struct {
	fs FileSystem
}

func (d *CRIODetector) Name() string {
	return "cri-o-marker"
}

func (d *CRIODetector) Priority() int {
	return 95
}

func (d *CRIODetector) Detect(ctx context.Context) (*Detection, error) {
	// Check for CRI-O specific files
	if _, err := d.fs.Stat("/var/run/crio/crio.sock"); err == nil {
		return &Detection{
			Type:       DetectionTypeRuntime,
			Value:      runtimeCRIO,
			Confidence: 0.95,
			Source:     d.Name(),
		}, nil
	}

	// Check cgroup for crio
	data, err := d.fs.ReadFile("/proc/self/cgroup")
	if err == nil && strings.Contains(string(data), "crio") {
		return &Detection{
			Type:       DetectionTypeRuntime,
			Value:      runtimeCRIO,
			Confidence: 0.90,
			Source:     d.Name(),
		}, nil
	}

	return nil, nil
}

// FirecrackerDetector detects Firecracker microVM
type FirecrackerDetector struct {
	fs FileSystem
}

func (d *FirecrackerDetector) Name() string {
	return "firecracker-dmi"
}

func (d *FirecrackerDetector) Priority() int {
	return 90
}

func (d *FirecrackerDetector) Detect(ctx context.Context) (*Detection, error) {
	// Firecracker sets specific DMI product name
	data, err := d.fs.ReadFile("/sys/class/dmi/id/product_name")
	if err == nil && strings.TrimSpace(string(data)) == "Firecracker" {
		return &Detection{
			Type:       DetectionTypeRuntime,
			Value:      runtimeFirecracker,
			Confidence: 0.98,
			Source:     d.Name(),
		}, nil
	}

	return nil, nil
}

// KataContainersDetector detects Kata Containers
type KataContainersDetector struct {
	fs FileSystem
}

func (d *KataContainersDetector) Name() string {
	return "kata-containers"
}

func (d *KataContainersDetector) Priority() int {
	return 90
}

func (d *KataContainersDetector) Detect(ctx context.Context) (*Detection, error) {
	// Check for Kata-specific markers
	if _, err := d.fs.Stat("/run/kata-containers"); err == nil {
		return &Detection{
			Type:       DetectionTypeRuntime,
			Value:      runtimeKata,
			Confidence: 0.95,
			Source:     d.Name(),
		}, nil
	}

	// Check DMI for QEMU (Kata uses QEMU)
	data, err := d.fs.ReadFile("/sys/class/dmi/id/product_name")
	if err == nil && strings.Contains(string(data), "QEMU") {
		// Lower confidence - could be any QEMU VM
		data2, err2 := d.fs.ReadFile("/proc/cpuinfo")
		if err2 == nil && strings.Contains(string(data2), "QEMU") {
			return &Detection{
				Type:       DetectionTypeRuntime,
				Value:      runtimeKata,
				Confidence: 0.60,
				Source:     d.Name(),
			}, nil
		}
	}

	return nil, nil
}

// GVisorDetector detects gVisor (runsc)
type GVisorDetector struct {
	fs FileSystem
}

func (d *GVisorDetector) Name() string {
	return "gvisor-marker"
}

func (d *GVisorDetector) Priority() int {
	return 90
}

func (d *GVisorDetector) Detect(ctx context.Context) (*Detection, error) {
	// gVisor shows up in cgroups
	data, err := d.fs.ReadFile("/proc/self/cgroup")
	if err == nil && (strings.Contains(string(data), "runsc") || strings.Contains(string(data), "gvisor")) {
		return &Detection{
			Type:       DetectionTypeRuntime,
			Value:      runtimeGVisor,
			Confidence: 0.95,
			Source:     d.Name(),
		}, nil
	}

	// Check for gVisor-specific /proc entries
	if _, err := d.fs.Stat("/proc/self/root/dev/gvisor"); err == nil {
		return &Detection{
			Type:       DetectionTypeRuntime,
			Value:      runtimeGVisor,
			Confidence: 0.98,
			Source:     d.Name(),
		}, nil
	}

	return nil, nil
}

// SysboxDetector detects Sysbox system containers
type SysboxDetector struct {
	fs FileSystem
}

func (d *SysboxDetector) Name() string {
	return "sysbox-marker"
}

func (d *SysboxDetector) Priority() int {
	return 90
}

func (d *SysboxDetector) Detect(ctx context.Context) (*Detection, error) {
	// Sysbox sets specific environment variable
	if _, ok := EnvironmentVariables["SYSBOX_CONTAINER"]; ok {
		return &Detection{
			Type:       DetectionTypeRuntime,
			Value:      runtimeSysbox,
			Confidence: 0.98,
			Source:     d.Name(),
		}, nil
	}

	// Check for Sysbox markers in cgroup
	data, err := d.fs.ReadFile("/proc/self/cgroup")
	if err == nil && strings.Contains(string(data), "sysbox") {
		return &Detection{
			Type:       DetectionTypeRuntime,
			Value:      runtimeSysbox,
			Confidence: 0.90,
			Source:     d.Name(),
		}, nil
	}

	return nil, nil
}

// AWS ECS Detector
type ECSDetector struct{}

func (d *ECSDetector) Name() string {
	return "aws-ecs"
}

func (d *ECSDetector) Priority() int {
	return 85
}

func (d *ECSDetector) Detect(ctx context.Context) (*Detection, error) {
	// ECS sets specific environment variables
	if _, ok := EnvironmentVariables["ECS_CONTAINER_METADATA_URI"]; ok {
		return &Detection{
			Type:       DetectionTypeScheduler,
			Value:      schedulerECS,
			Confidence: 0.98,
			Source:     d.Name(),
		}, nil
	}

	if _, ok := EnvironmentVariables["ECS_CONTAINER_METADATA_URI_V4"]; ok {
		return &Detection{
			Type:       DetectionTypeScheduler,
			Value:      schedulerECS,
			Confidence: 0.98,
			Source:     d.Name(),
		}, nil
	}

	return nil, nil
}

// AWS Fargate Detector
type FargateDetector struct{}

func (d *FargateDetector) Name() string {
	return "aws-fargate"
}

func (d *FargateDetector) Priority() int {
	return 85
}

func (d *FargateDetector) Detect(ctx context.Context) (*Detection, error) {
	// Fargate is ECS + specific launch type
	if launchType, ok := EnvironmentVariables["AWS_EXECUTION_ENV"]; ok {
		if strings.Contains(strings.ToLower(launchType), "fargate") {
			return &Detection{
				Type:       DetectionTypeScheduler,
				Value:      schedulerFargate,
				Confidence: 0.99,
				Source:     d.Name(),
			}, nil
		}
	}

	// Check for Fargate-specific metadata
	if _, ok := EnvironmentVariables["ECS_CONTAINER_METADATA_URI_V4"]; ok {
		if _, ok2 := EnvironmentVariables["AWS_EXECUTION_ENV"]; ok2 {
			return &Detection{
				Type:       DetectionTypeScheduler,
				Value:      schedulerFargate,
				Confidence: 0.85,
				Source:     d.Name(),
			}, nil
		}
	}

	return nil, nil
}

// Google Cloud Run Detector
type CloudRunDetector struct{}

func (d *CloudRunDetector) Name() string {
	return "google-cloud-run"
}

func (d *CloudRunDetector) Priority() int {
	return 85
}

func (d *CloudRunDetector) Detect(ctx context.Context) (*Detection, error) {
	// Cloud Run sets K_SERVICE environment variable
	if _, ok := EnvironmentVariables["K_SERVICE"]; ok {
		return &Detection{
			Type:       DetectionTypeScheduler,
			Value:      schedulerCloudRun,
			Confidence: 0.98,
			Source:     d.Name(),
		}, nil
	}

	// Also check for K_REVISION and K_CONFIGURATION
	if _, ok := EnvironmentVariables["K_REVISION"]; ok {
		return &Detection{
			Type:       DetectionTypeScheduler,
			Value:      schedulerCloudRun,
			Confidence: 0.95,
			Source:     d.Name(),
		}, nil
	}

	return nil, nil
}

// AWS Lambda Container Detector
type LambdaContainerDetector struct{}

func (d *LambdaContainerDetector) Name() string {
	return "aws-lambda-container"
}

func (d *LambdaContainerDetector) Priority() int {
	return 85
}

func (d *LambdaContainerDetector) Detect(ctx context.Context) (*Detection, error) {
	// Lambda sets specific environment variables
	if _, ok := EnvironmentVariables["AWS_LAMBDA_FUNCTION_NAME"]; ok {
		return &Detection{
			Type:       DetectionTypeScheduler,
			Value:      schedulerLambda,
			Confidence: 0.99,
			Source:     d.Name(),
		}, nil
	}

	if _, ok := EnvironmentVariables["LAMBDA_TASK_ROOT"]; ok {
		return &Detection{
			Type:       DetectionTypeScheduler,
			Value:      schedulerLambda,
			Confidence: 0.98,
			Source:     d.Name(),
		}, nil
	}

	return nil, nil
}

// Azure Container Instances Detector
type ACIDetector struct{}

func (d *ACIDetector) Name() string {
	return "azure-container-instances"
}

func (d *ACIDetector) Priority() int {
	return 85
}

func (d *ACIDetector) Detect(ctx context.Context) (*Detection, error) {
	// ACI sets specific environment variables
	if _, ok := EnvironmentVariables["ACI_RESOURCE_GROUP"]; ok {
		return &Detection{
			Type:       DetectionTypeScheduler,
			Value:      schedulerACI,
			Confidence: 0.98,
			Source:     d.Name(),
		}, nil
	}

	if _, ok := EnvironmentVariables["CONTAINER_GROUP_NAME"]; ok {
		return &Detection{
			Type:       DetectionTypeScheduler,
			Value:      schedulerACI,
			Confidence: 0.90,
			Source:     d.Name(),
		}, nil
	}

	return nil, nil
}

// Singularity/Apptainer Detector (HPC environments)
type SingularityDetector struct{}

func (d *SingularityDetector) Name() string {
	return "singularity-apptainer"
}

func (d *SingularityDetector) Priority() int {
	return 90
}

func (d *SingularityDetector) Detect(ctx context.Context) (*Detection, error) {
	// Singularity/Apptainer set specific environment variables
	if _, ok := EnvironmentVariables["SINGULARITY_CONTAINER"]; ok {
		return &Detection{
			Type:       DetectionTypeRuntime,
			Value:      "singularity",
			Confidence: 0.99,
			Source:     d.Name(),
		}, nil
	}

	if _, ok := EnvironmentVariables["APPTAINER_CONTAINER"]; ok {
		return &Detection{
			Type:       DetectionTypeRuntime,
			Value:      "apptainer",
			Confidence: 0.99,
			Source:     d.Name(),
		}, nil
	}

	if _, ok := EnvironmentVariables["SINGULARITY_NAME"]; ok {
		return &Detection{
			Type:       DetectionTypeRuntime,
			Value:      "singularity",
			Confidence: 0.95,
			Source:     d.Name(),
		}, nil
	}

	return nil, nil
}

// OCI Image Format Detector
type OCIImageDetector struct {
	fs FileSystem
}

func (d *OCIImageDetector) Name() string {
	return "oci-image-format"
}

func (d *OCIImageDetector) Priority() int {
	return 85
}

func (d *OCIImageDetector) Detect(ctx context.Context) (*Detection, error) {
	// Check for OCI-specific markers
	if _, err := d.fs.Stat("/var/lib/containers"); err == nil {
		return &Detection{
			Type:       DetectionTypeImageFormat,
			Value:      formatOCI,
			Confidence: 0.80,
			Source:     d.Name(),
		}, nil
	}

	// Podman typically uses OCI format
	if _, ok := EnvironmentVariables["PODMAN_SYSTEMD_UNIT"]; ok {
		return &Detection{
			Type:       DetectionTypeImageFormat,
			Value:      formatOCI,
			Confidence: 0.85,
			Source:     d.Name(),
		}, nil
	}

	return nil, nil
}

// Singularity Image Format Detector
type SingularityImageDetector struct{}

func (d *SingularityImageDetector) Name() string {
	return "singularity-image-format"
}

func (d *SingularityImageDetector) Priority() int {
	return 85
}

func (d *SingularityImageDetector) Detect(ctx context.Context) (*Detection, error) {
	if _, ok := EnvironmentVariables["SINGULARITY_CONTAINER"]; ok {
		return &Detection{
			Type:       DetectionTypeImageFormat,
			Value:      formatSingularity,
			Confidence: 0.95,
			Source:     d.Name(),
		}, nil
	}

	if _, ok := EnvironmentVariables["APPTAINER_CONTAINER"]; ok {
		return &Detection{
			Type:       DetectionTypeImageFormat,
			Value:      formatSingularity,
			Confidence: 0.95,
			Source:     d.Name(),
		}, nil
	}

	return nil, nil
}
