//go:build !linux && !darwin

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

// Package server provides a unified factory for creating socket servers
// across different network protocols on non-Linux, non-Darwin platforms.
//
// This package serves as a convenience wrapper that creates appropriate
// server implementations based on the specified network protocol. On platforms
// other than Linux and Darwin, only network-based protocols are supported:
//   - TCP, TCP4, TCP6: Connection-oriented network servers (see github.com/nabbar/golib/socket/server/tcp)
//   - UDP, UDP4, UDP6: Connectionless datagram network servers (see github.com/nabbar/golib/socket/server/udp)
//
// Note: UNIX domain sockets (NetworkUnix, NetworkUnixGram) are not available
// on this platform and will return an error if specified.
//
// All created servers implement the github.com/nabbar/golib/socket.Server interface,
// providing a consistent API regardless of the underlying protocol.
//
// Example:
//
//	handler := func(r socket.Reader, w socket.Writer) {
//	    defer r.Close()
//	    defer w.Close()
//	    io.Copy(w, r) // Echo server
//	}
//
//	server, err := server.New(nil, handler, protocol.NetworkTCP, ":8080", 0, -1)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer server.Close()
//
//	server.Listen(context.Background())
package server

import (
	"fmt"
	"os"

	libptc "github.com/nabbar/golib/network/protocol"
	libsck "github.com/nabbar/golib/socket"
	scksrt "github.com/nabbar/golib/socket/server/tcp"
	scksru "github.com/nabbar/golib/socket/server/udp"
)

// New creates a new socket server based on the specified network protocol.
//
// This factory function instantiates the appropriate server implementation
// for the given protocol type. On platforms other than Linux and Darwin,
// only TCP and UDP protocols are supported.
//
// Parameters:
//   - upd: Optional callback function invoked for each new connection (TCP) or
//     when the socket is created (UDP). Can be used to set socket options like
//     timeouts, buffer sizes, etc. Pass nil if not needed.
//   - handler: Required function to process connections or datagrams. For
//     TCP, it's called for each connection. For UDP, it handles all incoming
//     datagrams. The signature is: func(socket.Reader, socket.Writer)
//   - proto: Network protocol from github.com/nabbar/golib/network/protocol.
//     Supported values:
//   - NetworkTCP, NetworkTCP4, NetworkTCP6: TCP servers
//   - NetworkUDP, NetworkUDP4, NetworkUDP6: UDP servers
//     Note: UNIX domain sockets are NOT supported on this platform
//   - address: Address string in "[host]:port" format (e.g., ":8080", "0.0.0.0:8080")
//   - perm: Ignored on this platform (UNIX socket permissions not applicable)
//   - gid: Ignored on this platform (UNIX socket permissions not applicable)
//
// Returns:
//   - libsck.Server: A server instance implementing the socket.Server interface
//   - error: An error if the protocol is invalid/unsupported, address validation fails,
//     or server configuration fails
//
// Example:
//
//	// TCP server
//	handler := func(r socket.Reader, w socket.Writer) {
//	    defer r.Close()
//	    defer w.Close()
//	    // Handle connection...
//	}
//
//	server, err := New(nil, handler, protocol.NetworkTCP, ":8080", 0, -1)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer server.Close()
//
//	// UNIX sockets are not available on this platform
//	_, err = New(nil, handler, protocol.NetworkUnix, "/tmp/app.sock", 0600, -1) // Returns error
func New(upd libsck.UpdateConn, handler libsck.Handler, proto libptc.NetworkProtocol, address string, perm os.FileMode, gid int32) (libsck.Server, error) {
	switch proto {
	case libptc.NetworkTCP, libptc.NetworkTCP4, libptc.NetworkTCP6:
		s := scksrt.New(upd, handler)
		e := s.RegisterServer(address)
		return s, e
	case libptc.NetworkUDP, libptc.NetworkUDP4, libptc.NetworkUDP6:
		s := scksru.New(upd, handler)
		e := s.RegisterServer(address)
		return s, e
	}

	return nil, fmt.Errorf("invalid server protocol")
}
