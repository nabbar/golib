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

package tcp

import (
	"net"

	libatm "github.com/nabbar/golib/atomic"
	libptc "github.com/nabbar/golib/network/protocol"
	libsck "github.com/nabbar/golib/socket"
)

// ClientTCP represents a TCP client that implements the socket.Client interface.
//
// This interface extends github.com/nabbar/golib/socket.Client and provides
// all standard socket operations including:
//   - Connect(ctx) - Establish connection to server
//   - IsConnected() - Check connection status
//   - Read(p []byte) - Read data from connection
//   - Write(p []byte) - Write data to connection
//   - Close() - Close the connection
//   - Once(ctx, request, response) - One-shot request/response operation
//   - SetTLS(enable, config, serverName) - Configure TLS encryption
//   - RegisterFuncError(f) - Register error callback
//   - RegisterFuncInfo(f) - Register connection state callback
//
// All operations are thread-safe and use atomic operations internally.
// The client maintains connection state and automatically calls registered
// callbacks for errors and connection state changes.
//
// See github.com/nabbar/golib/socket package for interface details.
type ClientTCP interface {
	libsck.Client
}

// New creates a new TCP client for the specified address.
//
// The address parameter must be in the format "host:port", where:
//   - host can be a hostname, IPv4 address, IPv6 address (in brackets), or empty for localhost
//   - port must be a valid port number (1-65535)
//
// Examples of valid addresses:
//   - "localhost:8080"
//   - "192.168.1.1:9000"
//   - "[::1]:8080" (IPv6)
//   - ":8080" (binds to all interfaces)
//
// The client is created in a disconnected state. Use Connect() to establish
// the connection. The address is validated but no network connection is
// attempted during construction.
//
// Returns:
//   - ClientTCP: A new client instance if successful
//   - error: ErrAddress if address is empty or malformed, or a net.Error if
//     the address cannot be resolved as a valid TCP address
//
// Example:
//
//	client, err := tcp.New("localhost:8080")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Close()
func New(address string) (ClientTCP, error) {
	if len(address) < 1 {
		return nil, ErrAddress
	} else if _, err := net.ResolveTCPAddr(libptc.NetworkTCP.Code(), address); err != nil {
		return nil, err
	}

	c := &cli{m: libatm.NewMapAny[uint8]()}
	c.m.Store(keyNetAddr, address)

	return c, nil
}
