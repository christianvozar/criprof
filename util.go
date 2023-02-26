// Copyright © 2022-2023 Christian R. Vozar
// Licensed under the MIT License. All rights reserved.

package criprof

import (
	"os"
	"strings"
)

// environMap returns the results of os.Environ as a map.
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
