//go:build linux || darwin

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

package unixgram_test

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	libprm "github.com/nabbar/golib/file/perm"
	libptc "github.com/nabbar/golib/network/protocol"
	libsck "github.com/nabbar/golib/socket"
	sckcfg "github.com/nabbar/golib/socket/config"
	scksrv "github.com/nabbar/golib/socket/server/unixgram"
)

// Example_basicServer demonstrates the simplest Unix datagram server setup.
//
// This example shows minimal configuration for a datagram server
// that receives messages and logs them.
func Example_basicServer() {
	tmpDir := os.TempDir()
	sockPath := filepath.Join(tmpDir, "example_basic.sock")
	defer os.Remove(sockPath)

	// Create server configuration
	cfg := sckcfg.Server{
		Network:   libptc.NetworkUnixGram,
		Address:   sockPath,
		PermFile:  libprm.Perm(0600),
		GroupPerm: -1,
	}

	// Define simple handler
	handler := func(ctx libsck.Context) {
		defer ctx.Close()

		buf := make([]byte, 1024)
		n, err := ctx.Read(buf)
		if err != nil && err != io.EOF {
			return
		}

		fmt.Printf("Received: %s\n", buf[:n])
	}

	// Create server
	srv, err := scksrv.New(nil, handler, cfg)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Start server (would block in real usage)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	go func() {
		_ = srv.Listen(ctx)
	}()

	fmt.Println("Unix datagram server created successfully")
	// Output: Unix datagram server created successfully
}

// Example_datagramServer demonstrates a complete datagram logging server.
//
// This example shows a server that logs all received datagrams
// and demonstrates typical patterns for Unix datagram communication.
func Example_datagramServer() {
	tmpDir := os.TempDir()
	sockPath := filepath.Join(tmpDir, "example_log.sock")
	defer os.Remove(sockPath)

	cfg := sckcfg.Server{
		Network:   libptc.NetworkUnixGram,
		Address:   sockPath,
		PermFile:  libprm.Perm(0660), // Owner and group can write
		GroupPerm: -1,
	}

	// Logging handler
	handler := func(ctx libsck.Context) {
		defer ctx.Close()

		buf := make([]byte, 8192)
		for {
			n, err := ctx.Read(buf)
			if err != nil {
				if err == io.EOF || err == io.ErrClosedPipe {
					return
				}
				log.Printf("Read error: %v", err)
				return
			}

			// Log the datagram
			log.Printf("Datagram: %s", buf[:n])
		}
	}

	srv, err := scksrv.New(nil, handler, cfg)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Logging server configured at %s\n", sockPath)
	_ = srv
	// Output: Logging server configured at /tmp/example_log.sock
}

// Example_serverWithCallbacks demonstrates callback registration.
//
// This example shows how to register callbacks for error handling,
// datagram monitoring, and server lifecycle events.
func Example_serverWithCallbacks() {
	tmpDir := os.TempDir()
	sockPath := filepath.Join(tmpDir, "example_cb.sock")
	defer os.Remove(sockPath)

	cfg := sckcfg.Server{
		Network:   libptc.NetworkUnixGram,
		Address:   sockPath,
		PermFile:  libprm.Perm(0600),
		GroupPerm: -1,
	}

	handler := func(ctx libsck.Context) {
		buf := make([]byte, 1024)
		_, _ = ctx.Read(buf)
	}

	srv, err := scksrv.New(nil, handler, cfg)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Register error callback
	srv.RegisterFuncError(func(errs ...error) {
		for _, e := range errs {
			if e != nil {
				fmt.Printf("Server error: %v\n", e)
			}
		}
	})

	// Register datagram info callback
	srv.RegisterFuncInfo(func(local, remote net.Addr, state libsck.ConnState) {
		fmt.Printf("Datagram event: %s\n", state.String())
	})

	// Register server info callback
	srv.RegisterFuncInfoServer(func(msg string) {
		fmt.Printf("Server info: %s\n", msg)
	})

	fmt.Println("Callbacks registered successfully")
	// Output: Callbacks registered successfully
}

