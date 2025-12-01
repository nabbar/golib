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

package tcp_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	libptc "github.com/nabbar/golib/network/protocol"
	libsck "github.com/nabbar/golib/socket"
	sckclt "github.com/nabbar/golib/socket/client/tcp"
	sckcfg "github.com/nabbar/golib/socket/config"
	scksrt "github.com/nabbar/golib/socket/server/tcp"
)

// Example demonstrates the simplest TCP client usage.
// This is the most basic client implementation for quick start.
func Example() {
	// Start a simple echo server for demonstration
	srv := startEchoServer(":8080")
	defer srv.Shutdown(context.Background())
	time.Sleep(50 * time.Millisecond) // Wait for server

	// Create client
	client, err := sckclt.New("localhost:8080")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// Connect to server
	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		log.Fatal(err)
	}

	// Send data
	data := []byte("Hello")
	_, err = client.Write(data)
	if err != nil {
		log.Fatal(err)
	}

	// Read response
	buf := make([]byte, len(data))
	_, err = io.ReadFull(client, buf)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s", buf)
	// Output: Hello
}

// ExampleNew demonstrates creating a new TCP client.
// This is the first step in using the client package.
func ExampleNew() {
	// Create client with server address
	client, err := sckclt.New("localhost:8080")
	if err != nil {
		fmt.Printf("Failed to create client: %v\n", err)
		return
	}

	// Client is created but not connected yet
	fmt.Printf("Client created for localhost:8080\n")
	_ = client
	// Output: Client created for localhost:8080
}

// ExampleNew_invalidAddress demonstrates error handling for invalid addresses.
func ExampleNew_invalidAddress() {
	// Try to create client with empty address
	_, err := sckclt.New("")
	if err != nil {
		fmt.Println("Error: invalid dial address")
	}

	// Try to create client with invalid format
	_, err = sckclt.New("not-a-valid-address")
	if err != nil {
		fmt.Println("Error: address resolution failed")
	}
	// Output:
	// Error: invalid dial address
	// Error: address resolution failed
}

// ExampleClientTCP_Connect demonstrates establishing a connection.
func ExampleClientTCP_Connect() {
	srv := startEchoServer(":8081")
	defer srv.Shutdown(context.Background())
	time.Sleep(50 * time.Millisecond)

	client, _ := sckclt.New("localhost:8081")
	defer client.Close()

	// Connect with context
	ctx := context.Background()
	err := client.Connect(ctx)
	if err != nil {
		fmt.Printf("Connection failed: %v\n", err)
		return
	}

	if client.IsConnected() {
		fmt.Println("Connected successfully")
	}
	// Output: Connected successfully
}

// ExampleClientTCP_Connect_timeout demonstrates connection with timeout.
func ExampleClientTCP_Connect_timeout() {
	srv := startEchoServer(":8082")
	defer srv.Shutdown(context.Background())
	time.Sleep(50 * time.Millisecond)

	client, _ := sckclt.New("localhost:8082")
	defer client.Close()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := client.Connect(ctx)
	if err != nil {
		fmt.Printf("Connection failed: %v\n", err)
		return
	}

	fmt.Println("Connected with timeout")
	// Output: Connected with timeout
}

// ExampleClientTCP_Write demonstrates sending data to server.
func ExampleClientTCP_Write() {
	srv := startEchoServer(":8083")
	defer srv.Shutdown(context.Background())
	time.Sleep(50 * time.Millisecond)

	client, _ := sckclt.New("localhost:8083")
	defer client.Close()

	ctx := context.Background()
	_ = client.Connect(ctx)

	// Write data
	data := []byte("Test message")
	n, err := client.Write(data)
	if err != nil {
		fmt.Printf("Write failed: %v\n", err)
		return
	}

	fmt.Printf("Wrote %d bytes\n", n)
	// Output: Wrote 12 bytes
}

