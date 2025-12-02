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
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	libsck "github.com/nabbar/golib/socket"
	"github.com/nabbar/golib/socket/client/unix"
)

// Example demonstrates the simplest UNIX socket client usage.
// This is the most basic way to connect and send data to a UNIX socket server.
//
// Note: This example requires a running UNIX socket server at /tmp/app.sock.
func Example() {
	socketPath := "/tmp/example-basic.sock"

	// Create client
	client := unix.New(socketPath)
	if client == nil {
		log.Fatal("Invalid socket path")
	}
	defer client.Close()

	// Note: Server must be running at /tmp/example-basic.sock
	// In production, you would check for connection errors
	ctx := context.Background()
	_ = client.Connect(ctx)

	// Send data
	_, _ = client.Write([]byte("Hello, UNIX socket!"))

	// Output:
}

// Example_basicClient demonstrates basic client creation and usage.
// Shows the standard pattern of creating, connecting, and using a UNIX socket client.
func Example_basicClient() {
	socketPath := "/tmp/example-client.sock"

	// Create client
	client := unix.New(socketPath)
	if client == nil {
		fmt.Println("Invalid socket path")
		return
	}
	defer client.Close()

	// Connect to server
	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		fmt.Printf("Connect error (expected, server not running): %v\n", err != nil)
		return
	}

	// Send data
	data := []byte("Hello from client")
	n, err := client.Write(data)
	if err != nil {
		fmt.Println("Write failed")
		return
	}
	fmt.Printf("Sent %d bytes\n", n)

	// Output:
	// Connect error (expected, server not running): true
}

// Example_withCallbacks demonstrates using error and info callbacks.
// Callbacks are executed asynchronously and are useful for logging and monitoring.
func Example_withCallbacks() {
	socketPath := "/tmp/example-callbacks.sock"

	client := unix.New(socketPath)
	if client == nil {
		return
	}
	defer client.Close()

	var callbackCount int
	var mu sync.Mutex

	// Register error callback
	client.RegisterFuncError(func(errs ...error) {
		mu.Lock()
		defer mu.Unlock()
		callbackCount++
		for _, err := range errs {
			_ = err // Would log error in production
		}
	})

	// Register info callback
	client.RegisterFuncInfo(func(local, remote net.Addr, state libsck.ConnState) {
		mu.Lock()
		defer mu.Unlock()
		callbackCount++
		_ = state // Would log state change in production
	})

	ctx := context.Background()
	_ = client.Connect(ctx)

	// Allow callbacks to execute
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	count := callbackCount
	mu.Unlock()

	fmt.Printf("Callbacks executed: %v\n", count > 0)

	// Output:
	// Callbacks executed: true
}

// Example_oneShot demonstrates the Once() method for simple request/response.
// This is useful for fire-and-forget operations or simple queries.
func Example_oneShot() {
	socketPath := "/tmp/example-once.sock"

	client := unix.New(socketPath)
	if client == nil {
		return
	}

	request := bytes.NewBufferString("PING")

	// Once() handles connect, write, read, and close automatically
	err := client.Once(context.Background(), request, func(reader io.Reader) {
		response, _ := io.ReadAll(reader)
		fmt.Printf("Response: %s\n", string(response))
	})

	if err != nil {
		fmt.Println("Once failed (expected without server)")
	}

	// Note: Client is automatically closed after Once()
	fmt.Printf("Connection closed: %t\n", !client.IsConnected())

	// Output:
	// Once failed (expected without server)
	// Connection closed: true
}

// Example_contextTimeout demonstrates using context for operation timeouts.
// This is important for preventing operations from hanging indefinitely.
func Example_contextTimeout() {
	socketPath := "/tmp/example-timeout.sock"

	client := unix.New(socketPath)
	if client == nil {
		return
	}
	defer client.Close()

	// Create context with 5 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Connect will timeout if server doesn't respond
	err := client.Connect(ctx)

	if err != nil {
		fmt.Println("Connect failed (expected without server)")
	}

	// Output:
	// Connect failed (expected without server)
}

// Example_readWrite demonstrates bidirectional communication.
// Shows how to send data and receive responses.
func Example_readWrite() {
	socketPath := "/tmp/example-rw.sock"

	client := unix.New(socketPath)
	if client == nil {
		return
	}
	defer client.Close()

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		fmt.Println("Connect failed (expected without server)")
		return
	}

	// Send request
	request := []byte("GET /status")
	if _, err := client.Write(request); err != nil {
		fmt.Println("Write failed")
		return
	}

	// Read response
	response := make([]byte, 1024)
	n, err := client.Read(response)
	if err != nil {
		fmt.Println("Read failed")
		return
	}

	fmt.Printf("Received %d bytes\n", n)

	// Output:
	// Connect failed (expected without server)
}

// Example_stateMonitoring demonstrates monitoring connection state.
// Useful for health checks and connection management.
func Example_stateMonitoring() {
	socketPath := "/tmp/example-state.sock"

	client := unix.New(socketPath)
	if client == nil {
		return
	}
	defer client.Close()

	// Check initial state
	fmt.Printf("Initially connected: %t\n", client.IsConnected())

	// Connect
	ctx := context.Background()
	_ = client.Connect(ctx)

	// Check state after connect attempt
	fmt.Printf("After connect: %t\n", client.IsConnected())

	// Close
	_ = client.Close()

	// Check state after close
	fmt.Printf("After close: %t\n", client.IsConnected())

	// Output:
	// Initially connected: false
	// After connect: false
	// After close: false
}

