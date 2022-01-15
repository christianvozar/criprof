// Copyright © 2022 Christian R. Vozar ⚜
// Licensed under the MIT License. All rights reserved.

package criprof

import (
	"io/ioutil"
	"os"
	"regexp"
)

// IsContainer returns true if the application is running within a container
// runtime/engine.
func IsContainer() bool {
	if _, err := os.Stat("/.dockerinit"); err == nil {
		return true
	}

	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}

	if c := getContainerID(); c != "" {
		return true
	}

	return false
}

func getContainerID() string {
	dockerIDMatch := regexp.MustCompile(`cpu\:\/docker\/([0-9a-z]+)`)
	coreOSIDMatch := regexp.MustCompile(`cpuset\:\/system.slice\/docker-([0-9a-z]+)`)

	if _, err := os.Stat("/proc/self/cgroup"); os.IsExist(err) {
		cgroup, _ := ioutil.ReadFile("/proc/self/cgroup")
		strCgroup := string(cgroup)
		loc := dockerIDMatch.FindStringIndex(strCgroup)

		if loc != nil {
			return strCgroup[loc[0]+12 : loc[1]-2]
		}

		// cgroup not nil, not vanilla Docker. Check for CoreOS.
		loc = coreOSIDMatch.FindStringIndex(strCgroup)

		if loc != nil {
			return strCgroup[loc[0]+27:]
		}
	}

	return "undetermined"
}

func getHostname() string {
	if _, ok := EnvironmentVariables["HOSTNAME"]; ok {
		return EnvironmentVariables["HOSTNAME"]
	}

	return "undetermined"
}