// Example_socketConfiguration demonstrates custom socket options.
//
// This example shows how to use UpdateConn callback to configure
// socket buffer sizes and other options.
func Example_socketConfiguration() {
	tmpDir := os.TempDir()
	sockPath := filepath.Join(tmpDir, "example_cfg.sock")
	defer os.Remove(sockPath)

	cfg := sckcfg.Server{
		Network:   libptc.NetworkUnixGram,
		Address:   sockPath,
		PermFile:  libprm.Perm(0600),
		GroupPerm: -1,
	}

	// Configure socket with custom options
	updateConn := func(conn net.Conn) {
		if unixConn, ok := conn.(*net.UnixConn); ok {
			// Set large read buffer for high throughput
			_ = unixConn.SetReadBuffer(1024 * 1024) // 1MB

			// Set large write buffer
			_ = unixConn.SetWriteBuffer(1024 * 1024) // 1MB

			fmt.Println("Socket buffers configured")
		}
	}

	handler := func(ctx libsck.Context) {
		buf := make([]byte, 1024)
		_, _ = ctx.Read(buf)
	}

	srv, err := scksrv.New(updateConn, handler, cfg)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Server created with custom socket configuration")
	_ = srv
	// Output: Server created with custom socket configuration
}

// Example_gracefulShutdown demonstrates proper server shutdown.
//
// This example shows how to handle graceful shutdown using context
// cancellation and the Shutdown() method.
func Example_gracefulShutdown() {
	tmpDir := os.TempDir()
	sockPath := filepath.Join(tmpDir, "example_shutdown.sock")
	defer os.Remove(sockPath)

	cfg := sckcfg.Server{
		Network:   libptc.NetworkUnixGram,
		Address:   sockPath,
		PermFile:  libprm.Perm(0600),
		GroupPerm: -1,
	}

	handler := func(ctx libsck.Context) {
		buf := make([]byte, 1024)
		for {
			_, err := ctx.Read(buf)
			if err != nil {
				return // Exit on error
			}
		}
	}

	srv, err := scksrv.New(nil, handler, cfg)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Start server with cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start server in goroutine
	go func() {
		if err := srv.Listen(ctx); err != nil {
			log.Printf("Listen error: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(10 * time.Millisecond)

	// Trigger shutdown
	cancel()

	// Wait with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		fmt.Printf("Shutdown error: %v\n", err)
		return
	}

	fmt.Println("Server shut down gracefully")
	// Output: Server shut down gracefully
}

// Example_permissionsAndGroup demonstrates file permissions and group ownership.
//
// This example shows how to configure socket file permissions
// and group ownership for multi-user scenarios.
func Example_permissionsAndGroup() {
	tmpDir := os.TempDir()
	sockPath := filepath.Join(tmpDir, "example_perm.sock")
	defer os.Remove(sockPath)

	// Configuration with specific permissions
	cfg := sckcfg.Server{
		Network:   libptc.NetworkUnixGram,
		Address:   sockPath,
		PermFile:  libprm.Perm(0660), // rw-rw---- (owner and group)
		GroupPerm: -1,                // Use process's default group
	}

	handler := func(ctx libsck.Context) {
		buf := make([]byte, 1024)
		_, _ = ctx.Read(buf)
	}

	srv, err := scksrv.New(nil, handler, cfg)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Server configured with permissions 0660\n")
	_ = srv
	// Output: Server configured with permissions 0660
}

// Example_errorHandling demonstrates comprehensive error handling.
//
// This example shows how to handle various error scenarios including
// invalid configuration, socket errors, and I/O errors.
func Example_errorHandling() {
	tmpDir := os.TempDir()
	sockPath := filepath.Join(tmpDir, "example_error.sock")
	defer os.Remove(sockPath)

	// Attempt to create server with nil handler
	cfg := sckcfg.Server{
		Network:   libptc.NetworkUnixGram,
		Address:   sockPath,
		PermFile:  libprm.Perm(0600),
		GroupPerm: -1,
	}

	_, err := scksrv.New(nil, nil, cfg)
	if err != nil {
		fmt.Printf("Configuration error detected: %v\n", err)
		return
	}

	fmt.Println("This should not print")
	// Output: Configuration error detected: invalid handler
}

// Example_stateMonitoring demonstrates server state monitoring.
//
// This example shows how to monitor server state using IsRunning(),
// IsGone(), and OpenConnections() methods.
func Example_stateMonitoring() {
	tmpDir := os.TempDir()
	sockPath := filepath.Join(tmpDir, "example_state.sock")
	defer os.Remove(sockPath)

	cfg := sckcfg.Server{
		Network:   libptc.NetworkUnixGram,
		Address:   sockPath,
		PermFile:  libprm.Perm(0600),
		GroupPerm: -1,
	}

	handler := func(ctx libsck.Context) {
		buf := make([]byte, 1024)
		_, _ = ctx.Read(buf)
	}

	srv, err := scksrv.New(nil, handler, cfg)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Check initial state
	fmt.Printf("Running: %v\n", srv.IsRunning())
	fmt.Printf("Gone: %v\n", srv.IsGone())
	fmt.Printf("Connections: %d\n", srv.OpenConnections())

	// Start server
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = srv.Listen(ctx)
	}()

	// Wait for server to start
	time.Sleep(10 * time.Millisecond)

	fmt.Printf("After start - Running: %v\n", srv.IsRunning())
	fmt.Printf("After start - Gone: %v\n", srv.IsGone())
	fmt.Printf("After start - Connections: %d\n", srv.OpenConnections())

	// Shutdown
	cancel()
	time.Sleep(100 * time.Millisecond)

	fmt.Printf("After shutdown - Gone: %v\n", srv.IsGone())

	// Output:
	// Running: false
	// Gone: true
	// Connections: 0
	// After start - Running: true
	// After start - Gone: false
	// After start - Connections: 0
	// After shutdown - Gone: true
}

// Example_configFromStruct demonstrates using config.Server struct.
//
// This example shows the recommended way to create a server using
// the config package for type-safe configuration.
func Example_configFromStruct() {
	tmpDir := os.TempDir()
	sockPath := filepath.Join(tmpDir, "example_struct.sock")
	defer os.Remove(sockPath)

	// Create configuration
	cfg := sckcfg.Server{
		Network:   libptc.NetworkUnixGram,
		Address:   sockPath,
		PermFile:  libprm.Perm(0600),
		GroupPerm: -1,
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		fmt.Printf("Invalid configuration: %v\n", err)
		return
	}

	// Define handler
	handler := func(ctx libsck.Context) {
		buf := make([]byte, 1024)
		_, _ = ctx.Read(buf)
	}

	// Create server
	srv, err := scksrv.New(nil, handler, cfg)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Server configured for %s at %s\n",
		cfg.Network.String(), sockPath)
	_ = srv
	// Output: Server configured for unixgram at /tmp/example_struct.sock
}

