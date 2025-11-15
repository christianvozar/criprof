// Copyright © 2022-2023 Christian R. Vozar
// Licensed under the MIT License. All rights reserved.

package criprof

import (
	"encoding/json"
	"os"
	"testing"
)

func TestNew(t *testing.T) {
	inventory := New()

	if inventory == nil {
		t.Fatal("New() returned nil")
	}

	// Test that all fields are populated
	if inventory.PID == 0 {
		t.Error("New() inventory.PID is 0")
	}

	// PID should match current process
	expectedPID := os.Getpid()
	if inventory.PID != expectedPID {
		t.Errorf("New() inventory.PID = %v, expected %v", inventory.PID, expectedPID)
	}

	// Runtime should be set (even if undetermined)
	if inventory.Runtime == "" {
		t.Error("New() inventory.Runtime is empty")
	}

	// Scheduler should be set (even if undetermined)
	if inventory.Scheduler == "" {
		t.Error("New() inventory.Scheduler is empty")
	}

	// ImageFormat should be set (even if undetermined)
	if inventory.ImageFormat == "" {
		t.Error("New() inventory.ImageFormat is empty")
	}

	// ID should be set (even if undetermined)
	if inventory.ID == "" {
		t.Error("New() inventory.ID is empty")
	}
}

func TestInventoryJSON(t *testing.T) {
	inventory := New()
	jsonStr := inventory.JSON()

	// Verify we got a non-empty string
	if jsonStr == "" {
		t.Fatal("JSON() returned empty string")
	}

	// Verify it's valid JSON by unmarshaling it
	var parsed map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &parsed)
	if err != nil {
		t.Fatalf("JSON() returned invalid JSON: %v", err)
	}

	// Verify all expected fields are present
	requiredFields := []string{"hostname", "id", "image_format", "pid", "runtime", "scheduler"}
	for _, field := range requiredFields {
		if _, exists := parsed[field]; !exists {
			t.Errorf("JSON() missing required field: %s", field)
		}
	}

	// Verify PID is a number
	if pid, ok := parsed["pid"].(float64); !ok {
		t.Error("JSON() pid field is not a number")
	} else if int(pid) != inventory.PID {
		t.Errorf("JSON() pid = %v, expected %v", int(pid), inventory.PID)
	}

	// Verify runtime matches
	if runtime, ok := parsed["runtime"].(string); !ok {
		t.Error("JSON() runtime field is not a string")
	} else if runtime != inventory.Runtime {
		t.Errorf("JSON() runtime = %v, expected %v", runtime, inventory.Runtime)
	}
}

func TestInventoryJSONStructure(t *testing.T) {
	// Create an inventory with known values
	inventory := &Inventory{
		Hostname:    "test-host",
		ID:          "abc123",
		ImageFormat: "docker",
		PID:         12345,
		Runtime:     "docker",
		Scheduler:   "kubernetes",
	}

	jsonStr := inventory.JSON()

	// Unmarshal and verify each field
	var result Inventory
	err := json.Unmarshal([]byte(jsonStr), &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if result.Hostname != inventory.Hostname {
		t.Errorf("JSON hostname = %v, expected %v", result.Hostname, inventory.Hostname)
	}
	if result.ID != inventory.ID {
		t.Errorf("JSON id = %v, expected %v", result.ID, inventory.ID)
	}
	if result.ImageFormat != inventory.ImageFormat {
		t.Errorf("JSON image_format = %v, expected %v", result.ImageFormat, inventory.ImageFormat)
	}
	if result.PID != inventory.PID {
		t.Errorf("JSON pid = %v, expected %v", result.PID, inventory.PID)
	}
	if result.Runtime != inventory.Runtime {
		t.Errorf("JSON runtime = %v, expected %v", result.Runtime, inventory.Runtime)
	}
	if result.Scheduler != inventory.Scheduler {
		t.Errorf("JSON scheduler = %v, expected %v", result.Scheduler, inventory.Scheduler)
	}
}
