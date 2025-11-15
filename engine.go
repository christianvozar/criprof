// Copyright © 2022-2023 Christian R. Vozar
// Licensed under the MIT License. All rights reserved.

package criprof

import (
	"context"
	"os"
	"sort"
	"sync"
	"time"
)

// Engine coordinates multiple detectors to build an Inventory
type Engine struct {
	detectors []Detector
	cache     *DetectionCache
	mu        sync.RWMutex
}

// EngineConfig configures an Engine
type EngineConfig struct {
	// Detectors is the list of detectors to run
	Detectors []Detector

	// EnableCaching enables result caching to avoid re-detection
	EnableCaching bool

	// CacheTTL is how long cached results are valid
	CacheTTL time.Duration
}

// NewEngine creates a new detection engine with the given configuration
func NewEngine(config EngineConfig) *Engine {
	engine := &Engine{
		detectors: config.Detectors,
	}

	if config.EnableCaching {
		engine.cache = NewDetectionCache(config.CacheTTL)
	}

	// Sort detectors by priority (highest first)
	sort.Slice(engine.detectors, func(i, j int) bool {
		return engine.detectors[i].Priority() > engine.detectors[j].Priority()
	})

	return engine
}

// DetectAll runs all detectors and builds an Inventory from the results
//
// Detectors are run in priority order. For each detection type (runtime,
// scheduler, image format), the detection with the highest confidence is used.
//
// Results are cached if caching is enabled.
func (e *Engine) DetectAll(ctx context.Context) (*Inventory, error) {
	// Check cache first
	if e.cache != nil {
		if cached := e.cache.Get(); cached != nil {
			return cached, nil
		}
	}

	// Run all detectors and collect results
	results := make(map[DetectionType]*Detection)
	e.mu.RLock()
	detectors := e.detectors
	e.mu.RUnlock()

	for _, detector := range detectors {
		// Check if context was cancelled
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		detection, err := detector.Detect(ctx)
		if err != nil {
			// If error is context-related, propagate it
			if err == context.Canceled || err == context.DeadlineExceeded {
				return nil, err
			}
			// For other errors, log and continue with other detectors
			// In production, you might want to use a logger here
			continue
		}

		if detection == nil {
			continue
		}

		// Keep detection with highest confidence for each type
		existing := results[detection.Type]
		if existing == nil || detection.Confidence > existing.Confidence {
			results[detection.Type] = detection
		}
	}

	// Build inventory from results
	inventory := e.buildInventory(results)

	// Cache the result
	if e.cache != nil {
		e.cache.Set(inventory)
	}

	return inventory, nil
}

// buildInventory creates an Inventory from detection results
func (e *Engine) buildInventory(results map[DetectionType]*Detection) *Inventory {
	inv := &Inventory{
		PID:         os.Getpid(),
		Runtime:     "undetermined",
		Scheduler:   "undetermined",
		ImageFormat: "undetermined",
		ID:          "undetermined",
	}

	// Apply detected values
	if det := results[DetectionTypeRuntime]; det != nil {
		inv.Runtime = det.Value
	}
	if det := results[DetectionTypeScheduler]; det != nil {
		inv.Scheduler = det.Value
	}
	if det := results[DetectionTypeImageFormat]; det != nil {
		inv.ImageFormat = det.Value
	}

	// Get hostname
	if h, err := os.Hostname(); err == nil {
		inv.Hostname = h
	} else {
		inv.Hostname = "unknown"
	}

	// Get container ID (reuse existing logic)
	inv.ID = getContainerID()

	return inv
}

// InvalidateCache clears the cached inventory
func (e *Engine) InvalidateCache() {
	if e.cache != nil {
		e.cache.Invalidate()
	}
}

// DetectionCache caches detection results with TTL
type DetectionCache struct {
	inventory *Inventory
	timestamp time.Time
	ttl       time.Duration
	mu        sync.RWMutex
}

// NewDetectionCache creates a new cache with the given TTL
func NewDetectionCache(ttl time.Duration) *DetectionCache {
	return &DetectionCache{
		ttl: ttl,
	}
}

// Get retrieves the cached inventory if still valid
func (c *DetectionCache) Get() *Inventory {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.inventory == nil {
		return nil
	}

	// Check if cache expired
	if time.Since(c.timestamp) > c.ttl {
		return nil
	}

	return c.inventory
}

// Set stores an inventory in the cache
func (c *DetectionCache) Set(inv *Inventory) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.inventory = inv
	c.timestamp = time.Now()
}

// Invalidate clears the cache
func (c *DetectionCache) Invalidate() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.inventory = nil
}
