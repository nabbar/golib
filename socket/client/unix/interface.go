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

package unix

import (
	libatm "github.com/nabbar/golib/atomic"
	libsck "github.com/nabbar/golib/socket"
)

// ClientUnix represents a UNIX domain socket client that implements the socket.Client interface.
//
// This interface extends github.com/nabbar/golib/socket.Client and provides
// all standard socket operations for UNIX domain socket communication:
//   - Connect(ctx) - Establish connection to UNIX socket
//   - IsConnected() - Check if socket is connected
//   - Read(p []byte) - Read data from socket
//   - Write(p []byte) - Write data to socket
//   - Close() - Close the socket connection
//   - Once(ctx, request, response) - One-shot request/response operation
//   - SetTLS(enable, config, serverName) - No-op for UNIX sockets (always returns nil)
//   - RegisterFuncError(f) - Register error callback
//   - RegisterFuncInfo(f) - Register connection info callback
//
// UNIX socket characteristics:
//   - Connection-oriented: Reliable, ordered, bidirectional like TCP
//   - Local only: Cannot communicate across network
//   - Filesystem-based: Uses file paths instead of IP:port
//   - Fast: No network stack overhead, kernel-space only
//   - Secure: File permissions control access
//   - No fragmentation: Unlike UDP, no message size limits
//
// All operations are thread-safe and use atomic operations internally.
// The client manages connection state and cleanup automatically.
//
// Use cases:
//   - Docker container to host communication
//   - Microservices on same machine
//   - Database client connections (PostgreSQL, MySQL)
//   - System daemon control
//
// See github.com/nabbar/golib/socket package for interface details.
type ClientUnix interface {
	libsck.Client
}

// New creates a new UNIX domain socket client for the specified socket path.
//
// The unixfile parameter must be a valid filesystem path where the UNIX socket
// will exist or be created. Common locations:
//   - /tmp/app.sock - Temporary sockets
//   - /var/run/app.sock - System daemon sockets
//   - /run/user/$UID/app.sock - User-specific sockets
//   - ./app.sock - Relative path
//
// Path requirements:
//   - Must not be empty
//   - Maximum length typically 108 bytes (Linux UNIX_PATH_MAX)
//   - Parent directory must exist and be accessible
//   - No network address syntax (no IP:port)
//
// The client is created in a disconnected state. Use Connect() to establish
// the connection to the socket. The path is stored but not validated until
// Connect() is called.
//
// UNIX socket-specific notes:
//   - Socket file is created by the server, not the client
//   - File permissions control who can connect
//   - Socket file persists after server shutdown (must be cleaned up)
//   - Supports both stream (SOCK_STREAM) and datagram (SOCK_DGRAM) modes
//     (this implementation uses SOCK_STREAM)
//
// Returns:
//   - ClientUnix: A new client instance if successful
//   - nil: If unixfile is empty (use this to validate input)
//
// Example:
//
//	client := unix.New("/tmp/app.sock")
//	if client == nil {
//	    log.Fatal("Invalid socket path")
//	}
//	defer client.Close()
//
//	ctx := context.Background()
//	if err := client.Connect(ctx); err != nil {
//	    log.Fatal(err)
//	}
//
//	// Use client for read/write operations
//	data := []byte("Hello, UNIX!")
//	n, err := client.Write(data)
//	if err != nil {
//	    log.Fatal(err)
//	}
func New(unixfile string) ClientUnix {
	if len(unixfile) < 1 {
		return nil
	}

	c := &cli{m: libatm.NewMapAny[uint8]()}
	c.m.Store(keyNetAddr, unixfile)

	return c
}
