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

package unix_test

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"time"

	libdur "github.com/nabbar/golib/duration"
	libprm "github.com/nabbar/golib/file/perm"
	libptc "github.com/nabbar/golib/network/protocol"
	libsck "github.com/nabbar/golib/socket"
	sckcfg "github.com/nabbar/golib/socket/config"
	scksru "github.com/nabbar/golib/socket/server/unix"
)

// Example demonstrates a basic echo server.
// This is the simplest possible Unix socket server implementation.
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

	// Create temporary socket file
	tmpDir := os.TempDir()
	socketPath := filepath.Join(tmpDir, fmt.Sprintf("example-%d.sock", time.Now().UnixNano()))
	defer os.Remove(socketPath)

	// Create server configuration
	cfg := sckcfg.Server{
		Network:   libptc.NetworkUnix,
		Address:   socketPath,
		PermFile:  libprm.Perm(0600),
		GroupPerm: -1,
	}

	// Create server
	srv, err := scksru.New(nil, handler, cfg)
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

	// Create temporary socket file
	tmpDir := os.TempDir()
	socketPath := filepath.Join(tmpDir, fmt.Sprintf("complete-%d.sock", time.Now().UnixNano()))
	defer os.Remove(socketPath)

	// Create configuration with idle timeout
	cfg := sckcfg.Server{
		Network:        libptc.NetworkUnix,
		Address:        socketPath,
		PermFile:       libprm.Perm(0660),
		GroupPerm:      -1,
		ConIdleTimeout: libdur.Minutes(5),
	}

	// Create server
	srv, err := scksru.New(nil, handler, cfg)
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

// ExampleNew demonstrates creating a Unix socket server
func ExampleNew() {
	// Define connection handler
	handler := func(c libsck.Context) {
		defer func() { _ = c.Close() }()
		_, _ = io.Copy(c, c) // Echo
	}

	// Create temporary socket file
	tmpDir := os.TempDir()
	socketPath := filepath.Join(tmpDir, fmt.Sprintf("new-%d.sock", time.Now().UnixNano()))
	defer os.Remove(socketPath)

	// Create configuration
	cfg := sckcfg.Server{
		Network:   libptc.NetworkUnix,
		Address:   socketPath,
		PermFile:  libprm.Perm(0600),
		GroupPerm: -1,
	}

	// Create server
	srv, err := scksru.New(nil, handler, cfg)
	if err != nil {
		fmt.Printf("Failed to create server: %v\n", err)
		return
	}

	fmt.Printf("Server created successfully\n")
	_ = srv
	// Output: Server created successfully
}

