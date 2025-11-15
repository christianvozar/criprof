// Copyright © 2022-2023 Christian R. Vozar
// Licensed under the MIT License. All rights reserved.

package criprof

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"testing"
	"time"
)

// MockFileSystem implements FileSystem for testing
type MockFileSystem struct {
	files map[string]bool
	data  map[string][]byte
}

func (m *MockFileSystem) Stat(name string) (os.FileInfo, error) {
	if m.files != nil && m.files[name] {
		return nil, nil // File exists
	}
	return nil, os.ErrNotExist
}

func (m *MockFileSystem) ReadFile(name string) ([]byte, error) {
	if m.data != nil {
		if data, ok := m.data[name]; ok {
			return data, nil
		}
	}
	return nil, os.ErrNotExist
}

// MockNetwork implements Network for testing
type MockNetwork struct {
	dialResults map[string]error
	httpResults map[string]*http.Response
}

func (m *MockNetwork) DialTimeout(network, address string, timeout time.Duration) (net.Conn, error) {
	if m.dialResults != nil {
		if err, ok := m.dialResults[address]; ok {
			return nil, err
		}
	}
	return nil, errors.New("connection refused")
}

func (m *MockNetwork) HTTPGet(ctx context.Context, url string) (*http.Response, error) {
	if m.httpResults != nil {
		if resp, ok := m.httpResults[url]; ok {
			return resp, nil
		}
	}
	return nil, errors.New("not found")
}

// Test Docker File Detector
func TestDockerFileDetector(t *testing.T) {
	tests := []struct {
		name           string
		files          map[string]bool
		expectedValue  string
		expectedConf   float64
		shouldDetect   bool
	}{
		{
			name:          "detects .dockerenv",
			files:         map[string]bool{"/.dockerenv": true},
			expectedValue: "docker",
			expectedConf:  0.95,
			shouldDetect:  true,
		},
		{
			name:          "detects .dockerinit",
			files:         map[string]bool{"/.dockerinit": true},
			expectedValue: "docker",
			expectedConf:  0.90,
			shouldDetect:  true,
		},
		{
			name:         "no detection when files absent",
			files:        map[string]bool{},
			shouldDetect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := &DockerFileDetector{
				fs: &MockFileSystem{files: tt.files},
			}

			detection, err := detector.Detect(context.Background())

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.shouldDetect {
				if detection == nil {
					t.Fatal("expected detection, got nil")
				}
				if detection.Value != tt.expectedValue {
					t.Errorf("expected value %s, got %s", tt.expectedValue, detection.Value)
				}
				if detection.Confidence != tt.expectedConf {
					t.Errorf("expected confidence %f, got %f", tt.expectedConf, detection.Confidence)
				}
				if detection.Type != DetectionTypeRuntime {
					t.Errorf("expected type Runtime, got %v", detection.Type)
				}
			} else {
				if detection != nil {
					t.Errorf("expected no detection, got %+v", detection)
				}
			}
		})
	}
}

// Test Docker Cgroup Detector
func TestDockerCgroupDetector(t *testing.T) {
	tests := []struct {
		name         string
		cgroupData   string
		shouldDetect bool
	}{
		{
			name:         "detects docker in cgroup",
			cgroupData:   "1:name=systemd:/docker/abc123",
			shouldDetect: true,
		},
		{
			name:         "no detection without docker",
			cgroupData:   "1:name=systemd:/user.slice",
			shouldDetect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := &DockerCgroupDetector{
				fs: &MockFileSystem{
					data: map[string][]byte{
						"/proc/self/cgroup": []byte(tt.cgroupData),
					},
				},
			}

			detection, err := detector.Detect(context.Background())

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.shouldDetect && detection == nil {
				t.Error("expected detection, got nil")
			}
			if !tt.shouldDetect && detection != nil {
				t.Errorf("expected no detection, got %+v", detection)
			}
		})
	}
}

// Test Kubernetes Service Account Detector
func TestKubernetesServiceAccountDetector(t *testing.T) {
	detector := &KubernetesServiceAccountDetector{
		fs: &MockFileSystem{
			files: map[string]bool{
				"/run/secrets/kubernetes.io/serviceaccount/token": true,
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

	if detection.Value != "kubernetes" {
		t.Errorf("expected kubernetes, got %s", detection.Value)
	}

	if detection.Confidence != 0.99 {
		t.Errorf("expected confidence 0.99, got %f", detection.Confidence)
	}

	if detection.Type != DetectionTypeScheduler {
		t.Errorf("expected type Scheduler, got %v", detection.Type)
	}
}

// Test WASM Detector
func TestWASMDetector(t *testing.T) {
	detector := &WASMDetector{}

	detection, err := detector.Detect(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Detection depends on build target
	// In normal builds, should not detect
	// When built for WASM, should detect with confidence 1.0
	if detection != nil {
		if detection.Confidence != 1.0 {
			t.Errorf("WASM detection should have confidence 1.0, got %f", detection.Confidence)
		}
	}
}

// Test Detection Type Stringer
func TestDetectionTypeString(t *testing.T) {
	tests := []struct {
		dt       DetectionType
		expected string
	}{
		{DetectionTypeRuntime, "runtime"},
		{DetectionTypeScheduler, "scheduler"},
		{DetectionTypeImageFormat, "image_format"},
		{DetectionType(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.dt.String(); got != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, got)
			}
		})
	}
}
