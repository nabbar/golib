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
	"errors"
	"io"
	"net"
	"strings"
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
// # Internal Logic and Flow
//
//	1. [Initialization]: Validates address and handler.
//	2. [Listener Setup]: Creates net.Listener (optional TLS).
//	3. [Idle Manager]: If configured, starts the global sckidl.Manager to monitor timeouts.
//	4. [State Update]: Sets o.run = true.
//	5. [Shutdown Watcher]: Spawns a background goroutine to close the listener on ctx.Done() or setGone().
//	6. [Accept Loop]: Blocks on l.Accept().
//	   - On Success: Spawns o.Conn() in a new goroutine.
//	   - On Error: Checks if error is expected (closed listener) or fatal.
//	7. [Cleanup]: Closes listener, stops idle manager, sets o.run = false.
//
// # Graceful Shutdown Mechanism
//
// The server implements a two-stage shutdown:
//   - Stage 1: Close the listener to stop accepting new connections.
//   - Stage 2: Wait for active connections to finish (managed in Shutdown() method).
//
// # Error Handling and Return Values
//
//   - Returns ctx.Err() if context was canceled.
//   - Returns nil if the server was stopped cleanly via Close() or Shutdown().
//   - Returns any fatal error from net.Listen or l.Accept.
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

	// Start Idle Manager if timeout is defined
	if o.id != nil {
		if e = o.id.Start(ctx); e != nil {
			o.fctError(e)
		}
	}

	// Prepare for a new listen cycle
	o.gon.Store(false)
	o.gnc.Store(make(chan struct{}))

	// Channel to signal that the loop has finished
	done := make(chan struct{})

	defer func() {
		o.fctInfoSrv("closing listen socket '%s %s'", libptc.NetworkTCP.String(), a)

		if l != nil {
			_ = l.Close()
		}

		if o.id != nil {
			_ = o.id.Stop(context.Background())
		}

		o.run.Store(false)
		o.setGone()
		close(done)
	}()

	// Shutdown Watcher: handle signals to unblock Accept()
	go func() {
		select {
		case <-ctx.Done():
		case <-o.getGoneChan():
		case <-done: // prevent leaking if loop exits for other reasons
			return
		}

		if l != nil {
			_ = l.Close()
		}
	}()

	o.run.Store(true)

	// Main Accept Loop
	for {
		conn, err := l.Accept()
		if err != nil {
			// Check if the context was canceled - we must return the error for tests
			if ctx.Err() != nil {
				return ctx.Err()
			}

			// Check if we are shutting down via setGone
			if o.IsGone() {
				return nil
			}

			// For newer Go versions, net.ErrClosed is preferred
			if errors.Is(err, net.ErrClosed) {
				return nil
			}

			// compatibility for older Go or specific wrappers
			if strings.Contains(err.Error(), "use of closed network connection") {
				return nil
			}

			o.fctError(err)
			continue
		}

		if conn == nil {
			continue
		}

		// Handle each connection in its own goroutine
		go func(c net.Conn) {
			lc := c.LocalAddr()
			rc := c.RemoteAddr()

			defer o.fctInfo(lc, rc, libsck.ConnectionClose)
			o.fctInfo(lc, rc, libsck.ConnectionNew)

			o.Conn(ctx, c, t)
		}(conn)
	}
}

// Conn handles a single client connection in its own goroutine.
// This method is called automatically by Listen() for each accepted connection.
//
// # Connection Initialization Dataflow
//
//	1. [Counter]: Increment atomic connection count (nc).
//	2. [User Hook]: Execute UpdateConn callback (upd) to tune socket.
//	3. [TCP Tuning]:
//	   - Enable TCP_NODELAY (NoDelay) for lower latency.
//	   - Configure TCP Keep-Alive if idle timeout > 30s.
//	4. [Context Setup]:
//	   - Get sCtx from sync.Pool (recycle memory).
//	   - Create connection-specific cancellation context.
//	5. [Idle Registration]: Add connection to centralized Idle Manager (id).
//	6. [Handler Execution]: Spawn user HandlerFunc (hdl) in a new goroutine.
//	7. [Monitoring]: Wait for context termination or server shutdown signal.
//	8. [Cleanup]:
//	   - Unregister from Idle Manager.
//	   - Close context and socket.
//	   - Put sCtx back to sync.Pool.
//	   - Decrement connection count.
//
// # Performance Tuning
//
// This method implements several optimizations for high-throughput servers:
//   - TCP_NODELAY: Disabled Nagle's algorithm to ensure immediate packet delivery.
//   - Keep-Alive Configuration: Explicitly sets Idle/Interval/Count to detect dead peers faster.
//   - Object Pooling: Uses sync.Pool for connection contexts to avoid GC churn.
//
// # Thread Safety
//
// Safe for concurrent calls. Each call operates on a unique connection and
// isolated pooled context.
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
	)

	defer func() {
		// Decrement active connection count
		o.nc.Add(-1)

		// Remove from global idle monitor
		if o.id != nil && dur > 0 {
			_ = o.id.Unregister(sx)
		}

		if cnl != nil {
			cnl()
		}

		if sx != nil {
			_ = sx.Close()
			o.putContext(sx) // Return sCtx to sync.Pool
		}

		if con != nil {
			_ = con.Close()
		}
	}()

	o.nc.Add(1) // Increment active connection count

	// Allow connection configuration before handling
	if o.upd != nil {
		o.upd(con)
	}

	// Apply low-level TCP optimizations
	if c, k := con.(*net.TCPConn); k {
		// Ensure low latency (Disable Nagle's algorithm)
		_ = c.SetNoDelay(true)

		// Set aggressive Keep-Alive if requested via idle timeout
		if dur > 30*time.Second {
			_ = c.SetKeepAlive(true)
			_ = c.SetKeepAlivePeriod(dur)
			_ = c.SetKeepAliveConfig(net.KeepAliveConfig{
				Enable:   true,
				Idle:     dur,
				Interval: 15 * time.Second,
				Count:    0,
			})
		}
	}

	if c, k := con.(io.ReadWriteCloser); k {
		// Create connection-specific context using memory pool
		ctx, cnl = context.WithCancel(ctx)
		sx = o.getContext(ctx, cnl, c, con.LocalAddr(), con.RemoteAddr(), isTls)
	} else {
		return
	}

	// Register with centralized Idle Manager
	if dur > 0 {
		_ = o.id.Register(sx)
	}

	// Start user handler logic
	if o.hdl != nil {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					librun.RecoveryCaller("golib/socket/server/tcp/handler", r)
				}
			}()

			o.hdl(sx)
		}()
	}

	// Block until connection context is closed, parent is canceled, or server is gone
	select {
	case <-ctx.Done():
		return
	case <-sx.Done():
		return
	case <-o.getGoneChan():
		return
	}
}
