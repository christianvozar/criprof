// Copyright © 2022-2023 Christian R. Vozar
// Licensed under the MIT License. All rights reserved.

package criprof

import (
	"os"
	"strings"
)

// environMap converts the os.Environ() slice to a map for efficient lookups.
// This is an unexported utility function used during package initialization.
//
// The function takes the KEY=VALUE pairs from os.Environ() and splits them
// into a map where keys are environment variable names and values are their
// corresponding values.
//
// Returns a map of environment variable names to values.
func environMap() map[string]string {
	env := os.Environ()

	// Create empty map to store environment variables.
	vars := make(map[string]string)

	for _, pair := range env {
		// Split each string into a key and a value.
		e := strings.Split(pair, "=")
		vars[e[0]] = e[1]
	}
	return vars
}
