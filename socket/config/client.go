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
	"net"
	"runtime"

	libtls "github.com/nabbar/golib/certificates"
	libptc "github.com/nabbar/golib/network/protocol"
)

// Client defines the configuration for creating a socket client.
//
// This structure provides a declarative way to specify client connection parameters
// before instantiation. It's particularly useful when loading configuration from
// external sources or when you need to validate settings before connecting.
//
// Example usage:
//
//	// TCP client configuration
//	cfg := config.Client{
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
//	cfg := config.Client{
//	    Network: protocol.NetworkUnix,
//	    Address: "/tmp/app.sock",
//	}
//
// The New() method validates the configuration and returns an appropriate
// client implementation based on the network protocol.
//
// See github.com/nabbar/golib/socket/client for more client examples.
type Client struct {
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

	// TLS provides Transport Layer Security configuration for the client.
	//
	// TLS is only supported for TCP-based protocols (NetworkTCP, NetworkTCP4, NetworkTCP6).
	// Attempting to enable TLS for other protocols will cause Validate() to return ErrInvalidTLSConfig.
	//
	// Fields:
	//   - Enabled: Set to true to enable TLS/SSL encryption
	//   - Config: Certificate configuration from github.com/nabbar/golib/certificates
	//   - ServerName: Server hostname for certificate validation (required when Enabled is true)
	//
	// Example:
	//   cfg := config.Client{
	//       Network: protocol.NetworkTCP,
	//       Address: "secure.example.com:443",
	//       TLS: struct{
	//           Enabled: true,
	//           Config: tlsConfig,
	//           ServerName: "secure.example.com",
	//       },
	//   }
	//
	// The Config must provide valid certificates and the ServerName must match
	// the server's certificate for successful validation.
	TLS struct {
		Enabled    bool
		Config     libtls.Config
		ServerName string
	}

	// defTls holds the default TLS configuration set via DefaultTLS().
	// This is merged with TLS.Config when GetTLS() is called.
	defTls libtls.TLSConfig
}

// Validate checks the client configuration for correctness and compatibility.
//
// This method performs several validation checks:
//   - Verifies that the network protocol is supported
//   - Validates address format for the specified protocol
//   - Checks platform compatibility (Unix sockets not supported on Windows)
//   - Validates TLS configuration if enabled
//
// Returns an error if:
//   - The protocol is unsupported (returns ErrInvalidProtocol)
//   - The address format is invalid for the protocol
//   - Unix sockets are used on Windows (returns ErrInvalidProtocol)
//   - TLS is enabled but improperly configured (returns ErrInvalidTLSConfig)
//   - TLS is enabled for non-TCP protocols (returns ErrInvalidTLSConfig)
//
// TLS-specific validation ensures:
//   - Config.New() returns a valid TLS configuration
//   - ServerName is specified
//   - Config.TLS(ServerName) returns a valid tls.Config
//
// Example:
//
//	cfg := config.Client{Network: protocol.NetworkTCP, Address: "localhost:8080"}
//	if err := cfg.Validate(); err != nil {
//	    log.Fatal("Invalid configuration:", err)
//	}
func (o *Client) Validate() error {
	switch o.Network {
	case libptc.NetworkUnix:
		if runtime.GOOS == "windows" {
			return ErrInvalidProtocol
		}
	case libptc.NetworkUnixGram:
		if runtime.GOOS == "windows" {
			return ErrInvalidProtocol
		}
	case libptc.NetworkTCP, libptc.NetworkTCP4, libptc.NetworkTCP6:
		if _, err := net.ResolveTCPAddr(libptc.NetworkTCP.Code(), o.Address); err != nil {
			return err
		}
	case libptc.NetworkUDP, libptc.NetworkUDP4, libptc.NetworkUDP6:
		if _, err := net.ResolveUDPAddr(libptc.NetworkUDP.Code(), o.Address); err != nil {
			return err
		}
	default:
		return ErrInvalidProtocol
	}

	if !o.TLS.Enabled {
		return nil
	}

	switch o.Network {
	case libptc.NetworkTCP, libptc.NetworkTCP4, libptc.NetworkTCP6:
		if len(o.TLS.ServerName) < 1 {
			return ErrInvalidTLSConfig
		} else {
			return nil
		}
	default:
		return ErrInvalidTLSConfig
	}
}

// DefaultTLS sets a default TLS configuration that will be merged with TLS.Config.
//
// This method is useful when you want to provide fallback or base TLS settings
// that will be combined with the specific configuration in TLS.Config.
//
// The provided configuration will be stored and used by GetTLS() to create
// the final TLS configuration via Config.NewFrom().
//
// Parameters:
//   - t: Base TLS configuration from github.com/nabbar/golib/certificates
//
// Example:
//
//	srv := &config.Server{...}
//	srv.DefaultTLS(baseTLSConfig)
//	// Later, GetTLS() will merge TLS.Config with baseTLSConfig
//
// See GetTLS() for how this default is applied.
func (o *Client) DefaultTLS(t libtls.TLSConfig) {
	o.defTls = t
}

// GetTLS returns the TLS configuration with defaults applied.
//
// This method checks if TLS is enabled and returns the merged TLS configuration
// by combining TLS.Config with the default set via DefaultTLS().
//
// Returns:
//   - bool: true if TLS is enabled, false otherwise
//   - TLSConfig: The merged TLS configuration, or nil if TLS is disabled
//
// The returned configuration is created via Config.NewFrom(defTls), which
// merges the specific TLS.Config settings with the default configuration.
//
// Example:
//
//	if enabled, tlsConfig := srv.GetTLS(); enabled {
//	    // Use tlsConfig for TLS connections
//	}
//
// See DefaultTLS() for setting the default configuration.
func (o *Client) GetTLS() (bool, libtls.TLSConfig, string) {
	if !o.TLS.Enabled {
		return false, nil, ""
	}
	return true, o.TLS.Config.NewFrom(o.defTls), o.TLS.ServerName
}
