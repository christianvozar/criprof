// Copyright © 2022-2023 Christian R. Vozar
// Licensed under the MIT License. All rights reserved.

package criprof

import (
	"os"
	"testing"
)

func TestEnvironMap(t *testing.T) {
	// Set a test environment variable
	testKey := "CRIPROF_TEST_VAR"
	testValue := "test_value"
	os.Setenv(testKey, testValue)
	defer os.Unsetenv(testKey)

	// Call environMap
	result := environMap()

	// Verify it returns a map
	if result == nil {
		t.Fatal("environMap() returned nil")
	}

	// Verify our test variable is in the map
	if val, exists := result[testKey]; !exists {
		t.Errorf("environMap() missing key %s", testKey)
	} else if val != testValue {
		t.Errorf("environMap()[%s] = %v, expected %v", testKey, val, testValue)
	}

	// Verify PATH exists (should be in any environment)
	if _, exists := result["PATH"]; !exists {
		t.Error("environMap() missing PATH variable")
	}
}

func TestEnvironmentVariablesInitialization(t *testing.T) {
	// EnvironmentVariables should be populated at init
	if EnvironmentVariables == nil {
		t.Fatal("EnvironmentVariables is nil after package initialization")
	}

	// Should contain common environment variables
	if len(EnvironmentVariables) == 0 {
		t.Error("EnvironmentVariables is empty")
	}
}
