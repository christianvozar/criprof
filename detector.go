// Copyright © 2022-2023 Christian R. Vozar
// Licensed under the MIT License. All rights reserved.

package criprof

import (
	"context"
)

// DetectionType represents the type of detection being performed
type DetectionType int

const (
	// DetectionTypeRuntime detects container runtime engines
	DetectionTypeRuntime DetectionType = iota
	// DetectionTypeScheduler detects orchestration platforms
	DetectionTypeScheduler
	// DetectionTypeImageFormat detects container image formats
	DetectionTypeImageFormat
)

// String returns a string representation of the DetectionType
func (dt DetectionType) String() string {
	switch dt {
	case DetectionTypeRuntime:
		return "runtime"
	case DetectionTypeScheduler:
		return "scheduler"
	case DetectionTypeImageFormat:
		return "image_format"
	default:
		return "unknown"
	}
}

// Detection represents a single detection result from a detector
type Detection struct {
	// Type indicates what kind of detection this is (runtime, scheduler, etc.)
	Type DetectionType

	// Value is the detected value (e.g., "docker", "kubernetes")
	Value string

	// Confidence is a score from 0.0 to 1.0 indicating how confident
	// this detection is. Higher values mean more definitive evidence.
	Confidence float64

	// Source identifies which detector produced this result
	Source string
}

// Detector defines the interface for detection strategies
//
// Each detector implements a specific detection method (e.g., checking for
// a particular file, environment variable, or network endpoint). Detectors
// are executed by the Engine in priority order.
type Detector interface {
	// Name returns a unique identifier for this detector
	Name() string

	// Detect performs the detection logic and returns a result
	//
	// The context can be used for cancellation and timeouts.
	// Returns nil detection (not an error) if nothing was detected.
	// Returns an error only if the detection logic itself failed.
	Detect(ctx context.Context) (*Detection, error)

	// Priority returns the execution priority (higher = runs earlier)
	//
	// Fast detectors (file checks) should have high priority (90-100)
	// Medium detectors (env vars) should have medium priority (50-80)
	// Slow detectors (network) should have low priority (1-40)
	Priority() int
}
