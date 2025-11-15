// Copyright © 2022-2023 Christian R. Vozar
// Licensed under the MIT License. All rights reserved.

package criprof

import (
	"context"
	"testing"
)

// Test Podman Detector
func TestPodmanDetector(t *testing.T) {
	tests := []struct {
		name         string
		files        map[string]bool
		data         map[string][]byte
		shouldDetect bool
		expectedConf float64
	}{
		{
			name:         "detects Podman with containerenv file containing 'podman'",
			files:        map[string]bool{"/run/.containerenv": true},
			data:         map[string][]byte{"/run/.containerenv": []byte("engine=\"podman-1.9.3\"")},
			shouldDetect: true,
			expectedConf: 0.95,
		},
		{
			name:         "detects generic containerenv with lower confidence",
			files:        map[string]bool{"/run/.containerenv": true},
			data:         map[string][]byte{"/run/.containerenv": []byte("engine=\"unknown\"")},
			shouldDetect: true,
			expectedConf: 0.70,
		},
		{
			name:         "no detection when files absent",
			files:        map[string]bool{},
			shouldDetect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := &PodmanDetector{
				fs: &MockFileSystem{
					files: tt.files,
					data:  tt.data,
				},
			}

			detection, err := detector.Detect(context.Background())

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.shouldDetect {
				if detection == nil {
					t.Fatal("expected detection, got nil")
				}
				if detection.Value != "podman" {
					t.Errorf("expected value podman, got %s", detection.Value)
				}
				if detection.Confidence != tt.expectedConf {
					t.Errorf("expected confidence %f, got %f", tt.expectedConf, detection.Confidence)
				}
			} else {
				if detection != nil {
					t.Errorf("expected no detection, got %+v", detection)
				}
			}
		})
	}
}

// Test CRI-O Detector
func TestCRIODetector(t *testing.T) {
	tests := []struct {
		name         string
		files        map[string]bool
		data         map[string][]byte
		shouldDetect bool
		expectedConf float64
	}{
		{
			name:         "detects CRI-O socket",
			files:        map[string]bool{"/var/run/crio/crio.sock": true},
			shouldDetect: true,
			expectedConf: 0.95,
		},
		{
			name:         "detects CRI-O in cgroup",
			data:         map[string][]byte{"/proc/self/cgroup": []byte("1:name=systemd:/crio/abc123")},
			shouldDetect: true,
			expectedConf: 0.90,
		},
		{
			name:         "no detection when markers absent",
			files:        map[string]bool{},
			shouldDetect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := &CRIODetector{
				fs: &MockFileSystem{
					files: tt.files,
					data:  tt.data,
				},
			}

			detection, err := detector.Detect(context.Background())

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.shouldDetect {
				if detection == nil {
					t.Fatal("expected detection, got nil")
				}
				if detection.Value != "cri-o" {
					t.Errorf("expected value cri-o, got %s", detection.Value)
				}
				if detection.Confidence != tt.expectedConf {
					t.Errorf("expected confidence %f, got %f", tt.expectedConf, detection.Confidence)
				}
			} else {
				if detection != nil {
					t.Errorf("expected no detection, got %+v", detection)
				}
			}
		})
	}
}

// Test Firecracker Detector
func TestFirecrackerDetector(t *testing.T) {
	detector := &FirecrackerDetector{
		fs: &MockFileSystem{
			data: map[string][]byte{
				"/sys/class/dmi/id/product_name": []byte("Firecracker\n"),
			},
		},
	}

	detection, err := detector.Detect(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if detection == nil {
		t.Fatal("expected detection, got nil")
	}

	if detection.Value != "firecracker" {
		t.Errorf("expected firecracker, got %s", detection.Value)
	}

	if detection.Confidence != 0.98 {
		t.Errorf("expected confidence 0.98, got %f", detection.Confidence)
	}
}

// Test Kata Containers Detector
func TestKataContainersDetector(t *testing.T) {
	tests := []struct {
		name         string
		files        map[string]bool
		data         map[string][]byte
		shouldDetect bool
		expectedConf float64
	}{
		{
			name:         "detects Kata directory",
			files:        map[string]bool{"/run/kata-containers": true},
			shouldDetect: true,
			expectedConf: 0.95,
		},
		{
			name: "detects QEMU with lower confidence",
			data: map[string][]byte{
				"/sys/class/dmi/id/product_name": []byte("QEMU Virtual Machine"),
				"/proc/cpuinfo":                  []byte("model name: QEMU Virtual CPU"),
			},
			shouldDetect: true,
			expectedConf: 0.60,
		},
		{
			name:         "no detection when markers absent",
			files:        map[string]bool{},
			shouldDetect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := &KataContainersDetector{
				fs: &MockFileSystem{
					files: tt.files,
					data:  tt.data,
				},
			}

			detection, err := detector.Detect(context.Background())

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.shouldDetect {
				if detection == nil {
					t.Fatal("expected detection, got nil")
				}
				if detection.Value != "kata" {
					t.Errorf("expected value kata, got %s", detection.Value)
				}
				if detection.Confidence != tt.expectedConf {
					t.Errorf("expected confidence %f, got %f", tt.expectedConf, detection.Confidence)
				}
			} else {
				if detection != nil {
					t.Errorf("expected no detection, got %+v", detection)
				}
			}
		})
	}
}

