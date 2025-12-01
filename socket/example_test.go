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
 */

package socket_test

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	libsck "github.com/nabbar/golib/socket"
)

// ExampleConnState demonstrates the string representation of connection states.
// This is useful for logging and monitoring connection lifecycle events.
func ExampleConnState() {
	// Display all connection states
	states := []libsck.ConnState{
		libsck.ConnectionDial,
		libsck.ConnectionNew,
		libsck.ConnectionRead,
		libsck.ConnectionHandler,
		libsck.ConnectionWrite,
		libsck.ConnectionClose,
	}

	for _, state := range states {
		fmt.Println(state.String())
	}

	// Output:
	// Dial Connection
	// New Connection
	// Read Incoming Stream
	// Run HandlerFunc
	// Write Outgoing Steam
	// Close Connection
}

// ExampleErrorFilter demonstrates filtering expected network errors.
// This is useful during connection cleanup and shutdown sequences.
func ExampleErrorFilter() {
	// Simulate various errors
	normalErr := fmt.Errorf("connection timeout")
	closedErr := fmt.Errorf("use of closed network connection")

	// Filter errors
	if err := libsck.ErrorFilter(normalErr); err != nil {
		fmt.Println("Error occurred:", err.Error())
	}

	// This error is filtered out (returns nil)
	if err := libsck.ErrorFilter(closedErr); err != nil {
		fmt.Println("Error occurred:", err.Error())
	} else {
		fmt.Println("Closed connection error filtered")
	}

	// Nil errors are handled safely
	if err := libsck.ErrorFilter(nil); err != nil {
		fmt.Println("Error occurred:", err.Error())
	} else {
		fmt.Println("No error")
	}

	// Output:
	// Error occurred: connection timeout
	// Closed connection error filtered
	// No error
}

// ExampleFuncError demonstrates registering an error callback for socket operations.
// This pattern is used in both client and server implementations.
func ExampleFuncError() {
	// Define error handler
	errHandler := func(errs ...error) {
		for _, err := range errs {
			// Filter expected errors
			if err := libsck.ErrorFilter(err); err != nil {
				log.Printf("Socket error: %v", err)
			}
		}
	}

	// Simulate error handling
	errHandler(fmt.Errorf("connection refused"))
	errHandler(fmt.Errorf("use of closed network connection")) // Filtered
	errHandler(fmt.Errorf("timeout"), fmt.Errorf("network unreachable"))

	fmt.Println("Error handler demonstrated")

	// Output:
	// Error handler demonstrated
}

// ExampleFuncInfo demonstrates tracking connection state changes.
// This is useful for monitoring, logging, and debugging.
func ExampleFuncInfo() {
	// Define connection info handler
	infoHandler := func(local, remote net.Addr, state libsck.ConnState) {
		fmt.Printf("[%s] %s -> %s\n", state.String(), remote.String(), local.String())
	}

	// Simulate connection lifecycle
	localAddr := &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080}
	remoteAddr := &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 54321}

	infoHandler(localAddr, remoteAddr, libsck.ConnectionNew)
	infoHandler(localAddr, remoteAddr, libsck.ConnectionRead)
	infoHandler(localAddr, remoteAddr, libsck.ConnectionWrite)
	infoHandler(localAddr, remoteAddr, libsck.ConnectionClose)

	// Output:
	// [New Connection] 127.0.0.1:54321 -> 127.0.0.1:8080
	// [Read Incoming Stream] 127.0.0.1:54321 -> 127.0.0.1:8080
	// [Write Outgoing Steam] 127.0.0.1:54321 -> 127.0.0.1:8080
	// [Close Connection] 127.0.0.1:54321 -> 127.0.0.1:8080
}

// ExampleHandlerFunc demonstrates a basic request handler for socket servers.
// This shows the typical pattern for handling connections.
func ExampleHandlerFunc() {
	// Define handler function
	handler := func(ctx libsck.Context) {
		// Check if connection is active
		if !ctx.IsConnected() {
			return
		}

		// Log connection info
		fmt.Printf("Handling connection from %s\n", ctx.RemoteHost())

		// Read request
		buf := make([]byte, 1024)
		n, err := ctx.Read(buf)
		if err != nil {
			if err != io.EOF {
				fmt.Printf("Read error: %v\n", err)
			}
			return
		}

		// Process request
		request := string(buf[:n])
		fmt.Printf("Received: %s\n", request)

		// Send response
		response := []byte("Response: " + request)
		if _, err := ctx.Write(response); err != nil {
			fmt.Printf("Write error: %v\n", err)
			return
		}

		fmt.Println("Request handled successfully")
	}

	// Handler is now ready to be used in server configuration
	_ = handler
	fmt.Println("Handler function created")

	// Output:
	// Handler function created
}

