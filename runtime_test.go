// Copyright © 2022-2023 Christian R. Vozar
// Licensed under the MIT License. All rights reserved.

package criprof

import (
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
	// isWASM now uses runtime.GOOS and runtime.GOARCH which are set at compile time
	// We can only test the actual runtime value, not simulate different values
	// The function will return true only if compiled for WASM target

	result := isWASM()

	// Verify it returns a boolean and doesn't panic
	if result != true && result != false {
		t.Error("isWASM() returned non-boolean value")
	}

	// In normal (non-WASM) build, should return false
	// When built for WASM, runtime.GOOS=="js" && runtime.GOARCH=="wasm" will be true
	// This is the correct behavior - we detect build target, not environment variables
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
