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
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync/atomic"
	"syscall"

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
func (o *srv) getSocketPerm() libprm.Perm {
	if p, e := libprm.ParseInt64(o.sp.Load()); e != nil {
		return libprm.Perm(0770)
	} else {
		return p
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

// Listen starts the Unix datagram socket server and begins accepting datagrams.
// This method blocks until the server is shut down via context cancellation,
// Shutdown(), or Close().
//
// Lifecycle:
//  1. Validates configuration (socket path and handler must be set)
//  2. Creates Unix socket file with permissions and group ownership
//  3. Creates datagram listener socket (SOCK_DGRAM)
//  4. Invokes UpdateConn callback if registered
//  5. Creates Reader/Writer wrappers for the handler
//  6. Sets server to running state
//  7. Starts single handler goroutine (processes all datagrams)
//  8. Waits for shutdown signal
//  9. Cleans up socket file and returns
//
// The handler function runs in a single goroutine and receives:
//   - Reader: Reads incoming datagrams (ReadFrom under the hood)
//   - Writer: Sends response datagrams (WriteTo to last sender)
//
// Datagram handling:
//   - Unlike connection-oriented sockets, there's one handler for all datagrams
//   - Each Read() receives a complete datagram from any sender
//   - Sender address is tracked atomically for response routing
//   - No per-sender state is maintained
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
//   - ErrContextClosed: If context was cancelled
//   - Any error from socket creation or file operations
//
// The server operates in connectionless mode - each datagram is independent.
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
	var (
		e error
		u string
		s = new(atomic.Bool)

		loc *net.UnixAddr
		con *net.UnixConn
		cnl context.CancelFunc
		cor libsck.Reader
		cow libsck.Writer
	)

	s.Store(false)

	if u, e = o.getSocketFile(); e != nil {
		o.fctError(e)
		return e
	} else if o.hdl == nil {
		o.fctError(ErrInvalidHandler)
		return ErrInvalidHandler
	} else if loc, e = net.ResolveUnixAddr(libptc.NetworkUnixGram.Code(), u); e != nil {
		o.fctError(e)
		return e
	} else if con, e = o.getListen(u, loc); e != nil {
		o.fctError(e)
		return e
	}

	if o.upd != nil {
		o.upd(con)
	}

	ctx, cnl = context.WithCancel(ctx)
	cor, cow = o.getReadWriter(ctx, con, loc)

	o.stp.Store(make(chan struct{}))
	o.run.Store(true)

	defer func() {
		// cancel context for connection
		cnl()

		// send info about connection closing
		o.fctInfo(loc, &net.UnixAddr{}, libsck.ConnectionClose)
		o.fctInfoSrv("closing listen socket '%s %s'", libptc.NetworkUnixGram.String(), u)

		// close connection
		_ = con.Close()

		if _, e = os.Stat(u); e == nil {
			o.fctError(os.Remove(u))
		}

		o.run.Store(false)
	}()

	// get handler or exit if nil
	go o.hdl(cor, cow)

	for {
		select {
		case <-ctx.Done():
			return ErrContextClosed
		case <-o.Done():
			return nil
		}
	}
}

// getReadWriter creates Reader and Writer wrappers for the Unix datagram connection.
// These wrappers provide a higher-level interface to the handler while managing
// Unix datagram-specific behavior.
//
// Parameters:
//   - ctx: Context for lifecycle management
//   - con: The Unix datagram connection to wrap
//   - loc: Local address (used for info callbacks)
//
// Reader behavior:
//   - Read() calls con.ReadFrom() to receive datagrams from any sender
//   - Stores the sender's address atomically for response routing
//   - Invokes ConnectionRead info callback
//   - Checks context cancellation before each read
//
// Writer behavior:
//   - Write() calls con.WriteTo() with the last sender's address
//   - Falls back to con.Write() if no sender address is available
//   - Invokes ConnectionWrite info callback
//   - Checks context cancellation before each write
//
// Both Reader and Writer:
//   - Implement Close() to shut down the Unix datagram connection
//   - Implement IsAlive() to check connection and context health
//   - Implement Done() returning the context's Done channel
//
// The remote address tracking allows responses to be sent back to the
// originating sender for each datagram, simulating request-response patterns
// over the connectionless protocol.
//
// Unlike connection-oriented Unix sockets:
//   - No half-close support (datagram sockets don't have separate read/write sides)
//   - Single connection handles all senders
//   - Sender address changes with each datagram
//
// This is an internal helper called by Listen().
//
// See github.com/nabbar/golib/socket.NewReader and NewWriter for the
// wrapper constructors.
func (o *srv) getReadWriter(ctx context.Context, con *net.UnixConn, loc net.Addr) (libsck.Reader, libsck.Writer) {
	var (
		re = &net.UDPAddr{}
		ra = new(atomic.Value)
		fg = func() net.Addr {
			if i := ra.Load(); i != nil {
				if v, k := i.(net.Addr); k {
					return v
				}
			}
			return &net.UnixAddr{}
		}
	)
	ra.Store(re)

	fctClose := func() error {
		o.fctInfo(loc, fg(), libsck.ConnectionClose)
		return libsck.ErrorFilter(con.Close())
	}

	rdr := libsck.NewReader(
		func(p []byte) (n int, err error) {
			if ctx.Err() != nil {
				_ = fctClose()
				return 0, ctx.Err()
			}

			var a net.Addr
			n, a, err = con.ReadFrom(p)

			if a != nil {
				ra.Store(a)
			} else {
				ra.Store(re)
			}

			o.fctInfo(loc, fg(), libsck.ConnectionRead)
			return n, err
		},
		fctClose,
		func() bool {
			if ctx.Err() != nil {
				_ = fctClose()
				return false
			}
			_, e := con.Read(nil)

			if e != nil {
				_ = fctClose()
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
				_ = fctClose()
				return 0, ctx.Err()
			}

			if a := fg(); a != nil && a != re {
				o.fctInfo(loc, a, libsck.ConnectionWrite)
				return con.WriteTo(p, a)
			}

			o.fctInfo(loc, fg(), libsck.ConnectionWrite)
			return con.Write(p)
		},
		fctClose,
		func() bool {
			if ctx.Err() != nil {
				_ = fctClose()
				return false
			}
			_, e := con.Write(nil)

			if e != nil {
				_ = fctClose()
			}

			return true
		},
		func() <-chan struct{} {
			return ctx.Done()
		},
	)

	return rdr, wrt
}
