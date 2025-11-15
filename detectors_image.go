// Copyright © 2022-2023 Christian R. Vozar
// Licensed under the MIT License. All rights reserved.

package criprof

import (
	"context"
)

// DockerImageDetector detects Docker image format via filesystem markers
type DockerImageDetector struct {
	fs FileSystem
}

func (d *DockerImageDetector) Name() string {
	return "docker-image-format"
}

func (d *DockerImageDetector) Priority() int {
	return 95
}

func (d *DockerImageDetector) Detect(ctx context.Context) (*Detection, error) {
	// Check /.dockerenv
	if _, err := d.fs.Stat("/.dockerenv"); err == nil {
		return &Detection{
			Type:       DetectionTypeImageFormat,
			Value:      formatDocker,
			Confidence: 0.95,
			Source:     d.Name(),
		}, nil
	}

	// Check /.dockerinit (legacy)
	if _, err := d.fs.Stat("/.dockerinit"); err == nil {
		return &Detection{
			Type:       DetectionTypeImageFormat,
			Value:      formatDocker,
			Confidence: 0.90,
			Source:     d.Name(),
		}, nil
	}

	return nil, nil
}

// CRIImageDetector detects CRI image format via filesystem markers
type CRIImageDetector struct {
	fs FileSystem
}

func (d *CRIImageDetector) Name() string {
	return "cri-image-format"
}

func (d *CRIImageDetector) Priority() int {
	return 95
}

func (d *CRIImageDetector) Detect(ctx context.Context) (*Detection, error) {
	if _, err := d.fs.Stat("/run/.containerenv"); err == nil {
		return &Detection{
			Type:       DetectionTypeImageFormat,
			Value:      formatCRI,
			Confidence: 0.90,
			Source:     d.Name(),
		}, nil
	}

	return nil, nil
}

// ACIEnvDetector detects ACI (App Container Image) format via environment variables
type ACIEnvDetector struct{}

func (d *ACIEnvDetector) Name() string {
	return "aci-env"
}

func (d *ACIEnvDetector) Priority() int {
	return 85
}

func (d *ACIEnvDetector) Detect(ctx context.Context) (*Detection, error) {
	if _, ok := EnvironmentVariables["AC_METADATA_URL"]; ok {
		return &Detection{
			Type:       DetectionTypeImageFormat,
			Value:      formatACI,
			Confidence: 0.95,
			Source:     d.Name(),
		}, nil
	}

	if _, ok := EnvironmentVariables["AC_APP_NAME"]; ok {
		return &Detection{
			Type:       DetectionTypeImageFormat,
			Value:      formatACI,
			Confidence: 0.95,
			Source:     d.Name(),
		}, nil
	}

	return nil, nil
}
