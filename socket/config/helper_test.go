/*
 * MIT License
 *
 * Copyright (c) 2025 Nicolas JUHEL
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 *
 */

// Test Helpers - Socket Configuration Package
//
// This file contains shared test utilities and helper functions used across
// the socket/config test suite.
//
// Contents:
//   - Global test context: Shared context for test lifecycle management
//   - Platform detection: isWindows(), skipIfWindows() for platform-specific tests
//   - Address generators: Functions to create valid/invalid test addresses
//   - Protocol lists: Helper functions returning protocol collections
//   - Configuration factories: createValidClientConfig(), createValidServerConfig()
//   - Assertion helpers: expectValidationError(), expectNoValidationError()
//   - Temporary paths: tmpSocketPath() for Unix socket test paths
//
// Helper Philosophy:
// These utilities reduce code duplication and improve test readability by
// abstracting common patterns. They provide consistent test data and
// simplify the setup/teardown of test fixtures.
//
// Usage:
// All test files in the package can import and use these helpers. The global
// context is initialized once in BeforeSuite and cleaned up in AfterSuite.
package config_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	libprm "github.com/nabbar/golib/file/perm"
	libptc "github.com/nabbar/golib/network/protocol"
	"github.com/nabbar/golib/socket/config"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	// ctx is the global test context used across all test suites
	ctx    context.Context
	cancel context.CancelFunc
)

// initTestContext initializes the global test context
func initTestContext() {
	ctx, cancel = context.WithCancel(context.Background())
}

// cleanupTestContext cancels the global test context
func cleanupTestContext() {
	if cancel != nil {
		cancel()
	}
}

// isWindows returns true if running on Windows
func isWindows() bool {
	return runtime.GOOS == "windows"
}

// skipIfWindows skips the current test if running on Windows
func skipIfWindows(msg string) {
	if isWindows() {
		Skip(fmt.Sprintf("Skipping on Windows: %s", msg))
	}
}

// tmpSocketPath generates a temporary Unix socket path
func tmpSocketPath(prefix string) string {
	return filepath.Join(os.TempDir(), fmt.Sprintf("%s_%d.sock", prefix, os.Getpid()))
}

// cleanupSocket removes a Unix socket file if it exists
func cleanupSocket(path string) {
	_ = os.Remove(path)
}

// createTempDir creates a temporary directory for testing
func createTempDir(pattern string) (string, error) {
	return os.MkdirTemp("", pattern)
}

// removeTempDir removes a temporary directory and all its contents
func removeTempDir(dir string) error {
	return os.RemoveAll(dir)
}

// validTCPAddresses returns a list of valid TCP addresses for testing
func validTCPAddresses() []string {
	return []string{
		"localhost:8080",
		"127.0.0.1:8080",
		":8080",
		"0.0.0.0:8080",
		"[::1]:8080",
		"[::]:8080",
	}
}

// invalidTCPAddresses returns a list of invalid TCP addresses for testing
func invalidTCPAddresses() []string {
	return []string{
		"",
		"localhost",
		"localhost:",
		":99999", // Port out of range
		"invalid:port",
		"[::1]", // Missing port
	}
}

// validUDPAddresses returns a list of valid UDP addresses for testing
func validUDPAddresses() []string {
	return []string{
		"localhost:9000",
		"127.0.0.1:9000",
		":9000",
		"0.0.0.0:9000",
		"[::1]:9000",
		"[::]:9000",
	}
}

// invalidUDPAddresses returns a list of invalid UDP addresses for testing
func invalidUDPAddresses() []string {
	return []string{
		"",
		"localhost",
		"localhost:",
		":99999", // Port out of range
		"invalid:port",
	}
}

// validUnixAddresses returns a list of valid Unix socket addresses for testing
func validUnixAddresses() []string {
	tmpDir := os.TempDir()
	return []string{
		filepath.Join(tmpDir, "test.sock"),
		filepath.Join(tmpDir, "app", "server.sock"),
		"./test.sock",
		"/tmp/test.sock",
	}
}