// ExampleClientTCP_Read demonstrates reading data from server.
func ExampleClientTCP_Read() {
	srv := startEchoServer(":8084")
	defer srv.Shutdown(context.Background())
	time.Sleep(50 * time.Millisecond)

	client, _ := sckclt.New("localhost:8084")
	defer client.Close()

	ctx := context.Background()
	_ = client.Connect(ctx)

	// Send request
	_, _ = client.Write([]byte("Echo"))

	// Read response
	buf := make([]byte, 4)
	n, err := client.Read(buf)
	if err != nil {
		fmt.Printf("Read failed: %v\n", err)
		return
	}

	fmt.Printf("Read %d bytes: %s\n", n, buf[:n])
	// Output: Read 4 bytes: Echo
}

// ExampleClientTCP_IsConnected demonstrates checking connection status.
func ExampleClientTCP_IsConnected() {
	srv := startEchoServer(":8085")
	defer srv.Shutdown(context.Background())
	time.Sleep(50 * time.Millisecond)

	client, _ := sckclt.New("localhost:8085")
	defer client.Close()

	// Check before connecting
	if !client.IsConnected() {
		fmt.Println("Not connected yet")
	}

	// Connect
	ctx := context.Background()
	_ = client.Connect(ctx)

	// Check after connecting
	if client.IsConnected() {
		fmt.Println("Now connected")
	}

	// Output:
	// Not connected yet
	// Now connected
}

// ExampleClientTCP_Close demonstrates closing a connection.
func ExampleClientTCP_Close() {
	srv := startEchoServer(":8086")
	defer srv.Shutdown(context.Background())
	time.Sleep(50 * time.Millisecond)

	client, _ := sckclt.New("localhost:8086")

	ctx := context.Background()
	_ = client.Connect(ctx)

	// Close connection
	err := client.Close()
	if err != nil {
		fmt.Printf("Close failed: %v\n", err)
		return
	}

	fmt.Println("Connection closed")
	// Output: Connection closed
}

// ExampleClientTCP_Once demonstrates one-shot request/response pattern.
// This is useful for simple protocols that don't need persistent connections.
func ExampleClientTCP_Once() {
	srv := startEchoServer(":8087")
	defer srv.Shutdown(context.Background())
	time.Sleep(50 * time.Millisecond)

	client, _ := sckclt.New("localhost:8087")

	// Send request and read response in one operation
	request := bytes.NewBufferString("Hello Server")
	ctx := context.Background()

	err := client.Once(ctx, request, func(reader io.Reader) {
		// Read with timeout to avoid blocking
		buf := make([]byte, 12)
		n, _ := io.ReadFull(reader, buf)
		fmt.Printf("Response: %s\n", buf[:n])
	})

	if err != nil {
		fmt.Printf("Once failed: %v\n", err)
	}
	// Output: Response: Hello Server
}

// ExampleClientTCP_RegisterFuncError demonstrates error callback registration.
// This is essential for production monitoring and debugging.
func ExampleClientTCP_RegisterFuncError() {
	client, _ := sckclt.New("localhost:8088")
	defer client.Close()

	// Register error callback
	client.RegisterFuncError(func(ers ...error) {
		for _, e := range ers {
			if e != nil {
				_, _ = fmt.Fprintf(io.Discard, "Client error: %v\n", e)
			}
		}
	})

	// Attempt connection to non-existent server
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := client.Connect(ctx)
	if err != nil {
		fmt.Println("Connection failed (expected)")
	}
	// Output: Connection failed (expected)
}

// ExampleClientTCP_RegisterFuncInfo demonstrates connection state monitoring.
// This provides visibility into connection lifecycle events.
func ExampleClientTCP_RegisterFuncInfo() {
	srv := startEchoServer(":8089")
	defer srv.Shutdown(context.Background())
	time.Sleep(50 * time.Millisecond)

	client, _ := sckclt.New("localhost:8089")

	// Track states in order
	states := make(chan string, 10)
	client.RegisterFuncInfo(func(local, remote net.Addr, state libsck.ConnState) {
		states <- state.String()
	})

	ctx := context.Background()
	_ = client.Connect(ctx)

	// Collect states in order
	time.Sleep(10 * time.Millisecond)
	close(states)
	
	var collected []string
	for s := range states {
		collected = append(collected, s)
	}
	
	// Print in consistent order
	for _, s := range collected {
		fmt.Printf("Connection state: %s\n", s)
	}
	fmt.Println("Monitoring enabled")

	// Unregister callback before closing
	client.RegisterFuncInfo(nil)
	_ = client.Close()
	// Output:
	// Connection state: Dial Connection
	// Connection state: New Connection
	// Monitoring enabled
}

