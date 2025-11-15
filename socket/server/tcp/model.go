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
	"net"
	"sync/atomic"
	"time"

	libtls "github.com/nabbar/golib/certificates"
	libptc "github.com/nabbar/golib/network/protocol"
	libsck "github.com/nabbar/golib/socket"
)

var (
	// closedChanStruct is a pre-closed channel used as a sentinel value
	// to indicate that a channel has been closed or should be treated as closed.
	// This avoids repeated channel allocations and provides a consistent closed state.
	closedChanStruct chan struct{}
)

// init initializes the package-level closedChanStruct sentinel channel.
func init() {
	closedChanStruct = make(chan struct{})
	close(closedChanStruct)
}

// srv is the internal implementation of the ServerTcp interface.
// It uses atomic operations for thread-safe state management and supports
// concurrent client connections with proper lifecycle management.
//
// All fields use atomic types or are immutable after construction to ensure
// thread safety without explicit locking.
type srv struct {
	ssl *atomic.Value     // TLS configuration (*tls.Config)
	upd libsck.UpdateConn // Connection update callback (optional)
	hdl libsck.Handler    // Connection handler function (required)
	msg *atomic.Value     // Message channel (chan []byte)
	stp *atomic.Value     // Stop listening channel (chan struct{})
	rst *atomic.Value     // Reset/gone channel (chan struct{})
	run *atomic.Bool      // Server is accepting connections flag
	gon *atomic.Bool      // Server is draining connections flag

	fe *atomic.Value // Error callback (FuncError)
	fi *atomic.Value // Connection info callback (FuncInfo)
	fs *atomic.Value // Server info callback (FuncInfoSrv)

	ad *atomic.Value // Server listen address (string)
	nc *atomic.Int64 // Active connection counter
}

// OpenConnections returns the current number of active client connections.
// This count is atomically maintained and safe to call from multiple goroutines.
//
// The count increments when a connection is accepted and decrements when
// the connection is fully closed (both read and write sides).
func (o *srv) OpenConnections() int64 {
	return o.nc.Load()
}

// IsRunning returns true if the server is currently accepting new connections.
// Returns false if the server has not started, is shutting down, or has stopped.
//
// This is safe to call concurrently and provides the server's accept loop state.
func (o *srv) IsRunning() bool {
	return o.run.Load()
}

// IsGone returns true if the server is in connection draining mode.
// When true, no new connections will be accepted and existing connections
// are being closed or allowed to finish gracefully.
//
// This state is set by calling StopGone() or Shutdown().
func (o *srv) IsGone() bool {
	return o.gon.Load()
}

// Done returns a channel that is closed when the server stops accepting connections.
// This channel is closed during shutdown but before all connections are drained.
//
// Use this to detect when Listen() has exited. For complete shutdown including
// all connections being closed, use Gone() instead.
//
// Returns a pre-closed channel if the server is nil or not initialized.
func (o *srv) Done() <-chan struct{} {
	if o == nil {
		return closedChanStruct
	}

	if i := o.stp.Load(); i != nil {
		if c, k := i.(chan struct{}); k {
			return c
		}
	}

	return closedChanStruct
}

// Gone returns a channel that is closed when all connections have been closed
// and the server is fully shutdown. This happens after Done() is closed.
//
// Use this to wait for complete connection draining during graceful shutdown.
// Returns a pre-closed channel if the server is already gone or nil.
func (o *srv) Gone() <-chan struct{} {
	if o == nil {
		return closedChanStruct
	}
	if o.IsGone() {
		return closedChanStruct
	} else if i := o.rst.Load(); i != nil {
		if g, k := i.(chan struct{}); k {
			return g
		}
	}

	return closedChanStruct
}

// Close performs an immediate shutdown of the server using a background context.
// This is equivalent to calling Shutdown(context.Background()).
//
// For controlled shutdown with a custom timeout, use Shutdown() directly.
func (o *srv) Close() error {
	return o.Shutdown(context.Background())
}

