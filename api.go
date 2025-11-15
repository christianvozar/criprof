// Copyright © 2022-2023 Christian R. Vozar
// Licensed under the MIT License. All rights reserved.

package criprof

import (
	"context"
	"sync"
	"time"
)

var (
	defaultEngine *Engine
	engineOnce    sync.Once
)

// NewWithContext creates an Inventory using the default engine with context support
//
// This allows for timeout and cancellation control:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//	inventory := criprof.NewWithContext(ctx)
func NewWithContext(ctx context.Context) *Inventory {
	engineOnce.Do(initDefaultEngine)

	inv, err := defaultEngine.DetectAll(ctx)
	if err != nil {
		// Return fallback inventory on error
		return &Inventory{
			PID:         0,
			Runtime:     "undetermined",
			Scheduler:   "undetermined",
			ImageFormat: "undetermined",
			ID:          "undetermined",
			Hostname:    "unknown",
		}
	}

	return inv
}

// NewWithEngine creates an Inventory using a custom engine
//
// This allows advanced users to configure custom detectors:
//
//	engine := criprof.NewEngine(criprof.EngineConfig{
//	    Detectors: criprof.DefaultDetectors(),
//	    EnableCaching: true,
//	    CacheTTL: 10 * time.Minute,
//	})
//	inventory, err := criprof.NewWithEngine(context.Background(), engine)
func NewWithEngine(ctx context.Context, engine *Engine) (*Inventory, error) {
	return engine.DetectAll(ctx)
}

// DefaultDetectors returns the default set of detectors
func DefaultDetectors() []Detector {
	fs := DefaultFileSystem{}
	net := DefaultNetwork{}
	timeout := 2 * time.Second

	return []Detector{
		// Runtime detectors (priority 100-80)
		&DockerFileDetector{fs: fs},
		&DockerCgroupDetector{fs: fs},
		&PodmanDetector{fs: fs},
		&CRIODetector{fs: fs},
		&ContainerdFileDetector{fs: fs},
		&RktEnvDetector{},
		&LXDSocketDetector{fs: fs},
		&OpenVZDetector{fs: fs},
		&FirecrackerDetector{fs: fs},
		&KataContainersDetector{fs: fs},
		&GVisorDetector{fs: fs},
		&SysboxDetector{fs: fs},
		&SingularityDetector{},
		&WASMDetector{},

		// Scheduler detectors (priority 95-80)
		&KubernetesServiceAccountDetector{fs: fs},
		&KubernetesEnvDetector{},
		&NomadEnvDetector{},
		&NomadHostnameDetector{},
		&MesosEnvDetector{},
		&MesosCgroupDetector{fs: fs},
		&ECSDetector{},
		&FargateDetector{},
		&CloudRunDetector{},
		&LambdaContainerDetector{},
		&ACIDetector{},

		// Network detectors (priority 20-10)
		&SwarmPortDetector{network: net, timeout: timeout},
		&KubernetesAPIDetector{network: net, timeout: timeout},

		// Image format detectors (priority 95-85)
		&DockerImageDetector{fs: fs},
		&CRIImageDetector{fs: fs},
		&ACIEnvDetector{},
		&OCIImageDetector{fs: fs},
		&SingularityImageDetector{},
	}
}

// FastDetectors returns only fast (non-network) detectors
//
// Use this when you want quick detection without network I/O:
//
//	engine := criprof.NewEngine(criprof.EngineConfig{
//	    Detectors: criprof.FastDetectors(),
//	})
func FastDetectors() []Detector {
	fs := DefaultFileSystem{}

	return []Detector{
		// Runtime detectors
		&DockerFileDetector{fs: fs},
		&DockerCgroupDetector{fs: fs},
		&PodmanDetector{fs: fs},
		&CRIODetector{fs: fs},
		&ContainerdFileDetector{fs: fs},
		&RktEnvDetector{},
		&LXDSocketDetector{fs: fs},
		&OpenVZDetector{fs: fs},
		&FirecrackerDetector{fs: fs},
		&KataContainersDetector{fs: fs},
		&GVisorDetector{fs: fs},
		&SysboxDetector{fs: fs},
		&SingularityDetector{},
		&WASMDetector{},

		// Scheduler detectors (no network)
		&KubernetesServiceAccountDetector{fs: fs},
		&KubernetesEnvDetector{},
		&NomadEnvDetector{},
		&NomadHostnameDetector{},
		&MesosEnvDetector{},
		&MesosCgroupDetector{fs: fs},
		&ECSDetector{},
		&FargateDetector{},
		&CloudRunDetector{},
		&LambdaContainerDetector{},
		&ACIDetector{},

		// Image format detectors
		&DockerImageDetector{fs: fs},
		&CRIImageDetector{fs: fs},
		&ACIEnvDetector{},
		&OCIImageDetector{fs: fs},
		&SingularityImageDetector{},
	}
}

// initDefaultEngine initializes the default engine singleton
func initDefaultEngine() {
	defaultEngine = NewEngine(EngineConfig{
		Detectors:     DefaultDetectors(),
		EnableCaching: true,
		CacheTTL:      5 * time.Minute,
	})
}

// InvalidateCache clears the default engine's cache
//
// Call this if the environment may have changed:
//
//	criprof.InvalidateCache()
//	inventory := criprof.New() // Will re-detect
func InvalidateCache() {
	if defaultEngine != nil {
		defaultEngine.InvalidateCache()
	}
}
