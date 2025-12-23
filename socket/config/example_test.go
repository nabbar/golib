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

package config_test

import (
	"fmt"
	"os"
	"runtime"

	libdur "github.com/nabbar/golib/duration"
	"github.com/nabbar/golib/socket/config"

	libptc "github.com/nabbar/golib/network/protocol"
)

// Example_basicTCPClient demonstrates the simplest TCP client configuration.
//
// This example shows minimal configuration for connecting to a TCP server.
func Example_basicTCPClient() {
	// Create a basic TCP client configuration
	cfg := config.Client{
		Network: libptc.NetworkTCP,
		Address: "localhost:8080",
	}

	// Validate the configuration
	if err := cfg.Validate(); err != nil {
		fmt.Printf("Invalid configuration: %v\n", err)
		return
	}

	fmt.Println("TCP client configuration validated successfully")
	// Output: TCP client configuration validated successfully
}

// Example_basicTCPServer demonstrates the simplest TCP server configuration.
//
// This example shows minimal configuration for a TCP server listening on all interfaces.
func Example_basicTCPServer() {
	// Create a basic TCP server configuration
	cfg := config.Server{
		Network: libptc.NetworkTCP,
		Address: ":8080",
	}

	// Validate the configuration
	if err := cfg.Validate(); err != nil {
		fmt.Printf("Invalid configuration: %v\n", err)
		return
	}

	fmt.Println("TCP server configuration validated successfully")
	// Output: TCP server configuration validated successfully
}

// Example_basicUDPClient demonstrates UDP client configuration.
//
// This example shows how to configure a UDP client for connectionless communication.
func Example_basicUDPClient() {
	// Create a UDP client configuration
	cfg := config.Client{
		Network: libptc.NetworkUDP,
		Address: "localhost:9000",
	}

	// Validate the configuration
	if err := cfg.Validate(); err != nil {
		fmt.Printf("Invalid configuration: %v\n", err)
		return
	}

	fmt.Println("UDP client configuration validated successfully")
	// Output: UDP client configuration validated successfully
}

// Example_basicUDPServer demonstrates UDP server configuration.
//
// This example shows how to configure a UDP server for datagram handling.
func Example_basicUDPServer() {
	// Create a UDP server configuration
	cfg := config.Server{
		Network: libptc.NetworkUDP,
		Address: ":9000",
	}

	// Validate the configuration
	if err := cfg.Validate(); err != nil {
		fmt.Printf("Invalid configuration: %v\n", err)
		return
	}

	fmt.Println("UDP server configuration validated successfully")
	// Output: UDP server configuration validated successfully
}

// Example_unixSocketClient demonstrates Unix domain socket client configuration.
//
// This example shows how to configure a client for Unix socket IPC.
// Note: Unix sockets are not available on Windows.
func Example_unixSocketClient() {
	if runtime.GOOS == "windows" {
		fmt.Println("Unix sockets not available on Windows")
		return
	}

	// Create a Unix socket client configuration
	cfg := config.Client{
		Network: libptc.NetworkUnix,
		Address: "/tmp/app.sock",
	}

	// Validate the configuration
	if err := cfg.Validate(); err != nil {
		fmt.Printf("Invalid configuration: %v\n", err)
		return
	}

	fmt.Println("Unix socket client configuration validated successfully")
	// Output: Unix socket client configuration validated successfully
}

// Example_unixSocketServer demonstrates Unix domain socket server configuration.
//
// This example shows how to configure a Unix socket server with file permissions.
// Note: Unix sockets are not available on Windows.
func Example_unixSocketServer() {
	if runtime.GOOS == "windows" {
		fmt.Println("Unix sockets not available on Windows")
		return
	}

	// Create a Unix socket server configuration with permissions
	cfg := config.Server{
		Network:   libptc.NetworkUnix,
		Address:   "/tmp/app.sock",
		PermFile:  0660, // Owner and group can read/write
		GroupPerm: -1,   // Use current process group
	}

	// Validate the configuration
	if err := cfg.Validate(); err != nil {
		fmt.Printf("Invalid configuration: %v\n", err)
		return
	}

	fmt.Println("Unix socket server configuration validated successfully")
	// Output: Unix socket server configuration validated successfully
}

// Example_serverWithIdleTimeout demonstrates server configuration with connection timeout.
//
// This example shows how to configure automatic disconnection of idle connections.
func Example_serverWithIdleTimeout() {
	// Create a server with idle timeout
	cfg := config.Server{
		Network:        libptc.NetworkTCP,
		Address:        ":8080",
		ConIdleTimeout: libdur.Minutes(5), // Close idle connections after 5 minutes
	}

	// Validate the configuration
	if err := cfg.Validate(); err != nil {
		fmt.Printf("Invalid configuration: %v\n", err)
		return
	}

	fmt.Println("Server with idle timeout configured successfully")
	// Output: Server with idle timeout configured successfully
}

// Example_invalidProtocol demonstrates error handling for invalid protocols.
//
// This example shows how to detect and handle protocol validation errors.
func Example_invalidProtocol() {
	// Create a client with zero-value protocol (invalid)
	cfg := config.Client{
		Network: libptc.NetworkProtocol(0),
		Address: "localhost:8080",
	}

	// Validate the configuration
	if err := cfg.Validate(); err != nil {
		fmt.Println("Invalid protocol error detected")
		return
	}

	fmt.Println("This should not be reached")
	// Output: Invalid protocol error detected
}

// Example_invalidAddress demonstrates error handling for invalid addresses.
//
// This example shows how validation catches malformed addresses.
func Example_invalidAddress() {
	// Create a client with invalid address format
	cfg := config.Client{
		Network: libptc.NetworkTCP,
		Address: "invalid-address-without-port",
	}

	// Validate the configuration
	if err := cfg.Validate(); err != nil {
		fmt.Println("Address validation failed as expected")
		return
	}

	fmt.Println("This should not be reached")
	// Output: Address validation failed as expected
}