// Test gVisor Detector
func TestGVisorDetector(t *testing.T) {
	tests := []struct {
		name         string
		files        map[string]bool
		data         map[string][]byte
		shouldDetect bool
		expectedConf float64
	}{
		{
			name:         "detects gvisor in cgroup",
			data:         map[string][]byte{"/proc/self/cgroup": []byte("1:name=systemd:/gvisor/abc123")},
			shouldDetect: true,
			expectedConf: 0.95,
		},
		{
			name:         "detects runsc in cgroup",
			data:         map[string][]byte{"/proc/self/cgroup": []byte("1:name=systemd:/runsc/abc123")},
			shouldDetect: true,
			expectedConf: 0.95,
		},
		{
			name:         "detects gvisor device",
			files:        map[string]bool{"/proc/self/root/dev/gvisor": true},
			shouldDetect: true,
			expectedConf: 0.98,
		},
		{
			name:         "no detection when markers absent",
			files:        map[string]bool{},
			shouldDetect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := &GVisorDetector{
				fs: &MockFileSystem{
					files: tt.files,
					data:  tt.data,
				},
			}

			detection, err := detector.Detect(context.Background())

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.shouldDetect {
				if detection == nil {
					t.Fatal("expected detection, got nil")
				}
				if detection.Value != "gvisor" {
					t.Errorf("expected value gvisor, got %s", detection.Value)
				}
				if detection.Confidence != tt.expectedConf {
					t.Errorf("expected confidence %f, got %f", tt.expectedConf, detection.Confidence)
				}
			} else {
				if detection != nil {
					t.Errorf("expected no detection, got %+v", detection)
				}
			}
		})
	}
}

// Test Sysbox Detector
func TestSysboxDetector(t *testing.T) {
	tests := []struct {
		name         string
		data         map[string][]byte
		shouldDetect bool
		expectedConf float64
		setup        func() func()
	}{
		{
			name: "detects Sysbox environment variable",
			setup: func() func() {
				origEnv := EnvironmentVariables
				EnvironmentVariables = map[string]string{
					"SYSBOX_CONTAINER": "true",
				}
				return func() { EnvironmentVariables = origEnv }
			},
			shouldDetect: true,
			expectedConf: 0.98,
		},
		{
			name:         "detects Sysbox in cgroup",
			data:         map[string][]byte{"/proc/self/cgroup": []byte("1:name=systemd:/sysbox/abc123")},
			shouldDetect: true,
			expectedConf: 0.90,
		},
		{
			name:         "no detection when markers absent",
			shouldDetect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				cleanup := tt.setup()
				defer cleanup()
			}

			detector := &SysboxDetector{
				fs: &MockFileSystem{
					data: tt.data,
				},
			}

			detection, err := detector.Detect(context.Background())

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.shouldDetect {
				if detection == nil {
					t.Fatal("expected detection, got nil")
				}
				if detection.Value != "sysbox" {
					t.Errorf("expected value sysbox, got %s", detection.Value)
				}
				if detection.Confidence != tt.expectedConf {
					t.Errorf("expected confidence %f, got %f", tt.expectedConf, detection.Confidence)
				}
			} else {
				if detection != nil {
					t.Errorf("expected no detection, got %+v", detection)
				}
			}
		})
	}
}

