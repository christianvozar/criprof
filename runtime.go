// Copyright © 2022 Christian R. Vozar ⚜
// Licensed under the MIT License. All rights reserved.

package criprof

import (
	"io/ioutil"
	"os"
	"strings"
)

const (
	runtimeDocker       = "docker"
	runtimeRkt          = "rkt"
	runtimeRunC         = "runc"
	runtimeContainerD   = "containerd"
	runtimeLXC          = "lxc"
	runtimeLXD          = "lxd"
	runtimeOpenVZ       = "openvz"
	runtimeUndetermined = "undetermined"
)

func getRuntime() string {
	if _, err := os.Stat("/.dockerinit"); err == nil {
		return runtimeDocker
	}

	if _, err := os.Stat("/.dockerenv"); err == nil {
		return runtimeDocker
	}

	cgroup, _ := ioutil.ReadFile("/proc/self/cgroup")
	if strings.Contains(string(cgroup), "docker") {
		return runtimeDocker
	}

	if _, ok := EnvironmentVariables["AC_METADATA_URL"]; ok {
		return runtimeRkt
	}

	if _, ok := EnvironmentVariables["AC_APP_NAME"]; ok {
		return runtimeRkt
	}

	if _, err := os.Stat("/dev/lxd/sock"); err == nil {
		return runtimeLXD
	}

	return runtimeUndetermined
}
