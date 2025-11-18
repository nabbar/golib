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

package config

import (
	"os"

	libptc "github.com/nabbar/golib/network/protocol"
	libsck "github.com/nabbar/golib/socket"
	scksrv "github.com/nabbar/golib/socket/server"
)

// ServerConfig defines the configuration for creating a socket server.
//
// This structure provides a declarative way to specify server parameters
// before instantiation. It's particularly useful when loading configuration from
// external sources (config files, environment variables, etc.) or when you need
// to validate settings before starting the server.
//
// The configuration supports all server protocols (TCP, UDP, Unix, Unixgram) with
// protocol-specific settings like file permissions for Unix domain sockets.
//
// Example usage:
//
//	// TCP server configuration
//	cfg := config.ServerConfig{
//	    Network: protocol.NetworkTCP,
//	    Address: ":8080",
//	}
//
//	handler := func(r socket.Reader, w socket.Writer) {
//	    defer r.Close()
//	    defer w.Close()
//	    io.Copy(w, r) // Echo
//	}
//
//	server, err := cfg.New(nil, handler)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer server.Close()
//
//	server.Listen(context.Background())
//
//	// Unix socket server with permissions
//	cfg := config.ServerConfig{
//	    Network:   protocol.NetworkUnix,
//	    Address:   "/tmp/app.sock",
//	    PermFile:  0600,        // Owner only
//	    GroupPerm: 1000,        // Specific group
//	}
//
// The New() method validates the configuration and returns an appropriate
// server implementation based on the network protocol.
//
// See github.com/nabbar/golib/socket/server for more server examples.
type ServerConfig struct {
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
	//   - TCP/UDP: "[host]:port" (e.g., ":8080", "0.0.0.0:8080", "localhost:9000")
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
	PermFile os.FileMode

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
}

// New creates and returns a socket server based on the configuration.
//
// This method validates the configuration and instantiates the appropriate
// server implementation based on the Network protocol. The server is ready
// to accept connections after calling Listen().
//
// Parameters:
//   - updateCon: Optional callback invoked when a new connection is established
//     (for connection-oriented protocols) or when the socket is created
//     (for datagram protocols). Can be used to set socket options. Pass nil if not needed.
//   - handler: Required function that processes each connection (TCP/Unix) or
//     all datagrams (UDP/Unixgram). Receives Reader and Writer interfaces.
//     The handler signature is: func(socket.Reader, socket.Writer)
//
// Returns:
//   - libsck.Server: A server implementation matching the configured protocol
//   - error: Configuration validation errors or instantiation failures
//
// Possible errors:
//   - Invalid or unsupported network protocol
//   - Empty or malformed address
//   - Address format mismatch with protocol (e.g., file path for TCP)
//   - Invalid group permission (>32767) for Unix sockets
//   - HandlerFunc is nil
//   - Address already in use (may not be detected until Listen())
//   - Permission denied for privileged ports (<1024) or file paths
//
// For Unix domain sockets:
//   - The socket file is created when Listen() is called
//   - PermFile and GroupPerm are applied to the socket file
//   - The file is automatically removed on shutdown
//
// The returned server must be explicitly started using Listen().
// Always call Close() or Shutdown() when done to release resources.
//
// Example:
//
//	// TCP server
//	cfg := ServerConfig{
//	    Network: protocol.NetworkTCP,
//	    Address: ":8080",
//	}
//
//	handler := func(r socket.Reader, w socket.Writer) {
//	    defer r.Close()
//	    defer w.Close()
//	    // Handle connection...
//	}
//
//	server, err := cfg.New(nil, handler)
//	if err != nil {
//	    return fmt.Errorf("create server: %w", err)
//	}
//	defer server.Close()
//
//	if err := server.Listen(context.Background()); err != nil {
//	    return fmt.Errorf("listen: %w", err)
//	}
//
//	// Unix socket with permissions
//	cfg := ServerConfig{
//	    Network:   protocol.NetworkUnix,
//	    Address:   "/tmp/app.sock",
//	    PermFile:  0600,
//	    GroupPerm: -1,
//	}
//
//	server, err := cfg.New(nil, handler)
//	// Socket file created on Listen() with specified permissions
//
// See github.com/nabbar/golib/socket/server for server usage examples.
// See github.com/nabbar/golib/socket.Server for the server interface.
// See github.com/nabbar/golib/socket.HandlerFunc for handler signature.
func (o ServerConfig) New(updateCon libsck.UpdateConn, handler libsck.HandlerFunc) (libsck.Server, error) {
	return scksrv.New(updateCon, handler, o.Network, o.Address, o.PermFile, o.GroupPerm)
}
