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
	"fmt"
	"net"
	"sync/atomic"
	"time"

	libatm "github.com/nabbar/golib/atomic"
	libtls "github.com/nabbar/golib/certificates"
	libptc "github.com/nabbar/golib/network/protocol"
	librun "github.com/nabbar/golib/runner"
	libsck "github.com/nabbar/golib/socket"
	sckcfg "github.com/nabbar/golib/socket/config"
)

// srv is the internal implementation of the ServerTcp interface that provides
// a concurrent TCP server with support for TLS, connection management, and
// graceful shutdown.
//
// # Thread Safety
//
// The srv type is designed to be safe for concurrent use. It uses atomic
// operations and immutable state to ensure thread safety without explicit
// locking in most cases. The following guarantees are provided:
//
//   - All exported methods are safe for concurrent calls from multiple goroutines
//   - Connection handling is isolated per connection with minimal shared state
//   - Atomic counters are used for connection tracking and state management
//   - The server can be safely started and stopped from different goroutines
//
// # Lifecycle Management
//
// The server follows a strict lifecycle:
//
//  1. Creation: New() initializes the server with configuration
//  2. Configuration: Optional TLS and callback registration
//  3. Listening: Listen() starts accepting connections
//  4. Running: Handles client connections concurrently
//  5. Shutdown: Graceful or immediate shutdown of all connections
//
// # Resource Management
//
// The server manages the following resources:
//   - Network listener (closed on shutdown)
//   - Active connections (tracked and closed on shutdown)
//   - Goroutines for connection handling (cleaned up on shutdown)
//   - TLS configuration (if enabled)
//
// It is the caller's responsibility to ensure proper cleanup by calling
// Shutdown() or Close() when the server is no longer needed.
type srv struct {
	ssl libatm.Value[*tls.Config] // TLS configuration (*tls.Config)
	upd libsck.UpdateConn         // Connection update callback (optional)
	hdl libsck.HandlerFunc        // Connection handler function (required)
	idl time.Duration             // idle connection timeout
	run *atomic.Bool              // Server is accepting connections flag
	gon *atomic.Bool              // Server is draining connections flag

	fe libatm.Value[libsck.FuncError]   // Error callback (FuncError)
	fi libatm.Value[libsck.FuncInfo]    // Connection info callback (FuncInfo)
	fs libatm.Value[libsck.FuncInfoSrv] // Server info callback (FuncInfoSrv)

	ad libatm.Value[string] // Server listen address (string)
	nc *atomic.Int64        // Active connection counter
}

func (o *srv) Listener() (network libptc.NetworkProtocol, listener string, tls bool) {
	if t := o.getTLS(); t != nil {
		if len(t.Certificates) > 0 {
			return libptc.NetworkTCP, o.getAddress(), true
		}
	}

	return libptc.NetworkTCP, o.getAddress(), false
}

// OpenConnections returns the current number of active client connections.
// This count is atomically maintained and safe to call from multiple goroutines.
//
// The count is incremented when a new connection is accepted and decremented
// when the connection is fully closed (both read and write sides). This can be
// used for monitoring and enforcing connection limits.
//
// # Performance
//
// This method uses atomic operations and is designed to be efficient even with
// a large number of concurrent connections. It does not block and has a
// constant time complexity of O(1).
//
// # Example
//
//	// Monitor active connections
//	go func() {
//	    ticker := time.NewTicker(10 * time.Second)
//	    defer ticker.Stop()
//
//	    for range ticker.C {
//	        count := server.OpenConnections()
//	        log.Printf("Active connections: %d", count)
//	    }
//	}()
//
// # Thread Safety
//
// This method is safe to call from multiple goroutines concurrently.
func (o *srv) OpenConnections() int64 {
	return o.nc.Load()
}

// IsRunning reports whether the server is currently accepting new connections.
//
// Returns:
//   - bool: true if the server is active and accepting connections,
//     false otherwise (not started, shutting down, or stopped)
//
// # State Transitions
//
// The server's running state follows this lifecycle:
//   - false → true: When Listen() is called successfully
//   - true → false: When Shutdown(), StopListen(), or Close() is called
//
// # Usage Notes
//
//   - This method is useful for health checks and monitoring
//   - A return value of false does not necessarily indicate an error condition;
//     it may simply mean the server is in the process of shutting down
//   - For a complete picture of server state, also check IsGone()
//
// # Example
//
//	if !server.IsRunning() {
//	    log.Println("Server is not running, attempting to start...")
//	    if err := server.Listen(ctx); err != nil {
//	        return fmt.Errorf("failed to start server: %w", err)
//	    }
//	}
//
// # Thread Safety
//
// This method uses atomic operations and is safe to call from multiple goroutines.
func (o *srv) IsRunning() bool {
	return o.run.Load()
}

