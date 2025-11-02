//go:build linux || darwin

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

// Package unixgram provides a UNIX domain datagram socket client implementation.
//
// This package implements the github.com/nabbar/golib/socket.Client interface
// for UNIX datagram socket connections, combining characteristics of both
// UNIX sockets and UDP:
//   - Connectionless: Like UDP, no persistent connection
//   - Message-oriented: Preserves message boundaries (datagrams)
//   - Unreliable: No guaranteed delivery or ordering
//   - Local only: Uses filesystem paths, not network addresses
//   - Fast: Kernel-space only, no network overhead
//
// Key features:
//   - Thread-safe connection management using atomic.Map
//   - Configurable error and info callbacks
//   - Context-aware operations
//   - Support for one-shot request/response operations
//   - No TLS support (not applicable to UNIX sockets)
//
// Datagram characteristics:
//   - Each write sends a complete datagram
//   - Each read receives a complete datagram
//   - No fragmentation or reassembly needed
//   - No guaranteed delivery (fire-and-forget)
//   - No guaranteed order (out-of-order delivery possible)
//
// Basic usage:
//
//	// Create a new UNIX datagram socket client
//	client := unixgram.New("/tmp/app.sock")
//	if client == nil {
//	    log.Fatal("Invalid socket path")
//	}
//	defer client.Close()
//
//	// Connect to server
//	ctx := context.Background()
//	if err := client.Connect(ctx); err != nil {
//	    log.Fatal(err)
//	}
//
//	// Send data (fire-and-forget)
//	n, err := client.Write([]byte("Hello"))
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// UNIX datagram sockets are ideal for:
//   - High-speed local IPC where reliability not critical
//   - Event logging and notifications
//   - Real-time data streaming
//   - Stateless request/response patterns
//
// See github.com/nabbar/golib/socket/client/unix for connection-oriented UNIX sockets.
// See github.com/nabbar/golib/socket/client/udp for UDP datagram sockets (network-based).
package unixgram

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

	// ErrAddress is returned by internal methods when the socket path is empty,
	// malformed, or cannot be accessed. The path must point to a valid or
	// creatable location in the filesystem (e.g., "/tmp/app.sock").
	ErrAddress = fmt.Errorf("invalid dial address")
)