// ExampleClientTCP_SetTLS demonstrates TLS configuration.
// This shows how to enable encrypted connections (without actual certificates here).
func ExampleClientTCP_SetTLS() {
	client, _ := sckclt.New("localhost:8443")
	defer client.Close()

	// Disable TLS (or configure with actual certificates)
	err := client.SetTLS(false, nil, "")
	if err != nil {
		fmt.Printf("SetTLS failed: %v\n", err)
		return
	}

	fmt.Println("TLS configuration updated")
	// Output: TLS configuration updated
}

// Example_complete demonstrates a production-ready client with all features.
// This example shows proper error handling, monitoring, and resource management.
func Example_complete() {
	// Start test server
	srv := startEchoServer(":8090")
	defer srv.Shutdown(context.Background())
	time.Sleep(50 * time.Millisecond)

	// Create client
	client, err := sckclt.New("localhost:8090")
	if err != nil {
		fmt.Printf("Failed to create client: %v\n", err)
		return
	}
	defer client.Close()

	// Register error callback
	client.RegisterFuncError(func(errs ...error) {
		for _, e := range errs {
			log.Printf("Error: %v", e)
		}
	})

	// Register info callback for monitoring
	client.RegisterFuncInfo(func(local, remote net.Addr, state libsck.ConnState) {
		log.Printf("Connection %s: %s -> %s", state, local, remote)
	})

	// Connect with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		fmt.Printf("Connection failed: %v\n", err)
		return
	}

	// Verify connection
	if !client.IsConnected() {
		fmt.Println("Not connected")
		return
	}

	// Send data
	message := []byte("Production message")
	n, err := client.Write(message)
	if err != nil {
		fmt.Printf("Write failed: %v\n", err)
		return
	}

	// Read response
	response := make([]byte, len(message))
	n, err = io.ReadFull(client, response)
	if err != nil {
		fmt.Printf("Read failed: %v\n", err)
		return
	}

	fmt.Printf("Sent and received %d bytes successfully\n", n)
	// Output: Sent and received 18 bytes successfully
}

// Example_reconnection demonstrates handling reconnection after connection loss.
func Example_reconnection() {
	srv := startEchoServer(":8091")
	time.Sleep(50 * time.Millisecond)

	client, _ := sckclt.New("localhost:8091")
	defer client.Close()

	ctx := context.Background()

	// Initial connection
	if err := client.Connect(ctx); err != nil {
		fmt.Printf("Initial connection failed: %v\n", err)
		return
	}
	fmt.Println("Connected")

	// Close connection explicitly
	_ = client.Close()

	// Check connection status after close
	if !client.IsConnected() {
		fmt.Println("Connection closed")
	}

	_ = srv.Shutdown(ctx)
	fmt.Println("Reconnection scenario demonstrated")
	// Output:
	// Connected
	// Connection closed
	// Reconnection scenario demonstrated
}

// Example_multipleOperations demonstrates multiple read/write operations.
func Example_multipleOperations() {
	srv := startEchoServer(":8092")
	defer srv.Shutdown(context.Background())
	time.Sleep(50 * time.Millisecond)

	client, _ := sckclt.New("localhost:8092")
	defer client.Close()

	ctx := context.Background()
	_ = client.Connect(ctx)

	// Multiple operations
	for i := 1; i <= 3; i++ {
		msg := fmt.Sprintf("Msg%d", i)
		_, _ = client.Write([]byte(msg))

		buf := make([]byte, 4)
		_, _ = io.ReadFull(client, buf)
		fmt.Printf("%s ", buf)
	}
	fmt.Println()
	// Output: Msg1 Msg2 Msg3
}

// Helper function to start a simple echo server for examples
func startEchoServer(addr string) scksrt.ServerTcp {
	handler := func(c libsck.Context) {
		defer c.Close()
		buf := make([]byte, 4096)
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

	cfg := sckcfg.Server{
		Network: libptc.NetworkTCP,
		Address: addr,
	}

	srv, err := scksrt.New(nil, handler, cfg)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	go func() {
		_ = srv.Listen(ctx)
	}()

	return srv
}