// IsGone returns true if the server is in connection draining mode.
// When true, no new connections will be accepted and existing connections
// are being closed or allowed to finish gracefully.
//
// This state is set by calling StopGone() or Shutdown().
func (o *srv) IsGone() bool {
	return o.gon.Load()
}

/*
// Gone returns a channel that is closed when all connections have been closed
// and the server is fully shutdown. This happens after Done() is closed.
//
// Use this to wait for complete connection draining during graceful shutdown.
// Returns a pre-closed channel if the server is already gone or nil.
func (o *srv) Gone() <-chan struct{} {
	if o == nil {
		return closedChanStruct
	} else if o.IsGone() {
		return closedChanStruct
	} else if i := o.rst.Load(); i != nil && i != closedChanStruct {
		return i
	} else {
		return closedChanStruct
	}
}
*/

// Close performs an immediate shutdown of the server using a background context.
// This is equivalent to calling Shutdown(context.Background()).
//
// For controlled shutdown with a custom timeout, use Shutdown() directly.
func (o *srv) Close() error {
	return o.Shutdown(context.Background())
}

// Shutdown performs a graceful server shutdown by first draining connections
// (StopGone) and then stopping the listener (StopListen).
//
// The method applies a 25-second timeout to the provided context and calls:
//  1. StopGone() - Wait for all connections to close
//  2. StopListen() - Stop accepting new connections
//
// Returns the error from StopListen() if it fails, otherwise returns the error
// from StopGone(). This ensures that listener errors take precedence.
//
// For immediate shutdown without waiting for connections, use Close() instead.
func (o *srv) Shutdown(ctx context.Context) error {
	if o == nil {
		return ErrInvalidInstance
	} else if !o.IsRunning() || o.IsGone() {
		return nil
	}

	o.gon.Store(true)

	var (
		tck = time.NewTicker(3 * time.Millisecond)
		cnl context.CancelFunc
	)

	ctx, cnl = context.WithTimeout(ctx, time.Second)
	defer func() {
		tck.Stop()
		cnl()
	}()

	for o.IsRunning() || o.OpenConnections() > 0 {
		select {
		case <-ctx.Done():
			return ErrShutdownTimeout
		case <-tck.C:
			break // nolint
		}
	}

	return nil
}

// SetTLS configures TLS/SSL encryption for the server.
//
// Parameters:
//   - enable: If false, sets a default TLS config with TLS 1.2-1.3 support but no certificates.
//     If true, validates and applies the provided config.
//   - config: TLS configuration from github.com/nabbar/golib/certificates.
//     Must contain at least one certificate pair when enable is true.
//
// This method must be called before Listen() to enable TLS. When enabled, the server
// will only accept TLS connections.
//
// Returns an error if:
//   - config is nil when enable is true
//   - config has no certificate pairs
//   - config.TlsConfig() returns nil
//
// See github.com/nabbar/golib/certificates.TLSConfig for config creation.
func (o *srv) SetTLS(enable bool, config libtls.TLSConfig) error {
	if !enable {
		// Store default TLS config without certificates
		o.ssl.Store(nil)
		return nil
	}

	// Validate TLS config and certificates
	if config == nil {
		return sckcfg.ErrInvalidTLSConfig
	} else if l := config.GetCertificatePair(); len(l) < 1 {
		return sckcfg.ErrInvalidTLSConfig
	} else if t := config.TlsConfig(""); t == nil {
		return sckcfg.ErrInvalidTLSConfig
	} else {
		o.ssl.Store(t)
		return nil
	}
}

// RegisterFuncError registers a callback function for error notifications.
// The callback is invoked whenever an error occurs during server operation,
// including connection errors, I/O errors, and listener errors.
//
// The function receives variadic errors and should not block as it's called
// from various goroutines. Pass nil to clear the callback.
//
// Thread-safe and can be called at any time, even while the server is running.
//
// See github.com/nabbar/golib/socket.FuncError for the callback signature.
func (o *srv) RegisterFuncError(f libsck.FuncError) {
	if o == nil {
		return
	}

	o.fe.Store(f)
}

