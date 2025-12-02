//go:build darwin

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

// Package server provides a unified factory for creating socket servers
// across different network protocols on Darwin (macOS) platforms.
//
// This package serves as a convenience wrapper that creates appropriate
// server implementations based on the specified network protocol:
//   - TCP, TCP4, TCP6: Connection-oriented network servers (see github.com/nabbar/golib/socket/server/tcp)
//   - UDP, UDP4, UDP6: Connectionless datagram network servers (see github.com/nabbar/golib/socket/server/udp)
//   - Unix: Connection-oriented UNIX domain socket servers (see github.com/nabbar/golib/socket/server/unix)
//   - UnixGram: Connectionless UNIX datagram socket servers (see github.com/nabbar/golib/socket/server/unixgram)
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
	libptc "github.com/nabbar/golib/network/protocol"
	librun "github.com/nabbar/golib/runner"
	libsck "github.com/nabbar/golib/socket"
	sckcfg "github.com/nabbar/golib/socket/config"
	scksrt "github.com/nabbar/golib/socket/server/tcp"
	scksru "github.com/nabbar/golib/socket/server/udp"
	scksrx "github.com/nabbar/golib/socket/server/unix"
	sckgrm "github.com/nabbar/golib/socket/server/unixgram"
)

// New creates a new socket server based on the specified network protocol.
//
// This factory function instantiates the appropriate server implementation
// for the given protocol type. On Darwin (macOS) platforms, all protocol types
// are supported, including UNIX domain sockets.
//
// Parameters:
//   - upd: Optional callback function invoked for each new connection (TCP/Unix)
//     or when the socket is created (UDP/Unixgram). Can be used to set socket
//     options like timeouts, buffer sizes, etc. Pass nil if not needed.
//   - handler: Required function to process connections or datagrams. For
//     connection-oriented protocols (TCP/Unix), it's called for each connection.
//     For datagram protocols (UDP/Unixgram), it handles all incoming datagrams.
//     The signature is: func(socket.Reader, socket.Writer)
//   - proto: Network protocol from github.com/nabbar/golib/network/protocol.
//     Supported values on Darwin:
//   - NetworkTCP, NetworkTCP4, NetworkTCP6: TCP servers
//   - NetworkUDP, NetworkUDP4, NetworkUDP6: UDP servers
//   - NetworkUnix: UNIX domain stream socket servers
//   - NetworkUnixGram: UNIX domain datagram socket servers
//   - address: Protocol-specific address string:
//   - TCP/UDP: "[host]:port" format (e.g., ":8080", "0.0.0.0:8080", "localhost:9000")
//   - UNIX: filesystem path (e.g., "/tmp/app.sock", "/var/run/app.sock")
//   - perm: File permissions for UNIX socket files (e.g., 0600, 0660, 0666).
//     Only applies to NetworkUnix and NetworkUnixGram. Ignored for TCP/UDP.
//     If set to 0, default permissions (0770) are applied.
//   - gid: Group ID for UNIX socket file ownership. Only applies to NetworkUnix
//     and NetworkUnixGram. Use -1 for the process's current group, or specify
//     a group ID (0-32767). Ignored for TCP/UDP.
//
// Returns:
//   - libsck.Server: A server instance implementing the socket.Server interface
//   - error: An error if the protocol is invalid, address validation fails,
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
//
//	// UNIX socket server with permissions
//	unixServer, err := New(nil, handler, protocol.NetworkUnix, "/tmp/app.sock", 0600, -1)
//	if err != nil {
//	    log.Fatal(err)
//	}
func New(upd libsck.UpdateConn, handler libsck.HandlerFunc, cfg sckcfg.Server) (libsck.Server, error) {
	defer func() {
		if r := recover(); r != nil {
			librun.RecoveryCaller("golib/socket/server", r)
		}
	}()

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	switch cfg.Network {
	case libptc.NetworkUnix:
		return scksrx.New(upd, handler, cfg)
	case libptc.NetworkUnixGram:
		return sckgrm.New(upd, handler, cfg)
	case libptc.NetworkTCP, libptc.NetworkTCP4, libptc.NetworkTCP6:
		return scksrt.New(upd, handler, cfg)
	case libptc.NetworkUDP, libptc.NetworkUDP4, libptc.NetworkUDP6:
		return scksru.New(upd, handler, cfg)
	default:
		return nil, sckcfg.ErrInvalidProtocol
	}
}
