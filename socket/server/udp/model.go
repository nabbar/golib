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
	"context"
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

// srv is the internal implementation of the ServerUdp interface.
// It uses atomic operations for thread-safe state management and operates
// in connectionless datagram mode.
//
// Unlike TCP servers, UDP servers:
//   - Do not maintain per-client connections
//   - Have a single handler processing all datagrams
//   - OpenConnections() returns 1 when running, 0 when stopped
//   - Cannot use TLS (SetTLS is a no-op)
//
// All fields use atomic types or are immutable after construction to ensure
// thread safety without explicit locking.
type srv struct {
	upd libsck.UpdateConn // Connection update callback (optional, called once on socket creation)
	hdl libsck.Handler    // Datagram handler function (required)
	msg *atomic.Value     // Message channel (chan []byte)
	stp *atomic.Value     // Stop listening channel (chan struct{})
	run *atomic.Bool      // Server is accepting datagrams flag

	fe *atomic.Value // Error callback (FuncError)
	fi *atomic.Value // Datagram info callback (FuncInfo)
	fs *atomic.Value // Server info callback (FuncInfoSrv)

	ad *atomic.Value // Server listen address (string)
}

// OpenConnections returns the connection count for the UDP server.
// Unlike TCP, UDP is connectionless, so this returns:
//   - 1 when the server is running (actively listening for datagrams)
//   - 0 when the server is stopped
//
// This is safe to call from multiple goroutines.
func (o *srv) OpenConnections() int64 {
	if o.IsRunning() {
		return 1
	}

	return 0
}

// IsRunning returns true if the server is currently accepting datagrams.
// Returns false if the server has not started, is shutting down, or has stopped.
//
// This is safe to call concurrently and provides the server's listener state.
func (o *srv) IsRunning() bool {
	return o.run.Load()
}

// IsGone returns true if the server has stopped accepting datagrams.
// For UDP servers, this is simply the inverse of IsRunning() since there
// are no persistent connections to drain.
//
// This state is set by calling Shutdown() or Close().
func (o *srv) IsGone() bool {
	return !o.IsRunning()
}

// Done returns a channel that is closed when the server stops accepting datagrams.
// This channel is closed during shutdown.
//
// Use this to detect when Listen() has exited.
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

// Gone returns a channel that indicates when the server is fully stopped.
// For UDP servers, this always returns a closed channel since there are no
// persistent connections to drain.
//
// Unlike TCP servers, UDP shutdown is immediate once the listener stops.
func (o *srv) Gone() <-chan struct{} {
	return closedChanStruct
}

// Close performs an immediate shutdown of the server using a background context.
// This is equivalent to calling Shutdown(context.Background()).
//
// For controlled shutdown with a custom timeout, use Shutdown() directly.
func (o *srv) Close() error {
	return o.Shutdown(context.Background())
}

// StopListen signals the server to stop accepting datagrams and waits
// for the listener to exit. The Done() channel is closed when this completes.
//
// The method uses a 10-second timeout (overriding the provided context) and polls
// every 5ms until IsRunning() returns false. Returns ErrShutdownTimeout if the
// listener doesn't stop within the timeout.
//
// For UDP servers, this is typically fast since there are no connections to drain.
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

// Shutdown performs a graceful server shutdown by stopping the listener.
//
// The method applies a 25-second timeout to the provided context and calls
// StopListen(). For UDP servers, this is equivalent to StopListen() since
// there are no persistent connections to drain.
//
// Returns any error from StopListen().
func (o *srv) Shutdown(ctx context.Context) error {
	if o == nil {
		return ErrInvalidInstance
	}

	var cnl context.CancelFunc
	ctx, cnl = context.WithTimeout(ctx, 25*time.Second)
	defer cnl()
	return o.StopListen(ctx)
}

// SetTLS is a no-op for UDP servers.
// UDP does not support TLS at the transport layer.
// Always returns nil regardless of parameters.
//
// For secure UDP communication, consider using DTLS (not implemented here)
// or application-level encryption.
func (o *srv) SetTLS(enable bool, config libtls.TLSConfig) error {
	return nil
}

// RegisterFuncError registers a callback function for error notifications.
// The callback is invoked whenever an error occurs during server operation,
// including datagram processing errors, I/O errors, and listener errors.
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

// RegisterFuncInfo registers a callback function for datagram events.
// The callback is invoked for each datagram event:
//   - ConnectionRead: Data read from a datagram
//   - ConnectionWrite: Data written to a datagram response
//
// Note: UDP is connectionless, so ConnectionNew and ConnectionClose events
// are not typically generated.
//
// The function receives local and remote addresses and the event state.
// Should not block as it's called from the handler goroutine.
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

// RegisterServer sets the UDP address for the server to listen on.
// Must be called before Listen().
//
// Address format:
//   - "host:port" - Listen on specific host (e.g., "localhost:8080")
//   - ":port" - Listen on all interfaces (e.g., ":8080")
//   - "0.0.0.0:port" - Explicitly bind to all IPv4 interfaces
//
// The address is validated using net.ResolveUDPAddr to ensure it's well-formed.
//
// Returns ErrInvalidAddress if the address is empty or cannot be parsed.
func (o *srv) RegisterServer(address string) error {
	if len(address) < 1 {
		return ErrInvalidAddress
	} else if _, err := net.ResolveUDPAddr(libptc.NetworkUDP.Code(), address); err != nil {
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

// fctInfo invokes the registered datagram info callback if one exists.
// Reports datagram events with local and remote addresses.
// Safely handles nil callbacks to prevent panics.
// This is an internal helper called from datagram handling.
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
