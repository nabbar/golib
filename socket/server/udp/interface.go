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
	"sync/atomic"

	libatm "github.com/nabbar/golib/atomic"
	libsck "github.com/nabbar/golib/socket"
	sckcfg "github.com/nabbar/golib/socket/config"
)

// ServerUdp defines the interface for a UDP server implementation.
// It extends the base github.com/nabbar/golib/socket.Server interface
// with UDP-specific functionality.
//
// The server operates in connectionless datagram mode:
//   - No persistent connections are maintained
//   - Each datagram is processed independently
//   - OpenConnections() always returns 0 (UDP is stateless)
//   - Graceful shutdown supported
//   - Callback registration for events and errors
//
// See github.com/nabbar/golib/socket.Server for inherited methods:
//   - Listen(context.Context) error - Start accepting datagrams
//   - Shutdown(context.Context) error - Graceful shutdown
//   - Close() error - Immediate shutdown
//   - IsRunning() bool - Check if server is accepting datagrams
//   - IsGone() bool - Check if server has stopped
//   - OpenConnections() int64 - Always returns 0 for UDP (stateless)
//   - SetTLS(bool, TLSConfig) - No-op for UDP (always no error)
//   - RegisterFuncError(FuncError) - Register error callback
//   - RegisterFuncInfo(FuncInfo) - Register datagram info callback
//   - RegisterFuncInfoSrv(FuncInfoSrv) - Register server info callback
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
//   - h: Required HandlerFunc function that processes each datagram.
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
// See github.com/nabbar/golib/socket.HandlerFunc and socket.UpdateConn for
// callback function signatures.
func New(upd libsck.UpdateConn, hdl libsck.HandlerFunc, cfg sckcfg.Server) (ServerUdp, error) {
	if e := cfg.Validate(); e != nil {
		return nil, e
	} else if hdl == nil {
		return nil, ErrInvalidHandler
	}

	var (
		dfe libsck.FuncError   = func(_ ...error) {}
		dfi libsck.FuncInfo    = func(_, _ net.Addr, _ libsck.ConnState) {}
		dfs libsck.FuncInfoSrv = func(_ string) {}
	)

	s := &srv{
		upd: upd,
		hdl: hdl,
		run: new(atomic.Bool),
		gon: new(atomic.Bool),
		fe:  libatm.NewValueDefault[libsck.FuncError](dfe, dfe),
		fi:  libatm.NewValueDefault[libsck.FuncInfo](dfi, dfi),
		fs:  libatm.NewValueDefault[libsck.FuncInfoSrv](dfs, dfs),
		ad:  libatm.NewValue[string](),
	}

	if e := s.RegisterServer(cfg.Address); e != nil {
		return nil, e
	}

	s.gon.Store(true)

	return s, nil
}
