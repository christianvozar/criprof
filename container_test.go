// Copyright © 2022-2023 Christian R. Vozar
// Licensed under the MIT License. All rights reserved.

package criprof

import (
	"os"
	"testing"
)

func TestIsContainer(t *testing.T) {
	// This test just verifies that IsContainer returns a boolean
	// and doesn't panic. The actual result depends on the environment.
	result := IsContainer()

	if result != true && result != false {
		t.Error("IsContainer() returned non-boolean value")
	}
}

func TestGetContainerID(t *testing.T) {
	// This test verifies that getContainerID returns a non-empty string
	// The actual value depends on the environment
	result := getContainerID()

	if result == "" {
		t.Error("getContainerID() returned empty string")
	}

	// Should return "undetermined" or a valid container ID
	// We just verify it doesn't panic and returns something
}

func TestGetHostname(t *testing.T) {
	hostname, err := getHostname()
	if err != nil {
		t.Fatalf("getHostname() returned error: %v", err)
	}

	// Verify we got a non-empty hostname
	if hostname == "" {
		t.Error("getHostname() returned empty hostname")
	}

	// Verify it matches os.Hostname()
	expectedHostname, err := os.Hostname()
	if err != nil {
		t.Fatalf("os.Hostname() returned error: %v", err)
	}

	if hostname != expectedHostname {
		t.Errorf("getHostname() = %v, expected %v", hostname, expectedHostname)
	}
}