// ExampleServerUnix_Listen demonstrates starting a server
func ExampleServerUnix_Listen() {
	handler := func(c libsck.Context) {
		defer func() { _ = c.Close() }()
	}

	tmpDir := os.TempDir()
	socketPath := filepath.Join(tmpDir, fmt.Sprintf("listen-%d.sock", time.Now().UnixNano()))
	defer os.Remove(socketPath)

	cfg := sckcfg.Server{
		Network:   libptc.NetworkUnix,
		Address:   socketPath,
		PermFile:  libprm.Perm(0600),
		GroupPerm: -1,
	}
	srv, _ := scksru.New(nil, handler, cfg)

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

// ExampleServerUnix_Shutdown demonstrates graceful shutdown
func ExampleServerUnix_Shutdown() {
	handler := func(c libsck.Context) {
		defer func() { _ = c.Close() }()
	}

	tmpDir := os.TempDir()
	socketPath := filepath.Join(tmpDir, fmt.Sprintf("shutdown-%d.sock", time.Now().UnixNano()))
	defer os.Remove(socketPath)

	cfg := sckcfg.Server{
		Network:   libptc.NetworkUnix,
		Address:   socketPath,
		PermFile:  libprm.Perm(0600),
		GroupPerm: -1,
	}
	srv, _ := scksru.New(nil, handler, cfg)

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

// ExampleServerUnix_RegisterFuncError demonstrates error callback registration
func ExampleServerUnix_RegisterFuncError() {
	handler := func(c libsck.Context) {
		defer func() { _ = c.Close() }()
	}

	tmpDir := os.TempDir()
	socketPath := filepath.Join(tmpDir, fmt.Sprintf("error-%d.sock", time.Now().UnixNano()))
	defer os.Remove(socketPath)

	cfg := sckcfg.Server{
		Network:   libptc.NetworkUnix,
		Address:   socketPath,
		PermFile:  libprm.Perm(0600),
		GroupPerm: -1,
	}
	srv, _ := scksru.New(nil, handler, cfg)

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

// ExampleServerUnix_RegisterFuncInfo demonstrates connection info callback
func ExampleServerUnix_RegisterFuncInfo() {
	handler := func(c libsck.Context) {
		defer func() { _ = c.Close() }()
	}

	tmpDir := os.TempDir()
	socketPath := filepath.Join(tmpDir, fmt.Sprintf("info-%d.sock", time.Now().UnixNano()))
	defer os.Remove(socketPath)

	cfg := sckcfg.Server{
		Network:   libptc.NetworkUnix,
		Address:   socketPath,
		PermFile:  libprm.Perm(0600),
		GroupPerm: -1,
	}
	srv, _ := scksru.New(nil, handler, cfg)

	// Register info callback
	srv.RegisterFuncInfo(func(local, remote net.Addr, state libsck.ConnState) {
		fmt.Printf("Connection event: %s\n", state)
	})

	fmt.Println("Info callback registered")
	_ = srv.Shutdown(context.Background())
	// Output: Info callback registered
}

// ExampleServerUnix_OpenConnections demonstrates connection tracking
func ExampleServerUnix_OpenConnections() {
	handler := func(c libsck.Context) {
		defer func() { _ = c.Close() }()
		time.Sleep(100 * time.Millisecond)
	}

	tmpDir := os.TempDir()
	socketPath := filepath.Join(tmpDir, fmt.Sprintf("connections-%d.sock", time.Now().UnixNano()))
	defer os.Remove(socketPath)

	cfg := sckcfg.Server{
		Network:   libptc.NetworkUnix,
		Address:   socketPath,
		PermFile:  libprm.Perm(0600),
		GroupPerm: -1,
	}
	srv, _ := scksru.New(nil, handler, cfg)

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

// ExampleServerUnix_SetTLS demonstrates TLS configuration (no-op for Unix sockets)
func ExampleServerUnix_SetTLS() {
	handler := func(c libsck.Context) {
		defer func() { _ = c.Close() }()
	}

	tmpDir := os.TempDir()
	socketPath := filepath.Join(tmpDir, fmt.Sprintf("tls-%d.sock", time.Now().UnixNano()))
	defer os.Remove(socketPath)

	cfg := sckcfg.Server{
		Network:   libptc.NetworkUnix,
		Address:   socketPath,
		PermFile:  libprm.Perm(0600),
		GroupPerm: -1,
	}
	srv, _ := scksru.New(nil, handler, cfg)

	// SetTLS is a no-op for Unix sockets (always returns nil)
	err := srv.SetTLS(false, nil)
	if err == nil {
		fmt.Println("TLS configuration updated (no-op for Unix sockets)")
	}

	_ = srv.Shutdown(context.Background())
	// Output: TLS configuration updated (no-op for Unix sockets)
}

// ExampleServerUnix_IsRunning demonstrates checking server status
func ExampleServerUnix_IsRunning() {
	handler := func(c libsck.Context) {
		defer func() { _ = c.Close() }()
	}

	tmpDir := os.TempDir()
	socketPath := filepath.Join(tmpDir, fmt.Sprintf("running-%d.sock", time.Now().UnixNano()))
	defer os.Remove(socketPath)

	cfg := sckcfg.Server{
		Network:   libptc.NetworkUnix,
		Address:   socketPath,
		PermFile:  libprm.Perm(0600),
		GroupPerm: -1,
	}
	srv, _ := scksru.New(nil, handler, cfg)

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

// ExampleServerUnix_idleTimeout demonstrates idle connection timeout
func ExampleServerUnix_idleTimeout() {
	handler := func(c libsck.Context) {
		defer func() { _ = c.Close() }()
		// Handler that doesn't read/write (connection will idle)
		time.Sleep(200 * time.Millisecond)
	}

	tmpDir := os.TempDir()
	socketPath := filepath.Join(tmpDir, fmt.Sprintf("idle-%d.sock", time.Now().UnixNano()))
	defer os.Remove(socketPath)

	cfg := sckcfg.Server{
		Network:        libptc.NetworkUnix,
		Address:        socketPath,
		PermFile:       libprm.Perm(0600),
		GroupPerm:      -1,
		ConIdleTimeout: libdur.ParseDuration(100 * time.Millisecond),
	}
	srv, _ := scksru.New(nil, handler, cfg)

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
	// UpdateConn callback to configure socket options
	upd := func(c net.Conn) {
		if unixConn, ok := c.(*net.UnixConn); ok {
			_ = unixConn.SetReadBuffer(8192)
			_ = unixConn.SetWriteBuffer(8192)
		}
	}

	handler := func(c libsck.Context) {
		defer func() { _ = c.Close() }()
	}

	tmpDir := os.TempDir()
	socketPath := filepath.Join(tmpDir, fmt.Sprintf("update-%d.sock", time.Now().UnixNano()))
	defer os.Remove(socketPath)

	cfg := sckcfg.Server{
		Network:   libptc.NetworkUnix,
		Address:   socketPath,
		PermFile:  libprm.Perm(0600),
		GroupPerm: -1,
	}

	srv, err := scksru.New(upd, handler, cfg)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Server with custom connection config created")
	_ = srv.Shutdown(context.Background())
	// Output: Server with custom connection config created
}

// ExampleServerUnix_monitoring demonstrates complete monitoring setup
func ExampleServerUnix_monitoring() {
	handler := func(c libsck.Context) {
		defer func() { _ = c.Close() }()
	}

	tmpDir := os.TempDir()
	socketPath := filepath.Join(tmpDir, fmt.Sprintf("monitor-%d.sock", time.Now().UnixNano()))
	defer os.Remove(socketPath)

	cfg := sckcfg.Server{
		Network:   libptc.NetworkUnix,
		Address:   socketPath,
		PermFile:  libprm.Perm(0600),
		GroupPerm: -1,
	}
	srv, _ := scksru.New(nil, handler, cfg)

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

// ExampleNew_permissions demonstrates file permission configuration
func ExampleNew_permissions() {
	handler := func(c libsck.Context) {
		defer func() { _ = c.Close() }()
	}

	tmpDir := os.TempDir()
	socketPath := filepath.Join(tmpDir, fmt.Sprintf("perms-%d.sock", time.Now().UnixNano()))
	defer os.Remove(socketPath)

	// Configure different permission modes
	cfg := sckcfg.Server{
		Network:   libptc.NetworkUnix,
		Address:   socketPath,
		PermFile:  libprm.Perm(0600), // Owner only
		GroupPerm: -1,                // Default group
	}

	srv, err := scksru.New(nil, handler, cfg)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Server with custom permissions created")
	_ = srv.Shutdown(context.Background())
	// Output: Server with custom permissions created
}

// ExampleServerUnix_contextValues demonstrates using context values
func ExampleServerUnix_contextValues() {
	type contextKey string
	const requestIDKey contextKey = "requestID"

	handler := func(c libsck.Context) {
		defer func() { _ = c.Close() }()

		// Access context value
		if requestID := c.Value(requestIDKey); requestID != nil {
			fmt.Printf("Processing request ID: %v\n", requestID)
		}
	}

	tmpDir := os.TempDir()
	socketPath := filepath.Join(tmpDir, fmt.Sprintf("context-%d.sock", time.Now().UnixNano()))
	defer os.Remove(socketPath)

	cfg := sckcfg.Server{
		Network:   libptc.NetworkUnix,
		Address:   socketPath,
		PermFile:  libprm.Perm(0600),
		GroupPerm: -1,
	}

	srv, _ := scksru.New(nil, handler, cfg)
	fmt.Println("Server with context values ready")
	_ = srv.Shutdown(context.Background())
	// Output: Server with context values ready
}
