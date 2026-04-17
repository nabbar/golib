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
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	libatm "github.com/nabbar/golib/atomic"
	libtls "github.com/nabbar/golib/certificates"
	libptc "github.com/nabbar/golib/network/protocol"
	librun "github.com/nabbar/golib/runner"
	libsck "github.com/nabbar/golib/socket"
	sckcfg "github.com/nabbar/golib/socket/config"
	sckidl "github.com/nabbar/golib/socket/idlemgr"
)

// srv is the internal concrete implementation of the ServerTcp interface.
// It manages the server's lifecycle, connection tracking, and resource pooling.
//
// # Architecture Overview
//
// The server is built around several key components designed for high performance
// and thread safety:
//
//  1. [State Management]: Uses atomic flags (run, gon) to track whether the
//     server is accepting new connections or draining existing ones.
//
//  2. [Resource Pooling]: Employs a sync.Pool for sCtx objects. This drastically
//     reduces GC overhead by reusing the connection context structures instead
//     of allocating them for every new connection.
//
//  3. [Idle Connection Management]: Integrates with sckidl.Manager. Instead of
//     spawning a time.Ticker per connection, all connections are registered to
//     this centralized manager which scans for inactivity.
//
//  4. [Broadcasting]: Uses an atomic Value containing a channel (gnc) to
//     broadcast the shutdown signal to all active listeners and connections.
//
// # Dataflow: Server Lifecycle
//
//	[New()] -> [SetTLS()] -> [Listen()]
//	                            |
//	                            v
//	                    [Accept Loop] --(new conn)--> [Conn() handler]
//	                            |                        |
//	                            v                        v
//	                    [Shutdown()] <---------- [User Handler Logic]
//	                            |
//	                            +--> [setGone()] -> [Close Listener]
//	                            +--> [Wait for NC == 0]
//	                            +--> [Exit]
//
// # Thread Safety Model
//
// All fields use atomic primitives (atomic.Bool, atomic.Int64, libatm.Value)
// or are read-only after initialization. No Mutex is used in the hot path
// to avoid lock contention.
type srv struct {
	ssl libatm.Value[*tls.Config] // Atomic storage for the TLS configuration
	upd libsck.UpdateConn         // Callback to configure net.Conn before handling
	hdl libsck.HandlerFunc        // User-defined logic to process connection data
	idl time.Duration             // Configured timeout for idle connections
	run *atomic.Bool              // Indicates if the listener loop is active
	gon *atomic.Bool              // Indicates if the server is in draining/shutdown mode

	fe libatm.Value[libsck.FuncError]   // Callback for error reporting
	fi libatm.Value[libsck.FuncInfo]    // Callback for connection lifecycle events
	fs libatm.Value[libsck.FuncInfoSrv] // Callback for server-level status updates

	ad  libatm.Value[string]        // Registered listen address (e.g., ":8080")
	gnc libatm.Value[chan struct{}] // Signal channel for graceful shutdown broadcast
	id  sckidl.Manager              // Centralized manager for connection inactivity
	nc  *atomic.Int64               // Counter of currently open connections
	pol *sync.Pool                  // Memory pool for sCtx recycling
}

// Listener returns the network type, the listen address, and whether TLS is enabled.
func (o *srv) Listener() (network libptc.NetworkProtocol, listener string, tls bool) {
	if t := o.getTLS(); t != nil {
		if len(t.Certificates) > 0 {
			return libptc.NetworkTCP, o.getAddress(), true
		}
	}

	return libptc.NetworkTCP, o.getAddress(), false
}

// OpenConnections returns the number of client connections currently handled by the server.
// This count includes connections in the process of being closed.
func (o *srv) OpenConnections() int64 {
	return o.nc.Load()
}

// IsRunning returns true if the server's listener is active and accepting connections.
func (o *srv) IsRunning() bool {
	return o.run.Load()
}

// IsGone returns true if the server has started its shutdown process.
// When Gone is true, the server stops accepting new connections and waits
// for active ones to complete.
func (o *srv) IsGone() bool {
	return o.gon.Load()
}

// setGone sets the internal 'gon' state and closes the broadcast channel.
func (o *srv) setGone() {
	if o == nil {
		return
	}

	// Swap returns the old value. Ensure close(ch) happens exactly once.
	if o.gon.Swap(true) {
		return
	}

	if ch := o.gnc.Load(); ch != nil {
		close(ch)
	}
}

// getGoneChan returns the channel that will be closed when the server starts shutting down.
func (o *srv) getGoneChan() <-chan struct{} {
	if o == nil {
		return nil
	}
	return o.gnc.Load()
}

// Close performs an immediate shutdown using a Background context.
func (o *srv) Close() error {
	return o.Shutdown(context.Background())
}

