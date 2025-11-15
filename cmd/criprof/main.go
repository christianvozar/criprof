// Copyright © 2022-2023 Christian R. Vozar
// Licensed under the MIT License. All rights reserved.

// criprof is a command-line tool for Container Runtime Interface (CRI) profiling
// and introspection.
//
// This CLI provides access to criprof functionality for detecting container runtimes,
// orchestrators, and image formats from the command line. It's useful for shell scripts,
// non-Go programs, and manual debugging of container environments.
//
// # Commands
//
//   criprof hints   - Display container runtime information as JSON
//   criprof version - Print version information
//
// # Usage
//
//	criprof hints
//
// For more information, visit: https://github.com/christianvozar/criprof
package main

import "github.com/christianvozar/criprof/cmd/criprof/cmd"

func main() {
	cmd.Execute()
}