// RegisterFuncInfo registers a callback function for connection state changes.
// The callback is invoked for each connection event:
//   - ConnectionNew: New connection accepted
//   - ConnectionRead: Data read from connection
//   - ConnectionWrite: Data written to connection
//   - ConnectionCloseRead: Read side closed
//   - ConnectionCloseWrite: Write side closed
//   - ConnectionClose: Connection fully closed
//
// The function receives local and remote addresses and the connection state.
// Should not block as it's called from connection handler goroutines.
// Pass nil to clear the callback.
//
// See github.com/nabbar/golib/socket.FuncInfo and ConnState for details.
func (o *srv) RegisterFuncInfo(f libsck.FuncInfo) {
	if o == nil {
		return
	}

	o.fi.Store(f)
}

// RegisterFuncInfoServer registers a callback function for server informational messages.
// The callback receives formatted string messages about server lifecycle events:
//   - Server starting/stopping
//   - Listener creation/closure
//   - Configuration changes
//
// Should not block as it's called from the server's main goroutines.
// Pass nil to clear the callback.
//
// See github.com/nabbar/golib/socket.FuncInfoSrv for the callback signature.
func (o *srv) RegisterFuncInfoServer(f libsck.FuncInfoSrv) {
	if o == nil {
		return
	}

	o.fs.Store(f)
}

// RegisterServer sets the TCP address for the server to listen on.
// Must be called before Listen().
//
// Address format:
//   - "host:port" - Listen on specific host (e.g., "localhost:8080")
//   - ":port" - Listen on all interfaces (e.g., ":8080")
//   - "0.0.0.0:port" - Explicitly bind to all IPv4 interfaces
//
// The address is validated using net.ResolveTCPAddr to ensure it's well-formed.
//
// Returns ErrInvalidAddress if the address is empty or cannot be parsed.
func (o *srv) RegisterServer(address string) error {
	if len(address) < 1 {
		return ErrInvalidAddress
	} else if _, err := net.ResolveTCPAddr(libptc.NetworkTCP.Code(), address); err != nil {
		return err
	}

	o.ad.Store(address)
	return nil
}

// fctError invokes the registered error callback if one exists.
// Safely handles nil server instances and nil errors.
// This is an internal helper used throughout the server for error reporting.
func (o *srv) fctError(e ...error) {
	defer func() {
		if r := recover(); r != nil {
			librun.RecoveryCaller("golib/socket/server/tcp/fctError", r)
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

// fctInfo invokes the registered connection info callback if one exists.
// Reports connection state changes with local and remote addresses.
// Safely handles nil callbacks to prevent panics.
// This is an internal helper called from connection lifecycle events.
func (o *srv) fctInfo(local, remote net.Addr, state libsck.ConnState) {
	defer func() {
		if r := recover(); r != nil {
			librun.RecoveryCaller("golib/socket/server/tcp/fctInfo", r)
		}
	}()

	if o == nil {
		return
	} else if f := o.fi.Load(); f != nil {
		f(local, remote, state)
	}
}

// fctInfoSrv invokes the registered server info callback if one exists.
// Formats the message with fmt.Sprintf before passing to the callback.
// Safely handles nil callbacks to prevent panics.
// This is an internal helper for server lifecycle logging.
func (o *srv) fctInfoSrv(msg string, args ...interface{}) {
	defer func() {
		if r := recover(); r != nil {
			librun.RecoveryCaller("golib/socket/server/tcp/fctInfoSrv", r)
		}
	}()

	if o == nil {
		return
	} else if f := o.fs.Load(); f != nil {
		f(fmt.Sprintf(msg, args...))
	}
}

// getTLS retrieves the current TLS configuration if TLS is enabled.
// Returns nil if:
//   - No TLS config has been set
//   - The stored config is not a valid *tls.Config
//   - The config has no certificates
//
// This is an internal helper used by the listener to determine if TLS should be used.
func (o *srv) getTLS() *tls.Config {
	i := o.ssl.Load()

	if i == nil {
		return nil
	} else if len(i.Certificates) < 1 {
		return nil
	} else {
		return i
	}
}

// idleTimeout returns the configured idle timeout duration for connections.
// This is an internal helper used by connection handling to determine if
// idle timeout monitoring should be enabled.
//
// Returns:
//   - time.Duration: The idle timeout duration, or 0 if disabled
//
// # Behavior
//
// The method returns 0 (disabled) in the following cases:
//   - If the server instance is nil
//   - If the configured timeout is less than 1 second
//
// Otherwise, it returns the configured idle timeout value.
//
// # Usage
//
// This method is called during connection setup to configure the idle
// timeout timer. A return value of 0 indicates that idle timeout
// monitoring should be disabled for the connection.
//
// The minimum threshold of 1 second prevents overly aggressive timeout
// values that could interfere with normal operation.
func (o *srv) idleTimeout() time.Duration {
	if o == nil {
		return 0
	} else if o.idl < time.Second {
		return 0
	} else {
		return o.idl
	}
}
