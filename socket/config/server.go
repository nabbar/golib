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

package config

import (
	"net"
	"runtime"
	"time"

	libtls "github.com/nabbar/golib/certificates"
	libprm "github.com/nabbar/golib/file/perm"
	libptc "github.com/nabbar/golib/network/protocol"
)

// Server defines the configuration for creating a socket server.
//
// This structure provides a declarative way to specify server parameters
// before instantiation. It's particularly useful when loading configuration
// from external sources or when you need to validate settings before starting.
//
// The server supports all socket types through the NetworkProtocol interface:
//   - TCP: Connection-oriented network server with multiple concurrent clients
//   - UDP: Connectionless network server with datagram handling
//   - Unix: Connection-oriented IPC server via filesystem sockets
//   - Unixgram: Connectionless IPC server via filesystem sockets
//
// Example TCP server:
//
//	cfg := config.Server{
//	    Network: protocol.NetworkTCP,
//	    Address: ":8080",
//	}
//	if err := cfg.Validate(); err != nil {
//	    log.Fatal(err)
//	}
//
// Example Unix socket server with permissions:
//
//	cfg := config.Server{
//	    Network: protocol.NetworkUnix,
//	    Address: "/tmp/app.sock",
//	    PermFile: 0660,
//	    GroupPerm: 1000,
//	}
//
// See github.com/nabbar/golib/socket/server for server implementations.
type Server struct {
	// Network specifies the transport protocol for the server.
	//
	// Supported values:
	//   - NetworkTCP: TCP/IP server (e.g., ":8080", "0.0.0.0:8080")
	//   - NetworkUDP: UDP/IP server (e.g., ":8080", "0.0.0.0:8080")
	//   - NetworkUnix: Unix domain stream server (e.g., "/tmp/app.sock")
	//   - NetworkUnixGram: Unix domain datagram server (e.g., "/tmp/app.sock")
	//
	// The protocol determines both the transport layer and addressing:
	//   - TCP: Connection-oriented, reliable, multiple concurrent clients
	//   - UDP: Connectionless, fast, stateless datagram handling
	//   - Unix: IPC stream sockets, connection-oriented, file permissions
	//   - Unixgram: IPC datagram sockets, connectionless, file permissions
	//
	// See github.com/nabbar/golib/network/protocol for protocol definitions.
	// See github.com/nabbar/golib/socket/server for implementation details.
	Network libptc.NetworkProtocol

	// Address specifies where the server should listen.
	//
	// Format depends on the Network protocol:
	//   - TCP/UDP: "[host]:port" (e.g., "0:8080", "0.0.0.0:8080", "localhost:9000")
	//   - Unix/Unixgram: file path (e.g., "/tmp/app.sock", "./socket")
	//
	// For network protocols (TCP/UDP):
	//   - Use ":port" to listen on all interfaces
	//   - Use "host:port" to listen on specific interface
	//   - Port must be in range 1-65535
	//   - Ports < 1024 require elevated privileges
	//
	// For Unix domain sockets:
	//   - Use absolute or relative file path
	//   - File must not exist (will be created)
	//   - Directory must be writable
	//   - File is removed on server shutdown
	//   - Maximum path length depends on OS (typically 108 bytes)
	//
	// Empty address will cause New() to return an error.
	Address string

	// PermFile specifies file permissions for Unix domain socket files.
	//
	// This field is only used for Unix and Unixgram protocols and is ignored
	// for TCP/UDP servers.
	//
	// Common permission values:
	//   - 0600: Owner read/write only (most secure)
	//   - 0660: Owner and group read/write
	//   - 0666: All users read/write (least secure, not recommended)
	//
	// The permissions control who can connect to the socket:
	//   - Read permission: Required to connect
	//   - Write permission: Required to send data
	//
	// If set to 0 (zero), a default permission of 0770 is applied.
	//
	// Example:
	//   PermFile: 0600  // Only process owner can connect
	//   PermFile: 0660  // Owner and group members can connect
	//
	// See os.FileMode for permission representation.
	PermFile libprm.Perm

	// GroupPerm specifies the group ownership for Unix domain socket files.
	//
	// This field is only used for Unix and Unixgram protocols and is ignored
	// for TCP/UDP servers.
	//
	// The value is a numeric group ID (GID) that will own the socket file.
	// This allows group-based access control in combination with PermFile.
	//
	// Special values:
	//   - -1: Use the process's current group (default)
	//   - 0-32767: Specific group ID
	//   - >32767: Will cause New() to return ErrInvalidGroup
	//
	// The process must have permission to change the group ownership,
	// either by:
	//   - Running as root
	//   - Being a member of the target group
	//
	// Example:
	//   GroupPerm: -1    // Use current process group
	//   GroupPerm: 1000  // Set to group 1000
	//
	// Combined with PermFile 0660, this enables group-based access control.
	GroupPerm int32

	// ConIdleTimeout specifies the maximum duration a connection can remain idle.
	//
	// This field is only used for connection-oriented protocols (TCP, Unix).
	// It is ignored for connectionless protocols (UDP, Unixgram).
	//
	// When set to a positive duration:
	//   - Connections with no activity for this duration will be closed
	//   - Helps prevent resource exhaustion from stale connections
	//   - Each connection has its own independent timeout
	//
	// Special values:
	//   - 0: No timeout, connections remain open indefinitely (default)
	//   - Negative: Invalid, treated as 0
	//
	// Example:
	//   ConIdleTimeout: 5 * time.Minute  // Close idle connections after 5 minutes
	//   ConIdleTimeout: 0                // Never timeout
	//
	// Note: This timeout is independent of read/write deadlines that may be
	// set on individual operations.
	ConIdleTimeout time.Duration

	// TLS provides Transport Layer Security configuration for the server.
	//
	// TLS is only supported for TCP-based protocols (NetworkTCP, NetworkTCP4, NetworkTCP6).
	// Attempting to enable TLS for other protocols will cause Validate() to return ErrInvalidTLSConfig.
	//
	// Fields:
	//   - Enable: Set to true to enable TLS/SSL encryption
	//   - Config: Certificate configuration from github.com/nabbar/golib/certificates
	//
	// Example:
	//   cfg := config.Server{
	//       Network: protocol.NetworkTCP,
	//       Address: ":8443",
	//       TLS: struct{
	//           Enable: true,
	//           Config: tlsConfig,
	//       },
	//   }
	//
	// When TLS is enabled:
	//   - Config must provide at least one valid certificate pair
	//   - All client connections will use TLS encryption
	//   - Clients must use TLS to connect
	//
	// Use DefaultTLS() to set a fallback TLS configuration that will be used
	// if Config doesn't provide all necessary settings.
	TLS struct {
		Enable bool
		Config libtls.Config
	}

	// defTls holds the default TLS configuration set via DefaultTLS().
	// This is merged with TLS.Config when GetTLS() is called.
	defTls libtls.TLSConfig
}

