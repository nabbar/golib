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

package udp

import (
	"net"

	libatm "github.com/nabbar/golib/atomic"
	libptc "github.com/nabbar/golib/network/protocol"
	libsck "github.com/nabbar/golib/socket"
)

// ClientUDP represents a UDP client that implements the socket.Client interface.
//
// This interface extends github.com/nabbar/golib/socket.Client and provides
// all standard socket operations for UDP datagram communication:
//   - Connect(ctx) - Associate socket with remote address
//   - IsConnected() - Check if socket is associated
//   - Read(p []byte) - Read datagram from socket
//   - Write(p []byte) - Write datagram to socket
//   - Close() - Close the socket
//   - Once(ctx, request, response) - One-shot request/response operation
//   - SetTLS(enable, config, serverName) - No-op for UDP (always returns nil)
//   - RegisterFuncError(f) - Register error callback
//   - RegisterFuncInfo(f) - Register datagram info callback
//
// Important UDP characteristics:
//   - Connectionless: No handshake or persistent connection state
//   - Unreliable: Datagrams may be lost without notification
//   - Unordered: Datagrams may arrive out of order
//   - Message boundaries: Each Write() sends one datagram
//   - No TLS: UDP doesn't support encryption natively (use DTLS if needed)
//
// All operations are thread-safe and use atomic operations internally.
// The client maintains minimal state for the associated remote address.
//
// See github.com/nabbar/golib/socket package for interface details.
type ClientUDP interface {
	libsck.Client
}

// New creates a new UDP client for the specified address.
//
// The address parameter must be in the format "host:port", where:
//   - host can be a hostname, IPv4 address, IPv6 address (in brackets), or empty
//   - port must be a valid port number (1-65535)
//
// Examples of valid addresses:
//   - "localhost:8080"
//   - "192.168.1.1:9000"
//   - "[::1]:8080" (IPv6)
//   - ":8080" (binds to all interfaces)
//
// The client is created in a disconnected state. Use Connect() to associate
// the socket with the remote address. The address is validated but no network
// operation is performed during construction.
//
// UDP-specific notes:
//   - "Connect" doesn't establish a connection, it associates the socket
//   - Maximum datagram size is typically 65507 bytes (65535 - 8 byte header - 20 byte IP header)
//   - Consider using smaller sizes (< 1472 bytes) to avoid IP fragmentation
//   - No guarantee of delivery or ordering
//
// Returns:
//   - ClientUDP: A new client instance if successful
//   - error: ErrAddress if address is empty or malformed, or a net.Error if
//     the address cannot be resolved as a valid UDP address
//
// Example:
//
//	client, err := udp.New("localhost:8080")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Close()
//
//	ctx := context.Background()
//	if err := client.Connect(ctx); err != nil {
//	    log.Fatal(err)
//	}
//
//	// Send small datagram
//	data := []byte("Hello, UDP!")
//	n, err := client.Write(data)
//	if err != nil {
//	    log.Fatal(err)
//	}
func New(address string) (ClientUDP, error) {
	if len(address) < 1 {
		return nil, ErrAddress
	} else if _, err := net.ResolveUDPAddr(libptc.NetworkUDP.Code(), address); err != nil {
		return nil, err
	}

	c := &cli{m: libatm.NewMapAny[uint8]()}
	c.m.Store(keyNetAddr, address)

	return c, nil
}
