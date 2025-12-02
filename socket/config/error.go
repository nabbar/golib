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
	"errors"
)

// MaxGID defines the maximum allowed Unix group ID for socket file ownership.
//
// This value represents the upper limit for the Server.GroupPerm field.
// Group IDs above this threshold will cause Server.Validate() to return ErrInvalidGroup.
//
// The value 32767 is chosen as a conservative limit that is compatible with most
// Unix-like systems, where traditional group IDs are typically stored as signed 16-bit integers.
const MaxGID = 32767

var (
	// ErrInvalidProtocol indicates that an unsupported or invalid network protocol was specified.
	//
	// This error is returned when:
	//   - The Network field contains an unrecognized protocol
	//   - Unix domain sockets are used on Windows (not supported)
	//   - The protocol is incompatible with the requested operation
	//
	// Valid protocols are defined in github.com/nabbar/golib/network/protocol:
	//   - NetworkTCP, NetworkTCP4, NetworkTCP6
	//   - NetworkUDP, NetworkUDP4, NetworkUDP6
	//   - NetworkUnix, NetworkUnixGram (not available on Windows)
	ErrInvalidProtocol = errors.New("invalid protocol")

	// ErrInvalidTLSConfig indicates that TLS/SSL configuration is invalid or incomplete.
	//
	// This error is returned when:
	//   - TLS is enabled but Config.New() returns nil
	//   - TLS is enabled but ServerName is empty (client only)
	//   - TLS configuration lacks required certificate pairs (server only)
	//   - TLS is enabled for non-TCP protocols (UDP, Unix sockets)
	//
	// TLS is only supported for TCP-based connections:
	//   - Client: Requires valid Config and ServerName
	//   - Server: Requires valid Config with at least one certificate pair
	ErrInvalidTLSConfig = errors.New("invalid TLS config")

	// ErrInvalidGroup indicates that the Unix group ID exceeds the maximum allowed value.
	//
	// This error is returned when:
	//   - Server.GroupPerm is greater than MaxGID (32767)
	//
	// Valid group ID values:
	//   - -1: Use the process's current group (default)
	//   - 0 to MaxGID: Specific group ID
	//
	// This validation only applies to Unix domain socket servers.
	ErrInvalidGroup = errors.New("invalid unix group for socket group permission")
)
