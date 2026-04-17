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
	libsck "github.com/nabbar/golib/socket"
	sckcfg "github.com/nabbar/golib/socket/config"
)

// ServerUdp defines the specialized interface for a UDP datagram server.
//
// # Interface Hierarchy
//
// This interface extends the core github.com/nabbar/golib/socket.Server
// but with several important semantic differences:
//
//   - Stateless Nature: Unlike a TCP server, a UDP server doesn't "maintain"
//     connections. OpenConnections() will always return 0 (or 1 if the
//     socket is listening).
//   - Datagram Management: The server is purely datagram-oriented; no
//     automatic handshake or flow control is provided by this layer.
//   - Synchronicity: Only one main handler goroutine is spawned per Listen().
//
// # Inherited Methods (from socket.Server)
//
//   - Listen(context.Context) error: Starts the UDP listener. Blocks until
//     the context is cancelled or Shutdown is called.
//   - Shutdown(context.Context) error: Gracefully closes the listener and
//     cleans up internal resources within the given context deadline.
//   - Close() error: Immediate, non-blocking shutdown.
//   - IsRunning() bool: Thread-safe check of whether the socket is listening.
//   - IsGone() bool: Thread-safe check of whether the server is in the "Stopped" state.
//   - RegisterFuncError(FuncError): Registers a callback for server-level errors.
//   - RegisterFuncInfo(FuncInfo): Registers a callback for connection events.
//   - RegisterFuncInfoSrv(FuncInfoSrv): Registers a callback for lifecycle messages.
type ServerUdp interface {
	libsck.Server

	// RegisterServer configures the UDP address the server will bind to.
	//
	// # Requirements
	//
	//   1. Must be called BEFORE Listen().
	//   2. Address must be in the format "host:port" or ":port".
	//
	// Parameters:
	//   - address: String representing the network address.
	//
	// Returns:
	//   - error: ErrInvalidAddress if the address is empty or malformed.
	RegisterServer(address string) error
}

// New creates and initializes a new UDP server instance.
//
// # Design Principle
//
// This constructor uses a combination of mandatory and optional parameters
// to ensure the server is always in a valid, startable state.
//
// Parameters:
//   - upd: Optional UpdateConn callback. Use this to configure the raw socket
//          (e.g., SetReadBuffer, SetWriteBuffer, JoinMulticastGroup).
//   - hdl: Mandatory HandlerFunc. This is the main entry point for datagram
//          processing. It is executed in a dedicated goroutine when Listen() starts.
//   - cfg: Server configuration struct, providing the initial address and network protocol.
//
// # Implementation Notes
//
// During initialization:
//   1. Atomic values are allocated for all status flags and callbacks.
//   2. A default (no-op) broadcast channel (gnc) is created to prevent nil dereferences.
//   3. RegisterServer is called with the address provided in cfg.
//   4. The server is initially marked as 'gon=true' (not running).
//
// Returns:
//   - ServerUdp: The initialized server instance.
//   - error: If validation fails (e.g., nil handler or invalid address).
func New(upd libsck.UpdateConn, hdl libsck.HandlerFunc, cfg sckcfg.Server) (ServerUdp, error) {
	if e := cfg.Validate(); e != nil {
		return nil, e
	} else if hdl == nil {
		return nil, ErrInvalidHandler
	}

	// Default no-op callbacks to avoid nil-checks in hot paths.
	var (
		dfe libsck.FuncError   = func(_ ...error) {}
		dfi libsck.FuncInfo    = func(_, _ net.Addr, _ libsck.ConnState) {}
		dfs libsck.FuncInfoSrv = func(_ string) {}
	)

	s := &srv{
		upd: upd,
		hdl: hdl,
		// Atomic values for thread-safe callback registration at runtime.
		fe:  libatm.NewValueDefault[libsck.FuncError](dfe, dfe),
		fi:  libatm.NewValueDefault[libsck.FuncInfo](dfi, dfi),
		fs:  libatm.NewValueDefault[libsck.FuncInfoSrv](dfs, dfs),
		ad:  libatm.NewValue[string](),
		// Initialized with a default channel that will be swapped upon Listen().
		gnc: libatm.NewValueDefault[chan struct{}](make(chan struct{}), make(chan struct{})),
	}

	if e := s.RegisterServer(cfg.Address); e != nil {
		return nil, e
	}

	// Initial state is "not running"
	s.gon.Store(true)

	return s, nil
}
