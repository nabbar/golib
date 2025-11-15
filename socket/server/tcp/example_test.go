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
	"log"
	"net"
	"time"

	libsck "github.com/nabbar/golib/socket"
	scksrv "github.com/nabbar/golib/socket/server/tcp"
)

// Example demonstrates how to create and start a basic TCP echo server.
func Example() {
	// Define the connection handler
	handler := func(r libsck.Reader, w libsck.Writer) {
		defer func() {
			_ = r.Close()
			_ = w.Close()
		}()

		// Echo all received data back to the client
		_, _ = io.Copy(w, r)
	}

	// Create the server with the handler
	srv := scksrv.New(nil, handler)

	// Register the listening address
	if err := srv.RegisterServer(":8080"); err != nil {
		log.Fatalf("Failed to register server: %v", err)
	}

	// Start the server (this blocks until shutdown)
	ctx := context.Background()
	if err := srv.Listen(ctx); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// ExampleNew_withCallbacks demonstrates creating a server with all callbacks registered.
func ExampleNew_withCallbacks() {
	// Define the connection handler
	handler := func(r libsck.Reader, w libsck.Writer) {
		defer func() {
			_ = r.Close()
			_ = w.Close()
		}()
		_, _ = io.Copy(w, r)
	}

	// Create the server
	srv := scksrv.New(nil, handler)

	// Register error callback
	srv.RegisterFuncError(func(errs ...error) {
		for _, err := range errs {
			if err != nil {
				log.Printf("Server error: %v", err)
			}
		}
	})

	// Register connection state callback
	srv.RegisterFuncInfo(func(local, remote net.Addr, state libsck.ConnState) {
		log.Printf("Connection %v -> %v: %v", remote, local, state)
	})

	// Register server info callback
	srv.RegisterFuncInfoServer(func(msg string) {
		log.Printf("Server info: %s", msg)
	})

	// Register the address
	_ = srv.RegisterServer(":8080")

	// Start listening
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_ = srv.Listen(ctx)
}

// ExampleNew_withUpdateConn demonstrates using the UpdateConn callback to configure connections.
func ExampleNew_withUpdateConn() {
	handler := func(r libsck.Reader, w libsck.Writer) {
		defer func() {
			_ = r.Close()
			_ = w.Close()
		}()
		_, _ = io.Copy(w, r)
	}

	// UpdateConn callback to configure each connection
	updateConn := func(conn net.Conn) {
		// Type assert to access TCP-specific methods
		if tcpConn, ok := conn.(*net.TCPConn); ok {
			_ = tcpConn.SetKeepAlive(true)
			_ = tcpConn.SetKeepAlivePeriod(30 * time.Second)
		}
		log.Printf("New connection configured: %v", conn)
	}

	// Create server with UpdateConn callback
	srv := scksrv.New(updateConn, handler)
	_ = srv.RegisterServer(":8080")

	ctx := context.Background()
	_ = srv.Listen(ctx)
}

// ExampleServerTcp_Shutdown demonstrates graceful server shutdown.
func ExampleServerTcp_Shutdown() {
	handler := func(r libsck.Reader, w libsck.Writer) {
		defer func() {
			_ = r.Close()
			_ = w.Close()
		}()
		_, _ = io.Copy(w, r)
	}

	srv := scksrv.New(nil, handler)
	_ = srv.RegisterServer(":8080")

	// Start server in a goroutine
	go func() {
		_ = srv.Listen(context.Background())
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Perform graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Shutdown error: %v", err)
	}

	fmt.Println("Server shutdown complete")
	// Output: Server shutdown complete
}

// ExampleServerTcp_OpenConnections demonstrates monitoring active connections.
func ExampleServerTcp_OpenConnections() {
	handler := func(r libsck.Reader, w libsck.Writer) {
		defer func() {
			_ = r.Close()
			_ = w.Close()
		}()
		time.Sleep(1 * time.Second) // Simulate work
	}

	srv := scksrv.New(nil, handler)
	_ = srv.RegisterServer(":8080")

	// Start server
	go func() {
		_ = srv.Listen(context.Background())
	}()

	time.Sleep(100 * time.Millisecond)

	// Check connection count
	count := srv.OpenConnections()
	fmt.Printf("Active connections: %d\n", count)

	// Shutdown
	_ = srv.Close()

	// Output: Active connections: 0
}

// ExampleServerTcp_IsRunning demonstrates checking server state.
func ExampleServerTcp_IsRunning() {
	handler := func(r libsck.Reader, w libsck.Writer) {
		defer func() {
			_ = r.Close()
			_ = w.Close()
		}()

		_, _ = io.Copy(w, r)
	}

	srv := scksrv.New(nil, handler)
	_ = srv.RegisterServer(":8080")

	// Check before starting
	fmt.Printf("Running before start: %v\n", srv.IsRunning())

	// Start in goroutine
	go func() {
		_ = srv.Listen(context.Background())
	}()

	time.Sleep(100 * time.Millisecond)
	fmt.Printf("Running after start: %v\n", srv.IsRunning())

	// Shutdown
	_ = srv.Close()
	time.Sleep(100 * time.Millisecond)
	fmt.Printf("Running after shutdown: %v\n", srv.IsRunning())

	// Output:
	// Running before start: false
	// Running after start: true
	// Running after shutdown: false
}
