// Copyright © 2022-2023 Christian R. Vozar
// Licensed under the MIT License. All rights reserved.

package criprof

import (
	"testing"
)

func TestGetScheduler(t *testing.T) {
	scheduler := getScheduler()

	// Scheduler should never be empty
	if scheduler == "" {
		t.Error("getScheduler() returned empty string")
	}

	// Scheduler should be one of the known values
	validSchedulers := map[string]bool{
		schedulerKubernetes:   true,
		schedulerNomad:        true,
		schedulerMesos:        true,
		schedulerSwarm:        true,
		schedulerUndetermined: true,
	}

	if !validSchedulers[scheduler] {
		t.Errorf("getScheduler() returned unknown scheduler: %s", scheduler)
	}
}

func TestIsSwarm(t *testing.T) {
	result := isSwarm()

	// On most development machines, this should be false
	// We're just testing that it doesn't panic and returns a bool
	if result != false && result != true {
		t.Error("isSwarm() returned non-boolean value")
	}
}

func TestIsKubernetes(t *testing.T) {
	result := isKubernetes()

	// We're testing that it doesn't panic and returns a bool
	if result != false && result != true {
		t.Error("isKubernetes() returned non-boolean value")
	}
}

func TestIsNomad(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() func()
		expected bool
	}{
		{
			name: "returns false in normal environment",
			setup: func() func() {
				return func() {}
			},
			expected: false,
		},
		{
			name: "detects Nomad from NOMAD_TASK_DIR",
			setup: func() func() {
				// Since isNomad() now uses the cached EnvironmentVariables map,
				// we need to update that instead of setting environment variables
				origEnv := EnvironmentVariables
				EnvironmentVariables = map[string]string{
					"NOMAD_TASK_DIR": "/tmp/nomad",
				}
				return func() {
					EnvironmentVariables = origEnv
				}
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := tt.setup()
			defer cleanup()

			result := isNomad()
			if result != tt.expected {
				t.Errorf("isNomad() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestIsMesos(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() func()
		expected bool
	}{
		{
			name: "returns false in normal environment",
			setup: func() func() {
				origEnv := EnvironmentVariables
				EnvironmentVariables = map[string]string{}
				return func() {
					EnvironmentVariables = origEnv
				}
			},
			expected: false,
		},
		{
			name: "detects Mesos from MESOS_TASK_ID",
			setup: func() func() {
				origEnv := EnvironmentVariables
				EnvironmentVariables = map[string]string{
					"MESOS_TASK_ID": "task-12345",
				}
				return func() {
					EnvironmentVariables = origEnv
				}
			},
			expected: true,
		},
		{
			name: "detects Mesos from MESOS_CONTAINER_NAME",
			setup: func() func() {
				origEnv := EnvironmentVariables
				EnvironmentVariables = map[string]string{
					"MESOS_CONTAINER_NAME": "container-12345",
				}
				return func() {
					EnvironmentVariables = origEnv
				}
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := tt.setup()
			defer cleanup()

			result := isMesos()
			if result != tt.expected {
				t.Errorf("isMesos() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestSchedulerConstants(t *testing.T) {
	// Verify scheduler constants are defined correctly
	schedulers := map[string]string{
		"kubernetes":   schedulerKubernetes,
		"nomad":        schedulerNomad,
		"mesos":        schedulerMesos,
		"swarm":        schedulerSwarm,
		"undetermined": schedulerUndetermined,
	}

	for expected, actual := range schedulers {
		if actual != expected {
			t.Errorf("Scheduler constant mismatch: got %s, expected %s", actual, expected)
		}
	}
}

func BenchmarkGetScheduler(b *testing.B) {
	for i := 0; i < b.N; i++ {
		getScheduler()
	}
}

func BenchmarkIsKubernetes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		isKubernetes()
	}
}

func BenchmarkIsSwarm(b *testing.B) {
	for i := 0; i < b.N; i++ {
		isSwarm()
	}
}

func BenchmarkIsNomad(b *testing.B) {
	for i := 0; i < b.N; i++ {
		isNomad()
	}
}

func BenchmarkIsMesos(b *testing.B) {
	for i := 0; i < b.N; i++ {
		isMesos()
	}
}
