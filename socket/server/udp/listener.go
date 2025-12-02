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
	"time"

	libptc "github.com/nabbar/golib/network/protocol"
	librun "github.com/nabbar/golib/runner"
	libsck "github.com/nabbar/golib/socket"
)

// getAddress retrieves the configured listen address from atomic storage.
// Returns an empty string if no address has been set via RegisterServer().
// This is an internal helper used by Listen().
func (o *srv) getAddress() string {
	if s := o.ad.Load(); len(s) > 0 {
		return s
	} else {
		return ""
	}
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
func (o *srv) getListen(addr string) (*net.UDPConn, error) {
	var (
		adr *net.UDPAddr
		lis *net.UDPConn
		err error
	)

	if adr, err = net.ResolveUDPAddr(libptc.NetworkUDP.Code(), addr); err != nil {
		return nil, err
	} else if lis, err = net.ListenUDP(libptc.NetworkUDP.Code(), adr); err != nil {
		if lis != nil {
			_ = lis.Close()
		}
		return nil, err
	} else {
		o.fctInfoSrv("starting listening socket '%s %s'", libptc.NetworkUDP.String(), addr)
	}

	return lis, nil
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
	defer func() {
		if r := recover(); r != nil {
			librun.RecoveryCaller("golib/socket/server/udp/listen", r)
		}
	}()

	var (
		e   error            // error
		a   = o.getAddress() // address
		sx  *sCtx            // socket context
		con *net.UDPConn     // udp con listener
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

		o.run.Store(false)
		o.gon.Store(true)
	}()

	if len(a) == 0 {
		o.fctError(ErrInvalidInstance)
		return ErrInvalidAddress
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
				librun.RecoveryCaller("golib/socket/server/udp/handler", r)
			}
		}()

		lc := conn.LocalAddr()
		rc := &net.UDPAddr{}

		if lc == nil {
			lc = &net.UDPAddr{}
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
