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
	"io"
	"net"
	"sync/atomic"
	"time"

	libptc "github.com/nabbar/golib/network/protocol"
	librun "github.com/nabbar/golib/runner"
	libsck "github.com/nabbar/golib/socket"
)

// getAddress retrieves the currently configured server listen address.
// This is an internal helper function used by the server during initialization
// and connection handling.
//
// Returns:
//   - string: The configured listen address in "host:port" format, or an empty
//     string if no address has been registered.
//
// The address is set using the RegisterServer() method and is used when starting
// the listener. This function provides thread-safe access to the address field.
//
// Example:
//
//	srv.RegisterServer(":8080")
//	addr := srv.getAddress() // Returns ":8080"
func (o *srv) getAddress() string {
	if s := o.ad.Load(); len(s) > 0 {
		return s
	}

	return ""
}

// getListen creates and initializes a TCP listener on the specified address.
// This is an internal helper function used by the Listen() method to set up
// the network listener with optional TLS encryption.
//
// Parameters:
//   - addr: The network address to listen on, in "host:port" format.
//     Use ":0" to let the OS choose an available port.
//
// Returns:
//   - net.Listener: The created TCP listener, possibly wrapped with TLS
//   - bool: true if TLS is enabled, false otherwise
//   - error: Any error that occurred during listener creation
//
// The function performs the following steps:
//  1. Creates a TCP listener on the specified address
//  2. If TLS is configured (via SetTLS), wraps the listener with tls.NewListener
//  3. Logs the listener creation via the server info callback
//
// This function is safe to call multiple times but each call will create
// a new listener. The caller is responsible for closing the returned listener.
//
// Example:
//
//	listener, isTLS, err := s.getListen(":8443")
//	if err != nil {
//	    return fmt.Errorf("failed to create listener: %w", err)
//	}
//	defer listener.Close()
func (o *srv) getListen(addr string) (net.Listener, bool, error) {
	var (
		lis net.Listener
		ssl = false
		err error
	)

	if lis, err = net.Listen(libptc.NetworkTCP.Code(), addr); err != nil {
		return lis, false, err
	} else if t := o.getTLS(); t != nil {
		// Wrap with TLS if configured
		lis = tls.NewListener(lis, t)
		ssl = true
	}

	if ssl {
		o.fctInfoSrv("starting listening socket 'TLS %s %s'", libptc.NetworkTCP.String(), addr)
	} else {
		o.fctInfoSrv("starting listening socket '%s %s'", libptc.NetworkTCP.String(), addr)
	}

	return lis, ssl, nil
}