// Test AWS ECS Detector
func TestECSDetector(t *testing.T) {
	tests := []struct {
		name         string
		setup        func() func()
		shouldDetect bool
	}{
		{
			name: "detects ECS_CONTAINER_METADATA_URI",
			setup: func() func() {
				origEnv := EnvironmentVariables
				EnvironmentVariables = map[string]string{
					"ECS_CONTAINER_METADATA_URI": "http://169.254.170.2/v3",
				}
				return func() { EnvironmentVariables = origEnv }
			},
			shouldDetect: true,
		},
		{
			name: "detects ECS_CONTAINER_METADATA_URI_V4",
			setup: func() func() {
				origEnv := EnvironmentVariables
				EnvironmentVariables = map[string]string{
					"ECS_CONTAINER_METADATA_URI_V4": "http://169.254.170.2/v4",
				}
				return func() { EnvironmentVariables = origEnv }
			},
			shouldDetect: true,
		},
		{
			name:         "no detection when env vars absent",
			shouldDetect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				cleanup := tt.setup()
				defer cleanup()
			}

			detector := &ECSDetector{}
			detection, err := detector.Detect(context.Background())

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.shouldDetect {
				if detection == nil {
					t.Fatal("expected detection, got nil")
				}
				if detection.Value != "ecs" {
					t.Errorf("expected ecs, got %s", detection.Value)
				}
				if detection.Confidence != 0.98 {
					t.Errorf("expected confidence 0.98, got %f", detection.Confidence)
				}
			} else {
				if detection != nil {
					t.Errorf("expected no detection, got %+v", detection)
				}
			}
		})
	}
}

// Test AWS Fargate Detector
func TestFargateDetector(t *testing.T) {
	tests := []struct {
		name         string
		setup        func() func()
		shouldDetect bool
		expectedConf float64
	}{
		{
			name: "detects Fargate with AWS_EXECUTION_ENV",
			setup: func() func() {
				origEnv := EnvironmentVariables
				EnvironmentVariables = map[string]string{
					"AWS_EXECUTION_ENV": "AWS_ECS_FARGATE",
				}
				return func() { EnvironmentVariables = origEnv }
			},
			shouldDetect: true,
			expectedConf: 0.99,
		},
		{
			name: "detects Fargate with metadata URI",
			setup: func() func() {
				origEnv := EnvironmentVariables
				EnvironmentVariables = map[string]string{
					"ECS_CONTAINER_METADATA_URI_V4": "http://169.254.170.2/v4",
					"AWS_EXECUTION_ENV":             "AWS_ECS_EC2",
				}
				return func() { EnvironmentVariables = origEnv }
			},
			shouldDetect: true,
			expectedConf: 0.85,
		},
		{
			name: "no detection when env vars absent",
			setup: func() func() {
				origEnv := EnvironmentVariables
				EnvironmentVariables = map[string]string{}
				return func() { EnvironmentVariables = origEnv }
			},
			shouldDetect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				cleanup := tt.setup()
				defer cleanup()
			}

			detector := &FargateDetector{}
			detection, err := detector.Detect(context.Background())

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.shouldDetect {
				if detection == nil {
					t.Fatal("expected detection, got nil")
				}
				if detection.Value != "fargate" {
					t.Errorf("expected fargate, got %s", detection.Value)
				}
				if detection.Confidence != tt.expectedConf {
					t.Errorf("expected confidence %f, got %f", tt.expectedConf, detection.Confidence)
				}
			} else {
				if detection != nil {
					t.Errorf("expected no detection, got %+v", detection)
				}
			}
		})
	}
}

// Test Google Cloud Run Detector
func TestCloudRunDetector(t *testing.T) {
	tests := []struct {
		name         string
		setup        func() func()
		shouldDetect bool
		expectedConf float64
	}{
		{
			name: "detects Cloud Run with K_SERVICE",
			setup: func() func() {
				origEnv := EnvironmentVariables
				EnvironmentVariables = map[string]string{
					"K_SERVICE": "my-service",
				}
				return func() { EnvironmentVariables = origEnv }
			},
			shouldDetect: true,
			expectedConf: 0.98,
		},
		{
			name: "detects Cloud Run with K_REVISION",
			setup: func() func() {
				origEnv := EnvironmentVariables
				EnvironmentVariables = map[string]string{
					"K_REVISION": "my-service-00001-abc",
				}
				return func() { EnvironmentVariables = origEnv }
			},
			shouldDetect: true,
			expectedConf: 0.95,
		},
		{
			name:         "no detection when env vars absent",
			shouldDetect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				cleanup := tt.setup()
				defer cleanup()
			}

			detector := &CloudRunDetector{}
			detection, err := detector.Detect(context.Background())

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.shouldDetect {
				if detection == nil {
					t.Fatal("expected detection, got nil")
				}
				if detection.Value != "cloud-run" {
					t.Errorf("expected cloud-run, got %s", detection.Value)
				}
				if detection.Confidence != tt.expectedConf {
					t.Errorf("expected confidence %f, got %f", tt.expectedConf, detection.Confidence)
				}
			} else {
				if detection != nil {
					t.Errorf("expected no detection, got %+v", detection)
				}
			}
		})
	}
}

