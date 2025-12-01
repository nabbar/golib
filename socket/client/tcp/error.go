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

// Package tcp provides a TCP client implementation with TLS support and callback mechanisms.
//
// This package implements the github.com/nabbar/golib/socket.Client interface
// for TCP network connections. It supports both plain TCP and TLS-encrypted connections,
// provides connection state callbacks, error handling callbacks, and maintains
// connection state using thread-safe atomic operations via github.com/nabbar/golib/atomic.
//
// Key features:
//   - Plain TCP and TLS-encrypted connections
//   - Thread-safe connection management using atomic.Map
//   - Configurable error and info callbacks
//   - Context-aware connection operations
//   - Support for one-shot request/response operations
//
// Basic usage:
//
//	// Create a new TCP client
//	client, err := tcp.New("localhost:8080")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Close()
//
//	// Connect to server
//	ctx := context.Background()
//	if err := client.Connect(ctx); err != nil {
//	    log.Fatal(err)
//	}
//
//	// Write data
//	n, err := client.Write([]byte("Hello"))
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// For TLS connections, see SetTLS method documentation.
package tcp

import "fmt"

var (
	// ErrInstance is returned when a nil client instance is used for operations.
	// This typically indicates a programming error where a method is called on
	// a nil pointer or an uninitialized client.
	ErrInstance = fmt.Errorf("invalid instance")

	// ErrConnection is returned when attempting to perform I/O operations
	// on a client that is not connected, or when the underlying connection
	// is nil or invalid. Check IsConnected() before performing operations.
	ErrConnection = fmt.Errorf("invalid connection")

	// ErrAddress is returned by New() when the provided dial address is empty,
	// malformed, or cannot be resolved as a valid TCP address. The address
	// must be in the format "host:port" (e.g., "localhost:8080", "192.168.1.1:9000").
	ErrAddress = fmt.Errorf("invalid dial address")
)
