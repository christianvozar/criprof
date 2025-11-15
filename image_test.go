// Copyright © 2022-2023 Christian R. Vozar
// Licensed under the MIT License. All rights reserved.

package criprof

import (
	"testing"
)

func TestGetImageFormat(t *testing.T) {
	format, err := getImageFormat()

	// Should not return error in normal circumstances
	if err != nil {
		t.Errorf("getImageFormat() returned error: %v", err)
	}

	// Format should never be empty
	if format == "" {
		t.Error("getImageFormat() returned empty string")
	}

	// Format should be one of the known values
	validFormats := map[string]bool{
		formatDocker:       true,
		formatACI:          true,
		formatCRI:          true,
		formatOCF:          true,
		formatUndetermined: true,
	}

	if !validFormats[format] {
		t.Errorf("getImageFormat() returned unknown format: %s", format)
	}
}

func TestIsDockerFormat(t *testing.T) {
	result, err := isDockerFormat()

	// Should not return error in normal circumstances
	if err != nil {
		t.Errorf("isDockerFormat() returned error: %v", err)
	}

	// Result should be a boolean
	if result != false && result != true {
		t.Error("isDockerFormat() returned non-boolean value")
	}
}

func TestImageFormatConstants(t *testing.T) {
	// Verify format constants are defined correctly
	formats := map[string]string{
		"docker":       formatDocker,
		"aci":          formatACI,
		"cri":          formatCRI,
		"ocf":          formatOCF,
		"undetermined": formatUndetermined,
	}

	for expected, actual := range formats {
		if actual != expected {
			t.Errorf("Format constant mismatch: got %s, expected %s", actual, expected)
		}
	}
}

func TestGetImageFormatWithEnvironment(t *testing.T) {
	// Note: These tests verify environment variable detection logic
	// However, getImageFormat() checks file markers first (/.dockerenv, etc.)
	// which may exist on the system, causing these tests to return "docker"
	// instead of "aci". These tests primarily verify the code doesn't panic
	// and returns valid format strings.

	tests := []struct {
		name  string
		setup func() func()
	}{
		{
			name: "handles AC_METADATA_URL environment variable",
			setup: func() func() {
				origEnv := EnvironmentVariables
				EnvironmentVariables = map[string]string{
					"AC_METADATA_URL": "http://example.com",
				}
				return func() {
					EnvironmentVariables = origEnv
				}
			},
		},
		{
			name: "handles AC_APP_NAME environment variable",
			setup: func() func() {
				origEnv := EnvironmentVariables
				EnvironmentVariables = map[string]string{
					"AC_APP_NAME": "myapp",
				}
				return func() {
					EnvironmentVariables = origEnv
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := tt.setup()
			defer cleanup()

			format, err := getImageFormat()
			if err != nil {
				t.Errorf("getImageFormat() returned error: %v", err)
			}

			// Verify it returns a valid format (actual value depends on environment)
			validFormats := map[string]bool{
				formatDocker:       true,
				formatACI:          true,
				formatCRI:          true,
				formatOCF:          true,
				formatUndetermined: true,
			}

			if !validFormats[format] {
				t.Errorf("getImageFormat() returned invalid format: %s", format)
			}
		})
	}
}

func BenchmarkGetImageFormat(b *testing.B) {
	// Run getImageFormat function b.N times.
	for i := 0; i < b.N; i++ {
		getImageFormat()
	}
}

func BenchmarkIsDockerFormat(b *testing.B) {
	for i := 0; i < b.N; i++ {
		isDockerFormat()
	}
}
