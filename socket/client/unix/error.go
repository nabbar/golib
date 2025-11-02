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

// Package unix provides a UNIX domain socket client implementation with callback mechanisms.
//
// This package implements the github.com/nabbar/golib/socket.Client interface
// for UNIX domain socket connections. UNIX sockets provide fast, reliable IPC
// (Inter-Process Communication) on the same machine:
//   - Connection-oriented (SOCK_STREAM, like TCP)
//   - Uses filesystem paths instead of network addresses
//   - No network overhead - kernel-space only communication
//   - Better performance than TCP for local communication
//   - Supports file permissions for access control
//
// Key features:
//   - Thread-safe connection management using atomic.Map
//   - Configurable error and info callbacks
//   - Context-aware operations
//   - Support for one-shot request/response operations
//   - No TLS support (not applicable to UNIX sockets)
//   - Automatic socket file cleanup
//
// Basic usage:
//
//	// Create a new UNIX socket client
//	client := unix.New("/tmp/app.sock")
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
//	// Send data
//	n, err := client.Write([]byte("Hello"))
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// UNIX sockets are ideal for:
//   - Local microservices communication
//   - Docker container communication
//   - Database connections (PostgreSQL, MySQL)
//   - System daemon IPC
//   - High-performance local RPC
//
// See github.com/nabbar/golib/socket/client/tcp for TCP client implementation.
// See github.com/nabbar/golib/socket/client/udp for UDP client implementation.
package unix

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