// ExampleContext demonstrates using the Context interface in a handler.
// This shows how to access connection state and perform I/O operations.
func ExampleContext() {
	// This example demonstrates Context usage patterns
	// In real usage, Context is provided by the server to HandlerFunc

	// Example handler showing Context usage
	handler := func(ctx libsck.Context) {
		// Check connection state
		if !ctx.IsConnected() {
			fmt.Println("Connection closed")
			return
		}

		// Get connection information
		fmt.Printf("Local: %s, Remote: %s\n", ctx.LocalHost(), ctx.RemoteHost())

		// Check for cancellation
		select {
		case <-ctx.Done():
			fmt.Println("Context canceled:", ctx.Err())
			return
		default:
		}

		// Perform I/O operations
		buf := make([]byte, 1024)
		if n, err := ctx.Read(buf); err == nil {
			fmt.Printf("Read %d bytes\n", n)
			ctx.Write([]byte("ACK"))
		}
	}

	_ = handler
	fmt.Println("Context usage demonstrated")

	// Output:
	// Context usage demonstrated
}

// ExampleServer demonstrates the Server interface usage pattern.
// This shows how servers are typically configured and started.
func ExampleServer() {
	// This demonstrates the Server interface pattern
	// Actual server creation is done via protocol-specific packages

	var srv libsck.Server // Created via server.New() or protocol-specific constructors

	if srv != nil {
		// Register error handler
		srv.RegisterFuncError(func(errs ...error) {
			for _, err := range errs {
				log.Printf("Error: %v", err)
			}
		})

		// Register connection state tracker
		srv.RegisterFuncInfo(func(local, remote net.Addr, state libsck.ConnState) {
			log.Printf("[%s] %s -> %s", state, remote, local)
		})

		// Start server with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// This would block until context is canceled or error occurs
		_ = srv.Listen(ctx)

		// Graceful shutdown
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		srv.Shutdown(shutdownCtx)
	}

	fmt.Println("Server interface demonstrated")

	// Output:
	// Server interface demonstrated
}

// ExampleClient demonstrates the Client interface usage pattern.
// This shows how clients are typically configured and used.
func ExampleClient() {
	// This demonstrates the Client interface pattern
	// Actual client creation is done via protocol-specific packages

	var cli libsck.Client // Created via client.New() or protocol-specific constructors

	if cli != nil {
		// Register error handler
		cli.RegisterFuncError(func(errs ...error) {
			for _, err := range errs {
				log.Printf("Error: %v", err)
			}
		})

		// Connect to server
		ctx := context.Background()
		if err := cli.Connect(ctx); err != nil {
			log.Fatal(err)
		}
		defer cli.Close()

		// Write request
		if _, err := cli.Write([]byte("Hello, server!\n")); err != nil {
			log.Fatal(err)
		}

		// Read response
		buf := make([]byte, 1024)
		if n, err := cli.Read(buf); err == nil {
			fmt.Printf("Response: %s", buf[:n])
		}
	}

	fmt.Println("Client interface demonstrated")

	// Output:
	// Client interface demonstrated
}

// ExampleHandler demonstrates using a stateful handler type instead of HandlerFunc.
// This is useful for handlers that need to maintain state or access dependencies.
func ExampleHandler() {
	// Define a stateful handler type
	type EchoHandler struct {
		prefix string
	}

	// Implement the Handler interface
	handleFunc := func(h *EchoHandler, ctx libsck.Context) {
		buf := make([]byte, 1024)
		n, err := ctx.Read(buf)
		if err != nil {
			return
		}

		// Use handler state
		response := []byte(h.prefix + string(buf[:n]))
		ctx.Write(response)
	}

	// Create handler instance
	handler := &EchoHandler{prefix: "Echo: "}

	// Use in server (would be: cfg.Handler = handler)
	_ = handleFunc
	_ = handler

	fmt.Println("Stateful handler demonstrated")

	// Output:
	// Stateful handler demonstrated
}

