// Copyright © 2022 Christian R. Vozar ⚜
// Licensed under the MIT License. All rights reserved.

package criprof

const (
	schedulerKubernetes   = "kubernetes"
	schedulerNomad        = "nomad"
	scehdulerMesos        = "mesos"
	schedulerUndetermined = "undetermined"
)

// getScheduler returns the identified scheduler, if detected.
func getScheduler() string {
	if _, ok := EnvironmentVariables["KUBERNETES_SERVICE_HOST"]; ok {
		return schedulerKubernetes
	}

	return schedulerUndetermined
}
