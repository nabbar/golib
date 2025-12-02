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

package unixgram

import (
	libatm "github.com/nabbar/golib/atomic"
	libsck "github.com/nabbar/golib/socket"
)

// ClientUnix represents a UNIX domain datagram socket client.
//
// This interface extends github.com/nabbar/golib/socket.Client and provides
// all standard socket operations for UNIX datagram socket communication:
//   - Connect(ctx) - Associate socket with UNIX path
//   - IsConnected() - Check if socket is associated
//   - Read(p []byte) - Receive a datagram
//   - Write(p []byte) - Send a datagram
//   - Close() - Close the socket
//   - Once(ctx, request, response) - One-shot datagram operation
//   - SetTLS(enable, config, serverName) - No-op for UNIX sockets (always returns nil)
//   - RegisterFuncError(f) - Register error callback
//   - RegisterFuncInfo(f) - Register connection info callback
//
// UNIX datagram socket characteristics:
//   - Connectionless: No persistent connection like TCP
//   - Message-oriented: Each write/read is a complete datagram
//   - Unreliable: No guaranteed delivery (like UDP)
//   - Unordered: No guaranteed ordering (like UDP)
//   - Local only: Cannot communicate across network
//   - Fast: Kernel-space only, no network stack
//   - File-based: Uses filesystem paths instead of IP:port
//
// Datagram size considerations:
//   - No automatic fragmentation (unlike TCP)
//   - System-dependent maximum size (typically 16KB-64KB)
//   - Smaller datagrams more reliable
//
// All operations are thread-safe and use atomic operations internally.
// The client manages connection state and cleanup automatically.
//
// Use cases:
//   - High-performance event logging
//   - Real-time metrics collection
//   - Stateless notifications
//   - When delivery guarantee not critical
//
// See github.com/nabbar/golib/socket package for interface details.
type ClientUnix interface {
	libsck.Client
}

// New creates a new UNIX domain datagram socket client for the specified socket path.
//
// The unixfile parameter must be a valid filesystem path where the UNIX datagram
// socket will exist or be created. Common locations:
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
// The client is created in a disconnected state. Use Connect() to associate
// the socket with the path. The path is stored but not validated until
// Connect() is called.
//
// UNIX datagram-specific notes:
//   - Socket file is created by the server, not the client
//   - File permissions control who can send datagrams
//   - Socket file persists after server shutdown (must be cleaned up)
//   - Uses SOCK_DGRAM mode (datagram, like UDP)
//
// Differences from UNIX stream sockets (unix package):
//   - No persistent connection (connectionless)
//   - Message boundaries preserved (datagrams)
//   - No delivery guarantee (unreliable)
//   - No ordering guarantee (unordered)
//   - Lower overhead (no connection management)
//
// Returns:
//   - ClientUnix: A new client instance if successful
//   - nil: If unixfile is empty (use this to validate input)
//
// Example:
//
//	client := unixgram.New("/tmp/app.sock")
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
//	// Send datagram (fire-and-forget)
//	data := []byte("Event: user login")
//	n, err := client.Write(data)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	// Note: No guarantee the datagram was received!
func New(unixfile string) ClientUnix {
	if len(unixfile) < 1 {
		return nil
	}

	c := &cli{m: libatm.NewMapAny[uint8]()}
	c.m.Store(keyNetAddr, unixfile)

	return c
}
