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
	"net"
	"sync/atomic"

	libptc "github.com/nabbar/golib/network/protocol"
	libsck "github.com/nabbar/golib/socket"
)

// getAddress retrieves the configured listen address from atomic storage.
// Returns an empty string if no address has been set via RegisterServer().
// This is an internal helper used by Listen().
func (o *srv) getAddress() string {
	f := o.ad.Load()

	if f != nil {
		return f.(string)
	}

	return ""
}

// getListen creates and binds a UDP listener socket on the specified address.
//
// The function:
//   - Resolves the address string to a *net.UDPAddr
//   - Creates a UDP socket listening on that address
//   - Invokes the server info callback with startup message
//   - Cleans up on errors
//
// Returns:
//   - *net.UDPAddr: The resolved local address
//   - *net.UDPConn: The active UDP connection/listener
//   - error: Any error during resolution or socket creation
//
// This is an internal helper called by Listen().
func (o *srv) getListen(addr string) (*net.UDPAddr, *net.UDPConn, error) {
	var (
		adr *net.UDPAddr
		lis *net.UDPConn
		err error
	)

	if adr, err = net.ResolveUDPAddr(libptc.NetworkUDP.Code(), addr); err != nil {
		return nil, nil, err
	} else if lis, err = net.ListenUDP(libptc.NetworkUDP.Code(), adr); err != nil {
		if lis != nil {
			_ = lis.Close()
		}
		return nil, nil, err
	} else {
		o.fctInfoSrv("starting listening socket '%s %s'", libptc.NetworkUDP.String(), addr)
	}

	return adr, lis, nil
}

// Listen starts the UDP server and begins accepting datagrams.
// This method blocks until the server is shut down via context cancellation,
// Shutdown(), or Close().
//
// Lifecycle:
//  1. Validates configuration (address and handler must be set)
//  2. Creates UDP listener socket
//  3. Invokes UpdateConn callback if registered
//  4. Creates Reader/Writer wrappers for the handler
//  5. Sets server to running state
//  6. Starts handler goroutine
//  7. Waits for shutdown signal
//  8. Cleans up and returns
//
// The handler function runs in a separate goroutine and receives:
//   - Reader: Reads incoming datagrams (ReadFrom under the hood)
//   - Writer: Sends response datagrams (WriteTo to last sender)
//
// Context handling:
//   - The provided context is used for the lifetime of the listener
//   - Context cancellation triggers immediate shutdown
//   - Done() channel is closed when Listen() exits
//
// Returns:
//   - ErrInvalidAddress: If RegisterServer() wasn't called
//   - ErrInvalidHandler: If no handler was provided to New()
//   - ErrContextClosed: If context was cancelled
//   - Any error from socket creation
//
// The server maintains no per-datagram state. Each datagram is processed
// independently by reading from the connection and writing responses back
// to the source address.
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
		s = new(atomic.Bool)
		a = o.getAddress()

		loc *net.UDPAddr
		con *net.UDPConn
		cnl context.CancelFunc
		cor libsck.Reader
		cow libsck.Writer
	)

	s.Store(false)

	if len(a) == 0 {
		o.fctError(ErrInvalidInstance)
		return ErrInvalidAddress
	} else if o.hdl == nil {
		o.fctError(ErrInvalidInstance)
		return ErrInvalidHandler
	} else if loc, con, e = o.getListen(a); e != nil {
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
		o.fctInfo(loc, &net.UDPAddr{}, libsck.ConnectionClose)
		o.fctInfoSrv("closing listen socket '%s %s'", libptc.NetworkUDP.String(), a)

		// close connection
		_ = con.Close()

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

// getReadWriter creates Reader and Writer wrappers for the UDP connection.
// These wrappers provide a higher-level interface to the handler while managing
// UDP-specific behavior.
//
// Parameters:
//   - ctx: Context for lifecycle management
//   - con: The UDP connection to wrap
//   - loc: Local address (used for info callbacks)
//
// Reader behavior:
//   - Read() calls con.ReadFrom() to receive datagrams
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
//   - Implement Close() to shut down the UDP connection
//   - Implement IsAlive() to check connection and context health
//   - Implement Done() returning the context's Done channel
//
// The remote address tracking allows responses to be sent back to the
// originating sender for each datagram, simulating request-response patterns
// over UDP's connectionless protocol.
//
// This is an internal helper called by Listen().
//
// See github.com/nabbar/golib/socket.NewReader and NewWriter for the
// wrapper constructors.
func (o *srv) getReadWriter(ctx context.Context, con *net.UDPConn, loc net.Addr) (libsck.Reader, libsck.Writer) {
	var (
		re = &net.UDPAddr{}
		ra = new(atomic.Value)
		fg = func() net.Addr {
			if i := ra.Load(); i != nil {
				if v, k := i.(net.Addr); k {
					return v
				}
			}
			return &net.UDPAddr{}
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
