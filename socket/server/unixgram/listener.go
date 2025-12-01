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
	"sync/atomic"
	"syscall"
	"time"

	libprm "github.com/nabbar/golib/file/perm"
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

	ctx, cnl = context.WithCancel(ctx)

	defer func() {
		if cnl != nil {
			cnl()
		}

		if sx != nil {
			_ = sx.Close()
		}

		if con != nil {
			_ = con.Close()
		}

		// Remove socket file on shutdown
		if a != "" {
			_ = os.Remove(a)
		}

		o.run.Store(false)
		o.gon.Store(true)
	}()

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

	sx = &sCtx{
		loc: "",
		ctx: ctx,
		cnl: cnl,
		con: con,
		clo: new(atomic.Bool),
	}

	if l := con.LocalAddr(); l == nil {
		sx.loc = ""
	} else {
		sx.loc = l.String()
	}

	o.gon.Store(false)
	o.run.Store(true)
	time.Sleep(time.Millisecond)

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

	// get handler or exit if nil
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

		time.Sleep(time.Millisecond)
		o.hdl(sx)
	}(con)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-cG:
			return nil
		}
	}
}
