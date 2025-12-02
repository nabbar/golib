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
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	libprm "github.com/nabbar/golib/file/perm"
	libptc "github.com/nabbar/golib/network/protocol"
	libsck "github.com/nabbar/golib/socket"
	sckclt "github.com/nabbar/golib/socket/client/unixgram"
	sckcfg "github.com/nabbar/golib/socket/config"
	scksrv "github.com/nabbar/golib/socket/server/unixgram"
)

// Example_basicClient demonstrates the simplest Unix datagram client setup.
//
// This example shows minimal configuration for a datagram client
// that sends a message to a server.
func Example_basicClient() {
	tmpDir := os.TempDir()
	sockPath := filepath.Join(tmpDir, "example_basic.sock")
	defer os.Remove(sockPath)

	// Create client
	client := sckclt.New(sockPath)
	if client == nil {
		fmt.Println("Failed to create client")
		return
	}

	fmt.Println("Unix datagram client created successfully")
	// Output: Unix datagram client created successfully
}

// Example_sendDatagram demonstrates sending a datagram to a server.
//
// This example shows how to connect and send data using a Unix datagram client.
// Note: Unix datagram sockets are fire-and-forget; no response is expected.
func Example_sendDatagram() {
	tmpDir := os.TempDir()
	sockPath := filepath.Join(tmpDir, "example_send.sock")
	defer os.Remove(sockPath)

	// Start a simple receiving server
	cfg := sckcfg.Server{
		Network:   libptc.NetworkUnixGram,
		Address:   sockPath,
		PermFile:  libprm.Perm(0600),
		GroupPerm: -1,
	}

	handler := func(ctx libsck.Context) {
		defer ctx.Close()
		buf := make([]byte, 1024)
		n, _ := ctx.Read(buf)
		// Server receives but doesn't respond (datagram is one-way)
		if n > 0 {
			log.Printf("Server received: %s", buf[:n])
		}
	}

	srv, _ := scksrv.New(nil, handler, cfg)
	srvCtx, srvCancel := context.WithCancel(context.Background())
	defer srvCancel()

	go func() {
		_ = srv.Listen(srvCtx)
	}()

	// Wait for server to start
	time.Sleep(50 * time.Millisecond)

	// Create and connect client
	client := sckclt.New(sockPath)
	if client == nil {
		fmt.Println("Failed to create client")
		return
	}
	defer client.Close()

	ctx := context.Background()
	err := client.Connect(ctx)
	if err != nil {
		fmt.Printf("Connect error: %v\n", err)
		return
	}

	// Send datagram (fire-and-forget, no response expected)
	data := []byte("Hello, Server!")
	n, err := client.Write(data)
	if err != nil {
		fmt.Printf("Write error: %v\n", err)
		return
	}

	fmt.Printf("Sent %d bytes\n", n)
	// Output: Sent 14 bytes
}

// Example_clientWithCallbacks demonstrates callback registration.
//
// This example shows how to register callbacks for error handling
// and operation monitoring.
func Example_clientWithCallbacks() {
	tmpDir := os.TempDir()
	sockPath := filepath.Join(tmpDir, "example_cb.sock")
	defer os.Remove(sockPath)

	// Create client
	client := sckclt.New(sockPath)
	if client == nil {
		fmt.Println("Failed to create client")
		return
	}

	// Register error callback
	client.RegisterFuncError(func(errs ...error) {
		for _, e := range errs {
			if e != nil {
				fmt.Printf("Error occurred: %v\n", e)
			}
		}
	})

	// Register info callback
	client.RegisterFuncInfo(func(local, remote net.Addr, state libsck.ConnState) {
		fmt.Printf("State change: %s\n", state.String())
	})

	fmt.Println("Callbacks registered successfully")
	// Output: Callbacks registered successfully
}

// Example_contextIntegration demonstrates context usage for timeouts.
//
// This example shows how to use context for controlling connection attempts.
func Example_contextIntegration() {
	tmpDir := os.TempDir()
	sockPath := filepath.Join(tmpDir, "example_ctx.sock")
	defer os.Remove(sockPath)

	// Create client
	client := sckclt.New(sockPath)
	if client == nil {
		fmt.Println("Failed to create client")
		return
	}
	defer client.Close()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Try to connect (will timeout since no server)
	err := client.Connect(ctx)
	if err != nil {
		fmt.Println("Connect timed out as expected")
	}

	// Output: Connect timed out as expected
}

// Example_errorHandling demonstrates comprehensive error handling.
//
// This example shows how to handle various error scenarios.
func Example_errorHandling() {
	// Invalid socket path (empty)
	client := sckclt.New("")
	if client == nil {
		fmt.Println("Empty path rejected")
	}

	// Valid client but not connected
	tmpDir := os.TempDir()
	sockPath := filepath.Join(tmpDir, "example_err.sock")
	defer os.Remove(sockPath)

	client = sckclt.New(sockPath)
	if client == nil {
		fmt.Println("Failed to create client")
		return
	}

	// Try to write before connect
	_, err := client.Write([]byte("test"))
	if err != nil {
		fmt.Println("Write without connect failed as expected")
	}

	// Try to close before connect
	err = client.Close()
	if err != nil {
		fmt.Println("Close without connect failed as expected")
	}

	// Output:
	// Empty path rejected
	// Write without connect failed as expected
	// Close without connect failed as expected
}