// Test AWS Lambda Container Detector
func TestLambdaContainerDetector(t *testing.T) {
	tests := []struct {
		name         string
		setup        func() func()
		shouldDetect bool
		expectedConf float64
	}{
		{
			name: "detects Lambda with AWS_LAMBDA_FUNCTION_NAME",
			setup: func() func() {
				origEnv := EnvironmentVariables
				EnvironmentVariables = map[string]string{
					"AWS_LAMBDA_FUNCTION_NAME": "my-function",
				}
				return func() { EnvironmentVariables = origEnv }
			},
			shouldDetect: true,
			expectedConf: 0.99,
		},
		{
			name: "detects Lambda with LAMBDA_TASK_ROOT",
			setup: func() func() {
				origEnv := EnvironmentVariables
				EnvironmentVariables = map[string]string{
					"LAMBDA_TASK_ROOT": "/var/task",
				}
				return func() { EnvironmentVariables = origEnv }
			},
			shouldDetect: true,
			expectedConf: 0.98,
		},
		{
			name:         "no detection when env vars absent",
			shouldDetect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				cleanup := tt.setup()
				defer cleanup()
			}

			detector := &LambdaContainerDetector{}
			detection, err := detector.Detect(context.Background())

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.shouldDetect {
				if detection == nil {
					t.Fatal("expected detection, got nil")
				}
				if detection.Value != "lambda" {
					t.Errorf("expected lambda, got %s", detection.Value)
				}
				if detection.Confidence != tt.expectedConf {
					t.Errorf("expected confidence %f, got %f", tt.expectedConf, detection.Confidence)
				}
			} else {
				if detection != nil {
					t.Errorf("expected no detection, got %+v", detection)
				}
			}
		})
	}
}

// Test Azure Container Instances Detector
func TestACIDetector(t *testing.T) {
	tests := []struct {
		name         string
		setup        func() func()
		shouldDetect bool
		expectedConf float64
	}{
		{
			name: "detects ACI with ACI_RESOURCE_GROUP",
			setup: func() func() {
				origEnv := EnvironmentVariables
				EnvironmentVariables = map[string]string{
					"ACI_RESOURCE_GROUP": "my-rg",
				}
				return func() { EnvironmentVariables = origEnv }
			},
			shouldDetect: true,
			expectedConf: 0.98,
		},
		{
			name: "detects ACI with CONTAINER_GROUP_NAME",
			setup: func() func() {
				origEnv := EnvironmentVariables
				EnvironmentVariables = map[string]string{
					"CONTAINER_GROUP_NAME": "my-group",
				}
				return func() { EnvironmentVariables = origEnv }
			},
			shouldDetect: true,
			expectedConf: 0.90,
		},
		{
			name:         "no detection when env vars absent",
			shouldDetect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				cleanup := tt.setup()
				defer cleanup()
			}

			detector := &ACIDetector{}
			detection, err := detector.Detect(context.Background())

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.shouldDetect {
				if detection == nil {
					t.Fatal("expected detection, got nil")
				}
				if detection.Value != "aci" {
					t.Errorf("expected aci, got %s", detection.Value)
				}
				if detection.Confidence != tt.expectedConf {
					t.Errorf("expected confidence %f, got %f", tt.expectedConf, detection.Confidence)
				}
			} else {
				if detection != nil {
					t.Errorf("expected no detection, got %+v", detection)
				}
			}
		})
	}
}

