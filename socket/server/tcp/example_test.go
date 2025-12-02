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

package tcp_test

import (
	"context"
	"fmt"
	"io"
	"net"
	"time"

	libptc "github.com/nabbar/golib/network/protocol"
	libsck "github.com/nabbar/golib/socket"
	sckcfg "github.com/nabbar/golib/socket/config"
	scksrt "github.com/nabbar/golib/socket/server/tcp"
)

// Example demonstrates a basic echo server.
// This is the simplest possible TCP server implementation.
func Example() {
	// Create handler function that echoes back received data
	handler := func(c libsck.Context) {
		defer func() { _ = c.Close() }()
		buf := make([]byte, 1024)
		for {
			n, err := c.Read(buf)
			if err != nil {
				return
			}
			if n > 0 {
				_, _ = c.Write(buf[:n])
			}
		}
	}

	// Create server configuration
	cfg := sckcfg.Server{
		Network: libptc.NetworkTCP,
		Address: ":8080",
	}

	// Create server
	srv, err := scksrt.New(nil, handler, cfg)
	if err != nil {
		panic(err)
	}

	// Start server
	ctx := context.Background()
	go func() {
		_ = srv.Listen(ctx)
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Shutdown after demonstration
	_ = srv.Shutdown(ctx)
	// Output:
}

// Example_complete demonstrates a production-ready server with all features.
// This example shows error handling, monitoring, graceful shutdown, and logging.
func Example_complete() {
	// Handler with proper error handling
	handler := func(c libsck.Context) {
		defer func() { _ = c.Close() }()

		buf := make([]byte, 4096)
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

	// Create configuration with idle timeout
	cfg := sckcfg.Server{
		Network:        libptc.NetworkTCP,
		Address:        ":8081",
		ConIdleTimeout: 5 * time.Minute,
	}

	// Create server
	srv, err := scksrt.New(nil, handler, cfg)
	if err != nil {
		fmt.Printf("Failed to create server: %v\n", err)
		return
	}

	// Register monitoring callbacks
	srv.RegisterFuncError(func(errs ...error) {
		for _, e := range errs {
			fmt.Printf("Server error: %v\n", e)
		}
	})

	srv.RegisterFuncInfo(func(local, remote net.Addr, state libsck.ConnState) {
		fmt.Printf("Connection %s from %s\n", state, remote)
	})

	// Start server
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := srv.Listen(ctx); err != nil {
			fmt.Printf("Server stopped: %v\n", err)
		}
	}()

	// Wait for server to be ready
	time.Sleep(50 * time.Millisecond)
	fmt.Printf("Server running with %d connections\n", srv.OpenConnections())

	// Graceful shutdown
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

// ExampleNew demonstrates creating a TCP server
func ExampleNew() {
	// Define connection handler
	handler := func(c libsck.Context) {
		defer func() { _ = c.Close() }()
		_, _ = io.Copy(c, c) // Echo
	}

	// Create configuration
	cfg := sckcfg.Server{
		Network: libptc.NetworkUDP,
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

// ExampleServerTcp_Listen demonstrates starting a server
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

// ExampleServerTcp_Shutdown demonstrates graceful shutdown
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

	// Graceful shutdown
	err := srv.Shutdown(ctx)
	if err == nil {
		fmt.Println("Server shut down successfully")
	}
	// Output: Server shut down successfully
}

// ExampleServerTcp_RegisterFuncError demonstrates error callback registration
func ExampleServerTcp_RegisterFuncError() {
	handler := func(c libsck.Context) {
		defer func() { _ = c.Close() }()
	}

	cfg := sckcfg.Server{
		Network: libptc.NetworkTCP,
		Address: ":9003",
	}
	srv, _ := scksrt.New(nil, handler, cfg)

	// Register error callback
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

// ExampleServerTcp_RegisterFuncInfo demonstrates connection info callback
func ExampleServerTcp_RegisterFuncInfo() {
	handler := func(c libsck.Context) {
		defer func() { _ = c.Close() }()
	}

	cfg := sckcfg.Server{
		Network: libptc.NetworkTCP,
		Address: ":9004",
	}
	srv, _ := scksrt.New(nil, handler, cfg)

	// Register info callback
	srv.RegisterFuncInfo(func(local, remote net.Addr, state libsck.ConnState) {
		fmt.Printf("Connection event: %s on %s->(%s)%s\n", state, remote, local.Network(), local.String())
	})

	fmt.Println("Info callback registered")
	_ = srv.Shutdown(context.Background())
	// Output: Info callback registered
}

// ExampleServerTcp_OpenConnections demonstrates connection tracking
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

	count := srv.OpenConnections()
	fmt.Printf("Active connections: %d\n", count)

	_ = srv.Shutdown(ctx)
	// Output: Active connections: 0
}

// ExampleServerTcp_SetTLS demonstrates TLS configuration
func ExampleServerTcp_SetTLS() {
	handler := func(c libsck.Context) {
		defer func() { _ = c.Close() }()
	}

	cfg := sckcfg.Server{
		Network: libptc.NetworkTCP,
		Address: ":9443",
	}
	srv, _ := scksrt.New(nil, handler, cfg)

	// Disable TLS (or configure with certificates)
	err := srv.SetTLS(false, nil)
	if err == nil {
		fmt.Println("TLS configuration updated")
	}

	_ = srv.Shutdown(context.Background())
	// Output: TLS configuration updated
}

// ExampleServerTcp_IsRunning demonstrates checking server status
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

// ExampleServerTcp_idleTimeout demonstrates idle connection timeout
func ExampleServerTcp_idleTimeout() {
	handler := func(c libsck.Context) {
		defer func() { _ = c.Close() }()
		// Handler that doesn't read/write (connection will idle)
		time.Sleep(200 * time.Millisecond)
	}

	cfg := sckcfg.Server{
		Network:        libptc.NetworkTCP,
		Address:        ":9007",
		ConIdleTimeout: 100 * time.Millisecond,
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

// ExampleNew_withUpdateConn demonstrates custom connection configuration
func ExampleNew_withUpdateConn() {
	// UpdateConn callback to configure TCP keepalive
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

// ExampleServerTcp_monitoring demonstrates complete monitoring setup
func ExampleServerTcp_monitoring() {
	handler := func(c libsck.Context) {
		defer func() { _ = c.Close() }()
	}

	cfg := sckcfg.Server{
		Network: libptc.NetworkTCP,
		Address: ":9009",
	}
	srv, _ := scksrt.New(nil, handler, cfg)

	// Register all callbacks
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

// ExampleNew_simpleProtocol demonstrates a simple line-based protocol
func ExampleNew_simpleProtocol() {
	// Handler for a simple line-based protocol (newline-delimited)
	handler := func(c libsck.Context) {
		defer func() { _ = c.Close() }()

		buf := make([]byte, 1024)
		for {
			n, err := c.Read(buf)
			if err != nil {
				return
			}

			// Echo each line back
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

// ExampleServerTcp_contextValues demonstrates using context values
func ExampleServerTcp_contextValues() {
	type contextKey string
	const userIDKey contextKey = "userID"

	handler := func(c libsck.Context) {
		defer func() { _ = c.Close() }()

		// Access context value
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
