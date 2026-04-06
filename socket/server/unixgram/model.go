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
	"sync"
	"sync/atomic"
	"time"

	libatm "github.com/nabbar/golib/atomic"
	libtls "github.com/nabbar/golib/certificates"
	libprm "github.com/nabbar/golib/file/perm"
	libptc "github.com/nabbar/golib/network/protocol"
	librun "github.com/nabbar/golib/runner"
	libsck "github.com/nabbar/golib/socket"
)

// sckFile represents the filesystem attributes for a Unix socket.
type sckFile struct {
	File string
	Perm libprm.Perm
	GID  int32
}

// srv is the internal implementation of the ServerUnixGram interface.
//
// # Internal Architecture and Performance
//
//  1. Atomic State Flags: Uses atomic.Bool ('run' and 'gon') for lock-free state
//     transitions, ensuring high concurrency without synchronization overhead.
//
//  2. Resource Pooling: Uses a sync.Pool ('pol') to recycle sCtx instances.
//     In high-throughput datagram processing, this dramatically reduces memory
//     allocations and Garbage Collector (GC) pressure.
//
//  3. Event-Driven Shutdown: Uses a broadcast channel ('gnc') wrapped in a
//     libatm.Value. When the server shuts down, the channel is closed, providing
//     instant notification to any blocked goroutines.
//
// # Design Constraints:
//   - Connectionless (SOCK_DGRAM): OpenConnections() always returns 0.
//   - Single Handler: A single handler processes all datagrams sequentially or concurrently
//     as implemented by the user.
//   - No TLS: Unix domain sockets do not support TLS at the transport layer.
type srv struct {
	upd libsck.UpdateConn  // Connection update callback
	hdl libsck.HandlerFunc // Main datagram handler function
	run atomic.Bool        // True if the server is actively listening
	gon atomic.Bool        // True if the server has entered its shutdown cycle

	fe libatm.Value[libsck.FuncError]   // Error reporting callback
	fi libatm.Value[libsck.FuncInfo]    // Datagram event callback
	fs libatm.Value[libsck.FuncInfoSrv] // Server lifecycle callback

	ad  libatm.Value[sckFile]       // Atomic storage for address and file permissions
	gnc libatm.Value[chan struct{}] // Atomic channel for shutdown broadcast
	pol *sync.Pool                  // sync.Pool for recycling sCtx instances
}

// Listener returns information about the current server listener.
// For Unix Datagram, TLS is always false.
func (o *srv) Listener() (network libptc.NetworkProtocol, listener string, tls bool) {
	a := o.ad.Load()
	return libptc.NetworkUnixGram, a.File, false
}

// OpenConnections returns 0 for a Unix datagram server.
//
// Unlike Stream sockets, Datagram sockets are connectionless and do not track
// individual "sessions" or "connections".
func (o *srv) OpenConnections() int64 {
	return 0
}

// IsRunning returns true if the server is in the "Listen" state and accepting datagrams.
func (o *srv) IsRunning() bool {
	return o.run.Load()
}

// IsGone returns true if the server has completed its shutdown cycle or has been closed.
func (o *srv) IsGone() bool {
	return o.gon.Load()
}

// setGone marks the server as gone and triggers the shutdown broadcast.
//
// Behavior:
//   - Uses Swap(true) to ensure the broadcast channel is closed exactly once.
//   - Closes the 'gnc' channel to instantaneously signal all listening goroutines.
func (o *srv) setGone() {
	if o == nil {
		return
	}

	// Swap returns the old value. If it was already true, we do nothing.
	if o.gon.Swap(true) {
		return
	}

	if ch := o.gnc.Load(); ch != nil {
		close(ch)
	}
}

// getGoneChan returns a read-only channel used to monitor the server's shutdown state.
func (o *srv) getGoneChan() <-chan struct{} {
	if o == nil {
		return nil
	}
	return o.gnc.Load()
}

// Close initiates an immediate shutdown of the server.
func (o *srv) Close() error {
	return o.Shutdown(context.Background())
}

// Shutdown performs a graceful stop of the Unix domain datagram server.
//
// Lifecycle:
//  1. Sets the 'gone' state and broadcasts the signal via the channel.
//  2. Monitors the 'running' state until the listener goroutine exits.
//  3. Respects the timeout of the provided context (returns ErrShutdownTimeout).
//
// Parameters:
//   - ctx: Context controlling the shutdown timeout.
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

	ctx, cnl = context.WithTimeout(ctx, time.Second) // #nosec
	defer func() {
		tck.Stop()
		cnl()
	}()

	for o.IsRunning() {
		select {
		case <-ctx.Done():
			return ErrShutdownTimeout
		case <-tck.C:
			// wait for listener to exit
		}
	}

	return nil
}

// SetTLS is a no-op for Unix domain sockets.
// Always returns nil.
func (o *srv) SetTLS(enable bool, config libtls.TLSConfig) error {
	return nil
}

// RegisterFuncError registers a callback to handle asynchronous errors.
func (o *srv) RegisterFuncError(f libsck.FuncError) {
	if o == nil {
		return
	}

	o.fe.Store(f)
}

// RegisterFuncInfo registers a callback for datagram-related events.
func (o *srv) RegisterFuncInfo(f libsck.FuncInfo) {
	if o == nil {
		return
	}

	o.fi.Store(f)
}

// RegisterFuncInfoServer registers a callback for server-wide lifecycle events.
func (o *srv) RegisterFuncInfoServer(f libsck.FuncInfoSrv) {
	if o == nil {
		return
	}

	o.fs.Store(f)
}

// RegisterSocket defines the filesystem path and security settings for the socket.
//
// Parameters:
//   - unixFile: The path to the socket (e.g., "/tmp/server.sock").
//   - perm: Permissions applied via chmod.
//   - gid: Group ownership applied via chown.
func (o *srv) RegisterSocket(unixFile string, perm libprm.Perm, gid int32) error {
	if len(unixFile) < 1 {
		return ErrInvalidUnixFile
	} else if _, err := net.ResolveUnixAddr(libptc.NetworkUnixGram.Code(), unixFile); err != nil {
		return err
	} else if gid > MaxGID {
		return ErrInvalidGroup
	}

	o.ad.Store(sckFile{File: unixFile, Perm: perm, GID: gid})
	return nil
}

// Internal Helper: fctError invokes the error callback safely.
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

// Internal Helper: fctInfo invokes the info callback safely.
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

// Internal Helper: fctInfoSrv invokes the server lifecycle callback safely.
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

// getContext retrieves an sCtx instance, either from the pool or by creating a new one.
func (o *srv) getContext(ctx context.Context, cnl context.CancelFunc, con *net.UnixConn, loc string) *sCtx {
	if o == nil || o.pol == nil {
		return &sCtx{}
	}

	if i := o.pol.Get(); i != nil {
		if c, ok := i.(*sCtx); ok {
			c.reset(ctx, cnl, con, loc)
			return c
		}
	}

	return &sCtx{}
}

// putContext returns an sCtx instance to the sync.Pool for reuse.
func (o *srv) putContext(c *sCtx) {
	if o == nil || o.pol == nil || c == nil {
		return
	}

	o.pol.Put(c)
}