// StopGone triggers connection draining and waits for all active connections to close.
// It signals all connection handlers to terminate via the Gone() channel and polls
// until OpenConnections() reaches zero.
//
// The method uses a 10-second timeout (overriding the provided context) and polls
// every 5ms. Returns ErrGoneTimeout if connections don't close within the timeout.
//
// This is typically called as part of Shutdown() but can be invoked independently
// for connection draining without stopping the listener.
//
// The method is safe against double-close panics using defer/recover.
func (o *srv) StopGone(ctx context.Context) error {
	if o == nil {
		return ErrInvalidInstance
	}

	// Set the gone flag to signal all connection handlers
	o.gon.Store(true)

	if i := o.rst.Load(); i != nil {
		if c, k := i.(chan struct{}); k && c != closedChanStruct {
			// Use defer recover to handle potential double close
			func() {
				defer func() {
					_ = recover() // Ignore panic from closing already closed channel
				}()
				close(c)
			}()
		}
	}
	o.rst.Store(closedChanStruct)

	var (
		tck = time.NewTicker(5 * time.Millisecond)
		cnl context.CancelFunc
	)

	ctx, cnl = context.WithTimeout(ctx, 10*time.Second)

	defer func() {
		tck.Stop()
		cnl()
	}()

	for {
		select {
		case <-ctx.Done():
			return ErrGoneTimeout
		case <-tck.C:
			if o.OpenConnections() > 0 {
				continue
			}
			return nil
		}
	}

}

// StopListen signals the server to stop accepting new connections and waits
// for the accept loop to exit. The Done() channel is closed when this completes.
//
// The method uses a 10-second timeout (overriding the provided context) and polls
// every 5ms until IsRunning() returns false. Returns ErrShutdownTimeout if the
// listener doesn't stop within the timeout.
//
// Existing connections are not affected and will continue to run.
// Use StopGone() or Shutdown() to drain connections.
//
// The method is safe against double-close panics using defer/recover.
func (o *srv) StopListen(ctx context.Context) error {
	if o == nil {
		return ErrInvalidInstance
	}

	if i := o.stp.Load(); i != nil {
		if c, k := i.(chan struct{}); k && c != closedChanStruct {
			// Use defer recover to handle potential double close
			func() {
				defer func() {
					_ = recover() // Ignore panic from closing already closed channel
				}()
				close(c)
			}()
		}
	}
	o.stp.Store(closedChanStruct)

	var (
		tck = time.NewTicker(5 * time.Millisecond)
		cnl context.CancelFunc
	)

	ctx, cnl = context.WithTimeout(ctx, 10*time.Second)

	defer func() {
		tck.Stop()
		cnl()
	}()

	for {
		select {
		case <-ctx.Done():
			return ErrShutdownTimeout
		case <-tck.C:
			if o.IsRunning() {
				continue
			}
			return nil
		}
	}

}

// Shutdown performs a graceful server shutdown by first draining connections
// (StopGone) and then stopping the listener (StopListen).
//
// The method applies a 25-second timeout to the provided context and calls:
//  1. StopGone() - Wait for all connections to close
//  2. StopListen() - Stop accepting new connections
//
// Returns the error from StopListen() if it fails, otherwise returns the error
// from StopGone(). This ensures that listener errors take precedence.
//
// For immediate shutdown without waiting for connections, use Close() instead.
func (o *srv) Shutdown(ctx context.Context) error {
	if o == nil {
		return ErrInvalidInstance
	}

	var cnl context.CancelFunc
	ctx, cnl = context.WithTimeout(ctx, 25*time.Second)
	defer cnl()

	// First drain connections, then stop listener
	e := o.StopGone(ctx)
	if err := o.StopListen(ctx); err != nil {
		return err
	} else {
		return e
	}
}

// SetTLS configures TLS/SSL encryption for the server.
//
// Parameters:
//   - enable: If false, sets a default TLS config with TLS 1.2-1.3 support but no certificates.
//     If true, validates and applies the provided config.
//   - config: TLS configuration from github.com/nabbar/golib/certificates.
//     Must contain at least one certificate pair when enable is true.
//
// This method must be called before Listen() to enable TLS. When enabled, the server
// will only accept TLS connections.
//
// Returns an error if:
//   - config is nil when enable is true
//   - config has no certificate pairs
//   - config.TlsConfig() returns nil
//
// See github.com/nabbar/golib/certificates.TLSConfig for config creation.
func (o *srv) SetTLS(enable bool, config libtls.TLSConfig) error {
	if !enable {
		// Store default TLS config without certificates
		o.ssl.Store(&tls.Config{
			MinVersion: tls.VersionTLS12,
			MaxVersion: tls.VersionTLS13,
		})
		return nil
	}

	// Validate TLS config and certificates
	if config == nil {
		return fmt.Errorf("invalid tls config")
	} else if l := config.GetCertificatePair(); len(l) < 1 {
		return fmt.Errorf("invalid tls config, missing certificates pair")
	} else if t := config.TlsConfig(""); t == nil {
		return fmt.Errorf("invalid tls config")
	} else {
		o.ssl.Store(t)
		return nil
	}
}

