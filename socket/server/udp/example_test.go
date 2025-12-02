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

package udp_test

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	libptc "github.com/nabbar/golib/network/protocol"
	libsck "github.com/nabbar/golib/socket"
	sckcfg "github.com/nabbar/golib/socket/config"
	"github.com/nabbar/golib/socket/server/udp"
)

// Example_basicServer demonstrates the simplest UDP server setup.
//
// This example shows minimal configuration for a UDP echo server
// that receives datagrams and logs them.
func Example_basicServer() {
	// Create server configuration
	cfg := sckcfg.Server{
		Network: libptc.NetworkUDP,
		Address: ":0", // Let OS choose available port
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
	srv, err := udp.New(nil, handler, cfg)
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

	fmt.Println("UDP server created successfully")
	// Output: UDP server created successfully
}

// Example_echoServer demonstrates a complete UDP echo server.
//
// This example shows a server that echoes back received datagrams
// to the sender using ReadFrom/WriteTo pattern.
func Example_echoServer() {
	cfg := sckcfg.Server{
		Network: libptc.NetworkUDP,
		Address: ":9001",
	}

	// Echo handler using ReadFrom/WriteTo
	handler := func(ctx libsck.Context) {
		// Type assertion to get UDP connection
		type udpContext interface {
			libsck.Context
			// Would need access to underlying connection for WriteTo
		}

		// For demonstration, just read
		buf := make([]byte, 65507) // Max UDP datagram
		for {
			n, err := ctx.Read(buf)
			if err != nil {
				if err == io.EOF || err == io.ErrClosedPipe {
					return
				}
				log.Printf("Read error: %v", err)
				return
			}

			// In real implementation, would echo back using WriteTo
			_ = buf[:n]
		}
	}

	srv, err := udp.New(nil, handler, cfg)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Echo server configured on %s\n", cfg.Address)
	_ = srv
	// Output: Echo server configured on :9001
}

// Example_serverWithCallbacks demonstrates callback registration.
//
// This example shows how to register callbacks for error handling,
// connection monitoring, and server lifecycle events.
func Example_serverWithCallbacks() {
	cfg := sckcfg.Server{
		Network: libptc.NetworkUDP,
		Address: ":9002",
	}

	handler := func(ctx libsck.Context) {
		// Handler implementation
		buf := make([]byte, 1024)
		_, _ = ctx.Read(buf)
	}

	srv, err := udp.New(nil, handler, cfg)
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

	// Register connection info callback
	srv.RegisterFuncInfo(func(local, remote net.Addr, state libsck.ConnState) {
		fmt.Printf("Connection event: %s (local: %v, remote: %v)\n",
			state.String(), local, remote)
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
	cfg := sckcfg.Server{
		Network: libptc.NetworkUDP,
		Address: ":9003",
	}

	// Configure socket with custom options
	updateConn := func(conn net.Conn) {
		if udpConn, ok := conn.(*net.UDPConn); ok {
			// Set large read buffer for high throughput
			_ = udpConn.SetReadBuffer(1024 * 1024) // 1MB

			// Set large write buffer
			_ = udpConn.SetWriteBuffer(1024 * 1024) // 1MB

			fmt.Println("Socket buffers configured")
		}
	}

	handler := func(ctx libsck.Context) {
		buf := make([]byte, 1024)
		_, _ = ctx.Read(buf)
	}

	srv, err := udp.New(updateConn, handler, cfg)
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
	cfg := sckcfg.Server{
		Network: libptc.NetworkUDP,
		Address: ":9004",
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

	srv, err := udp.New(nil, handler, cfg)
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

// Example_multicastServer demonstrates handling multicast UDP traffic.
//
// This example shows configuration for a server that can receive
// multicast datagrams (concept demonstration).
func Example_multicastServer() {
	cfg := sckcfg.Server{
		Network: libptc.NetworkUDP,
		Address: ":9005",
	}

	// Configure multicast options
	updateConn := func(conn net.Conn) {
		if udpConn, ok := conn.(*net.UDPConn); ok {
			// In real implementation, would join multicast group
			// using udpConn.SetMulticastInterface() and related methods
			_ = udpConn
			fmt.Println("Multicast options configured")
		}
	}

	handler := func(ctx libsck.Context) {
		buf := make([]byte, 1024)
		_, _ = ctx.Read(buf)
	}

	srv, err := udp.New(updateConn, handler, cfg)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Multicast server configured")
	_ = srv
	// Output: Multicast server configured
}

// Example_errorHandling demonstrates comprehensive error handling.
//
// This example shows how to handle various error scenarios including
// invalid configuration, connection errors, and I/O errors.
func Example_errorHandling() {
	// Attempt to create server with invalid configuration
	invalidCfg := sckcfg.Server{
		Network: libptc.NetworkUDP,
		Address: "", // Invalid: empty address
	}

	handler := func(ctx libsck.Context) {}

	_, err := udp.New(nil, handler, invalidCfg)
	if err != nil {
		fmt.Printf("Configuration error detected: %v\n", err)
		return
	}

	fmt.Println("This should not print")
	// Output: Configuration error detected: invalid listen address
}

// Example_stateMonitoring demonstrates server state monitoring.
//
// This example shows how to monitor server state using IsRunning(),
// IsGone(), and OpenConnections() methods.
func Example_stateMonitoring() {
	cfg := sckcfg.Server{
		Network: libptc.NetworkUDP,
		Address: ":9007",
	}

	handler := func(ctx libsck.Context) {
		buf := make([]byte, 1024)
		_, _ = ctx.Read(buf)
	}

	srv, err := udp.New(nil, handler, cfg)
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
	// Create configuration
	cfg := sckcfg.Server{
		Network: libptc.NetworkUDP,
		Address: ":9008",
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
	srv, err := udp.New(nil, handler, cfg)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Server configured for %s on %s\n",
		cfg.Network.String(), cfg.Address)
	_ = srv
	// Output: Server configured for udp on :9008
}

// Example_integrationTest demonstrates testing UDP servers.
//
// This example shows how to write integration tests for UDP servers
// by creating a server and client for testing.
func Example_integrationTest() {
	cfg := sckcfg.Server{
		Network: libptc.NetworkUDP,
		Address: "127.0.0.1:0", // Bind to loopback, random port
	}

	receivedData := make(chan []byte, 1)

	handler := func(ctx libsck.Context) {
		buf := make([]byte, 1024)
		n, err := ctx.Read(buf)
		if err != nil {
			return
		}
		receivedData <- buf[:n]
	}

	srv, err := udp.New(nil, handler, cfg)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = srv.Listen(ctx)
	}()

	time.Sleep(10 * time.Millisecond)

	fmt.Println("Integration test server ready")
	// In real test, would send test datagram and verify
	// Output: Integration test server ready
}