// Test Singularity Detector
func TestSingularityDetector(t *testing.T) {
	tests := []struct {
		name         string
		setup        func() func()
		shouldDetect bool
		expectedVal  string
		expectedConf float64
	}{
		{
			name: "detects Singularity with SINGULARITY_CONTAINER",
			setup: func() func() {
				origEnv := EnvironmentVariables
				EnvironmentVariables = map[string]string{
					"SINGULARITY_CONTAINER": "/path/to/container.sif",
				}
				return func() { EnvironmentVariables = origEnv }
			},
			shouldDetect: true,
			expectedVal:  "singularity",
			expectedConf: 0.99,
		},
		{
			name: "detects Apptainer with APPTAINER_CONTAINER",
			setup: func() func() {
				origEnv := EnvironmentVariables
				EnvironmentVariables = map[string]string{
					"APPTAINER_CONTAINER": "/path/to/container.sif",
				}
				return func() { EnvironmentVariables = origEnv }
			},
			shouldDetect: true,
			expectedVal:  "apptainer",
			expectedConf: 0.99,
		},
		{
			name: "detects Singularity with SINGULARITY_NAME",
			setup: func() func() {
				origEnv := EnvironmentVariables
				EnvironmentVariables = map[string]string{
					"SINGULARITY_NAME": "my-container",
				}
				return func() { EnvironmentVariables = origEnv }
			},
			shouldDetect: true,
			expectedVal:  "singularity",
			expectedConf: 0.95,
		},
		{
			name:         "no detection when env vars absent",
			shouldDetect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				cleanup := tt.setup()
				defer cleanup()
			}

			detector := &SingularityDetector{}
			detection, err := detector.Detect(context.Background())

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.shouldDetect {
				if detection == nil {
					t.Fatal("expected detection, got nil")
				}
				if detection.Value != tt.expectedVal {
					t.Errorf("expected %s, got %s", tt.expectedVal, detection.Value)
				}
				if detection.Confidence != tt.expectedConf {
					t.Errorf("expected confidence %f, got %f", tt.expectedConf, detection.Confidence)
				}
			} else {
				if detection != nil {
					t.Errorf("expected no detection, got %+v", detection)
				}
			}
		})
	}
}

// Test OCI Image Detector
func TestOCIImageDetector(t *testing.T) {
	tests := []struct {
		name         string
		files        map[string]bool
		setup        func() func()
		shouldDetect bool
		expectedConf float64
	}{
		{
			name:         "detects OCI from /var/lib/containers",
			files:        map[string]bool{"/var/lib/containers": true},
			shouldDetect: true,
			expectedConf: 0.80,
		},
		{
			name: "detects OCI from Podman env var",
			setup: func() func() {
				origEnv := EnvironmentVariables
				EnvironmentVariables = map[string]string{
					"PODMAN_SYSTEMD_UNIT": "podman.service",
				}
				return func() { EnvironmentVariables = origEnv }
			},
			shouldDetect: true,
			expectedConf: 0.85,
		},
		{
			name:         "no detection when markers absent",
			files:        map[string]bool{},
			shouldDetect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				cleanup := tt.setup()
				defer cleanup()
			}

			detector := &OCIImageDetector{
				fs: &MockFileSystem{
					files: tt.files,
				},
			}

			detection, err := detector.Detect(context.Background())

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.shouldDetect {
				if detection == nil {
					t.Fatal("expected detection, got nil")
				}
				if detection.Value != "oci" {
					t.Errorf("expected oci, got %s", detection.Value)
				}
				if detection.Confidence != tt.expectedConf {
					t.Errorf("expected confidence %f, got %f", tt.expectedConf, detection.Confidence)
				}
				if detection.Type != DetectionTypeImageFormat {
					t.Errorf("expected type ImageFormat, got %v", detection.Type)
				}
			} else {
				if detection != nil {
					t.Errorf("expected no detection, got %+v", detection)
				}
			}
		})
	}
}

// Test Singularity Image Detector
func TestSingularityImageDetector(t *testing.T) {
	tests := []struct {
		name         string
		setup        func() func()
		shouldDetect bool
	}{
		{
			name: "detects Singularity image format",
			setup: func() func() {
				origEnv := EnvironmentVariables
				EnvironmentVariables = map[string]string{
					"SINGULARITY_CONTAINER": "/path/to/container.sif",
				}
				return func() { EnvironmentVariables = origEnv }
			},
			shouldDetect: true,
		},
		{
			name: "detects Apptainer image format",
			setup: func() func() {
				origEnv := EnvironmentVariables
				EnvironmentVariables = map[string]string{
					"APPTAINER_CONTAINER": "/path/to/container.sif",
				}
				return func() { EnvironmentVariables = origEnv }
			},
			shouldDetect: true,
		},
		{
			name:         "no detection when env vars absent",
			shouldDetect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				cleanup := tt.setup()
				defer cleanup()
			}

			detector := &SingularityImageDetector{}
			detection, err := detector.Detect(context.Background())

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.shouldDetect {
				if detection == nil {
					t.Fatal("expected detection, got nil")
				}
				if detection.Value != "singularity" {
					t.Errorf("expected singularity, got %s", detection.Value)
				}
				if detection.Confidence != 0.95 {
					t.Errorf("expected confidence 0.95, got %f", detection.Confidence)
				}
				if detection.Type != DetectionTypeImageFormat {
					t.Errorf("expected type ImageFormat, got %v", detection.Type)
				}
			} else {
				if detection != nil {
					t.Errorf("expected no detection, got %+v", detection)
				}
			}
		})
	}
}