// Shutdown performs a graceful shutdown of the TCP server.
//
// It follows these steps:
//  1. Calls setGone() to signal all components to stop.
//  2. Closes the listener to stop accepting new connections.
//  3. Periodically checks OpenConnections() until it reaches zero or the context expires.
//
// Returns ErrShutdownTimeout if the context deadline is reached before all connections are closed.
func (o *srv) Shutdown(ctx context.Context) error {
	if o == nil {
		return ErrInvalidInstance
	} else if !o.IsRunning() || o.IsGone() {
		return nil
	}

	o.setGone()

	var (
		tck = time.NewTicker(3 * time.Millisecond)
		cnl context.CancelFunc
	)

	// Apply a safeguard timeout (e.g. 1s) if the provided context doesn't have one
	ctx, cnl = context.WithTimeout(ctx, time.Second) // #nosec
	defer func() {
		tck.Stop()
		cnl()
	}()

	// Wait loop for connection draining
	for o.IsRunning() || o.OpenConnections() > 0 {
		select {
		case <-ctx.Done():
			return ErrShutdownTimeout
		case <-tck.C:
			break // nolint
		}
	}

	return nil
}

// SetTLS updates the server's TLS configuration.
// Must be called before Listen() to take effect.
func (o *srv) SetTLS(enable bool, config libtls.TLSConfig) error {
	if !enable {
		o.ssl.Store(nil)
		return nil
	}

	if config == nil {
		return sckcfg.ErrInvalidTLSConfig
	} else if l := config.GetCertificatePair(); len(l) < 1 {
		return sckcfg.ErrInvalidTLSConfig
	} else if t := config.TlsConfig(""); t == nil {
		return sckcfg.ErrInvalidTLSConfig
	} else {
		o.ssl.Store(t)
		return nil
	}
}

// RegisterFuncError registers a callback for asynchronous error reporting.
func (o *srv) RegisterFuncError(f libsck.FuncError) {
	if o == nil {
		return
	}
	o.fe.Store(f)
}

// RegisterFuncInfo registers a callback for connection-level events (New, Close, I/O).
func (o *srv) RegisterFuncInfo(f libsck.FuncInfo) {
	if o == nil {
		return
	}
	o.fi.Store(f)
}

// RegisterFuncInfoServer registers a callback for server-level events (Listen start, Shutdown).
func (o *srv) RegisterFuncInfoServer(f libsck.FuncInfoSrv) {
	if o == nil {
		return
	}
	o.fs.Store(f)
}

// RegisterServer sets the network address to listen on.
func (o *srv) RegisterServer(address string) error {
	if len(address) < 1 {
		return ErrInvalidAddress
	} else if _, err := net.ResolveTCPAddr(libptc.NetworkTCP.Code(), address); err != nil {
		return err
	}

	o.ad.Store(address)
	return nil
}

// fctError internal helper to safely trigger the error callback.
func (o *srv) fctError(e ...error) {
	defer func() {
		if r := recover(); r != nil {
			librun.RecoveryCaller("golib/socket/server/tcp/fctError", r)
		}
	}()

	if o == nil || len(e) < 1 {
		return
	}

	var ok = false
	for _, err := range e {
		if err != nil {
			ok = true
			break
		}
	}

	if ok {
		if f := o.fe.Load(); f != nil {
			f(e...)
		}
	}
}

// fctInfo internal helper to safely trigger the connection lifecycle callback.
func (o *srv) fctInfo(local, remote net.Addr, state libsck.ConnState) {
	defer func() {
		if r := recover(); r != nil {
			librun.RecoveryCaller("golib/socket/server/tcp/fctInfo", r)
		}
	}()

	if o != nil {
		if f := o.fi.Load(); f != nil {
			f(local, remote, state)
		}
	}
}

// fctInfoSrv internal helper to safely trigger the server status callback.
func (o *srv) fctInfoSrv(msg string, args ...interface{}) {
	defer func() {
		if r := recover(); r != nil {
			librun.RecoveryCaller("golib/socket/server/tcp/fctInfoSrv", r)
		}
	}()

	if o != nil {
		if f := o.fs.Load(); f != nil {
			f(fmt.Sprintf(msg, args...))
		}
	}
}

// getTLS returns the current tls.Config or nil if TLS is disabled.
func (o *srv) getTLS() *tls.Config {
	i := o.ssl.Load()

	if i == nil || len(i.Certificates) < 1 {
		return nil
	}
	return i
}

// idleTimeout returns the duration after which an inactive connection is dropped.
func (o *srv) idleTimeout() time.Duration {
	if o == nil || o.idl < time.Second {
		return 0
	}
	return o.idl
}

// getContext fetches an sCtx from the memory pool or allocates a new one if empty.
// It automatically reinitializes the structure for a new connection.
func (o *srv) getContext(ctx context.Context, cnl context.CancelFunc, con io.ReadWriteCloser, l, r net.Addr, t bool) *sCtx {
	if o == nil || o.pol == nil {
		return &sCtx{}
	}

	if i := o.pol.Get(); i != nil {
		if c, ok := i.(*sCtx); ok {
			c.reset(ctx, cnl, con, l, r, t)
			return c
		}
	}

	return &sCtx{}
}

// putContext returns an sCtx to the sync.Pool for later reuse.
func (o *srv) putContext(c *sCtx) {
	if o == nil || o.pol == nil || c == nil {
		return
	}

	o.pol.Put(c)
}