// RegisterFuncError registers a callback function for error notifications.
// The callback is invoked whenever an error occurs during server operation,
// including connection errors, I/O errors, and listener errors.
//
// The function receives variadic errors and should not block as it's called
// from various goroutines. Pass nil to clear the callback.
//
// Thread-safe and can be called at any time, even while the server is running.
//
// See github.com/nabbar/golib/socket.FuncError for the callback signature.
func (o *srv) RegisterFuncError(f libsck.FuncError) {
	if o == nil {
		return
	}

	o.fe.Store(f)
}

// RegisterFuncInfo registers a callback function for connection state changes.
// The callback is invoked for each connection event:
//   - ConnectionNew: New connection accepted
//   - ConnectionRead: Data read from connection
//   - ConnectionWrite: Data written to connection
//   - ConnectionCloseRead: Read side closed
//   - ConnectionCloseWrite: Write side closed
//   - ConnectionClose: Connection fully closed
//
// The function receives local and remote addresses and the connection state.
// Should not block as it's called from connection handler goroutines.
// Pass nil to clear the callback.
//
// See github.com/nabbar/golib/socket.FuncInfo and ConnState for details.
func (o *srv) RegisterFuncInfo(f libsck.FuncInfo) {
	if o == nil {
		return
	}

	o.fi.Store(f)
}

// RegisterFuncInfoServer registers a callback function for server informational messages.
// The callback receives formatted string messages about server lifecycle events:
//   - Server starting/stopping
//   - Listener creation/closure
//   - Configuration changes
//
// Should not block as it's called from the server's main goroutines.
// Pass nil to clear the callback.
//
// See github.com/nabbar/golib/socket.FuncInfoSrv for the callback signature.
func (o *srv) RegisterFuncInfoServer(f libsck.FuncInfoSrv) {
	if o == nil {
		return
	}

	o.fs.Store(f)
}

// RegisterServer sets the TCP address for the server to listen on.
// Must be called before Listen().
//
// Address format:
//   - "host:port" - Listen on specific host (e.g., "localhost:8080")
//   - ":port" - Listen on all interfaces (e.g., ":8080")
//   - "0.0.0.0:port" - Explicitly bind to all IPv4 interfaces
//
// The address is validated using net.ResolveTCPAddr to ensure it's well-formed.
//
// Returns ErrInvalidAddress if the address is empty or cannot be parsed.
func (o *srv) RegisterServer(address string) error {
	if len(address) < 1 {
		return ErrInvalidAddress
	} else if _, err := net.ResolveTCPAddr(libptc.NetworkTCP.Code(), address); err != nil {
		return err
	}

	o.ad.Store(address)
	return nil
}

// fctError invokes the registered error callback if one exists.
// Safely handles nil server instances and nil errors.
// This is an internal helper used throughout the server for error reporting.
func (o *srv) fctError(e error) {
	if o == nil {
		return
	}

	if e == nil {
		return
	}

	v := o.fe.Load()
	if v != nil {
		v.(libsck.FuncError)(e)
	}
}

// fctInfo invokes the registered connection info callback if one exists.
// Reports connection state changes with local and remote addresses.
// Safely handles nil callbacks to prevent panics.
// This is an internal helper called from connection lifecycle events.
func (o *srv) fctInfo(local, remote net.Addr, state libsck.ConnState) {
	if o == nil {
		return
	}

	v := o.fi.Load()
	if v != nil {
		if fn, ok := v.(libsck.FuncInfo); ok && fn != nil {
			fn(local, remote, state)
		}
	}
}

// fctInfoSrv invokes the registered server info callback if one exists.
// Formats the message with fmt.Sprintf before passing to the callback.
// Safely handles nil callbacks to prevent panics.
// This is an internal helper for server lifecycle logging.
func (o *srv) fctInfoSrv(msg string, args ...interface{}) {
	if o == nil {
		return
	}

	v := o.fs.Load()
	if v != nil {
		if fn, ok := v.(libsck.FuncInfoSrv); ok && fn != nil {
			fn(fmt.Sprintf(msg, args...))
		}
	}
}

// getTLS retrieves the current TLS configuration if TLS is enabled.
// Returns nil if:
//   - No TLS config has been set
//   - The stored config is not a valid *tls.Config
//   - The config has no certificates
//
// This is an internal helper used by the listener to determine if TLS should be used.
func (o *srv) getTLS() *tls.Config {
	i := o.ssl.Load()

	if i == nil {
		return nil
	} else if t, k := i.(*tls.Config); !k {
		return nil
	} else if len(t.Certificates) < 1 {
		return nil
	} else {
		return t
	}
}
