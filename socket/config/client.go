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

// Package config provides configuration structures for creating socket clients and servers.
//
// This package offers a declarative configuration approach for socket connections,
// allowing you to define client and server settings before instantiation.
//
// It supports all socket types through the NetworkProtocol interface:
//   - TCP: Connection-oriented, reliable network sockets
//   - UDP: Connectionless, fast network sockets
//   - Unix: Connection-oriented IPC via filesystem sockets
//   - Unixgram: Connectionless IPC via filesystem sockets
//
// The configuration structs are typically used in scenarios where socket parameters
// are loaded from external sources (config files, environment variables, etc.)
// and need to be validated and instantiated separately.
//
// See github.com/nabbar/golib/socket/client for client implementations.
// See github.com/nabbar/golib/socket/server for server implementations.
// See github.com/nabbar/golib/network/protocol for supported protocols.
package config

import (
	libptc "github.com/nabbar/golib/network/protocol"
	libsck "github.com/nabbar/golib/socket"
	sckclt "github.com/nabbar/golib/socket/client"
)

// ClientConfig defines the configuration for creating a socket client.
//
// This structure provides a declarative way to specify client connection parameters
// before instantiation. It's particularly useful when loading configuration from
// external sources or when you need to validate settings before connecting.
//
// Example usage:
//
//	// TCP client configuration
//	cfg := config.ClientConfig{
//	    Network: protocol.NetworkTCP,
//	    Address: "localhost:8080",
//	}
//	client, err := cfg.New()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Close()
//
//	// Unix socket client configuration
//	cfg := config.ClientConfig{
//	    Network: protocol.NetworkUnix,
//	    Address: "/tmp/app.sock",
//	}
//
// The New() method validates the configuration and returns an appropriate
// client implementation based on the network protocol.
//
// See github.com/nabbar/golib/socket/client for more client examples.
type ClientConfig struct {
	// Network specifies the transport protocol to use for the connection.
	//
	// Supported values:
	//   - NetworkTCP: TCP/IP network socket (e.g., "localhost:8080", "192.168.1.1:9000")
	//   - NetworkUDP: UDP/IP network socket (e.g., "localhost:8080", "192.168.1.1:9000")
	//   - NetworkUnix: Unix domain stream socket (e.g., "/tmp/app.sock", "./socket")
	//   - NetworkUnixGram: Unix domain datagram socket (e.g., "/tmp/app.sock")
	//
	// The protocol determines both the transport layer and the addressing scheme.
	// See github.com/nabbar/golib/network/protocol for protocol definitions.
	Network libptc.NetworkProtocol

	// Address specifies the destination to connect to.
	//
	// Format depends on the Network protocol:
	//   - TCP/UDP: "host:port" (e.g., "localhost:8080", "192.168.1.1:9000")
	//   - Unix/Unixgram: file path (e.g., "/tmp/app.sock", "./socket")
	//
	// For network protocols (TCP/UDP):
	//   - Use "host:port" format
	//   - Host can be hostname, IPv4, or IPv6 address
	//   - Port must be in range 1-65535
	//
	// For Unix domain sockets:
	//   - Use absolute or relative file path
	//   - Path must be accessible (read/write permissions)
	//   - Maximum path length depends on OS (typically 108 bytes)
	//
	// Empty address will cause New() to return an error.
	Address string
}

// New creates and returns a socket client based on the configuration.
//
// This method validates the configuration and instantiates the appropriate
// client implementation based on the Network protocol. The returned client
// is ready to use but not yet connected.
//
// Returns:
//   - libsck.Client: A client implementation matching the configured protocol
//   - error: Configuration validation errors or instantiation failures
//
// Possible errors:
//   - Invalid or unsupported network protocol
//   - Empty or malformed address
//   - Address format mismatch with protocol (e.g., file path for TCP)
//   - Resource allocation failures
//
// The returned client must be explicitly connected using Connect() before use.
// Always call Close() when done to release resources.
//
// Example:
//
//	cfg := ClientConfig{
//	    Network: protocol.NetworkTCP,
//	    Address: "localhost:8080",
//	}
//
//	client, err := cfg.New()
//	if err != nil {
//	    return fmt.Errorf("create client: %w", err)
//	}
//	defer client.Close()
//
//	if err := client.Connect(ctx); err != nil {
//	    return fmt.Errorf("connect: %w", err)
//	}
//
// See github.com/nabbar/golib/socket/client for client usage examples.
// See github.com/nabbar/golib/socket.Client for the client interface.
func (o ClientConfig) New() (libsck.Client, error) {
	return sckclt.New(o.Network, o.Address)
}
