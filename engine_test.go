// Copyright © 2022-2023 Christian R. Vozar
// Licensed under the MIT License. All rights reserved.

package criprof

import (
	"context"
	"testing"
	"time"
)

// TestEngineDetectAll tests the detection engine
func TestEngineDetectAll(t *testing.T) {
	mockFS := &MockFileSystem{
		files: map[string]bool{
			"/.dockerenv": true,
			"/run/secrets/kubernetes.io/serviceaccount/token": true,
		},
	}

	detectors := []Detector{
		&DockerFileDetector{fs: mockFS},
		&KubernetesServiceAccountDetector{fs: mockFS},
		&DockerImageDetector{fs: mockFS},
	}

	engine := NewEngine(EngineConfig{
		Detectors:     detectors,
		EnableCaching: false,
	})

	inventory, err := engine.DetectAll(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if inventory == nil {
		t.Fatal("expected inventory, got nil")
	}

	if inventory.Runtime != "docker" {
		t.Errorf("expected runtime docker, got %s", inventory.Runtime)
	}

	if inventory.Scheduler != "kubernetes" {
		t.Errorf("expected scheduler kubernetes, got %s", inventory.Scheduler)
	}

	if inventory.ImageFormat != "docker" {
		t.Errorf("expected image format docker, got %s", inventory.ImageFormat)
	}
}

// TestEngineWithCaching tests cache functionality
func TestEngineWithCaching(t *testing.T) {
	callCount := 0

	// Create a custom detector that tracks calls
	detector := &testDetector{
		onDetect: func() {
			callCount++
		},
	}

	engine := NewEngine(EngineConfig{
		Detectors:     []Detector{detector},
		EnableCaching: true,
		CacheTTL:      1 * time.Second,
	})

	// First call should execute detector
	_, err := engine.DetectAll(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if callCount != 1 {
		t.Errorf("expected 1 detector call, got %d", callCount)
	}

	// Second call should use cache
	_, err = engine.DetectAll(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if callCount != 1 {
		t.Errorf("expected cached result (1 call), got %d calls", callCount)
	}

	// Wait for cache to expire
	time.Sleep(1100 * time.Millisecond)

	// Third call should execute detector again
	_, err = engine.DetectAll(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if callCount != 2 {
		t.Errorf("expected 2 detector calls after cache expiry, got %d", callCount)
	}
}

// TestEngineInvalidateCache tests cache invalidation
func TestEngineInvalidateCache(t *testing.T) {
	callCount := 0

	detector := &testDetector{
		onDetect: func() {
			callCount++
		},
	}

	engine := NewEngine(EngineConfig{
		Detectors:     []Detector{detector},
		EnableCaching: true,
		CacheTTL:      10 * time.Second,
	})

	// First call
	_, _ = engine.DetectAll(context.Background())

	// Second call uses cache
	_, _ = engine.DetectAll(context.Background())

	if callCount != 1 {
		t.Errorf("expected 1 call, got %d", callCount)
	}

	// Invalidate cache
	engine.InvalidateCache()

	// Third call should execute detector
	_, _ = engine.DetectAll(context.Background())

	if callCount != 2 {
		t.Errorf("expected 2 calls after invalidation, got %d", callCount)
	}
}

// TestEngineContextCancellation tests context cancellation
func TestEngineContextCancellation(t *testing.T) {
	detector := &slowDetector{
		delay: 2 * time.Second,
	}

	engine := NewEngine(EngineConfig{
		Detectors:     []Detector{detector},
		EnableCaching: false,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := engine.DetectAll(ctx)

	if err == nil {
		t.Error("expected context deadline error, got nil")
	}

	if err != context.DeadlineExceeded {
		t.Errorf("expected DeadlineExceeded, got %v", err)
	}
}

// TestEngineConfidenceScoring tests that highest confidence wins
func TestEngineConfidenceScoring(t *testing.T) {
	detectors := []Detector{
		&fakeDetector{
			detection: &Detection{
				Type:       DetectionTypeRuntime,
				Value:      "runtime1",
				Confidence: 0.5,
				Source:     "low-confidence",
			},
		},
		&fakeDetector{
			detection: &Detection{
				Type:       DetectionTypeRuntime,
				Value:      "runtime2",
				Confidence: 0.9,
				Source:     "high-confidence",
			},
		},
	}

	engine := NewEngine(EngineConfig{
		Detectors:     detectors,
		EnableCaching: false,
	})

	inventory, err := engine.DetectAll(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if inventory.Runtime != "runtime2" {
		t.Errorf("expected highest confidence runtime2, got %s", inventory.Runtime)
	}
}

// TestEnginePrioritySorting tests that detectors run in priority order
func TestEnginePrioritySorting(t *testing.T) {
	var executionOrder []int

	detectors := []Detector{
		&priorityDetector{priority: 50, onRun: func() { executionOrder = append(executionOrder, 50) }},
		&priorityDetector{priority: 100, onRun: func() { executionOrder = append(executionOrder, 100) }},
		&priorityDetector{priority: 75, onRun: func() { executionOrder = append(executionOrder, 75) }},
	}

	engine := NewEngine(EngineConfig{
		Detectors:     detectors,
		EnableCaching: false,
	})

	_, _ = engine.DetectAll(context.Background())

	// Should execute in order: 100, 75, 50
	expected := []int{100, 75, 50}
	if len(executionOrder) != len(expected) {
		t.Fatalf("expected %d executions, got %d", len(expected), len(executionOrder))
	}

	for i, p := range expected {
		if executionOrder[i] != p {
			t.Errorf("execution order[%d]: expected %d, got %d", i, p, executionOrder[i])
		}
	}
}

// Helper test detectors

type testDetector struct {
	onDetect func()
}

func (d *testDetector) Name() string                                   { return "test-detector" }
func (d *testDetector) Priority() int                                  { return 50 }
func (d *testDetector) Detect(ctx context.Context) (*Detection, error) {
	if d.onDetect != nil {
		d.onDetect()
	}
	return nil, nil
}

type slowDetector struct {
	delay time.Duration
}

func (d *slowDetector) Name() string    { return "slow-detector" }
func (d *slowDetector) Priority() int   { return 50 }
func (d *slowDetector) Detect(ctx context.Context) (*Detection, error) {
	select {
	case <-time.After(d.delay):
		return nil, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

type fakeDetector struct {
	detection *Detection
}

func (d *fakeDetector) Name() string    { return "fake-detector" }
func (d *fakeDetector) Priority() int   { return 50 }
func (d *fakeDetector) Detect(ctx context.Context) (*Detection, error) {
	return d.detection, nil
}

type priorityDetector struct {
	priority int
	onRun    func()
}

func (d *priorityDetector) Name() string    { return "priority-detector" }
func (d *priorityDetector) Priority() int   { return d.priority }
func (d *priorityDetector) Detect(ctx context.Context) (*Detection, error) {
	if d.onRun != nil {
		d.onRun()
	}
	return nil, nil
}