// invalidUnixAddresses returns a list of invalid Unix socket addresses for testing
func invalidUnixAddresses() []string {
	return []string{
		"", // Empty address
	}
}

// allNetworkProtocols returns all supported network protocols
func allNetworkProtocols() []libptc.NetworkProtocol {
	return []libptc.NetworkProtocol{
		libptc.NetworkTCP,
		libptc.NetworkTCP4,
		libptc.NetworkTCP6,
		libptc.NetworkUDP,
		libptc.NetworkUDP4,
		libptc.NetworkUDP6,
		libptc.NetworkUnix,
		libptc.NetworkUnixGram,
	}
}

// tcpProtocols returns TCP-based network protocols
func tcpProtocols() []libptc.NetworkProtocol {
	return []libptc.NetworkProtocol{
		libptc.NetworkTCP,
		libptc.NetworkTCP4,
		libptc.NetworkTCP6,
	}
}

// udpProtocols returns UDP-based network protocols
func udpProtocols() []libptc.NetworkProtocol {
	return []libptc.NetworkProtocol{
		libptc.NetworkUDP,
		libptc.NetworkUDP4,
		libptc.NetworkUDP6,
	}
}

// unixProtocols returns Unix socket protocols
func unixProtocols() []libptc.NetworkProtocol {
	return []libptc.NetworkProtocol{
		libptc.NetworkUnix,
		libptc.NetworkUnixGram,
	}
}

// validGroupIDs returns valid Unix group IDs for testing
func validGroupIDs() []int32 {
	return []int32{
		-1,   // Use current group
		0,    // Root group
		1000, // Common user group
		config.MaxGID,
	}
}

// invalidGroupIDs returns invalid Unix group IDs for testing
func invalidGroupIDs() []int32 {
	return []int32{
		config.MaxGID + 1,
		99999,
		-2, // Invalid negative value
	}
}

// validFilePermissions returns valid file permissions for testing
func validFilePermissions() []libprm.Perm {
	return []libprm.Perm{
		0600, // Owner read/write
		0660, // Owner and group read/write
		0666, // All read/write
		0700, // Owner all permissions
		0770, // Owner and group all permissions
	}
}

// protocolSupportsAddress checks if a protocol supports a given address format
func protocolSupportsAddress(proto libptc.NetworkProtocol, addr string) bool {
	cfg := config.Client{
		Network: proto,
		Address: addr,
	}
	return cfg.Validate() == nil
}

// createValidClientConfig creates a valid client configuration for the given protocol
func createValidClientConfig(proto libptc.NetworkProtocol) config.Client {
	var addr string
	switch proto {
	case libptc.NetworkTCP, libptc.NetworkTCP4, libptc.NetworkTCP6:
		addr = "localhost:8080"
	case libptc.NetworkUDP, libptc.NetworkUDP4, libptc.NetworkUDP6:
		addr = "localhost:9000"
	case libptc.NetworkUnix, libptc.NetworkUnixGram:
		addr = tmpSocketPath("client")
	default:
		addr = "localhost:8080"
	}

	return config.Client{
		Network: proto,
		Address: addr,
	}
}

// createValidServerConfig creates a valid server configuration for the given protocol
func createValidServerConfig(proto libptc.NetworkProtocol) config.Server {
	var addr string
	switch proto {
	case libptc.NetworkTCP, libptc.NetworkTCP4, libptc.NetworkTCP6:
		addr = ":8080"
	case libptc.NetworkUDP, libptc.NetworkUDP4, libptc.NetworkUDP6:
		addr = ":9000"
	case libptc.NetworkUnix, libptc.NetworkUnixGram:
		addr = tmpSocketPath("server")
	default:
		addr = ":8080"
	}

	return config.Server{
		Network: proto,
		Address: addr,
	}
}

// expectValidationError is a helper to check for specific validation errors
func expectValidationError(err error, expectedErr error) {
	ExpectWithOffset(1, err).To(HaveOccurred())
	ExpectWithOffset(1, err).To(MatchError(expectedErr))
}

// expectNoValidationError is a helper to check that validation succeeds
func expectNoValidationError(err error) {
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
}
