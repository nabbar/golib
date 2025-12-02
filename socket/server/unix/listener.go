//go:build linux || darwin

/*
 * MIT License
 *
 * Copyright (c) 2025 Nicolas JUHEL
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

package unix

import (
	"context"
	"errors"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync/atomic"
	"syscall"
	"time"

	libprm "github.com/nabbar/golib/file/perm"
	libptc "github.com/nabbar/golib/network/protocol"
	librun "github.com/nabbar/golib/runner"
	libsck "github.com/nabbar/golib/socket"
)

// getSocketFile retrieves and validates the configured Unix socket file path.
// Returns the path after validation via checkFile(), or os.ErrNotExist if not set.
// This is an internal helper used by Listen().
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
// Returns the configured permissions or 0770 as default if parsing fails.
// Uses github.com/nabbar/golib/file/perm for permission parsing.
// This is an internal helper used by getListen().
func (o *srv) getSocketPerm() libprm.Perm {
	if a := o.ad.Load(); a == (sckFile{}) {
		return libprm.Perm(0770)
	} else if a.Perm == 0 {
		return libprm.Perm(0770)
	} else {
		return a.Perm
	}
}

// getSocketGroup retrieves the group ID for the Unix socket file.
// Returns:
//   - The configured GID if >= 0
//   - The process's current GID if <= MaxGID
//   - 0 (root) as fallback
//
// This is an internal helper used by getListen() for os.Chown().
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

// checkFile validates and prepares the Unix socket file path.
//
// The function:
//  1. Validates the path is not empty
//  2. Normalizes the path using filepath.Join
//  3. Checks if the file exists
//  4. Removes existing file if present (socket files must be recreated)
//
// Returns the normalized path and any error encountered.
// Returns no error if the file doesn't exist (ready to create).
//
// This is an internal helper that ensures the socket file is in a clean
// state before Listen() attempts to create it.
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

// Listen starts the Unix socket server and begins accepting connections.
// This method blocks until the server is shut down via context cancellation,
// Shutdown(), or Close().
//
// Lifecycle:
//  1. Validates configuration (socket path and handler must be set)
//  2. Creates Unix socket file with permissions and group ownership
//  3. Creates listener socket
//  4. Sets server to running state
//  5. Starts connection acceptance loop
//  6. Spawns handler goroutine for each accepted connection
//  7. Waits for shutdown signal
//  8. Cleans up socket file and returns
//
// The handler function runs in a separate goroutine per connection and receives:
//   - Reader: Reads data from the connection
//   - Writer: Sends data to the connection
//
// Context handling:
//   - The provided context is used for the lifetime of the listener
//   - Context cancellation triggers immediate shutdown
//   - Done() channel is closed when Listen() exits
//
// Socket file management:
//   - The socket file is created at the path specified in RegisterSocket()
//   - If the file exists, it's deleted before creating the new socket
//   - The file is removed during shutdown
//   - Permissions and group ownership are applied as configured
//
// Returns:
//   - ErrInvalidHandler: If no handler was provided to New()
//   - os.ErrNotExist: If RegisterSocket() wasn't called
//   - Any error from socket creation or file operations
//
// The server maintains per-connection state and tracks connection count atomically.
// Each connection persists until explicitly closed by either side.
//
// Example:
//
//	go func() {
//	    if err := srv.Listen(ctx); err != nil {
//	        log.Printf("Server error: %v", err)
//	    }
//	}()
//
// See github.com/nabbar/golib/socket.HandlerFunc for handler function signature.
func (o *srv) Listen(ctx context.Context) error {
	defer func() {
		if r := recover(); r != nil {
			librun.RecoveryCaller("golib/socket/server/unix/listen", r)
		}
	}()

	var (
		e error        // error
		l net.Listener // socket listener
		a string       // address
	)

	if a, e = o.getSocketFile(); e != nil {
		o.fctError(e)
		return e
	} else if o.hdl == nil {
		o.fctError(ErrInvalidHandler)
		return ErrInvalidHandler
	} else if l, e = o.getListen(a); e != nil {
		o.fctError(e)
		return e
	}

	defer func() {
		o.fctInfoSrv("closing listen socket '%s %s'", libptc.NetworkUnix.String(), a)

		if l != nil {
			o.fctError(l.Close())
		}

		// Remove socket file on shutdown
		if a != "" {
			_ = os.Remove(a)
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

					o.Conn(ctx, conn)
				}(c.c)
			}
		}
	}
}

// Conn handles an individual client connection.
// This method is called in a goroutine for each accepted connection.
//
// Lifecycle:
//  1. Increments connection counter atomically
//  2. Invokes UpdateConn callback if registered
//  3. Creates connection-specific context (cancellable)
//  4. Creates Reader/Writer wrappers with the connection
//  5. Starts the handler goroutine
//  6. Waits for connection close or server shutdown
//  7. Cleans up connection resources
//  8. Decrements connection counter
//
// The connection remains active until:
//   - The client closes the connection
//   - The handler closes Reader/Writer
//   - The context is cancelled
//   - StopGone() signals connection draining (via Gone() channel)
//
// Cleanup behavior:
//   - If IsGone() is true: waits 5 seconds to avoid blocking next connection
//   - Otherwise: waits 500ms for graceful connection closure
//
// The Reader and Writer wrappers support:
//   - Half-close (Unix sockets support CloseRead/CloseWrite)
//   - Automatic context cancellation when both sides close
//   - Connection state callbacks (read, write, close events)
//
// This is an internal method called by Listen() for each accepted connection.
//
// Parameters:
//   - ctx: Context for the connection lifetime (derived from Listen's context)
//   - con: The accepted net.Conn (typically *net.UnixConn)
func (o *srv) Conn(ctx context.Context, con net.Conn) {
	defer func() {
		if r := recover(); r != nil {
			librun.RecoveryCaller("golib/socket/server/unix/conn", r)
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
		sx.loc = l.String()
	}

	if r := con.RemoteAddr(); r == nil {
		sx.rem = ""
	} else {
		sx.rem = r.String()
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
					librun.RecoveryCaller("golib/socket/server/unix/handler", r)
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