// Example_errorHandling demonstrates proper error handling patterns.
// Shows how to handle different types of errors that can occur.
func Example_errorHandling() {
	// Test invalid path
	client := unix.New("")
	if client == nil {
		fmt.Println("Empty path rejected: true")
	}

	// Test connection error
	client = unix.New("/tmp/nonexistent-socket.sock")
	if client == nil {
		return
	}
	defer client.Close()

	ctx := context.Background()
	err := client.Connect(ctx)
	if err != nil {
		fmt.Println("Connection to nonexistent socket failed: true")
	}

	// Test operation without connection
	_, err = client.Write([]byte("test"))
	if err != nil {
		fmt.Println("Write without connection failed: true")
	}

	// Output:
	// Empty path rejected: true
	// Connection to nonexistent socket failed: true
	// Write without connection failed: true
}

// Example_socketPaths demonstrates various valid socket path formats.
// UNIX sockets can be placed in different locations depending on use case.
func Example_socketPaths() {
	// Temporary socket (cleaned up on reboot)
	tmpSocket := filepath.Join(os.TempDir(), "app.sock")
	client1 := unix.New(tmpSocket)
	fmt.Printf("Temp socket created: %t\n", client1 != nil)
	client1.Close()

	// System daemon socket (requires root or specific permissions)
	daemonSocket := "/var/run/daemon.sock"
	client2 := unix.New(daemonSocket)
	fmt.Printf("Daemon socket created: %t\n", client2 != nil)
	client2.Close()

	// User-specific socket
	userSocket := fmt.Sprintf("/run/user/%d/app.sock", os.Getuid())
	client3 := unix.New(userSocket)
	fmt.Printf("User socket created: %t\n", client3 != nil)
	client3.Close()

	// Relative path socket
	relSocket := "./local.sock"
	client4 := unix.New(relSocket)
	fmt.Printf("Relative socket created: %t\n", client4 != nil)
	client4.Close()

	// Output:
	// Temp socket created: true
	// Daemon socket created: true
	// User socket created: true
	// Relative socket created: true
}

// Example_reconnection demonstrates handling connection failures and reconnection.
// Important for robust applications that need to handle server restarts.
func Example_reconnection() {
	socketPath := "/tmp/example-reconnect.sock"

	client := unix.New(socketPath)
	if client == nil {
		return
	}
	defer client.Close()

	ctx := context.Background()

	// Try to connect (will fail without server)
	err := client.Connect(ctx)
	if err != nil {
		fmt.Println("Initial connect failed: true")
	}

	// Simulate waiting for server to start
	time.Sleep(100 * time.Millisecond)

	// Try to reconnect
	err = client.Connect(ctx)
	if err != nil {
		fmt.Println("Reconnect failed: true")
	}

	// Output:
	// Initial connect failed: true
	// Reconnect failed: true
}

// Example_tlsNoOp demonstrates that SetTLS is a no-op for UNIX sockets.
// UNIX sockets don't support TLS; use filesystem permissions for security.
func Example_tlsNoOp() {
	socketPath := "/tmp/example-tls.sock"

	client := unix.New(socketPath)
	if client == nil {
		return
	}
	defer client.Close()

	// SetTLS always returns nil for UNIX sockets
	err := client.SetTLS(true, nil, "")
	fmt.Printf("SetTLS returns nil: %t\n", err == nil)

	// UNIX sockets use filesystem permissions for security
	fmt.Println("Use filesystem permissions instead of TLS")

	// Output:
	// SetTLS returns nil: true
	// Use filesystem permissions instead of TLS
}

// Example_multipleMessages demonstrates sending multiple messages over same connection.
// Connection-oriented sockets maintain state across multiple operations.
func Example_multipleMessages() {
	socketPath := "/tmp/example-multi.sock"

	client := unix.New(socketPath)
	if client == nil {
		return
	}
	defer client.Close()

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		fmt.Println("Connect failed (expected without server)")
		return
	}

	// Send multiple messages
	messages := []string{"MSG1", "MSG2", "MSG3"}
	for i, msg := range messages {
		_, err := client.Write([]byte(msg))
		if err != nil {
			fmt.Printf("Message %d sent: false\n", i+1)
		} else {
			fmt.Printf("Message %d sent: true\n", i+1)
		}
	}

	// Output:
	// Connect failed (expected without server)
}

// Example_callbackOrdering demonstrates callback execution order.
// Shows when callbacks are triggered during the connection lifecycle.
//
// Note: Callbacks are asynchronous, so exact order may vary.
func Example_callbackOrdering() {
	socketPath := "/tmp/example-order.sock"

	client := unix.New(socketPath)
	if client == nil {
		return
	}
	defer client.Close()

	var (
		events []string
		mu     sync.Mutex
	)

	client.RegisterFuncInfo(func(local, remote net.Addr, state libsck.ConnState) {
		mu.Lock()
		events = append(events, state.String())
		mu.Unlock()
	})

	ctx := context.Background()
	_ = client.Connect(ctx)

	// Allow callbacks to execute (they're async)
	time.Sleep(50 * time.Millisecond)

	// Note: Exact order may vary due to async execution
	mu.Lock()
	eventCount := len(events)
	mu.Unlock()

	fmt.Printf("Events captured: %v\n", eventCount > 0)

	// Output:
	// Events captured: true
}
