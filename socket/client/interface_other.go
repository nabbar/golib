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

// Package client provides a unified factory for creating socket clients
// across different network protocols on non-Linux, non-Darwin platforms.
//
// This package serves as a convenience wrapper that creates appropriate
// client implementations based on the specified network protocol. On platforms
// other than Linux and Darwin, only network-based protocols are supported:
//   - TCP, TCP4, TCP6: Connection-oriented network sockets (see github.com/nabbar/golib/socket/client/tcp)
//   - UDP, UDP4, UDP6: Connectionless datagram network sockets (see github.com/nabbar/golib/socket/client/udp)
//
// Note: UNIX domain sockets (NetworkUnix, NetworkUnixGram) are not available
// on this platform and will return an error if specified.
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
)

// New creates a new socket client based on the specified network protocol.
//
// This factory function instantiates the appropriate client implementation
// for the given protocol type. On platforms other than Linux and Darwin,
// only TCP and UDP protocols are supported.
//
// Parameters:
//   - proto: Network protocol from github.com/nabbar/golib/network/protocol package.
//     Supported values:
//   - NetworkTCP, NetworkTCP4, NetworkTCP6: TCP clients
//   - NetworkUDP, NetworkUDP4, NetworkUDP6: UDP clients
//     Note: UNIX domain sockets are NOT supported on this platform
//   - address: Protocol-specific address string in "host:port" format
//     (e.g., "localhost:8080", "192.168.1.1:9000")
//
// Returns:
//   - libsck.Client: A client instance implementing the socket.Client interface
//   - error: An error if the protocol is invalid/unsupported or address validation fails
//
// Example:
//
//	// Create TCP client
//	client, err := New(protocol.NetworkTCP, "localhost:8080")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Close()
//
//	// UNIX sockets are not available
//	_, err = New(protocol.NetworkUnix, "/tmp/app.sock") // Returns error
func New(proto libptc.NetworkProtocol, address string) (libsck.Client, error) {
	switch proto {
	case libptc.NetworkTCP, libptc.NetworkTCP4, libptc.NetworkTCP6:
		return sckclt.New(address)
	case libptc.NetworkUDP, libptc.NetworkUDP4, libptc.NetworkUDP6:
		return sckclu.New(address)
	default:
		return nil, fmt.Errorf("invalid client protocol")
	}
}