// ExampleUpdateConn demonstrates customizing connection settings.
// This callback is used to configure socket options before use.
func ExampleUpdateConn() {
	// Define connection customization function
	updateConn := func(conn net.Conn) {
		// Set connection deadlines
		if tcpConn, ok := conn.(*net.TCPConn); ok {
			// Disable Nagle's algorithm for low latency
			_ = tcpConn.SetNoDelay(true)

			// Set keepalive
			_ = tcpConn.SetKeepAlive(true)
			_ = tcpConn.SetKeepAlivePeriod(30 * time.Second)

			// Set buffer sizes
			_ = tcpConn.SetReadBuffer(64 * 1024)
			_ = tcpConn.SetWriteBuffer(64 * 1024)
		}

		// Set overall deadlines
		_ = conn.SetDeadline(time.Now().Add(5 * time.Minute))
	}

	_ = updateConn
	fmt.Println("Connection customization demonstrated")

	// Output:
	// Connection customization demonstrated
}

// ExampleResponse demonstrates processing server responses on the client side.
// This callback is used with Client.Once() for request/response patterns.
func ExampleResponse() {
	// Define response handler
	responseHandler := func(r io.Reader) {
		buf := make([]byte, 1024)
		n, err := r.Read(buf)
		if err != nil && err != io.EOF {
			log.Printf("Response error: %v", err)
			return
		}

		response := string(buf[:n])
		fmt.Printf("Server response: %s\n", response)

		// Process response data
		// Parse, validate, etc.
	}

	_ = responseHandler
	fmt.Println("Response handler demonstrated")

	// Output:
	// Response handler demonstrated
}

// Example_errorHandling demonstrates comprehensive error handling patterns.
// This shows best practices for handling errors in socket operations.
func Example_errorHandling() {
	// Define comprehensive error handler
	handleErrors := func(operation string, errs ...error) {
		for _, err := range errs {
			// Filter expected errors
			if err := libsck.ErrorFilter(err); err != nil {
				// Log with context
				fmt.Printf("[%s] Error: %v\n", operation, err)

				// Handle specific error types
				if netErr, ok := err.(net.Error); ok {
					if netErr.Timeout() {
						fmt.Println("Network timeout")
					}
					if netErr.Temporary() {
						fmt.Println("Temporary error, can retry")
					}
				}
			}
		}
	}

	// Simulate various error scenarios
	handleErrors("connect", fmt.Errorf("connection refused"))
	handleErrors("read", fmt.Errorf("use of closed network connection")) // Filtered
	handleErrors("write", fmt.Errorf("broken pipe"))

	// Output:
	// [connect] Error: connection refused
	// [write] Error: broken pipe
}

// Example_connectionLifecycle demonstrates a complete connection lifecycle
// with proper state tracking and resource management.
func Example_connectionLifecycle() {
	// Track connection state
	states := make([]libsck.ConnState, 0)

	// Info callback to track states
	trackState := func(local, remote net.Addr, state libsck.ConnState) {
		states = append(states, state)
	}

	// Simulate lifecycle
	localAddr := &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080}
	remoteAddr := &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 54321}

	// Client dialing
	trackState(localAddr, remoteAddr, libsck.ConnectionDial)

	// Connection established
	trackState(localAddr, remoteAddr, libsck.ConnectionNew)

	// Request processing
	trackState(localAddr, remoteAddr, libsck.ConnectionRead)
	trackState(localAddr, remoteAddr, libsck.ConnectionHandler)
	trackState(localAddr, remoteAddr, libsck.ConnectionWrite)

	// Connection teardown
	trackState(localAddr, remoteAddr, libsck.ConnectionCloseRead)
	trackState(localAddr, remoteAddr, libsck.ConnectionCloseWrite)
	trackState(localAddr, remoteAddr, libsck.ConnectionClose)

	// Display lifecycle
	fmt.Println("Connection lifecycle:")
	for i, state := range states {
		fmt.Printf("%d. %s\n", i+1, state.String())
	}

	// Output:
	// Connection lifecycle:
	// 1. Dial Connection
	// 2. New Connection
	// 3. Read Incoming Stream
	// 4. Run HandlerFunc
	// 5. Write Outgoing Steam
	// 6. Close Incoming Stream
	// 7. Close Outgoing Stream
	// 8. Close Connection
}
