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
	"sync/atomic"

	libsck "github.com/nabbar/golib/socket"
)

// ServerUdp defines the interface for a UDP server implementation.
// It extends the base github.com/nabbar/golib/socket.Server interface
// with UDP-specific functionality.
//
// The server operates in connectionless datagram mode:
//   - No persistent connections are maintained
//   - Each datagram is processed independently
//   - OpenConnections() always returns 1 when server is running
//   - Graceful shutdown supported
//   - Callback registration for events and errors
//
// See github.com/nabbar/golib/socket.Server for inherited methods:
//   - Listen(context.Context) error - Start accepting datagrams
//   - Shutdown(context.Context) error - Graceful shutdown
//   - Close() error - Immediate shutdown
//   - IsRunning() bool - Check if server is accepting datagrams
//   - IsGone() bool - Check if server has stopped
//   - OpenConnections() int64 - Returns 1 if running, 0 if stopped
//   - Done() <-chan struct{} - Channel closed when server stops
//   - SetTLS(bool, TLSConfig) error - No-op for UDP (always returns nil)
//   - RegisterFuncError(FuncError) - Register error callback
//   - RegisterFuncInfo(FuncInfo) - Register datagram info callback
//   - RegisterFuncInfoServer(FuncInfoSrv) - Register server info callback
type ServerUdp interface {
	libsck.Server

	// RegisterServer sets the UDP address for the server to listen on.
	// The address must be in the format "host:port" or ":port" to bind to all interfaces.
	//
	// Example addresses:
	//   - "127.0.0.1:8080" - Listen on localhost port 8080
	//   - ":8080" - Listen on all interfaces port 8080
	//   - "0.0.0.0:8080" - Explicitly listen on all IPv4 interfaces
	//
	// This method must be called before Listen(). Returns ErrInvalidAddress
	// if the address is empty or malformed.
	RegisterServer(address string) error
}

// New creates a new UDP server instance.
//
// Parameters:
//   - u: Optional UpdateConn callback invoked when the UDP socket is created.
//     Can be used to set socket options (e.g., buffer sizes). Pass nil if not needed.
//     Note: For UDP, this is called once per Listen(), not per datagram.
//   - h: Required Handler function that processes each datagram.
//     Receives Reader and Writer interfaces for the datagram.
//     The handler runs in its own goroutine for each Listen() call.
//
// The returned server must have RegisterServer() called to set the listen address
// before calling Listen().
//
// Example usage:
//
//	handler := func(r socket.Reader, w socket.Writer) {
//	    defer r.Close()
//	    defer w.Close()
//	    buf := make([]byte, 65507) // Max UDP datagram size
//	    n, _ := r.Read(buf)
//	    w.Write(buf[:n]) // Echo back
//	}
//
//	srv := udp.New(nil, handler)
//	srv.RegisterServer(":8080")
//	srv.Listen(context.Background())
//
// The server is safe for concurrent use and manages lifecycle properly.
// Unlike TCP, UDP servers maintain no per-client state.
//
// See github.com/nabbar/golib/socket.Handler and socket.UpdateConn for
// callback function signatures.
func New(u libsck.UpdateConn, h libsck.Handler) ServerUdp {
	c := new(atomic.Value)
	c.Store(make(chan []byte))

	s := new(atomic.Value)
	s.Store(make(chan struct{}))

	return &srv{
		upd: u,
		hdl: h,
		msg: c,
		stp: s,
		run: new(atomic.Bool),
		fe:  new(atomic.Value),
		fi:  new(atomic.Value),
		fs:  new(atomic.Value),
		ad:  new(atomic.Value),
	}
}
