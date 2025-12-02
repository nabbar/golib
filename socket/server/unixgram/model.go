//go:build linux || darwin

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

package unixgram

import (
	"context"
	"fmt"
	"net"
	"sync/atomic"
	"time"

	libatm "github.com/nabbar/golib/atomic"
	libtls "github.com/nabbar/golib/certificates"
	libprm "github.com/nabbar/golib/file/perm"
	libptc "github.com/nabbar/golib/network/protocol"
	librun "github.com/nabbar/golib/runner"
	libsck "github.com/nabbar/golib/socket"
)

type sckFile struct {
	File string
	Perm libprm.Perm
	GID  int32
}

// srv is the internal implementation of the ServerUnixGram interface.
// It uses atomic operations for thread-safe state management and operates
// in connectionless datagram mode (SOCK_DGRAM).
//
// Unlike connection-oriented Unix sockets (unix package), Unix datagram servers:
//   - Do not maintain per-client connections
//   - Have a single handler processing all datagrams
//   - OpenConnections() returns 1 when running, 0 when stopped
//   - Cannot use TLS (SetTLS is a no-op)
//
// All fields use atomic types or are immutable after construction to ensure
// thread safety without explicit locking.
type srv struct {
	upd libsck.UpdateConn  // Connection update callback (optional, called once on socket creation)
	hdl libsck.HandlerFunc // Datagram handler function (required)
	run *atomic.Bool       // Server is accepting connections flag
	gon *atomic.Bool       // Server is draining connections flag

	fe libatm.Value[libsck.FuncError]   // Error callback (FuncError)
	fi libatm.Value[libsck.FuncInfo]    // Connection info callback (FuncInfo)
	fs libatm.Value[libsck.FuncInfoSrv] // Server info callback (FuncInfoSrv)

	ad libatm.Value[sckFile] // Server listen address (string)
}

func (o *srv) Listener() (network libptc.NetworkProtocol, listener string, tls bool) {
	a := o.ad.Load()
	return libptc.NetworkUnixGram, a.File, false
}

// OpenConnections returns the connection count for the Unix datagram server.
// Unlike connection-oriented sockets, Unix datagram is connectionless, so this returns:
//   - 1 when the server is running (actively listening for datagrams)
//   - 0 when the server is stopped
//
// This is safe to call from multiple goroutines.
func (o *srv) OpenConnections() int64 {
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
// For Unix datagram servers, this is simply the inverse of IsRunning() since there
// are no persistent connections to drain.
//
// This state is set by calling Shutdown() or Close().
func (o *srv) IsGone() bool {
	return o.gon.Load()
}

// Close performs an immediate shutdown of the server using a background context.
// This is equivalent to calling Shutdown(context.Background()).
//
// For controlled shutdown with a custom timeout, use Shutdown() directly.
func (o *srv) Close() error {
	return o.Shutdown(context.Background())
}

// Shutdown performs a graceful server shutdown by stopping the listener.
//
// The method applies a 25-second timeout to the provided context and calls
// StopListen(). For Unix datagram servers, this is equivalent to StopListen() since
// there are no persistent connections to drain.
//
// The socket file is removed during shutdown.
//
// Returns any error from StopListen().
func (o *srv) Shutdown(ctx context.Context) error {
	if o == nil {
		return ErrInvalidInstance
	} else if !o.IsRunning() || o.IsGone() {
		return nil
	}

	o.gon.Store(true)

	var (
		tck = time.NewTicker(3 * time.Millisecond)
		cnl context.CancelFunc
	)

	ctx, cnl = context.WithTimeout(ctx, time.Second)
	defer func() {
		tck.Stop()
		cnl()
	}()

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

// SetTLS is a no-op for Unix domain socket servers.
// Unix domain sockets do not support TLS at the transport layer.
// Always returns nil regardless of parameters.
//
// For secure Unix socket communication, consider using file permissions
// to restrict access or application-level encryption.
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
// Note: Unix datagram is connectionless, so ConnectionNew and ConnectionClose events
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
//   - Socket file creation/removal
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

// RegisterSocket sets the Unix socket file path, permissions, and group ownership.
// Must be called before Listen().
//
// Parameters:
//   - unixFile: Path to the socket file (e.g., "/tmp/app.sock", "./app.sock")
//   - perm: File permissions (e.g., 0600 for user-only, 0660 for user+group)
//   - gid: Group ID for the socket file, or -1 to use the process's default group
//
// The socket file:
//   - Will be created when Listen() is called
//   - Will be removed on server shutdown
//   - Will be deleted if it exists before creating the new socket
//
// File permissions control who can send datagrams to the socket:
//   - 0600: Only the socket owner can send
//   - 0660: Owner and group members can send
//   - 0666: Anyone can send (use with caution)
//
// The address is validated using net.ResolveUnixAddr to ensure it's well-formed.
//
// Returns ErrInvalidGroup if gid exceeds MaxGID (32767).
func (o *srv) RegisterSocket(unixFile string, perm libprm.Perm, gid int32) error {
	if _, err := net.ResolveUnixAddr(libptc.NetworkUnixGram.Code(), unixFile); err != nil {
		return err
	} else if gid > MaxGID {
		return ErrInvalidGroup
	}

	o.ad.Store(sckFile{File: unixFile, Perm: perm, GID: gid})
	return nil
}

// fctError invokes the registered error callback if one exists.
// Safely handles nil server instances and nil errors.
// This is an internal helper used throughout the server for error reporting.
func (o *srv) fctError(e ...error) {
	defer func() {
		if r := recover(); r != nil {
			librun.RecoveryCaller("golib/socket/server/unixgram/fctErro", r)
		}
	}()

	if o == nil {
		return
	} else if len(e) < 1 {
		return
	}

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

// fctInfo invokes the registered datagram info callback if one exists.
// Reports datagram events with local and remote addresses.
// Safely handles nil callbacks to prevent panics.
// This is an internal helper called from datagram handling.
func (o *srv) fctInfo(local, remote net.Addr, state libsck.ConnState) {
	defer func() {
		if r := recover(); r != nil {
			librun.RecoveryCaller("golib/socket/server/unixgram/fctInfo", r)
		}
	}()

	if o == nil {
		return
	} else if f := o.fi.Load(); f != nil {
		f(local, remote, state)
	}
}

// fctInfoSrv invokes the registered server info callback if one exists.
// Formats the message with fmt.Sprintf before passing to the callback.
// Safely handles nil callbacks to prevent panics.
// This is an internal helper for server lifecycle logging.
func (o *srv) fctInfoSrv(msg string, args ...interface{}) {
	defer func() {
		if r := recover(); r != nil {
			librun.RecoveryCaller("golib/socket/server/unixgram/fctInfoSrv", r)
		}
	}()

	if o == nil {
		return
	} else if f := o.fs.Load(); f != nil {
		f(fmt.Sprintf(msg, args...))
	}
}
