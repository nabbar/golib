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

package unix

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"os"
	"path/filepath"
	"sync/atomic"
	"syscall"
	"time"

	libprm "github.com/nabbar/golib/file/perm"
	libptc "github.com/nabbar/golib/network/protocol"
	libsck "github.com/nabbar/golib/socket"
)

// getSocketFile retrieves and validates the configured Unix socket file path.
// Returns the path after validation via checkFile(), or os.ErrNotExist if not set.
// This is an internal helper used by Listen().
func (o *srv) getSocketFile() (string, error) {
	f := o.sf.Load()
	if f != nil {
		return o.checkFile(f.(string))
	}

	return "", os.ErrNotExist
}

// getSocketPerm retrieves the configured file permissions for the Unix socket.
// Returns the configured permissions or 0770 as default if parsing fails.
// Uses github.com/nabbar/golib/file/perm for permission parsing.
// This is an internal helper used by getListen().
func (o *srv) getSocketPerm() os.FileMode {
	if p, e := libprm.ParseInt64(o.sp.Load()); e != nil {
		return os.FileMode(0770)
	} else {
		return p.FileMode()
	}
}

// getSocketGroup retrieves the group ID for the Unix socket file.
// Returns:
//   - The configured GID if >= 0
//   - The process's current GID if <= maxGID
//   - 0 (root) as fallback
//
// This is an internal helper used by getListen() for os.Chown().
func (o *srv) getSocketGroup() int {
	p := o.sg.Load()
	if p >= 0 {
		return int(p)
	}

	gid := syscall.Getgid()
	if gid <= maxGID {
		return gid
	}

	return 0
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
		return unixFile, fmt.Errorf("missing socket file path")
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

// getListen creates and configures the Unix socket listener.
//
// The function:
//  1. Sets umask to apply configured permissions
//  2. Creates the Unix socket listener with net.Listen()
//  3. Verifies and corrects file permissions with os.Chmod() if needed
//  4. Changes file group ownership with os.Chown() if needed
//  5. Invokes the server info callback with startup message
//
// The umask is temporarily modified to ensure correct permissions are applied
// during socket creation, then restored to the original value.
//
// Parameters:
//   - uxf: Unix socket file path
//
// Returns:
//   - net.Listener: The active Unix socket listener
//   - error: Any error during socket creation or configuration
//
// This is an internal helper called by Listen().
//
// See syscall.Umask, os.Chmod, and os.Chown for permission management.
func (o *srv) getListen(uxf string) (net.Listener, error) {
	var (
		err error
		prm = o.getSocketPerm()
		grp = o.getSocketGroup()
		old int
		inf fs.FileInfo
		lis net.Listener
	)

	old = syscall.Umask(int(prm))
	defer func() {
		syscall.Umask(old)
	}()

	if lis, err = net.Listen(libptc.NetworkUnix.Code(), uxf); err != nil {
		return nil, err
	} else if inf, err = os.Stat(uxf); err != nil {
		_ = lis.Close()
		return nil, err
	} else if inf.Mode() != prm {
		if err = os.Chmod(uxf, prm); err != nil {
			_ = lis.Close()
			return nil, err
		}
	}

	if stt, ok := inf.Sys().(*syscall.Stat_t); ok {
		if int(stt.Gid) != grp {
			if err = os.Chown(uxf, syscall.Getuid(), grp); err != nil {
				_ = lis.Close()
				return nil, err
			}
		}
	}

	o.fctInfoSrv("starting listening socket '%s %s'", libptc.NetworkUnix.String(), uxf)
	return lis, nil
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
// See github.com/nabbar/golib/socket.Handler for handler function signature.
func (o *srv) Listen(ctx context.Context) error {
	var (
		e error
		f string
		l net.Listener
		s = new(atomic.Bool)
	)

	if f, e = o.getSocketFile(); e != nil {
		o.fctError(e)
		return e
	} else if o.hdl == nil {
		o.fctError(ErrInvalidHandler)
		return ErrInvalidHandler
	} else if l, e = o.getListen(f); e != nil {
		o.fctError(e)
		return e
	}

	s.Store(false)

	defer func() {
		o.fctInfoSrv("closing listen socket '%s %s'", libptc.NetworkUnix.String(), f)

		if l != nil {
			_ = l.Close()
		}

		if _, e = os.Stat(f); e == nil {
			o.fctError(os.Remove(f))
		}

		o.run.Store(false)
	}()

	o.rst.Store(make(chan struct{}))
	o.stp.Store(make(chan struct{}))
	o.run.Store(true)
	o.gon.Store(false)

	go func() {
		defer func() {
			s.Store(true)

			if l != nil {
				o.fctError(l.Close())
			}

			go func() {
				_ = o.Shutdown(context.Background())
			}()
		}()

		select {
		case <-ctx.Done():
			return
		case <-o.Done():
			return
		}
	}()

	// Accept new connection or stop if context or shutdown trigger
	for l != nil && !s.Load() {
		if co, ce := l.Accept(); ce != nil && !s.Load() {
			o.fctError(ce)
		} else if co != nil {
			o.fctInfo(co.LocalAddr(), co.RemoteAddr(), libsck.ConnectionNew)
			go o.Conn(ctx, co)
		}
	}

	return nil
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
	var (
		cnl context.CancelFunc
		cor libsck.Reader
		cow libsck.Writer
	)

	o.nc.Add(1) // inc nb connection

	if o.upd != nil {
		o.upd(con)
	}

	ctx, cnl = context.WithCancel(ctx)
	cor, cow = o.getReadWriter(ctx, cnl, con)

	defer func() {
		// cancel context for connection
		cnl()

		// dec nb connection
		o.nc.Add(-1)

		// close connection writer
		_ = cow.Close()

		// delay stopping for 5 seconds to avoid blocking next connection
		if o.IsGone() {
			// if connection is closed
			time.Sleep(5 * time.Second)
		} else {
			// if connection is not closed in 5 seconds
			time.Sleep(500 * time.Millisecond)
		}

		// close connection reader
		_ = cor.Close()

		// send info about connection closing
		o.fctInfo(con.LocalAddr(), con.RemoteAddr(), libsck.ConnectionClose)

		// close connection
		_ = con.Close()
	}()

	// get handler or exit if nil
	if o.hdl == nil {
		return
	} else {
		go o.hdl(cor, cow)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-o.Gone():
			return
		}
	}
}

// getReadWriter creates Reader and Writer wrappers for a Unix socket connection.
// These wrappers provide a higher-level interface to the handler while managing
// Unix socket-specific behavior like half-close support.
//
// Parameters:
//   - ctx: Context for connection lifetime management
//   - cnl: Context cancel function (called when connection fully closes)
//   - con: The Unix socket connection to wrap
//
// Reader behavior:
//   - Read() calls con.Read() to receive data
//   - Invokes ConnectionRead info callback
//   - Checks context cancellation before each read
//   - Close() supports half-close via CloseRead() on *net.UnixConn
//   - Cancels context when both read and write sides are closed
//
// Writer behavior:
//   - Write() calls con.Write() to send data
//   - Invokes ConnectionWrite info callback
//   - Checks context cancellation before each write
//   - Close() supports half-close via CloseWrite() on *net.UnixConn
//   - Cancels context when both read and write sides are closed
//
// Both Reader and Writer:
//   - Implement Close() with proper half-close semantics
//   - Implement IsAlive() to check connection and context health
//   - Implement Done() returning the context's Done channel
//
// Half-close support:
// Unix sockets support independent closing of read and write directions.
// The context is only cancelled after BOTH sides are closed, allowing
// proper shutdown handshakes where one side can continue reading after
// stopping writes, or vice versa.
//
// This is an internal helper called by Conn().
//
// See github.com/nabbar/golib/socket.NewReader and NewWriter for the
// wrapper constructors.
func (o *srv) getReadWriter(ctx context.Context, cnl context.CancelFunc, con net.Conn) (libsck.Reader, libsck.Writer) {
	var (
		rc = new(atomic.Bool)
		rw = new(atomic.Bool)
	)

	rdrClose := func() error {
		defer func() {
			if rw.Load() {
				cnl()
			}
		}()

		if cr, ok := con.(*net.UnixConn); ok {
			rc.Store(true)
			o.fctInfo(con.LocalAddr(), con.RemoteAddr(), libsck.ConnectionCloseRead)
			return libsck.ErrorFilter(cr.CloseRead())
		} else {
			rc.Store(true)
			rw.Store(true)
			o.fctInfo(con.LocalAddr(), con.RemoteAddr(), libsck.ConnectionClose)
			return libsck.ErrorFilter(con.Close())
		}
	}

	wrtClose := func() error {
		defer func() {
			if rc.Load() {
				cnl()
			}
		}()

		if cr, ok := con.(*net.UnixConn); ok {
			rw.Store(true)
			o.fctInfo(con.LocalAddr(), con.RemoteAddr(), libsck.ConnectionCloseWrite)
			return libsck.ErrorFilter(cr.CloseRead())
		} else {
			rc.Store(true)
			rw.Store(true)
			o.fctInfo(con.LocalAddr(), con.RemoteAddr(), libsck.ConnectionClose)
			return libsck.ErrorFilter(con.Close())
		}
	}

	rdr := libsck.NewReader(
		func(p []byte) (n int, err error) {
			if ctx.Err() != nil {
				_ = rdrClose()
				return 0, ctx.Err()
			}
			o.fctInfo(con.LocalAddr(), con.RemoteAddr(), libsck.ConnectionRead)
			return con.Read(p)
		},
		rdrClose,
		func() bool {
			if ctx.Err() != nil {
				_ = rdrClose()
				return false
			}

			_, e := con.Write(nil)

			if e != nil {
				_ = rdrClose()
				return false
			}

			return true
		},
		func() <-chan struct{} {
			return ctx.Done()
		},
	)

	wrt := libsck.NewWriter(
		func(p []byte) (n int, err error) {
			if ctx.Err() != nil {
				_ = wrtClose()
				return 0, ctx.Err()
			}
			o.fctInfo(con.LocalAddr(), con.RemoteAddr(), libsck.ConnectionWrite)
			return con.Write(p)
		},
		wrtClose,
		func() bool {
			if ctx.Err() != nil {
				_ = wrtClose()
				return false
			}

			_, e := con.Write(nil)

			if e != nil {
				_ = wrtClose()
				return false
			}

			return true
		},
		func() <-chan struct{} {
			return ctx.Done()
		},
	)

	return rdr, wrt
}
