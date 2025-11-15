//go:build darwin

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

// Package client provides a unified factory for creating socket clients
// across different network protocols on Darwin platforms.
//
// This package serves as a convenience wrapper that creates appropriate
// client implementations based on the specified network protocol:
//   - TCP, TCP4, TCP6: Connection-oriented network sockets (see github.com/nabbar/golib/socket/client/tcp)
//   - UDP, UDP4, UDP6: Connectionless datagram network sockets (see github.com/nabbar/golib/socket/client/udp)
//   - Unix: Connection-oriented UNIX domain sockets (see github.com/nabbar/golib/socket/client/unix)
//   - UnixGram: Connectionless UNIX datagram sockets (see github.com/nabbar/golib/socket/client/unixgram)
//
// All created clients implement the github.com/nabbar/golib/socket.Client interface,
// providing a consistent API regardless of the underlying protocol.
//
// Example:
//
//	client, err := client.New(protocol.NetworkTCP, "localhost:8080")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Close()
package client

import (
	"fmt"

	libptc "github.com/nabbar/golib/network/protocol"
	libsck "github.com/nabbar/golib/socket"
	sckclt "github.com/nabbar/golib/socket/client/tcp"
	sckclu "github.com/nabbar/golib/socket/client/udp"
	sckclx "github.com/nabbar/golib/socket/client/unix"
	sckgrm "github.com/nabbar/golib/socket/client/unixgram"
)

// New creates a new socket client based on the specified network protocol.
//
// This factory function instantiates the appropriate client implementation
// for the given protocol type. On Darwin (macOS) platforms, all protocol
// types are supported.
//
// Parameters:
//   - proto: Network protocol from github.com/nabbar/golib/network/protocol package.
//     Supported values:
//   - NetworkTCP, NetworkTCP4, NetworkTCP6: TCP clients
//   - NetworkUDP, NetworkUDP4, NetworkUDP6: UDP clients
//   - NetworkUnix: UNIX domain stream socket clients
//   - NetworkUnixGram: UNIX domain datagram socket clients
//   - address: Protocol-specific address string:
//   - TCP/UDP: "host:port" format (e.g., "localhost:8080", "192.168.1.1:9000")
//   - UNIX: filesystem path (e.g., "/tmp/app.sock")
//
// Returns:
//   - libsck.Client: A client instance implementing the socket.Client interface
//   - error: An error if the protocol is invalid or address validation fails
//
// Example:
//
//	// Create TCP client
//	client, err := New(protocol.NetworkTCP, "localhost:8080")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Create UNIX socket client
//	unixClient, err := New(protocol.NetworkUnix, "/tmp/app.sock")
//	if err != nil {
//	    log.Fatal(err)
//	}
func New(proto libptc.NetworkProtocol, address string) (libsck.Client, error) {
	switch proto {
	case libptc.NetworkUnix:
		return sckclx.New(address), nil
	case libptc.NetworkUnixGram:
		return sckgrm.New(address), nil
	case libptc.NetworkTCP, libptc.NetworkTCP4, libptc.NetworkTCP6:
		return sckclt.New(address)
	case libptc.NetworkUDP, libptc.NetworkUDP4, libptc.NetworkUDP6:
		return sckclu.New(address)
	default:
		return nil, fmt.Errorf("invalid client protocol")
	}
}
