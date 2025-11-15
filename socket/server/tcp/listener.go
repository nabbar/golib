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
	"context"
	"crypto/tls"
	"net"
	"sync/atomic"
	"time"

	libptc "github.com/nabbar/golib/network/protocol"
	libsck "github.com/nabbar/golib/socket"
)

// getAddress retrieves the configured server listen address.
// Returns an empty string if no address has been registered.
// This is an internal helper used by Listen().
func (o *srv) getAddress() string {
	f := o.ad.Load()

	if f != nil {
		return f.(string)
	}

	return ""
}

// getListen creates a TCP listener on the specified address.
// If TLS is configured (getTLS() returns non-nil), wraps the listener with tls.NewListener.
//
// Logs the listener creation via the server info callback, indicating whether
// TLS is enabled.
//
// Returns the listener and any error from net.Listen().
// This is an internal helper called by Listen().
func (o *srv) getListen(addr string) (net.Listener, error) {
	var (
		lis net.Listener
		err error
	)

	if lis, err = net.Listen(libptc.NetworkTCP.Code(), addr); err != nil {
		return lis, err
	} else if t := o.getTLS(); t != nil {
		// Wrap with TLS if configured
		lis = tls.NewListener(lis, t)
		o.fctInfoSrv("starting listening socket 'TLS %s %s'", libptc.NetworkTCP.String(), addr)
	} else {
		o.fctInfoSrv("starting listening socket '%s %s'", libptc.NetworkTCP.String(), addr)
	}

	return lis, nil
}

// Listen starts the TCP server and begins accepting client connections.
// This is the main server loop and blocks until the server is shut down or the context is cancelled.
//
// The method performs the following:
//  1. Validates that an address has been registered via RegisterServer()
//  2. Validates that a handler was provided to New()
//  3. Creates the TCP listener (with TLS if configured)
//  4. Sets up signal channels for graceful shutdown
//  5. Spawns a goroutine to monitor context cancellation and shutdown signals
//  6. Enters an accept loop to handle incoming connections
//
// Each accepted connection is handled in a separate goroutine via Conn().
// The server can be stopped by:
//   - Calling Shutdown(), StopListen(), or Close()
//   - Cancelling the provided context
//
// Returns:
//   - ErrInvalidAddress if no address has been registered
//   - ErrInvalidHandler if no handler was provided to New()
//   - Any error from net.Listen() during listener creation
//   - nil when the server exits cleanly
//
// The method is safe to call only once per server instance. Calling it
// multiple times concurrently will result in undefined behavior.
//
// See Conn() for per-connection handling and github.com/nabbar/golib/socket.Handler
// for the handler function signature.
func (o *srv) Listen(ctx context.Context) error {
	var (
		e error              // error
		l net.Listener       // socket listener
		a = o.getAddress()   // address
		s = new(atomic.Bool) // shutdown signal flag
	)

	// Validate configuration
	if len(a) == 0 {
		o.fctError(ErrInvalidHandler)
		return ErrInvalidAddress
	} else if o.hdl == nil {
		o.fctError(ErrInvalidHandler)
		return ErrInvalidHandler
	} else if l, e = o.getListen(a); e != nil {
		o.fctError(e)
		return e
	}

	s.Store(false)

	defer func() {
		o.fctInfoSrv("closing listen socket '%s %s'", libptc.NetworkTCP.String(), a)

		if l != nil {
			_ = l.Close()
		}

		go func() {
			_ = o.StopGone(context.Background())
		}()

		o.run.Store(false)
	}()

	o.rst.Store(make(chan struct{}))
	o.stp.Store(make(chan struct{}))
	o.run.Store(true)
	o.gon.Store(false)

	go func() {
		defer func() {
			s.Store(true)

			if l != nil {
				o.fctError(l.Close())
			}

			go func() {
				_ = o.Shutdown(context.Background())
			}()
		}()

		select {
		case <-ctx.Done():
			return
		case <-o.Done():
			return
		}
	}()

	// Accept new connection or stop if context or shutdown trigger
	for l != nil && !s.Load() {
		if co, ce := l.Accept(); ce != nil && !s.Load() {
			o.fctError(ce)
		} else if co != nil {
			o.fctInfo(co.LocalAddr(), co.RemoteAddr(), libsck.ConnectionNew)
			go o.Conn(ctx, co)
		}
	}

	return nil
}

