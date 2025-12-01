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
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	libsck "github.com/nabbar/golib/socket"
	"github.com/nabbar/golib/socket/client/udp"
)

// Example_basicClient demonstrates the simplest UDP client setup.
//
// This example shows minimal configuration for a UDP client
// that sends a datagram to a remote server.
func Example_basicClient() {
	// Create client
	client, err := udp.New("localhost:8080")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer client.Close()

	// Connect (associate socket with remote address)
	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Send datagram
	data := []byte("Hello, UDP!")
	n, err := client.Write(data)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Sent %d bytes\n", n)
	// Output: Sent 11 bytes
}

// Example_clientWithCallbacks demonstrates callback registration.
//
// This example shows how to register callbacks for error handling
// and operation monitoring.
func Example_clientWithCallbacks() {
	client, err := udp.New("localhost:8081")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer client.Close()

	// Register error callback
	client.RegisterFuncError(func(errs ...error) {
		for _, e := range errs {
			if e != nil {
				fmt.Printf("Client error: %v\n", e)
			}
		}
	})

	// Register info callback
	client.RegisterFuncInfo(func(local, remote net.Addr, state libsck.ConnState) {
		fmt.Printf("State: %s\n", state.String())
	})

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Callbacks registered")
	// Note: Callback state changes are asynchronous and output may vary
}

// Example_oneShotRequest demonstrates one-shot request/response pattern.
//
// This example shows the Once() method for simple request/response
// operations that don't require persistent socket association.
func Example_oneShotRequest() {
	client, err := udp.New("localhost:8082")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Prepare request
	request := bytes.NewBufferString("QUERY")

	// Send and optionally receive response
	err = client.Once(ctx, request, func(reader io.Reader) {
		// In real usage, would read response here
		// For example only, we skip the read
		fmt.Println("Request sent successfully")
	})

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Socket automatically closed after Once()
	fmt.Printf("Connection closed: %v\n", !client.IsConnected())
	// Output:
	// Request sent successfully
	// Connection closed: true
}

// Example_fireAndForget demonstrates fire-and-forget pattern.
//
// This example shows sending datagrams without waiting for responses,
// which is a common UDP pattern.
func Example_fireAndForget() {
	client, err := udp.New("localhost:8083")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer client.Close()

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Send multiple datagrams
	messages := []string{"msg1", "msg2", "msg3"}
	for _, msg := range messages {
		_, err := client.Write([]byte(msg))
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}
	}

	fmt.Printf("Sent %d messages\n", len(messages))
	// Note: This example requires a UDP server listening on localhost:8083
}

// Example_contextTimeout demonstrates context timeout handling.
//
// This example shows how to use context timeouts to limit
// operation duration.
func Example_contextTimeout() {
	client, err := udp.New("localhost:8084")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer client.Close()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			fmt.Println("Connection timeout")
			return
		}
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Connected successfully")
	// Output: Connected successfully
}

// Example_datagramSizeManagement demonstrates proper datagram sizing.
//
// This example shows how to manage datagram sizes to avoid
// IP fragmentation and ensure reliable delivery.
func Example_datagramSizeManagement() {
	client, err := udp.New("localhost:8085")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer client.Close()

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Safe datagram size (well below 1472 byte Ethernet MTU limit)
	const maxSafeSize = 1400

	data := make([]byte, maxSafeSize)
	for i := range data {
		data[i] = byte(i % 256)
	}

	n, err := client.Write(data)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Sent datagram of %d bytes (safe size)\n", n)
	// Output: Sent datagram of 1400 bytes (safe size)
}

// Example_errorHandling demonstrates comprehensive error handling.
//
// This example shows how to handle various error scenarios including
// connection errors and I/O errors.
func Example_errorHandling() {
	// Attempt to create client with invalid address
	_, err := udp.New("")
	if err != nil {
		fmt.Printf("Creation error: %v\n", err)
	}

	// Valid client
	client, err := udp.New("localhost:8086")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer client.Close()

	// Attempt operation before connection
	_, err = client.Write([]byte("test"))
	if err != nil {
		fmt.Printf("Write error before connect: %v\n", err)
	}

	// Output:
	// Creation error: invalid dial address
	// Write error before connect: invalid connection
}

// Example_stateMonitoring demonstrates client state monitoring.
//
// This example shows how to check client connection state using
// the IsConnected() method.
func Example_stateMonitoring() {
	client, err := udp.New("localhost:8087")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Check initial state
	fmt.Printf("Initially connected: %v\n", client.IsConnected())

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("After connect: %v\n", client.IsConnected())

	client.Close()
	fmt.Printf("After close: %v\n", client.IsConnected())

	// Output:
	// Initially connected: false
	// After connect: true
	// After close: false
}

// Example_multipleMessages demonstrates sending multiple messages.
//
// This example shows sending a sequence of datagrams over the same
// associated socket.
func Example_multipleMessages() {
	client, err := udp.New("localhost:8088")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer client.Close()

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Send sequence of messages
	count := 0
	for i := 0; i < 5; i++ {
		msg := fmt.Sprintf("Message %d", i)
		n, err := client.Write([]byte(msg))
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}
		if n > 0 {
			count++
		}
	}

	fmt.Printf("Successfully sent %d messages\n", count)
	// Note: This example requires a UDP server listening on localhost:8088
}

// Example_reconnection demonstrates reconnecting a client.
//
// This example shows that calling Connect() multiple times
// replaces the existing socket association.
func Example_reconnection() {
	client, err := udp.New("localhost:8089")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer client.Close()

	ctx := context.Background()

	// Initial connection
	if err := client.Connect(ctx); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("First connect: %v\n", client.IsConnected())

	// Reconnect (replaces socket)
	if err := client.Connect(ctx); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Second connect: %v\n", client.IsConnected())

	// Output:
	// First connect: true
	// Second connect: true
}

// Example_tlsNoOp demonstrates that TLS is not supported.
//
// This example shows that SetTLS() is a no-op for UDP clients
// since UDP doesn't support TLS natively.
func Example_tlsNoOp() {
	client, err := udp.New("localhost:8090")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer client.Close()

	// SetTLS is a no-op for UDP (returns nil)
	err = client.SetTLS(true, nil, "localhost")
	if err == nil {
		fmt.Println("TLS not supported (no-op, returns nil)")
	}

	// Output: TLS not supported (no-op, returns nil)
}

// Example_ipv6Address demonstrates using IPv6 addresses.
//
// This example shows how to create a client with an IPv6 address.
func Example_ipv6Address() {
	// IPv6 addresses must be enclosed in brackets
	client, err := udp.New("[::1]:8091")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer client.Close()

	fmt.Println("IPv6 client created successfully")
	// Output: IPv6 client created successfully
}

// Example_emptyWrite demonstrates writing empty datagrams.
//
// This example shows that writing empty data is allowed
// and can be used as a keepalive or ping mechanism.
func Example_emptyWrite() {
	client, err := udp.New("localhost:8092")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer client.Close()

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Write empty datagram
	n, err := client.Write([]byte{})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Sent empty datagram: %d bytes\n", n)
	// Output: Sent empty datagram: 0 bytes
}

// Example_callbackOrdering demonstrates callback execution order.
//
// This example shows the order in which callbacks are triggered
// during client lifecycle operations.
func Example_callbackOrdering() {
	client, err := udp.New("localhost:8093")
	if err != nil {
		log.Fatal(err)
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
	fmt.Printf("Events captured: %d\n", eventCount)
	// Output: Events captured: 2
}
