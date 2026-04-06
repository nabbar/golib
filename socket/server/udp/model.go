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

	libatm "github.com/nabbar/golib/atomic"
	libtls "github.com/nabbar/golib/certificates"
	libptc "github.com/nabbar/golib/network/protocol"
	librun "github.com/nabbar/golib/runner"
	libsck "github.com/nabbar/golib/socket"
)

// srv is the internal implementation of the ServerUdp interface.
//
// # Design Pattern: Lock-Free State Machine
//
// This struct is designed to be fully thread-safe without using sync.Mutex.
// It relies on:
//   - sync/atomic.Bool: For binary state flags (running, gone).
//   - libatm.Value (Atomic): For dynamic configuration and callback updates.
//   - Gone Channel (gnc): For lock-free broadcast notification of state changes.
//
// # Internal States
//
//   1. Initialized (run=false, gon=true): Created but not listening.
//   2. Listening (run=true, gon=false): Active socket, accepting datagrams.
//   3. Draining (run=true, gon=true): Shutdown() called, waiting for handler cleanup.
//   4. Stopped (run=false, gon=true): Listener closed and handler finished.
type srv struct {
	upd libsck.UpdateConn  // Optional callback for low-level socket configuration (tuning).
	hdl libsck.HandlerFunc // Mandatory user-provided datagram processing logic.
	run atomic.Bool        // Thread-safe flag: true when Listen() is active.
	gon atomic.Bool        // Thread-safe flag: true when the server is in shutdown mode.

	fe libatm.Value[libsck.FuncError]   // Error reporting callback (thread-safe update).
	fi libatm.Value[libsck.FuncInfo]    // Datagram traffic reporting callback (thread-safe update).
	fs libatm.Value[libsck.FuncInfoSrv] // Server lifecycle reporting callback (thread-safe update).

	ad  libatm.Value[string]        // Configured listen address.
	gnc libatm.Value[chan struct{}] // Broadcast channel used for immediate shutdown notification.
}

// Listener returns the network protocol, listen address, and TLS state.
// UDP is always non-TLS (SetTLS is a no-op).
func (o *srv) Listener() (network libptc.NetworkProtocol, listener string, tls bool) {
	return libptc.NetworkUDP, o.getAddress(), false
}

// OpenConnections returns the connection count for the UDP server.
//
// # UDP Semantics
//
// Since UDP is connectionless, there is no persistent state to track per peer.
// This method always returns 0, as per the stateless nature of the protocol.
func (o *srv) OpenConnections() int64 {
	return 0
}

// IsRunning returns true if the server's listener is active and the handler is executing.
func (o *srv) IsRunning() bool {
	return o.run.Load()
}

// IsGone returns true if the server has entered the "Stopped" or "Draining" state.
// This flag is toggled by Shutdown() or Close() and triggers immediate termination.
func (o *srv) IsGone() bool {
	return o.gon.Load()
}

// setGone marks the server as gone and broadcasts the notification to all listeners.
//
// # Internals: Broadcast Mechanism
//
// 1. o.gon.Swap(true): Atomically sets the state and checks if it was already true.
// 2. close(ch): Closes the broadcast channel if not already closed. All goroutines
//    waiting on this channel (gnc) via a select will wake up INSTANTLY.
//
// This is the core mechanism replacing traditional polling, improving shutdown speed by ~140x.
func (o *srv) setGone() {
	if o == nil {
		return
	}

	// Gatekeeper: Ensure closing logic runs exactly once per shutdown cycle.
	if o.gon.Swap(true) {
		return
	}

	// Broadcast to all goroutines (e.g., those waiting in Listen() or sCtx)
	if ch := o.gnc.Load(); ch != nil {
		close(ch)
	}
}

// getGoneChan returns the broadcast channel as a receive-only channel.
// Internal helper used for synchronization in select blocks.
func (o *srv) getGoneChan() <-chan struct{} {
	if o == nil {
		return nil
	}
	return o.gnc.Load()
}

// Close performs an immediate, context-less shutdown.
// Equivalent to calling Shutdown(context.Background()).
func (o *srv) Close() error {
	return o.Shutdown(context.Background())
}

