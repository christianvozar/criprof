// Copyright © 2022 Christian R. Vozar ⚜
// Licensed under the MIT License. All rights reserved.

package criprof

import (
	"io/ioutil"
	"os"
	"strings"
)

const (
	formatDocker       = "docker"
	formatACI          = "aci"
	formatCRI          = "cri"
	formatOCF          = "ocf"
	formatUndetermined = "undetermined"
)

func getImageFormat() string {
	if _, err := os.Stat("/.dockerinit"); err == nil {
		return formatDocker
	}

	if _, err := os.Stat("/.dockerenv"); err == nil {
		return formatDocker
	}

	cgroup, _ := ioutil.ReadFile("/proc/self/cgroup")
	if strings.Contains(string(cgroup), "docker") {
		return formatDocker
	}

	if _, ok := EnvironmentVariables["AC_METADATA_URL"]; ok {
		return formatACI
	}

	if _, ok := EnvironmentVariables["AC_APP_NAME"]; ok {
		return formatACI
	}

	return formatUndetermined
}