// Example_metricsCollector demonstrates a metrics collection server.
//
// This example shows how to build a metrics collection system
// that receives metrics from multiple local processes.
func Example_metricsCollector() {
	tmpDir := os.TempDir()
	sockPath := filepath.Join(tmpDir, "metrics.sock")
	defer os.Remove(sockPath)

	cfg := sckcfg.Server{
		Network:   libptc.NetworkUnixGram,
		Address:   sockPath,
		PermFile:  libprm.Perm(0666), // Allow all processes to send
		GroupPerm: -1,
	}

	// Metrics storage (in real implementation, use proper storage)
	metricsCount := 0

	handler := func(ctx libsck.Context) {
		defer ctx.Close()

		buf := make([]byte, 2048)
		for {
			n, err := ctx.Read(buf)
			if err != nil {
				return
			}

			// Parse and store metric (simplified)
			metricsCount++
			fmt.Printf("Metric received: %s\n", buf[:n])
		}
	}

	srv, err := scksrv.New(nil, handler, cfg)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Metrics collector configured")
	_ = srv
	// Output: Metrics collector configured
}

// Example_logAggregator demonstrates a log aggregation server.
//
// This example shows how to build a centralized logging system
// using Unix datagram sockets.
func Example_logAggregator() {
	tmpDir := os.TempDir()
	sockPath := filepath.Join(tmpDir, "logs.sock")
	defer os.Remove(sockPath)

	cfg := sckcfg.Server{
		Network:   libptc.NetworkUnixGram,
		Address:   sockPath,
		PermFile:  libprm.Perm(0660), // Owner and group
		GroupPerm: -1,
	}

	handler := func(ctx libsck.Context) {
		defer ctx.Close()

		buf := make([]byte, 8192)
		for {
			n, err := ctx.Read(buf)
			if err != nil {
				return
			}

			// Write to log file (simplified)
			logEntry := buf[:n]
			fmt.Printf("LOG: %s\n", logEntry)
		}
	}

	srv, err := scksrv.New(nil, handler, cfg)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Log aggregator configured")
	_ = srv
	// Output: Log aggregator configured
}