// Conn handles a single client connection in its own goroutine.
// This method is called automatically by Listen() for each accepted connection.
//
// The connection lifecycle:
//  1. Increments the active connection counter
//  2. Invokes the UpdateConn callback (if registered) to configure the connection
//  3. Creates a child context for connection-specific cancellation
//  4. Wraps the connection in Reader/Writer interfaces via getReadWriter()
//  5. Spawns the user's handler function in a goroutine
//  6. Monitors for context cancellation or server shutdown (Gone signal)
//  7. Cleans up and decrements the connection counter on exit
//
// The method handles both graceful and ungraceful connection termination:
//   - During normal shutdown: brief delay (500ms) before cleanup
//   - During draining (IsGone): longer delay (5s) to allow final I/O
//
// Connection state changes are reported via the registered FuncInfo callback.
// The handler receives Reader and Writer interfaces that support:
//   - Partial close (CloseRead/CloseWrite for TCP connections)
//   - Context-aware I/O operations
//   - Connection liveness checking
//
// This method should not be called directly by users. It's invoked automatically
// by Listen() for each new connection.
//
// See getReadWriter() for the Reader/Writer implementation and
// github.com/nabbar/golib/socket.Handler for the handler signature.
func (o *srv) Conn(ctx context.Context, con net.Conn) {
	var (
		cnl context.CancelFunc
		cor libsck.Reader
		cow libsck.Writer
	)

	o.nc.Add(1) // Increment active connection count

	// Allow connection configuration before handling
	if o.upd != nil {
		o.upd(con)
	}

	// Create connection-specific context
	ctx, cnl = context.WithCancel(ctx)
	cor, cow = o.getReadWriter(ctx, cnl, con)

	defer func() {
		// cancel context for connection
		cnl()

		// dec nb connection
		o.nc.Add(-1)

		// close connection writer
		_ = cow.Close()

		// delay stopping for 5 seconds to avoid blocking next connection
		if o.IsGone() {
			// if connection is closed
			time.Sleep(5 * time.Second)
		} else {
			// if connection is not closed in 5 seconds
			time.Sleep(500 * time.Millisecond)
		}

		// close connection reader
		_ = cor.Close()

		// send info about connection closing
		o.fctInfo(con.LocalAddr(), con.RemoteAddr(), libsck.ConnectionClose)

		// close connection
		_ = con.Close()
	}()

	// get handler or exit if nil
	if o.hdl == nil {
		return
	} else {
		go o.hdl(cor, cow)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-o.Gone():
			return
		}
	}
}

// getReadWriter creates Reader and Writer interfaces for a connection.
// These interfaces wrap the net.Conn with context-aware operations and
// support partial connection closure (CloseRead/CloseWrite).
//
// The implementation:
//   - Tracks read and write side closure states atomically
//   - Cancels the connection context when both sides are closed
//   - Reports connection state changes via fctInfo callbacks
//   - Filters errors through libsck.ErrorFilter to normalize I/O errors
//   - Supports graceful half-close for TCP connections
//
// For TCP connections (*net.TCPConn), CloseRead and CloseWrite are used
// to close individual sides. For other connection types, Close() is called
// which closes both sides.
//
// The Reader and Writer interfaces provide:
//   - Read([]byte) (int, error) - Context-aware read with state reporting
//   - Write([]byte) (int, error) - Context-aware write with state reporting
//   - Close() error - Closes the read or write side
//   - IsAlive() bool - Checks if the connection is still usable
//   - Done() <-chan struct{} - Returns the connection context's Done channel
//
// This is an internal method called by Conn() to wrap connections.
//
// See github.com/nabbar/golib/socket.NewReader and NewWriter for the
// interface constructors, and github.com/nabbar/golib/socket.ErrorFilter
// for error normalization.
func (o *srv) getReadWriter(ctx context.Context, cnl context.CancelFunc, con net.Conn) (libsck.Reader, libsck.Writer) {
	var (
		rc = new(atomic.Bool) // Read side closed flag
		rw = new(atomic.Bool) // Write side closed flag
	)

	// rdrClose handles read side closure
	rdrClose := func() error {
		defer func() {
			if rw.Load() {
				cnl()
			}
		}()

		if cr, ok := con.(*net.TCPConn); ok {
			rc.Store(true)
			o.fctInfo(con.LocalAddr(), con.RemoteAddr(), libsck.ConnectionCloseRead)
			return libsck.ErrorFilter(cr.CloseRead())
		} else {
			rc.Store(true)
			rw.Store(true)
			o.fctInfo(con.LocalAddr(), con.RemoteAddr(), libsck.ConnectionClose)
			return libsck.ErrorFilter(con.Close())
		}
	}

	wrtClose := func() error {
		defer func() {
			if rc.Load() {
				cnl()
			}
		}()

		if cr, ok := con.(*net.TCPConn); ok {
			rw.Store(true)
			o.fctInfo(con.LocalAddr(), con.RemoteAddr(), libsck.ConnectionCloseWrite)
			return libsck.ErrorFilter(cr.CloseRead())
		} else {
			rc.Store(true)
			rw.Store(true)
			o.fctInfo(con.LocalAddr(), con.RemoteAddr(), libsck.ConnectionClose)
			return libsck.ErrorFilter(con.Close())
		}
	}

	rdr := libsck.NewReader(
		func(p []byte) (n int, err error) {
			if ctx.Err() != nil {
				_ = rdrClose()
				return 0, ctx.Err()
			}
			o.fctInfo(con.LocalAddr(), con.RemoteAddr(), libsck.ConnectionRead)
			return con.Read(p)
		},
		rdrClose,
		func() bool {
			if ctx.Err() != nil {
				_ = rdrClose()
				return false
			}

			_, e := con.Write(nil)

			if e != nil {
				_ = rdrClose()
				return false
			}

			return true
		},
		func() <-chan struct{} {
			return ctx.Done()
		},
	)

	wrt := libsck.NewWriter(
		func(p []byte) (n int, err error) {
			if ctx.Err() != nil {
				_ = wrtClose()
				return 0, ctx.Err()
			}
			o.fctInfo(con.LocalAddr(), con.RemoteAddr(), libsck.ConnectionWrite)
			return con.Write(p)
		},
		wrtClose,
		func() bool {
			if ctx.Err() != nil {
				_ = wrtClose()
				return false
			}

			_, e := con.Write(nil)

			if e != nil {
				_ = wrtClose()
				return false
			}

			return true
		},
		func() <-chan struct{} {
			return ctx.Done()
		},
	)

	return rdr, wrt
}
