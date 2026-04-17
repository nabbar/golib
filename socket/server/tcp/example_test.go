/*
 * MIT License
 *
 * Copyright (c) 2022 Nicolas JUHEL
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

// Package tcp_test provides runnable examples for the TCP server.
//
// # Examples Overview
//
// These examples demonstrate how to integrate the TCP server into your application. 
// They cover everything from a minimal echo server to a full-featured production setup 
// with TLS, monitoring, and graceful shutdown.
//
// # Design Principles for Production
//
// When using this server in a production environment, we recommend:
//   1. Context Propagation: Always pass a context to Listen() to control the server lifecycle.
//   2. Monitoring: Register callbacks for Error, Info, and InfoServer to gain visibility.
//   3. Timeouts: Configure ConIdleTimeout to prevent resource exhaustion from leaking clients.
//   4. TLS: Enable TLS with a secure configuration from the certificates package.
//   5. Handlers: Use the sCtx structure (passed as libsck.Context) to read and write safely.
package tcp_test

import (
	"context"
	"fmt"
	"io"
	"net"
	"time"

	libdur "github.com/nabbar/golib/duration"
	libptc "github.com/nabbar/golib/network/protocol"
	libsck "github.com/nabbar/golib/socket"
	sckcfg "github.com/nabbar/golib/socket/config"
	scksrt "github.com/nabbar/golib/socket/server/tcp"
)

// Example demonstrates a minimal TCP echo server.
// This is the simplest possible implementation for development and testing.
func Example() {
	// 1. Define the handler function
	handler := func(c libsck.Context) {
		// Ensure connection is closed after the handler exits.
		defer func() { _ = c.Close() }()

		buf := make([]byte, 1024)
		for {
			// Read from the context-aware wrapper.
			n, err := c.Read(buf)
			if err != nil {
				return
			}
			// Write received data back (Echo).
			if n > 0 {
				_, _ = c.Write(buf[:n])
			}
		}
	}

	// 2. Create server configuration
	cfg := sckcfg.Server{
		Network: libptc.NetworkTCP,
		Address: ":8080",
	}

	// 3. Instantiate the server
	srv, err := scksrt.New(nil, handler, cfg)
	if err != nil {
		panic(err)
	}

	// 4. Start listening in a background goroutine
	ctx := context.Background()
	go func() {
		_ = srv.Listen(ctx)
	}()

	// Wait briefly for the listener to bind.
	time.Sleep(100 * time.Millisecond)

	// 5. Gracefully shutdown the server.
	_ = srv.Shutdown(ctx)
	// Output:
}

// Example_complete demonstrates a robust, production-ready TCP server.
// This example includes error handling, connection monitoring, and graceful shutdown.
func Example_complete() {
	// Advanced handler with activity check
	handler := func(c libsck.Context) {
		defer func() { _ = c.Close() }()

		buf := make([]byte, 4096)
		// Use IsConnected() to check the socket state in a loop.
		for c.IsConnected() {
			n, err := c.Read(buf)
			if err != nil {
				return
			}

			if n > 0 {
				if _, err := c.Write(buf[:n]); err != nil {
					return
				}
			}
		}
	}

	// Configure with a 5-minute idle timeout (centralized management).
	cfg := sckcfg.Server{
		Network:        libptc.NetworkTCP,
		Address:        ":8081",
		ConIdleTimeout: libdur.Minutes(5),
	}

	// Create server
	srv, err := scksrt.New(nil, handler, cfg)
	if err != nil {
		fmt.Printf("Failed to create server: %v\n", err)
		return
	}

	// Register monitoring callbacks for observability
	srv.RegisterFuncError(func(errs ...error) {
		for _, e := range errs {
			fmt.Printf("Server error: %v\n", e)
		}
	})

	srv.RegisterFuncInfo(func(local, remote net.Addr, state libsck.ConnState) {
		fmt.Printf("Connection %s from %s\n", state, remote)
	})

	// Start server with context support
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := srv.Listen(ctx); err != nil {
			fmt.Printf("Server stopped: %v\n", err)
		}
	}()

	// Active polling for server readiness
	for i := 0; i < 25; i++ {
		if srv.IsRunning() {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if !srv.IsRunning() {
		fmt.Println("Server not started")
	}

	fmt.Printf("Server running with %d connections\n", srv.OpenConnections())

	// Shutdown with a specific timeout for connection draining
	shutdownCtx, shutdownCancel := context.WithTimeout(
		context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		fmt.Printf("Shutdown error: %v\n", err)
	}

	fmt.Println("Server stopped gracefully")
	// Output:
	// Server running with 0 connections
	// Server stopped gracefully
}

// ExampleNew shows how to initialize a server and handle initial validation errors.
func ExampleNew() {
	handler := func(c libsck.Context) {
		defer func() { _ = c.Close() }()
		_, _ = io.Copy(c, c)
	}

	// Note: Providing a non-TCP protocol will fail validation for scksrt.New
	cfg := sckcfg.Server{
		Network: libptc.NetworkTCP,
		Address: ":9000",
	}

	// Create server
	srv, err := scksrt.New(nil, handler, cfg)
	if err != nil {
		fmt.Printf("Failed to create server: %v\n", err)
		return
	}

	fmt.Printf("Server created successfully\n")
	_ = srv
	// Output: Server created successfully
}

// ExampleServerTcp_Listen shows how to start the listener loop in the background.
func ExampleServerTcp_Listen() {
	handler := func(c libsck.Context) {
		defer func() { _ = c.Close() }()
	}

	cfg := sckcfg.Server{
		Network: libptc.NetworkTCP,
		Address: ":9001",
	}
	srv, _ := scksrt.New(nil, handler, cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Start server in background
	go func() {
		_ = srv.Listen(ctx)
	}()

	// Wait for server to start
	time.Sleep(50 * time.Millisecond)

	if srv.IsRunning() {
		fmt.Println("Server is running")
	}

	// Cleanup
	_ = srv.Shutdown(context.Background())
	// Output: Server is running
}

// ExampleServerTcp_Shutdown shows how to perform a graceful drain of active connections.
func ExampleServerTcp_Shutdown() {
	handler := func(c libsck.Context) {
		defer func() { _ = c.Close() }()
	}

	cfg := sckcfg.Server{
		Network: libptc.NetworkTCP,
		Address: ":9002",
	}
	srv, _ := scksrt.New(nil, handler, cfg)

	ctx := context.Background()
	go func() {
		_ = srv.Listen(ctx)
	}()

	time.Sleep(50 * time.Millisecond)

	// Shutdown blocks until all connections are finished or the context expires.
	err := srv.Shutdown(ctx)
	if err == nil {
		fmt.Println("Server shut down successfully")
	}
	// Output: Server shut down successfully
}

// ExampleServerTcp_RegisterFuncError shows how to capture server-level errors asynchronously.
func ExampleServerTcp_RegisterFuncError() {
	handler := func(c libsck.Context) {
		defer func() { _ = c.Close() }()
	}

	cfg := sckcfg.Server{
		Network: libptc.NetworkTCP,
		Address: ":9003",
	}
	srv, _ := scksrt.New(nil, handler, cfg)

	// Error callback is triggered for port binding issues, handshake errors, etc.
	srv.RegisterFuncError(func(errs ...error) {
		for _, err := range errs {
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		}
	})

	fmt.Println("Error callback registered")
	_ = srv.Shutdown(context.Background())
	// Output: Error callback registered
}

// ExampleServerTcp_RegisterFuncInfo shows how to monitor connection New/Close events.
func ExampleServerTcp_RegisterFuncInfo() {
	handler := func(c libsck.Context) {
		defer func() { _ = c.Close() }()
	}

	cfg := sckcfg.Server{
		Network: libptc.NetworkTCP,
		Address: ":9004",
	}
	srv, _ := scksrt.New(nil, handler, cfg)

	// Info callback provides observability for connection status.
	srv.RegisterFuncInfo(func(local, remote net.Addr, state libsck.ConnState) {
		fmt.Printf("Connection event: %s on %s->(%s)%s\n", state, remote, local.Network(), local.String())
	})

	fmt.Println("Info callback registered")
	_ = srv.Shutdown(context.Background())
	// Output: Info callback registered
}

// ExampleServerTcp_OpenConnections shows how to query the current connection count.
func ExampleServerTcp_OpenConnections() {
	handler := func(c libsck.Context) {
		defer func() { _ = c.Close() }()
		time.Sleep(100 * time.Millisecond)
	}

	cfg := sckcfg.Server{
		Network: libptc.NetworkTCP,
		Address: ":9005",
	}
	srv, _ := scksrt.New(nil, handler, cfg)

	ctx := context.Background()
	go func() {
		_ = srv.Listen(ctx)
	}()

	time.Sleep(50 * time.Millisecond)

	// Returns the current value of the atomic connection counter.
	count := srv.OpenConnections()
	fmt.Printf("Active connections: %d\n", count)

	_ = srv.Shutdown(ctx)
	// Output: Active connections: 0
}

// ExampleServerTcp_SetTLS shows how to dynamically enable or update TLS settings.
func ExampleServerTcp_SetTLS() {
	handler := func(c libsck.Context) {
		defer func() { _ = c.Close() }()
	}

	cfg := sckcfg.Server{
		Network: libptc.NetworkTCP,
		Address: ":9443",
	}
	srv, _ := scksrt.New(nil, handler, cfg)

	// SetTLS can be used to update certificates or disable encryption before Listen.
	err := srv.SetTLS(false, nil)
	if err == nil {
		fmt.Println("TLS configuration updated")
	}

	_ = srv.Shutdown(context.Background())
	// Output: TLS configuration updated
}

// ExampleServerTcp_IsRunning shows how to check if the server is actively listening.
func ExampleServerTcp_IsRunning() {
	handler := func(c libsck.Context) {
		defer func() { _ = c.Close() }()
	}

	cfg := sckcfg.Server{
		Network: libptc.NetworkTCP,
		Address: ":9006",
	}
	srv, _ := scksrt.New(nil, handler, cfg)

	if !srv.IsRunning() {
		fmt.Println("Server is not running")
	}

	ctx := context.Background()
	go func() {
		_ = srv.Listen(ctx)
	}()

	time.Sleep(50 * time.Millisecond)

	if srv.IsRunning() {
		fmt.Println("Server is now running")
	}

	_ = srv.Shutdown(ctx)
	// Output:
	// Server is not running
	// Server is now running
}

// ExampleServerTcp_idleTimeout shows how inactivity thresholds are enforced.
func ExampleServerTcp_idleTimeout() {
	handler := func(c libsck.Context) {
		defer func() { _ = c.Close() }()
		// Connection remains idle. The idlemgr will terminate it.
		time.Sleep(200 * time.Millisecond)
	}

	cfg := sckcfg.Server{
		Network:        libptc.NetworkTCP,
		Address:        ":9007",
		ConIdleTimeout: libdur.ParseDuration(100 * time.Millisecond),
	}
	srv, _ := scksrt.New(nil, handler, cfg)

	ctx := context.Background()
	go func() {
		_ = srv.Listen(ctx)
	}()

	time.Sleep(50 * time.Millisecond)
	fmt.Println("Server with idle timeout running")

	_ = srv.Shutdown(ctx)
	// Output: Server with idle timeout running
}

// ExampleNew_withUpdateConn shows how to tune low-level socket options.
func ExampleNew_withUpdateConn() {
	// Use UpdateConn for advanced OS-level tuning.
	upd := func(c net.Conn) {
		if tcpConn, ok := c.(*net.TCPConn); ok {
			_ = tcpConn.SetKeepAlive(true)
			_ = tcpConn.SetKeepAlivePeriod(30 * time.Second)
		}
	}

	handler := func(c libsck.Context) {
		defer func() { _ = c.Close() }()
	}

	cfg := sckcfg.Server{
		Network: libptc.NetworkTCP,
		Address: ":9008",
	}

	srv, err := scksrt.New(upd, handler, cfg)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Server with custom connection config created")
	_ = srv.Shutdown(context.Background())
	// Output: Server with custom connection config created
}

// ExampleServerTcp_monitoring shows a full setup of observability callbacks.
func ExampleServerTcp_monitoring() {
	handler := func(c libsck.Context) {
		defer func() { _ = c.Close() }()
	}

	cfg := sckcfg.Server{
		Network: libptc.NetworkTCP,
		Address: ":9009",
	}
	srv, _ := scksrt.New(nil, handler, cfg)

	// Register all available notification hooks.
	srv.RegisterFuncError(func(errs ...error) {
		fmt.Println("Error callback registered")
	})

	srv.RegisterFuncInfo(func(local, remote net.Addr, state libsck.ConnState) {
		fmt.Println("Connection callback registered")
	})

	srv.RegisterFuncInfoServer(func(msg string) {
		fmt.Println("Server info callback registered")
	})

	fmt.Println("All monitoring callbacks configured")
	_ = srv.Shutdown(context.Background())
	// Output: All monitoring callbacks configured
}

// ExampleNew_simpleProtocol shows a line-based TCP server implementation.
func ExampleNew_simpleProtocol() {
	handler := func(c libsck.Context) {
		defer func() { _ = c.Close() }()

		buf := make([]byte, 1024)
		for {
			n, err := c.Read(buf)
			if err != nil {
				return
			}

			// Process and response logic (Echo for simplicity)
			if n > 0 {
				_, _ = c.Write(buf[:n])
			}
		}
	}

	cfg := sckcfg.Server{
		Network: libptc.NetworkTCP,
		Address: ":9010",
	}

	srv, err := scksrt.New(nil, handler, cfg)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Line-based protocol server created")
	_ = srv.Shutdown(context.Background())
	// Output: Line-based protocol server created
}

// ExampleServerTcp_contextValues shows how to use the context for request-scoped data.
func ExampleServerTcp_contextValues() {
	type contextKey string
	const userIDKey contextKey = "userID"

	handler := func(c libsck.Context) {
		defer func() { _ = c.Close() }()

		// The context wrapper (sCtx) delegates Value() calls to the parent context.
		if userID := c.Value(userIDKey); userID != nil {
			fmt.Printf("Processing request for user: %v\n", userID)
		}
	}

	cfg := sckcfg.Server{
		Network: libptc.NetworkTCP,
		Address: ":9011",
	}

	srv, _ := scksrt.New(nil, handler, cfg)
	fmt.Println("Server with context values ready")
	_ = srv.Shutdown(context.Background())
	// Output: Server with context values ready
}
