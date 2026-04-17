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
	"errors"
	"net"
	"os"
	"path/filepath"
	"syscall"

	libprm "github.com/nabbar/golib/file/perm"
	librun "github.com/nabbar/golib/runner"
	libsck "github.com/nabbar/golib/socket"
)

// getSocketFile retrieves and validates the configured Unix socket file path.
// This internal helper is responsible for identifying the location where
// the SOCK_DGRAM endpoint will be created in the filesystem.
func (o *srv) getSocketFile() (string, error) {
	if o == nil {
		return "", ErrInvalidUnixFile
	} else if a := o.ad.Load(); a == (sckFile{}) {
		return "", ErrInvalidUnixFile
	} else if len(a.File) < 1 {
		return "", ErrInvalidUnixFile
	} else {
		return o.checkFile(a.File)
	}
}

// getSocketPerm retrieves the configured file permissions for the Unix socket.
// If not explicitly set, it defaults to 0770 (rwxrwx---), which allows the
// owner and the associated group to communicate via the socket.
func (o *srv) getSocketPerm() libprm.Perm {
	if a := o.ad.Load(); a == (sckFile{}) {
		return libprm.Perm(0770)
	} else if a.Perm == 0 {
		return libprm.Perm(0770)
	} else {
		return a.Perm
	}
}

// getSocketGroup determines the Group ID (GID) that will be applied to the
// socket file. It follows a hierarchy:
//  1. Configured GID (if >= 0).
//  2. Current process's GID (syscall.Getgid).
//  3. -1 as fallback (root or default).
func (o *srv) getSocketGroup() int {
	if a := o.ad.Load(); a == (sckFile{}) {
		return -1
	} else if a.GID >= 0 {
		return int(a.GID)
	} else if gid := syscall.Getgid(); gid <= MaxGID {
		return gid
	}

	return -1
}

// checkFile performs filesystem-level preparation for the socket.
//
// Behavior:
//  - Normalizes the path using filepath.Join.
//  - If a file already exists at that path, it is removed (os.Remove)
//    to ensure the socket can be bound to a fresh endpoint.
func (o *srv) checkFile(unixFile string) (string, error) {
	if len(unixFile) < 1 {
		return unixFile, ErrInvalidUnixFile
	} else {
		unixFile = filepath.Join(filepath.Dir(unixFile), filepath.Base(unixFile))
	}

	if _, e := os.Stat(unixFile); e != nil && !errors.Is(e, os.ErrNotExist) {
		return unixFile, e
	} else if e != nil {
		return unixFile, nil
	} else if e = os.Remove(unixFile); e != nil {
		return unixFile, e
	}

	return unixFile, nil
}

// Listen starts the Unix domain datagram socket server and blocks until termination.
//
// # Internal Lifecycle Diagram
//
//	[Listen Called]
//	       |
//	       v
//	[Check File] -> [net.ListenUnixgram] -> [Chmod/Chown Socket]
//	       |
//	       |------> [Setup gnc Channel] (Broadcast for Instant Shutdown)
//	       |
//	       v
//	[UpdateConn Callback] (Optional)
//	       |
//	       v
//	[Start Single Handler Goroutine] <--- [sCtx from sync.Pool]
//	       |
//	       |------> [select] (Wait for ctx.Done or gnc Channel)
//	       |
//	[Shutdown Triggered]
//	       |
//	       v
//	[Close net.UnixConn] -> [Remove Socket File] -> [Recycle sCtx]
//
// # Key Implementation Details
//
// 1. Instant Shutdown: Uses the 'gnc' channel for instantaneous broadcast.
//    Unlike older versions using polling tickers, this eliminates latency
//    and CPU overhead during the shutdown transition.
//
// 2. Resource Pooling: Contexts are retrieved from a sync.Pool (via o.getContext)
//    to minimize memory allocation during rapid datagram bursts.
//
// 3. Robust Cleanup: A deferred function ensures that the server's state is reset
//    to 'gone', the connection is closed, and the socket file is removed from the filesystem,
//    even in case of panics.
//
// # Returns:
//   - ctx.Err() if the provided context is canceled.
//   - nil on graceful shutdown.
//   - Relevant network or filesystem errors if startup fails.
func (o *srv) Listen(ctx context.Context) error {
	defer func() {
		if r := recover(); r != nil {
			librun.RecoveryCaller("golib/socket/server/unixgram/listen", r)
		}
	}()

	var (
		e   error         // error
		a   string        // address
		sx  *sCtx         // socket context
		con *net.UnixConn // udp con listener
		cnl context.CancelFunc
	)

	if a, e = o.getSocketFile(); e != nil {
		o.fctError(e)
		return e
	} else if o.hdl == nil {
		o.fctError(ErrInvalidInstance)
		return ErrInvalidHandler
	} else if con, e = o.getListen(a); e != nil {
		o.fctError(e)
		return e
	} else if o.upd != nil {
		o.upd(con)
	}

	// Prepare for a new listen cycle
	o.gon.Store(false)
	o.gnc.Store(make(chan struct{}))

	// Channel to signal that the loop has finished
	done := make(chan struct{})

	defer func() {
		o.run.Store(false)

		if con != nil {
			_ = con.Close()
		}

		if sx != nil {
			_ = sx.Close()
			o.putContext(sx)
		}

		// Remove socket file on shutdown
		if a != "" {
			_ = os.Remove(a)
		}

		o.setGone()
		close(done)
	}()

	// Create connection-specific context
	ctx, cnl = context.WithCancel(ctx)
	sx = o.getContext(ctx, cnl, con, con.LocalAddr().String())

	// Single goroutine to handle shutdown signals and unblock Accept()
	// This ensures that if the server is stopped from another goroutine,
	// the blocked Read() calls in the handler are immediately unblocked by closing the connection.
	go func() {
		select {
		case <-ctx.Done():
		case <-o.getGoneChan():
		case <-done: // prevent leaking if loop exits for other reasons
			return
		}

		if con != nil {
			_ = con.Close()
		}
	}()

	o.run.Store(true)

	// In Unix Datagram mode, a single handler goroutine is started.
	// This handler is expected to loop and read multiple datagrams from the connection.
	go func(conn net.Conn) {
		defer func() {
			if r := recover(); r != nil {
				librun.RecoveryCaller("golib/socket/server/unixgram/handler", r)
			}
		}()

		lc := conn.LocalAddr()
		rc := &net.UnixAddr{}

		if lc == nil {
			lc = &net.UnixAddr{}
		}

		defer o.fctInfo(lc, rc, libsck.ConnectionClose)
		o.fctInfo(lc, rc, libsck.ConnectionNew)

		o.hdl(sx)
	}(con)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-o.getGoneChan():
		return nil
	case <-done:
		return nil
	}
}
