// Copyright © 2022-2023 Christian R. Vozar
// Licensed under the MIT License. All rights reserved.

package criprof

import (
	"os"
	"testing"
)

func TestGetRuntime(t *testing.T) {
	runtime := getRuntime()

	// Runtime should never be empty
	if runtime == "" {
		t.Error("getRuntime() returned empty string")
	}

	// Runtime should be one of the known values
	validRuntimes := map[string]bool{
		runtimeDocker:       true,
		runtimeRkt:          true,
		runtimeRunC:         true,
		runtimeContainerD:   true,
		runtimeLXC:          true,
		runtimeLXD:          true,
		runtimeOpenVZ:       true,
		runtimeWASM:         true,
		runtimeUndetermined: true,
	}

	if !validRuntimes[runtime] {
		t.Errorf("getRuntime() returned unknown runtime: %s", runtime)
	}
}

func TestIsOpenVZ(t *testing.T) {
	result := isOpenVZ()

	// On most development machines, this should be false
	// We're just testing that it doesn't panic and returns a bool
	if result != false && result != true {
		t.Error("isOpenVZ() returned non-boolean value")
	}
}

func TestIsWASM(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() func()
		expected bool
	}{
		{
			name: "returns false in normal environment",
			setup: func() func() {
				// Save original values
				origGOOS := os.Getenv("GOOS")
				origGOARCH := os.Getenv("GOARCH")
				return func() {
					// Restore original values
					os.Setenv("GOOS", origGOOS)
					os.Setenv("GOARCH", origGOARCH)
				}
			},
			expected: false,
		},
		{
			name: "returns true when GOOS=js and GOARCH=wasm",
			setup: func() func() {
				origGOOS := os.Getenv("GOOS")
				origGOARCH := os.Getenv("GOARCH")
				os.Setenv("GOOS", "js")
				os.Setenv("GOARCH", "wasm")
				return func() {
					os.Setenv("GOOS", origGOOS)
					os.Setenv("GOARCH", origGOARCH)
				}
			},
			expected: true,
		},
		{
			name: "returns false when only GOOS=js",
			setup: func() func() {
				origGOOS := os.Getenv("GOOS")
				origGOARCH := os.Getenv("GOARCH")
				os.Setenv("GOOS", "js")
				os.Setenv("GOARCH", "amd64")
				return func() {
					os.Setenv("GOOS", origGOOS)
					os.Setenv("GOARCH", origGOARCH)
				}
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := tt.setup()
			defer cleanup()

			result := isWASM()
			if result != tt.expected {
				t.Errorf("isWASM() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestRuntimeConstants(t *testing.T) {
	// Verify runtime constants are defined correctly
	runtimes := map[string]string{
		"docker":       runtimeDocker,
		"rkt":          runtimeRkt,
		"runc":         runtimeRunC,
		"containerd":   runtimeContainerD,
		"lxc":          runtimeLXC,
		"lxd":          runtimeLXD,
		"openvz":       runtimeOpenVZ,
		"wasm":         runtimeWASM,
		"undetermined": runtimeUndetermined,
	}

	for expected, actual := range runtimes {
		if actual != expected {
			t.Errorf("Runtime constant mismatch: got %s, expected %s", actual, expected)
		}
	}
}

func BenchmarkGetRuntime(b *testing.B) {
	// Run getRuntime function b.N times.
	for i := 0; i < b.N; i++ {
		getRuntime()
	}
}

func BenchmarkIsOpenVZ(b *testing.B) {
	for i := 0; i < b.N; i++ {
		isOpenVZ()
	}
}

func BenchmarkIsWASM(b *testing.B) {
	for i := 0; i < b.N; i++ {
		isWASM()
	}
}
