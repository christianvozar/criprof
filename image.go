// Copyright © 2022-2023 Christian R. Vozar ⚜
// Licensed under the MIT License. All rights reserved.

package criprof

import (
	"fmt"
	"os"
)

// Detectable image formats
const (
	formatDocker       = "docker"       // Docker image format
	formatACI          = "aci"          // App Container Image format
	formatCRI          = "cri"          // Container Runtime Interface format
	formatOCF          = "ocf"          // Open Container Format
	formatUndetermined = "undetermined" // Undetermined image format
)

// getImageFormat returns the format of the container image currently running.
func getImageFormat() (string, error) {
	// Check if Docker format
	if _, err := isDockerFormat(); err == nil {
		return formatDocker, nil
	}

	// Check if /run/.containerenv file exists hinting CRI image.
	if _, err := os.Stat("/run/.containerenv"); err == nil {
		return formatCRI, nil
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("failed to check /run/.containerenv file: %v", err)
	}

	// Check if AC_METADATA_URL environment variable is set hinting ACI image.
	if _, ok := EnvironmentVariables["AC_METADATA_URL"]; ok {
		return formatACI, nil
	}

	// Check if AC_APP_NAME environment variable is set hinting ACI image.
	if _, ok := EnvironmentVariables["AC_APP_NAME"]; ok {
		return formatACI, nil
	}

	// Undetermined format.
	return formatUndetermined, nil
}

func isDockerFormat() (bool, error) {
	_, err := os.Stat("/.dockerinit")
	if err == nil {
		return true, nil
	} else if !os.IsNotExist(err) {
		return false, fmt.Errorf("failed to check /.dockerinit file: %v", err)
	}

	_, err = os.Stat("/.dockerenv")
	if err == nil {
		return true, nil
	} else if !os.IsNotExist(err) {
		return false, fmt.Errorf("failed to check /.dockerenv file: %v", err)
	}

	return false, nil
}
