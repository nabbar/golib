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

package tcp

import (
	"context"
	"crypto/tls"
	"net"
	"sync"
	"sync/atomic"

	libatm "github.com/nabbar/golib/atomic"
	durbig "github.com/nabbar/golib/duration/big"
	libsck "github.com/nabbar/golib/socket"
	sckcfg "github.com/nabbar/golib/socket/config"
	sckidl "github.com/nabbar/golib/socket/idlemgr"
)

// ServerTcp defines the interface for a high-performance TCP server.
// It extends the base libsck.Server interface with specific TCP functionality.
//
// A ServerTcp provides a concurrent TCP server that handles client connections
// using a customizable handler function. It supports TLS encryption, idle
// connection management, and graceful shutdown.
//
// # Thread Safety
//
// All implementations of ServerTcp MUST be safe for concurrent use by multiple
// goroutines. All methods can be called simultaneously from different threads.
//
// # Lifecycle Management
//
//  1. Creation: New() initializes the server with configuration and handler.
//  2. Configuration: via SetTLS() and RegisterFunc* methods (Error, Info, InfoServer).
//  3. Bind: via RegisterServer() to set the listen address.
//  4. Operation: via Listen() to start accepting connections.
//  5. Shutdown: via Shutdown() (graceful) or Close() (immediate).
//
// # Example Usage (Echo Server)
//
//	hdl := func(ctx libsck.Context) {
//	    defer ctx.Close()
//	    io.Copy(ctx, ctx) // Echo data back
//	}
//	cfg := sckcfg.DefaultServer(":8080")
//	srv, _ := tcp.New(nil, hdl, cfg)
//	srv.Listen(context.Background())
type ServerTcp interface {
	libsck.Server

	// RegisterServer sets the TCP address for the server to listen on.
	// The address should be in "host:port" format (e.g., "localhost:8080" or ":8080").
	// Must be called before Listen(). Returns ErrInvalidAddress if the input 
	// is malformed.
	RegisterServer(address string) error
}

// New creates and initializes a new TCP server instance with the provided configuration.
//
// # Configuration and Initialization Dataflow
//
//  1. Validation: cfg.Validate() ensures basic parameters (address, timeouts) are sound.
//  2. Defaults: Default TLS versions (1.2/1.3) and empty callbacks are set.
//  3. Structure: The srv internal structure is allocated.
//  4. Resource Pooling: The sync.Pool for sCtx recycling is initialized.
//  5. Idle Manager: If ConIdleTimeout > 0, an sckidl.Manager is started to handle timeouts.
//  6. Binding: RegisterServer() is called with the address from the config.
//  7. Security: SetTLS() is called with TLS settings from the config.
//  8. State: gon is set to true (server is ready to be started).
//
// # Parameters
//
//   - upd: Optional callback to configure each net.Conn (e.g., buffer sizes) before handling.
//   - hdl: Required handler function that will be called for each new connection.
//   - cfg: Server configuration structure including address, TLS settings, and timeouts.
//
// # Returns
//
//   - ServerTcp: The initialized server instance.
//   - error: Initialization errors (ErrInvalidHandler, ErrInvalidAddress, sckidl errors).
func New(upd libsck.UpdateConn, hdl libsck.HandlerFunc, cfg sckcfg.Server) (ServerTcp, error) {
	// ... implementation ...
	if e := cfg.Validate(); e != nil {
		return nil, e
	} else if hdl == nil {
		return nil, ErrInvalidHandler
	}

	var (
		ssl = &tls.Config{
			MinVersion: tls.VersionTLS12,
			MaxVersion: tls.VersionTLS13,
		}
		dfe libsck.FuncError   = func(_ ...error) {}
		dfi libsck.FuncInfo    = func(_, _ net.Addr, _ libsck.ConnState) {}
		dfs libsck.FuncInfoSrv = func(_ string) {}
	)

	s := &srv{
		ssl: libatm.NewValueDefault[*tls.Config](ssl, ssl),
		upd: upd,
		hdl: hdl,
		idl: 0,
		run: new(atomic.Bool),
		gon: new(atomic.Bool),
		fe:  libatm.NewValueDefault[libsck.FuncError](dfe, dfe),
		fi:  libatm.NewValueDefault[libsck.FuncInfo](dfi, dfi),
		fs:  libatm.NewValueDefault[libsck.FuncInfoSrv](dfs, dfs),
		ad:  libatm.NewValue[string](),
		gnc: libatm.NewValueDefault[chan struct{}](make(chan struct{}), make(chan struct{})),
		id:  nil,
		nc:  new(atomic.Int64),
		pol: &sync.Pool{
			New: func() interface{} {
				return &sCtx{}
			},
		},
	}

	if c := cfg.ConIdleTimeout.Seconds(); c > 0 {
		s.idl = cfg.ConIdleTimeout.Time()

		i, e := sckidl.New(context.Background(), durbig.Seconds(c), durbig.Seconds(1))
		if e != nil {
			return nil, e
		}
		s.id = i
	}

	if e := s.RegisterServer(cfg.Address); e != nil {
		return nil, e
	}

	k, t := cfg.GetTLS()
	if e := s.SetTLS(k, t); e != nil {
		return nil, e
	}

	s.gon.Store(true)

	return s, nil
}