// Example_eventLogging demonstrates using Unix datagram for event logging.
//
// This is a typical use case for fire-and-forget event notifications.
func Example_eventLogging() {
	tmpDir := os.TempDir()
	sockPath := filepath.Join(tmpDir, "example_log.sock")
	defer os.Remove(sockPath)

	// Start logging server
	cfg := sckcfg.Server{
		Network:   libptc.NetworkUnixGram,
		Address:   sockPath,
		PermFile:  libprm.Perm(0600),
		GroupPerm: -1,
	}

	handler := func(ctx libsck.Context) {
		defer ctx.Close()
		buf := make([]byte, 8192)
		// Read single datagram per invocation
		n, err := ctx.Read(buf)
		if err == nil && n > 0 {
			log.Printf("LOG: %s", buf[:n])
		}
	}

	srv, _ := scksrv.New(nil, handler, cfg)
	srvCtx, srvCancel := context.WithCancel(context.Background())
	defer srvCancel()

	go func() {
		_ = srv.Listen(srvCtx)
	}()

	time.Sleep(50 * time.Millisecond)

	// Create logging client
	client := sckclt.New(sockPath)
	if client == nil {
		fmt.Println("Failed to create client")
		return
	}
	defer client.Close()

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		fmt.Printf("Connect error: %v\n", err)
		return
	}

	// Send log events (fire-and-forget, each Write sends one datagram)
	events := []string{
		"INFO: Application started",
		"WARN: High memory usage",
		"ERROR: Database connection failed",
	}

	for _, event := range events {
		_, _ = client.Write([]byte(event))
		time.Sleep(10 * time.Millisecond) // Brief delay between datagrams
	}

	fmt.Println("Log events sent")
	time.Sleep(100 * time.Millisecond) // Allow server to process

	// Output: Log events sent
}

// Example_metricsCollection demonstrates using Unix datagram for metrics.
//
// This shows a common pattern for local metrics collection.
func Example_metricsCollection() {
	tmpDir := os.TempDir()
	sockPath := filepath.Join(tmpDir, "example_metrics.sock")
	defer os.Remove(sockPath)

	// Start metrics server
	cfg := sckcfg.Server{
		Network:   libptc.NetworkUnixGram,
		Address:   sockPath,
		PermFile:  libprm.Perm(0660),
		GroupPerm: -1,
	}

	handler := func(ctx libsck.Context) {
		defer func() {
			_ = ctx.Close()
		}()

		for {
			buf := make([]byte, 1024)
			n, e := ctx.Read(buf)

			if n > 0 {
				fmt.Printf("Metric: %s\n", buf[:n])
			}
			if e != nil {
				break
			}
		}
	}

	srv, _ := scksrv.New(nil, handler, cfg)
	srvCtx, srvCancel := context.WithCancel(context.Background())
	defer srvCancel()

	go func() {
		_ = srv.Listen(srvCtx)
	}()

	time.Sleep(50 * time.Millisecond)

	// Create metrics client
	client := sckclt.New(sockPath)
	if client == nil {
		fmt.Println("Failed to create client")
		return
	}
	defer client.Close()

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		fmt.Printf("Connect error: %v\n", err)
		return
	}

	// Send metrics
	metrics := []string{
		"http.requests:1|c",
		"cpu.usage:45.2|g",
		"response.time:123|ms",
	}

	for _, metric := range metrics {
		_, _ = client.Write([]byte(metric))
		time.Sleep(10 * time.Millisecond)
	}

	time.Sleep(200 * time.Millisecond) // Allow processing

	// Output:
	// Metric: http.requests:1|c
	// Metric: cpu.usage:45.2|g
	// Metric: response.time:123|ms
}

// Example_reconnection demonstrates reconnecting a client.
//
// This shows how to handle reconnection scenarios.
func Example_reconnection() {
	tmpDir := os.TempDir()
	sockPath := filepath.Join(tmpDir, "example_reconnect.sock")
	defer os.Remove(sockPath)

	// Start server
	cfg := sckcfg.Server{
		Network:   libptc.NetworkUnixGram,
		Address:   sockPath,
		PermFile:  libprm.Perm(0600),
		GroupPerm: -1,
	}

	handler := func(ctx libsck.Context) {
		defer ctx.Close()
		buf := make([]byte, 1024)
		_, _ = ctx.Read(buf)
	}

	srv, _ := scksrv.New(nil, handler, cfg)
	srvCtx, srvCancel := context.WithCancel(context.Background())
	defer srvCancel()

	go func() {
		_ = srv.Listen(srvCtx)
	}()

	time.Sleep(50 * time.Millisecond)

	// Create client
	client := sckclt.New(sockPath)
	if client == nil {
		fmt.Println("Failed to create client")
		return
	}
	defer client.Close()

	ctx := context.Background()

	// First connection
	err := client.Connect(ctx)
	if err != nil {
		fmt.Printf("First connect error: %v\n", err)
		return
	}
	fmt.Println("First connection established")

	// Reconnect (replaces old connection)
	err = client.Connect(ctx)
	if err != nil {
		fmt.Printf("Reconnect error: %v\n", err)
		return
	}
	fmt.Println("Reconnection successful")

	// Output:
	// First connection established
	// Reconnection successful
}