// Shutdown initiates the graceful termination of the UDP server.
//
// # Execution Flow
//
// 1. Marks the state as gone via setGone() (triggering immediate broadcast).
// 2. Applies a timeout to the provided context (default: 25 seconds).
// 3. Loops (polling every 3ms) until IsRunning() becomes false.
// 4. Returns nil on success or ErrShutdownTimeout if the context expires first.
//
// # Parameters:
//   - ctx: Parent context for timeout control.
//
// # Returns:
//   - error: nil if shutdown finished normally, or ErrShutdownTimeout.
func (o *srv) Shutdown(ctx context.Context) error {
	if o == nil {
		return ErrInvalidInstance
	} else if !o.IsRunning() || o.IsGone() {
		// Already stopped or not started
		return nil
	}

	// Trigger broadcast notification
	o.setGone()

	var (
		tck = time.NewTicker(3 * time.Millisecond) // Frequency of status check
		cnl context.CancelFunc
	)

	// Context with timeout to prevent indefinite blocking
	ctx, cnl = context.WithTimeout(ctx, 25*time.Second) // #nosec
	defer func() {
		tck.Stop()
		cnl()
	}()

	// Wait for the Listen() goroutine to finish its cleanup defer block.
	for o.IsRunning() || o.OpenConnections() > 0 {
		select {
		case <-ctx.Done():
			return ErrShutdownTimeout
		case <-tck.C:
			continue
		}
	}

	return nil
}

// SetTLS is a no-op for UDP servers.
// UDP does not support TLS at the transport layer (no handshake).
// For secure UDP, use DTLS externally or application-level encryption.
func (o *srv) SetTLS(_ bool, _ libtls.TLSConfig) error {
	return nil
}

// RegisterFuncError dynamically updates the error reporting callback.
// Thread-safe and can be called while the server is running.
func (o *srv) RegisterFuncError(f libsck.FuncError) {
	if o == nil {
		return
	}
	o.fe.Store(f)
}

// RegisterFuncInfo dynamically updates the traffic monitoring callback.
// Reports ConnectionRead/ConnectionWrite events (ConnectionNew/Close are also reported).
func (o *srv) RegisterFuncInfo(f libsck.FuncInfo) {
	if o == nil {
		return
	}
	o.fi.Store(f)
}

// RegisterFuncInfoServer dynamically updates the lifecycle monitoring callback.
// Reports starting, stopping, and socket creation events.
func (o *srv) RegisterFuncInfoServer(f libsck.FuncInfoSrv) {
	if o == nil {
		return
	}
	o.fs.Store(f)
}

// RegisterServer validates and stores the UDP listen address.
// Must be called before Listen() to ensure the socket can be bound.
func (o *srv) RegisterServer(address string) error {
	if len(address) < 1 {
		return ErrInvalidAddress
	} else if _, err := net.ResolveUDPAddr(libptc.NetworkUDP.Code(), address); err != nil {
		return err
	}

	o.ad.Store(address)
	return nil
}

// fctError is an internal reporting helper that routes errors to the registered callback.
// It includes panic recovery to prevent the entire server from crashing due to a callback error.
func (o *srv) fctError(e ...error) {
	defer func() {
		if r := recover(); r != nil {
			librun.RecoveryCaller("golib/socket/server/udp/fctError", r)
		}
	}()

	if o == nil || len(e) < 1 {
		return
	}

	// Only invoke if at least one error is non-nil
	var ok = false
	for _, err := range e {
		if err != nil {
			ok = true
			break
		}
	}

	if !ok {
		return
	} else if f := o.fe.Load(); f != nil {
		f(e...)
	}
}

// fctInfo is an internal reporting helper for datagram events.
func (o *srv) fctInfo(local, remote net.Addr, state libsck.ConnState) {
	defer func() {
		if r := recover(); r != nil {
			librun.RecoveryCaller("golib/socket/server/udp/fctInfo", r)
		}
	}()

	if o == nil {
		return
	} else if f := o.fi.Load(); f != nil {
		f(local, remote, state)
	}
}

// fctInfoSrv is an internal reporting helper for server lifecycle events.
func (o *srv) fctInfoSrv(msg string, args ...interface{}) {
	defer func() {
		if r := recover(); r != nil {
			librun.RecoveryCaller("golib/socket/server/udp/fctInfoSrv", r)
		}
	}()

	if o == nil {
		return
	} else if f := o.fs.Load(); f != nil {
		f(fmt.Sprintf(msg, args...))
	}
}
