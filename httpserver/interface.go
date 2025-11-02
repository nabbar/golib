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

package httpserver

import (
	"net/http"

	libatm "github.com/nabbar/golib/atomic"
	libctx "github.com/nabbar/golib/context"
	srvtps "github.com/nabbar/golib/httpserver/types"
	liblog "github.com/nabbar/golib/logger"
	logcfg "github.com/nabbar/golib/logger/config"
	montps "github.com/nabbar/golib/monitor/types"
	libsrv "github.com/nabbar/golib/runner"
	librun "github.com/nabbar/golib/runner/startStop"
	libver "github.com/nabbar/golib/version"
)

// Info provides read-only access to server identification and configuration information.
// It exposes essential metadata about the server instance without allowing modifications.
type Info interface {
	// GetName returns the unique identifier name of the server instance.
	GetName() string

	// GetBindable returns the local bind address (host:port) the server listens on.
	GetBindable() string

	// GetExpose returns the public-facing URL used to access this server externally.
	GetExpose() string

	// IsDisable returns true if the server is configured as disabled and should not start.
	IsDisable() bool

	// IsTLS returns true if the server is configured to use TLS/HTTPS.
	IsTLS() bool
}

// Server defines the complete interface for managing an HTTP server instance.
// It extends libsrv.Runner with HTTP-specific functionality including configuration,
// handler management, and monitoring capabilities. All operations are thread-safe.
type Server interface {
	// Server embeds the base server interface providing lifecycle management
	// methods: Start, Stop, Restart, IsRunning, Uptime, etc.
	libsrv.Runner

	// Info embeds server information access methods
	Info

	// Handler registers or updates the handler function that provides HTTP handlers.
	// The function should return a map of handler keys to http.Handler instances.
	Handler(h srvtps.FuncHandler)

	// Merge combines configuration from another server instance into this one.
	// This is useful for updating configuration dynamically without recreating the server.
	Merge(s Server, def liblog.FuncLog) error

	// GetConfig returns the current server configuration.
	// The returned pointer should not be modified directly; use SetConfig instead.
	GetConfig() *Config

	// SetConfig updates the server configuration with validation.
	// The server must be stopped before calling SetConfig to apply changes.
	SetConfig(cfg Config, defLog liblog.FuncLog) error

	// Monitor returns monitoring data for the server including health and metrics.
	// Requires a version parameter for versioning the monitoring data format.
	Monitor(vrs libver.Version) (montps.Monitor, error)

	// MonitorName returns the unique monitoring identifier for this server instance.
	MonitorName() string
}

// New creates and initializes a new HTTP server instance from the provided configuration.
// The configuration is validated before creating the server. Returns an error if
// configuration validation fails or if server initialization encounters an error.
//
// Parameters:
//   - cfg: Server configuration with all required fields populated
//   - defLog: Optional default logger function (can be nil)
//
// Returns:
//   - Server: Initialized server instance ready to start
//   - error: Configuration validation or initialization error
//
// Example:
//
//	cfg := httpserver.Config{
//	    Name:   "api-server",
//	    Listen: "127.0.0.1:8080",
//	    Expose: "http://localhost:8080",
//	}
//	cfg.RegisterHandlerFunc(handlerFunc)
//	srv, err := httpserver.New(cfg, nil)
func New(cfg Config, defLog liblog.FuncLog) (Server, error) {
	s := &srv{
		c: libctx.New[string](cfg.getParentContext),
		h: libatm.NewValue[srvtps.FuncHandler](),
		l: libatm.NewValue[liblog.FuncLog](),
		r: libatm.NewValue[librun.StartStop](),
		s: libatm.NewValue[*http.Server](),
	}

	_ = s.setLogger(nil, logcfg.Options{})
	s.Handler(cfg.getHandlerFunc)

	if e := s.SetConfig(cfg, defLog); e != nil {
		return nil, e
	}

	return s, nil
}