// Example_groupPermissionValidation demonstrates Unix socket group permission validation.
//
// This example shows how to configure and validate group ownership.
func Example_groupPermissionValidation() {
	if runtime.GOOS == "windows" {
		fmt.Println("Unix sockets not available on Windows")
		return
	}

	// Valid group permission
	validCfg := config.Server{
		Network:   libptc.NetworkUnix,
		Address:   "/tmp/app.sock",
		PermFile:  0660,
		GroupPerm: 1000, // Valid group ID
	}

	if err := validCfg.Validate(); err != nil {
		fmt.Printf("Unexpected error: %v\n", err)
		return
	}

	// Invalid group permission (exceeds MaxGID)
	invalidCfg := config.Server{
		Network:   libptc.NetworkUnix,
		Address:   "/tmp/app.sock",
		PermFile:  0660,
		GroupPerm: 99999, // Exceeds MaxGID
	}

	if err := invalidCfg.Validate(); err != nil {
		fmt.Println("Group permission validation failed as expected")
		return
	}

	fmt.Println("Valid group permission configured")
	// Output: Group permission validation failed as expected
}

// Example_configurationFromEnvironment demonstrates loading configuration from environment.
//
// This example shows a common pattern for 12-factor apps.
func Example_configurationFromEnvironment() {
	// Simulate reading from environment
	network := os.Getenv("SOCKET_NETWORK")
	if network == "" {
		network = "tcp" // Default
	}

	address := os.Getenv("SOCKET_ADDRESS")
	if address == "" {
		address = ":8080" // Default
	}

	// Parse network protocol
	var proto libptc.NetworkProtocol
	switch network {
	case "tcp":
		proto = libptc.NetworkTCP
	case "udp":
		proto = libptc.NetworkUDP
	case "unix":
		proto = libptc.NetworkUnix
	default:
		fmt.Printf("Unknown network type: %s\n", network)
		return
	}

	// Create configuration
	cfg := config.Server{
		Network: proto,
		Address: address,
	}

	// Validate before use
	if err := cfg.Validate(); err != nil {
		fmt.Printf("Invalid configuration: %v\n", err)
		return
	}

	fmt.Printf("Server configured for %s on %s\n", network, address)
	// Output: Server configured for tcp on :8080
}

// Example_multipleServers demonstrates configuring multiple server types.
//
// This example shows how to set up a service that listens on both network and Unix sockets.
func Example_multipleServers() {
	// Network server for remote access
	tcpCfg := config.Server{
		Network:        libptc.NetworkTCP,
		Address:        ":8080",
		ConIdleTimeout: libdur.Minutes(10),
	}

	if err := tcpCfg.Validate(); err != nil {
		fmt.Printf("TCP configuration error: %v\n", err)
		return
	}

	// Unix socket for local IPC (if not on Windows)
	var unixCfg *config.Server
	if runtime.GOOS != "windows" {
		unixCfg = &config.Server{
			Network:  libptc.NetworkUnix,
			Address:  "/tmp/app.sock",
			PermFile: 0660,
		}

		if err := unixCfg.Validate(); err != nil {
			fmt.Printf("Unix socket configuration error: %v\n", err)
			return
		}
	}

	fmt.Println("Multiple server configurations validated")
	if unixCfg != nil {
		fmt.Println("Both TCP and Unix socket servers ready")
	} else {
		fmt.Println("TCP server ready (Unix socket not available on this platform)")
	}
	// Output: Multiple server configurations validated
	// Both TCP and Unix socket servers ready
}

// Example_ipv6Configuration demonstrates configuring IPv6 sockets.
//
// This example shows how to configure servers for IPv6 networks.
func Example_ipv6Configuration() {
	// IPv6-specific server
	cfg := config.Server{
		Network: libptc.NetworkTCP6,
		Address: "[::1]:8080", // IPv6 loopback
	}

	if err := cfg.Validate(); err != nil {
		fmt.Printf("IPv6 configuration error: %v\n", err)
		return
	}

	fmt.Println("IPv6 server configuration validated")
	// Output: IPv6 server configuration validated
}

// Example_dualStackServer demonstrates configuring a dual-stack (IPv4 + IPv6) server.
//
// This example shows how to listen on both IPv4 and IPv6.
func Example_dualStackServer() {
	// Dual-stack server (listens on both IPv4 and IPv6)
	cfg := config.Server{
		Network: libptc.NetworkTCP,
		Address: ":8080", // Binds to both [::]:8080 and 0.0.0.0:8080
	}

	if err := cfg.Validate(); err != nil {
		fmt.Printf("Dual-stack configuration error: %v\n", err)
		return
	}

	fmt.Println("Dual-stack server configuration validated")
	// Output: Dual-stack server configuration validated
}

// Example_validationBeforeUse demonstrates best practices for configuration validation.
//
// This example shows the recommended pattern of validate-then-use.
func Example_validationBeforeUse() {
	configs := []config.Server{
		{Network: libptc.NetworkTCP, Address: ":8080"},
		{Network: libptc.NetworkTCP, Address: ":8081"},
		{Network: libptc.NetworkUDP, Address: ":9000"},
	}

	// Validate all configurations before starting any servers
	for i, cfg := range configs {
		if err := cfg.Validate(); err != nil {
			fmt.Printf("Configuration %d invalid: %v\n", i, err)
			return
		}
	}

	fmt.Printf("All %d configurations validated successfully\n", len(configs))
	// Output: All 3 configurations validated successfully
}
