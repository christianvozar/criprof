// Copyright © 2022-2023 Christian R. Vozar
// Licensed under the MIT License. All rights reserved.

package criprof

import (
	"context"
	"net"
	"net/http"
	"os"
	"time"
)

// FileSystem abstracts filesystem operations for testability
type FileSystem interface {
	// Stat returns file information
	Stat(name string) (os.FileInfo, error)

	// ReadFile reads the entire file
	ReadFile(name string) ([]byte, error)
}

// Network abstracts network operations for testability
type Network interface {
	// DialTimeout connects to an address with a timeout
	DialTimeout(network, address string, timeout time.Duration) (net.Conn, error)

	// HTTPGet performs an HTTP GET request with context
	HTTPGet(ctx context.Context, url string) (*http.Response, error)
}

// DefaultFileSystem implements FileSystem using the os package
type DefaultFileSystem struct{}

// Stat returns file information using os.Stat
func (DefaultFileSystem) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

// ReadFile reads a file using os.ReadFile
func (DefaultFileSystem) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}

// DefaultNetwork implements Network using net and http packages
type DefaultNetwork struct{}

// DialTimeout connects to an address with timeout using net.DialTimeout
func (DefaultNetwork) DialTimeout(network, address string, timeout time.Duration) (net.Conn, error) {
	return net.DialTimeout(network, address, timeout)
}

// HTTPGet performs an HTTP GET with context
func (DefaultNetwork) HTTPGet(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	return http.DefaultClient.Do(req)
}
