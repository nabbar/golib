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

	libptc "github.com/nabbar/golib/network/protocol"
	librun "github.com/nabbar/golib/runner"
	libsck "github.com/nabbar/golib/socket"
)

// getAddress retrieves the configured listen address from atomic storage.
// Internal helper used by Listen().
func (o *srv) getAddress() string {
	if s := o.ad.Load(); len(s) > 0 {
		return s
	} else {
		return ""
	}
}

// getListen creates and binds a UDP listener socket on the specified address.
//
// # Execution Steps
//
// 1. Resolves the address string to a *net.UDPAddr.
// 2. Creates a UDP socket (net.ListenUDP) bound to that address.
// 3. Reports a startup message via the server info callback.
//
// Returns:
//   - *net.UDPConn: The active UDP connection/listener.
//   - error: Any failure in address resolution or socket creation.
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

// Listen starts the UDP server and enters the main processing loop.
//
// # Design Pattern: Blocking-Wait
//
// This method blocks for the entire lifetime of the server. It manages the
// lifecycle of the underlying socket and the user-provided handler goroutine.
//
// # Detailed Lifecycle Sequence
//
//	  [Listen Called]
//	         │
//	         ▼
//	[1. Validate State] ─────────▶ Return Error (if addr or handler missing)
//	         │
//	         ▼
//	[2. Create Socket] ──────────▶ Return Error (if net.ListenUDP fails)
//	         │
//	         ▼
//	[3. UpdateConn Hook] ────────▶ Invoke user-provided tuning callback (if any)
//	         │
//	         ▼
//	[4. Reset Gone State] ───────▶ Mark gon=false, Create new 'gnc' channel
//	         │
//	         ▼
//	[5. Create sCtx Wrapper] ────▶ Wrap *net.UDPConn with context for handler
//	         │
//	         ▼
//	[6. Spawn Monitor] ──────────▶ Goroutine waiting for [Shutdown OR CtxDone]
//	         │
//	         ▼
//	[7. Spawn Handler] ──────────▶ Execute user-provided HandlerFunc(sCtx)
//	         │
//	         ▼
//	[8. Set Running=true]
//	         │
//	         ▼
//	[9. Block on gnc/CtxDone] ◀──▶ [Waiting for termination signal]
//	         │
//	         ▼
//	[10. Cleanup on Exit] ───────▶ Close socket, Mark gon=true, Set Running=false
//
// # Concurrency Model
//
// The 'srv' struct uses atomic flags to maintain thread-safe status during this flow.
// The new 'gnc' (Gone channel) broadcast mechanism allows the [6. Monitor]
// goroutine to unblock the main [Listen] thread instantly when Shutdown() is called.
//
// # Error Propagation
//
//   - Context cancellation: Returns the context error.
//   - Internal Shutdown: Returns nil.
//   - Startup errors: Returns the specific network or validation error.
func (o *srv) Listen(ctx context.Context) error {
	defer func() {
		if r := recover(); r != nil {
			librun.RecoveryCaller("golib/socket/server/udp/listen", r)
		}
	}()

	var (
		e   error            // internal error tracking
		a   = o.getAddress() // address from atomic storage
		sx  *sCtx            // contextual socket wrapper
		con *net.UDPConn     // the raw UDP socket
		cnl context.CancelFunc
	)

	// Step 1: Pre-requisite validation
	if len(a) == 0 {
		o.fctError(ErrInvalidInstance)
		return ErrInvalidAddress
	} else if o.hdl == nil {
		o.fctError(ErrInvalidInstance)
		return ErrInvalidHandler
	}

	// Step 2: Socket instantiation
	if con, e = o.getListen(a); e != nil {
		o.fctError(e)
		return e
	}

	// Step 3: Optional tuning hook
	if o.upd != nil {
		o.upd(con)
	}

	// Step 4: State preparation for the new cycle
	o.gon.Store(false)
	o.gnc.Store(make(chan struct{}))

	// Local synchronization channel to detect main loop exit
	done := make(chan struct{})

	// Step 10: Defer cleanup logic
	defer func() {
		if con != nil {
			_ = con.Close()
		}

		if sx != nil {
			_ = sx.Close()
		}

		o.run.Store(false)
		o.setGone() // Signal to all monitoring goroutines
		close(done)
	}()

	// Step 5: Wrapper creation
	ctx, cnl = context.WithCancel(ctx)
	sx = &sCtx{
		ctx: ctx,
		cnl: cnl,
		con: con,
	}

	// Cache local address for performance
	if l := con.LocalAddr(); l == nil {
		sx.loc = ""
	} else {
		sx.loc = l.String()
	}

	// Step 6: Dedicated monitor goroutine for unblocking
	// This goroutine listens for external shutdown signals (Shutdown/Close/CtxCancel)
	// and forcefully closes the socket to unblock any pending I/O operations.
	go func() {
		select {
		case <-ctx.Done(): // Context cancelled from outside
		case <-o.getGoneChan(): // Server.Shutdown() or Server.Close() called
		case <-done: // Listen() loop finished for other reasons (prevents leak)
			return
		}

		// Force-close the socket to ensure the main handler exits Read()
		if con != nil {
			_ = con.Close()
		}
	}()

	// Step 8: Mark server as active
	o.run.Store(true)

	// Step 7: Handler execution
	// The handler runs in a separate goroutine and is responsible for
	// datagram I/O via the 'sx' context.
	go func(conn net.Conn) {
		defer func() {
			if r := recover(); r != nil {
				librun.RecoveryCaller("golib/socket/server/udp/handler", r)
			}
		}()

		lc := conn.LocalAddr()
		rc := &net.UDPAddr{} // Default remote address for logging (UDP is connectionless)

		if lc == nil {
			lc = &net.UDPAddr{}
		}

		// Report connection lifecycle events
		defer o.fctInfo(lc, rc, libsck.ConnectionClose)
		o.fctInfo(lc, rc, libsck.ConnectionNew)

		// Execute the application business logic
		o.hdl(sx)
	}(con)

	// Step 9: Block and wait for termination
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-o.getGoneChan():
		// Graceful exit triggered by Shutdown() or Close()
		return nil
	case <-done:
		// Termination via internal logic or panic recovery
		return nil
	}
}
