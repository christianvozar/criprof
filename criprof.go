// Copyright © 2022 Christian R. Vozar ⚜
// Licensed under the MIT License. All rights reserved.

package criprof

import (
	"encoding/json"
	"fmt"
	"os"
)

// EnvironmentVariables is used to cache all environment variables read at
// execution.
var EnvironmentVariables map[string]string

func init() {
	EnvironmentVariables = environMap()
}

// Inventory holds an application's container and runtime information.
type Inventory struct {
	Hostname    string `json:"hostname"`
	ID          string `json:"id"`
	ImageFormat string `json:"image_format"`
	PID         int    `json:"pid"`
	Runtime     string `json:"runtime"`
	Scheduler   string `json:"scheduler"`
}

// New returns a new Inventory with populated values.
func New() *Inventory {
	return &Inventory{
		Hostname:    getHostname(),
		ID:          getContainerID(),
		ImageFormat: getImageFormat(),
		PID:         os.Getpid(),
		Runtime:     getRuntime(),
		Scheduler:   getScheduler(),
	}
}

// JSON returns the Inventory as JSON string.
func (i Inventory) JSON() string {
	j, err := json.Marshal(i)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	return string(j)
}
