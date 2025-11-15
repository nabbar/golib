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

// Package udp provides a UDP client implementation with callback mechanisms.
//
// This package implements the github.com/nabbar/golib/socket.Client interface
// for UDP network connections. UDP is a connectionless protocol, meaning:
//   - No persistent connection state is maintained
//   - Datagrams may be lost, duplicated, or arrive out of order
//   - No automatic retransmission or flow control
//   - Lower overhead than TCP, suitable for real-time applications
//
// Key features:
//   - Thread-safe connectionless datagram communication
//   - Thread-safe state management using atomic.Map
//   - Configurable error and info callbacks
//   - Context-aware operations
//   - Support for one-shot request/response operations
//   - No TLS support (UDP doesn't support TLS natively)
//
// Basic usage:
//
//	// Create a new UDP client
//	client, err := udp.New("localhost:8080")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Close()
//
//	// Connect to server (prepares the UDP socket)
//	ctx := context.Background()
//	if err := client.Connect(ctx); err != nil {
//	    log.Fatal(err)
//	}
//
//	// Send datagram
//	n, err := client.Write([]byte("Hello"))
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Note: UDP "Connect" doesn't establish a connection like TCP. It merely
// associates the socket with a remote address for subsequent operations.
//
// See github.com/nabbar/golib/socket/client/tcp for TCP client implementation.
package udp

import "fmt"

var (
	// ErrInstance is returned when a nil client instance is used for operations.
	// This typically indicates a programming error where a method is called on
	// a nil pointer or an uninitialized client.
	ErrInstance = fmt.Errorf("invalid instance")

	// ErrConnection is returned when attempting to perform I/O operations
	// on a client that hasn't called Connect(), or when the underlying socket
	// is nil or invalid. Call Connect() before performing operations.
	ErrConnection = fmt.Errorf("invalid connection")

	// ErrAddress is returned by New() when the provided address is empty,
	// malformed, or cannot be resolved as a valid UDP address. The address
	// must be in the format "host:port" (e.g., "localhost:8080", "192.168.1.1:9000").
	ErrAddress = fmt.Errorf("invalid dial address")
)