// Validate checks the server configuration for correctness and compatibility.
//
// This method performs several validation checks:
//   - Verifies that the network protocol is supported
//   - Validates address format for the specified protocol
//   - Checks platform compatibility (Unix sockets not supported on Windows)
//   - Validates group permissions for Unix sockets
//   - Validates TLS configuration if enabled
//
// Returns an error if:
//   - The protocol is unsupported (returns ErrInvalidProtocol)
//   - The address format is invalid for the protocol
//   - Unix sockets are used on Windows (returns ErrInvalidProtocol)
//   - GroupPerm exceeds MaxGID (returns ErrInvalidGroup)
//   - TLS is enabled but improperly configured (returns ErrInvalidTLSConfig)
//   - TLS is enabled for non-TCP protocols (returns ErrInvalidTLSConfig)
//
// TLS-specific validation ensures:
//   - Config.New() returns a valid TLS configuration
//   - Config.LenCertificatePair() returns at least 1 certificate pair
//
// Example:
//
//	cfg := config.Server{Network: protocol.NetworkTCP, Address: ":8080"}
//	if err := cfg.Validate(); err != nil {
//	    log.Fatal("Invalid configuration:", err)
//	}
func (o *Server) Validate() error {
	switch o.Network {
	case libptc.NetworkUnix:
		if runtime.GOOS == "windows" {
			return ErrInvalidProtocol
		} else if o.GroupPerm > MaxGID {
			return ErrInvalidGroup
		}
	case libptc.NetworkUnixGram:
		if runtime.GOOS == "windows" {
			return ErrInvalidProtocol
		} else if o.GroupPerm > MaxGID {
			return ErrInvalidGroup
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

	if !o.TLS.Enable {
		return nil
	}

	switch o.Network {
	case libptc.NetworkTCP, libptc.NetworkTCP4, libptc.NetworkTCP6:
		c := o.TLS.Config.New()
		if c.LenCertificatePair() < 1 {
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
func (o *Server) DefaultTLS(t libtls.TLSConfig) {
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
func (o *Server) GetTLS() (bool, libtls.TLSConfig) {
	if !o.TLS.Enable {
		return false, nil
	}
	return true, o.TLS.Config.NewFrom(o.defTls)
}