// Listen starts the TCP server and begins accepting client connections.
// This is the main server loop that runs until the context is cancelled or
// an unrecoverable error occurs.
//
// # Overview
//
// The Listen method performs the following sequence of operations:
//  1. Validates server configuration (address, handler)
//  2. Creates the TCP listener (with optional TLS)
//  3. Updates server state to "running"
//  4. Starts the accept loop in a separate goroutine
//  5. Waits for shutdown signals or context cancellation
//  6. Performs cleanup when stopping
//
// # Connection Handling
//
// Each incoming connection is handled in its own goroutine, allowing the server
// to handle multiple clients concurrently. The connection handling includes:
//   - TCP keepalive (if configured)
//   - TLS handshake (if enabled)
//   - Connection state tracking
//   - Error handling and recovery
//
// # Graceful Shutdown
//
// The server supports graceful shutdown through several mechanisms:
//   - Context cancellation: The provided context can be cancelled to initiate shutdown
//   - StopListen(): Stops accepting new connections
//   - Shutdown(): Stops accepting new connections and waits for active ones to complete
//   - Close(): Forcefully closes all connections immediately
//
// # Error Handling
//
// The following errors may be returned:
//   - ErrInvalidAddress: If no address has been registered
//   - ErrInvalidHandler: If no handler was provided to New()
//   - Network errors: If binding to the address fails
//   - TLS errors: If TLS configuration is invalid
//   - Context errors: If the context is cancelled
//
// # Example
//
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//
//	// Start the server in a goroutine
//	go func() {
//	    if err := srv.Listen(ctx); err != nil {
//	        log.Fatalf("Server error: %v", err)
//	    }
//	}()
//
//	// Later, to shut down:
//	// cancel() // or srv.Shutdown(context.Background())
//
// # Concurrency
//
// The server is designed to be safe for concurrent use. Multiple goroutines
// may call Listen(), StopListen(), Shutdown(), and other methods simultaneously.
//
// # Resource Management
//
// The server manages the following resources:
//   - Network listener (closed on shutdown)
//   - Active connections (closed on shutdown)
//   - Goroutines for connection handling (cleaned up on shutdown)
//
// It is the caller's responsibility to ensure proper cleanup by calling
// Shutdown() or Close() when the server is no longer needed.
//   - ErrInvalidHandler if no handler was provided to New()
//   - Any error from net.Listen() during listener creation
//   - nil when the server exits cleanly
//
// The method is safe to call only once per server instance. Calling it
// multiple times concurrently will result in undefined behavior.
//
// See Conn() for per-connection handling and github.com/nabbar/golib/socket.HandlerFunc
// for the handler function signature.
func (o *srv) Listen(ctx context.Context) error {
	defer func() {
		if r := recover(); r != nil {
			librun.RecoveryCaller("golib/socket/server/tcp/listen", r)
		}
	}()

	var (
		e error            // error
		l net.Listener     // socket listener
		t bool             // is tls
		a = o.getAddress() // address
	)

	// Validate configuration
	if len(a) == 0 {
		o.fctError(ErrInvalidAddress)
		return ErrInvalidAddress
	} else if o.hdl == nil {
		o.fctError(ErrInvalidHandler)
		return ErrInvalidHandler
	} else if l, t, e = o.getListen(a); e != nil {
		o.fctError(e)
		return e
	}

	defer func() {
		o.fctInfoSrv("closing listen socket '%s %s'", libptc.NetworkTCP.String(), a)

		if l != nil {
			o.fctError(l.Close())
		}

		o.run.Store(false)
		o.gon.Store(true)
	}()

	o.gon.Store(false)
	o.run.Store(true)
	time.Sleep(time.Millisecond)

	type cR struct {
		c net.Conn
		e error
	}

	// Create Channel to check server is Going to shutdown
	cG := make(chan bool, 1)
	go func() {
		tc := time.NewTicker(time.Millisecond)
		for {
			<-tc.C
			if o.IsGone() {
				cG <- true
				return
			}
		}
	}()

	for {
		// Create a channel to receive the accept result
		cC := make(chan cR, 1)

		// Start accept in a goroutine
		go func() {
			co, ce := l.Accept()
			cC <- cR{c: co, e: ce}
		}()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-cG:
			return nil
		case c := <-cC:
			if c.e != nil {
				o.fctError(c.e)
			} else if c.c == nil {
				// skip error message for invalid connection
			} else {
				go func(conn net.Conn) {
					lc := conn.LocalAddr()
					rc := conn.RemoteAddr()

					defer o.fctInfo(lc, rc, libsck.ConnectionClose)
					o.fctInfo(lc, rc, libsck.ConnectionNew)

					o.Conn(ctx, conn, t)
				}(c.c)
			}
		}
	}
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
// github.com/nabbar/golib/socket.HandlerFunc for the handler signature.
func (o *srv) Conn(ctx context.Context, con net.Conn, isTls bool) {
	defer func() {
		if r := recover(); r != nil {
			librun.RecoveryCaller("golib/socket/server/tcp/conn", r)
		}
	}()

	var (
		cnl context.CancelFunc
		dur = o.idleTimeout()

		sx *sCtx
		tc *time.Ticker
		tw = time.NewTicker(3 * time.Millisecond)
	)

	defer func() {
		// Decrement active connection count
		o.nc.Add(-1)

		if cnl != nil {
			cnl()
		}

		if sx != nil {
			_ = sx.Close()
		}

		if con != nil {
			_ = con.Close()
		}

		if tc != nil {
			tc.Stop()
		}

		if tw != nil {
			tw.Stop()
		}
	}()

	o.nc.Add(1) // Increment active connection count

	// Allow connection configuration before handling
	if o.upd != nil {
		o.upd(con)
	}

	// Create connection-specific context
	ctx, cnl = context.WithCancel(ctx)
	sx = &sCtx{
		ctx: ctx,
		cnl: cnl,
		clo: new(atomic.Bool),
	}

	if c, k := con.(io.ReadWriteCloser); k {
		sx.con = c
	} else {
		return
	}

	if l := con.LocalAddr(); l == nil {
		sx.loc = ""
	} else {
		sx.ptc = l.Network()
		sx.loc = l.String()
	}

	if r := con.RemoteAddr(); r == nil {
		sx.rem = ""
	} else {
		sx.rem = r.String()
	}

	if isTls {
		sx.ptc = sx.ptc + "/tls"
	}

	if dur > 0 {
		tc = time.NewTicker(dur)
		sx.rst = func() {
			tc.Reset(dur)
		}
	} else {
		tc = time.NewTicker(time.Hour)
		sx.rst = func() {
			tc.Reset(time.Hour)
		}
	}

	// get handler or exit if nil
	if o.hdl == nil {
		return
	} else {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					librun.RecoveryCaller("golib/socket/server/tcp/handler", r)
				}
			}()

			o.hdl(sx)
		}()
	}

	for ctx.Err() == nil && !o.IsGone() {
		select {
		case <-tc.C:
			if dur > 0 {
				return
			}
		case <-tw.C:
			// check ctx & gone
		}
	}
}
