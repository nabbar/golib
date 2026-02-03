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

package httpserver

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"

	libval "github.com/go-playground/validator/v10"

	libtls "github.com/nabbar/golib/certificates"
	libdur "github.com/nabbar/golib/duration"
	srvtps "github.com/nabbar/golib/httpserver/types"
	liblog "github.com/nabbar/golib/logger"
	logcfg "github.com/nabbar/golib/logger/config"
	moncfg "github.com/nabbar/golib/monitor/types"
)

const (
	cfgConfig        = "cfgConfig"
	cfgName          = "cfgName"
	cfgListen        = "cfgListen"
	cfgExpose        = "cfgExpose"
	cfgHandler       = "cfgHandler"
	cfgHandlerKey    = "cfgHandlerKey"
	cfgDisabled      = "cfgDisabled"
	cfgMonitor       = "cfgMonitor"
	cfgTLS           = "cfgTLS"
	cfgTLSMandatory  = "cfgTLSMandatory"
	cfgServerOptions = "cfgServerOptions"
)

// Config defines the complete HTTP server configuration including network settings,
// TLS options, timeouts, and HTTP/2 parameters. All fields are serializable to
// various formats (JSON, YAML, TOML) for externalized configuration.
type Config struct {

	// Name is the unique identifier for the server instance.
	// Multiple servers can be configured, each identified by a unique name.
	// If not defined, the listen address is used as the name.
	Name string `mapstructure:"name" json:"name" yaml:"name" toml:"name" validate:"required"`

	// Listen is the local bind address (host:port) for the server.
	// The server will bind to this address and listen for incoming connections.
	// Examples: "127.0.0.1:8080", "0.0.0.0:443", "localhost:3000"
	Listen string `mapstructure:"listen" json:"listen" yaml:"listen" toml:"listen" validate:"required,hostname_port"`

	// Expose is the public-facing URL used to access this server externally.
	// This allows using a single domain with multiple servers on different ports.
	// Examples: "http://localhost:8080", "https://api.example.com"
	Expose string `mapstructure:"expose" json:"expose" yaml:"expose" toml:"expose" validate:"required,url"`

	// HandlerKey associates this server with a specific handler from the handler map.
	// This enables multiple servers to use different handlers from a shared registry,
	// allowing different APIs to run on different ports with a single configuration.
	HandlerKey string `mapstructure:"handler_key" json:"handler_key" yaml:"handler_key" toml:"handler_key"`

	//private
	getTLSDefault libtls.FctTLSDefault

	//private
	getParentContext context.Context

	//private
	getHandlerFunc srvtps.FuncHandler

	// Disabled allows disabling a server without removing its configuration.
	// Useful for maintenance mode or gradual rollout scenarios.
	Disabled bool `mapstructure:"disabled" json:"disabled" yaml:"disabled" toml:"disabled"`

	// Monitor defines the monitoring configuration for health checks and metrics collection.
	// Enables integration with the monitoring system for server health tracking.
	Monitor moncfg.Config `mapstructure:"monitor" json:"monitor" yaml:"monitor" toml:"monitor"`

	// TLSMandatory requires valid TLS configuration for the server to start.
	// If true, the server will fail to start without proper TLS certificates.
	TLSMandatory bool `mapstructure:"tls_mandatory" json:"tls_mandatory" yaml:"tls_mandatory" toml:"tls_mandatory"`

	// TLS is the certificate configuration for HTTPS/TLS support.
	// Set InheritDefault to true to inherit from default TLS config, or provide
	// specific certificate paths. Leave empty to disable TLS for this server.
	TLS libtls.Config `mapstructure:"tls" json:"tls" yaml:"tls" toml:"tls"`

	/*** http options ***/

	// ReadTimeout is the maximum duration for reading the entire
	// request, including the body.
	//
	// Because ReadTimeout does not let Handlers make per-request
	// decisions on each request body's acceptable deadline or
	// upload rate, most users will prefer to use
	// ReadHeaderTimeout. It is valid to use them both.
	ReadTimeout libdur.Duration `mapstructure:"read_timeout" json:"read_timeout" yaml:"read_timeout" toml:"read_timeout"`

	// ReadHeaderTimeout is the amount of time allowed to read
	// request headers. The connection's read deadline is reset
	// after reading the headers and the Handler can decide what
	// is considered too slow for the body. If ReadHeaderTimeout
	// is zero, the value of ReadTimeout is used. If both are
	// zero, there is no timeout.
	ReadHeaderTimeout libdur.Duration `mapstructure:"read_header_timeout" json:"read_header_timeout" yaml:"read_header_timeout" toml:"read_header_timeout"`

	// WriteTimeout is the maximum duration before timing out
	// writes of the response. It is reset whenever a new
	// request's header is read. Like ReadTimeout, it does not
	// let Handlers make decisions on a per-request basis.
	WriteTimeout libdur.Duration `mapstructure:"write_timeout" json:"write_timeout" yaml:"write_timeout" toml:"write_timeout"`

	// MaxHeaderBytes controls the maximum number of bytes the
	// srv will read parsing the request header's keys and
	// values, including the request line. It does not limit the
	// size of the request body.
	// If zero, DefaultMaxHeaderBytes is used.
	MaxHeaderBytes int `mapstructure:"max_header_bytes" json:"max_header_bytes" yaml:"max_header_bytes" toml:"max_header_bytes"`

	/*** http2 options ***/

	// MaxHandlers limits the number of http.Handler ServeHTTP goroutines
	// which may run at a time over all connections.
	// Negative or zero no limit.
	MaxHandlers int `mapstructure:"max_handlers" json:"max_handlers" yaml:"max_handlers" toml:"max_handlers"`

	// MaxConcurrentStreams optionally specifies the number of
	// concurrent streams that each client may have open at a
	// time. This is unrelated to the number of http.Handler goroutines
	// which may be active globally, which is MaxHandlers.
	// If zero, MaxConcurrentStreams defaults to at least 100, per
	// the HTTP/2 spec's recommendations.
	MaxConcurrentStreams uint32 `mapstructure:"max_concurrent_streams" json:"max_concurrent_streams" yaml:"max_concurrent_streams" toml:"max_concurrent_streams"`

	// MaxReadFrameSize optionally specifies the largest frame
	// this srv is willing to read. A valid value is between
	// 16k and 16M, inclusive. If zero or otherwise invalid, a
	// default value is used.
	MaxReadFrameSize uint32 `mapstructure:"max_read_frame_size" json:"max_read_frame_size" yaml:"max_read_frame_size" toml:"max_read_frame_size"`

	// PermitProhibitedCipherSuites, if true, permits the use of
	// cipher suites prohibited by the HTTP/2 spec.
	PermitProhibitedCipherSuites bool `mapstructure:"permit_prohibited_cipher_suites" json:"permit_prohibited_cipher_suites" yaml:"permit_prohibited_cipher_suites" toml:"permit_prohibited_cipher_suites"`

	// IdleTimeout specifies how long until idle clients should be
	// closed with a GOAWAY frame. PING frames are not considered
	// activity for the purposes of IdleTimeout.
	IdleTimeout libdur.Duration `mapstructure:"idle_timeout" json:"idle_timeout" yaml:"idle_timeout" toml:"idle_timeout"`

	// MaxUploadBufferPerConnection is the size of the initial flow
	// control window for each connections. The HTTP/2 spec does not
	// allow this to be smaller than 65535 or larger than 2^32-1.
	// If the value is outside this range, a default value will be
	// used instead.
	MaxUploadBufferPerConnection int32 `mapstructure:"max_upload_buffer_per_connection" json:"max_upload_buffer_per_connection" yaml:"max_upload_buffer_per_connection" toml:"max_upload_buffer_per_connection"`

	// MaxUploadBufferPerStream is the size of the initial flow control
	// window for each stream. The HTTP/2 spec does not allow this to
	// be larger than 2^32-1. If the value is zero or larger than the
	// maximum, a default value will be used instead.
	MaxUploadBufferPerStream int32 `mapstructure:"max_upload_buffer_per_stream" json:"max_upload_buffer_per_stream" yaml:"max_upload_buffer_per_stream" toml:"max_upload_buffer_per_stream"`

	// DisableKeepAlive controls whether HTTP keep-alives are disabled.
	// By default, keep-alives are always enabled. Only very
	// resource-constrained environments or servers in the process of
	// shutting down should disable them.
	DisableKeepAlive bool `mapstructure:"disable_keep_alive" json:"disable_keep_alive" yaml:"disable_keep_alive" toml:"disable_keep_alive"`

	// Logger is used to define the logger options.
	Logger logcfg.Options `mapstructure:"logger" json:"logger" yaml:"logger" toml:"logger"`
}

// Clone creates a deep copy of the Config structure.
// All fields are copied, including function pointers.
func (c *Config) Clone() Config {
	return Config{
		Disabled:                     c.Disabled,
		getTLSDefault:                c.getTLSDefault,
		getParentContext:             c.getParentContext,
		ReadTimeout:                  c.ReadTimeout,
		ReadHeaderTimeout:            c.ReadHeaderTimeout,
		WriteTimeout:                 c.WriteTimeout,
		MaxHeaderBytes:               c.MaxHeaderBytes,
		MaxHandlers:                  c.MaxHandlers,
		MaxConcurrentStreams:         c.MaxConcurrentStreams,
		MaxReadFrameSize:             c.MaxReadFrameSize,
		PermitProhibitedCipherSuites: c.PermitProhibitedCipherSuites,
		IdleTimeout:                  c.IdleTimeout,
		MaxUploadBufferPerConnection: c.MaxUploadBufferPerConnection,
		MaxUploadBufferPerStream:     c.MaxUploadBufferPerStream,
		DisableKeepAlive:             c.DisableKeepAlive,
		Name:                         c.Name,
		Listen:                       c.Listen,
		Expose:                       c.Expose,
		HandlerKey:                   strings.ToLower(c.HandlerKey),
		TLSMandatory:                 c.TLSMandatory,
		TLS: libtls.Config{
			CurveList:            c.TLS.CurveList,
			CipherList:           c.TLS.CipherList,
			RootCA:               c.TLS.RootCA,
			ClientCA:             c.TLS.ClientCA,
			Certs:                c.TLS.Certs,
			VersionMin:           c.TLS.VersionMin,
			VersionMax:           c.TLS.VersionMax,
			AuthClient:           c.TLS.AuthClient,
			InheritDefault:       c.TLS.InheritDefault,
			DynamicSizingDisable: c.TLS.DynamicSizingDisable,
			SessionTicketDisable: c.TLS.SessionTicketDisable,
		},
		Monitor: c.Monitor.Clone(),
	}
}

// RegisterHandlerFunc registers a handler function that provides HTTP handlers.
// The function should return a map of handler keys to http.Handler instances.
func (c *Config) RegisterHandlerFunc(hdl srvtps.FuncHandler) {
	c.getHandlerFunc = hdl
}

// SetDefaultTLS registers a function that provides default TLS configuration.
// This is used when TLS.InheritDefault is enabled.
func (c *Config) SetDefaultTLS(f libtls.FctTLSDefault) {
	c.getTLSDefault = f
}

// SetContext registers a function that provides the parent context for the server.
// The context is used for lifecycle management and cancellation.
func (c *Config) SetContext(f context.Context) {
	c.getParentContext = f
}

// GetTLS returns the TLS configuration for the server.
// If InheritDefault is true, it merges with the default TLS configuration.
func (c *Config) GetTLS() libtls.TLSConfig {
	var def libtls.TLSConfig

	if c.TLS.InheritDefault && c.getTLSDefault != nil {
		def = c.getTLSDefault()
	}

	return c.TLS.NewFrom(def)
}

// CheckTLS validates the TLS configuration and returns it if valid.
// Returns an error if no certificates are defined.
func (c *Config) CheckTLS() (libtls.TLSConfig, error) {
	if ssl := c.GetTLS(); ssl.LenCertificatePair() < 1 {
		return nil, ErrorServerValidate.Error(fmt.Errorf("not certificates defined"))
	} else {
		return ssl, nil
	}
}

// IsTLS returns true if the server has a valid TLS configuration.
func (c *Config) IsTLS() bool {
	if _, err := c.CheckTLS(); err == nil {
		return true
	}

	return false
}

// GetListen parses and returns the listen address as a URL.
// Returns nil if the address is invalid.
func (c *Config) GetListen() *url.URL {
	var (
		err error
		add *url.URL
	)

	if c.Listen != "" {
		if add, err = url.Parse(c.Listen); err != nil {
			if host, prt, e := net.SplitHostPort(c.Listen); e == nil {
				add = &url.URL{
					Host: fmt.Sprintf("%s:%s", host, prt),
				}
			} else {
				add = nil
			}
		}
	}

	if add == nil && c.Expose != "" {
		if add, err = url.Parse(c.Expose); err != nil {
			add = nil
		}
	}

	return add
}

// GetExpose parses and returns the expose address as a URL.
// Falls back to the listen address with appropriate scheme if not set.
func (c *Config) GetExpose() *url.URL {
	var (
		err error
		add *url.URL
	)

	if c.Expose != "" {
		if add, err = url.Parse(c.Expose); err != nil {
			add = nil
		}
	}

	if add == nil {
		if add = c.GetListen(); add != nil {
			if c.IsTLS() {
				add.Scheme = "https"
			} else {
				add.Scheme = "http"
			}
		}
	}

	return add
}

// GetHandlerKey returns the handler key for this server configuration.
// Returns empty string if no specific key is set (uses default handler).
func (c *Config) GetHandlerKey() string {
	return c.HandlerKey
}

// Validate checks if the configuration is valid according to struct tag constraints.
// Returns an error describing all validation failures, or nil if valid.
func (c *Config) Validate() error {
	err := ErrorServerValidate.Error(nil)

	if er := libval.New().Struct(c); er != nil {
		if e, ok := er.(*libval.InvalidValidationError); ok {
			err.Add(e)
		}

		for _, e := range er.(libval.ValidationErrors) {
			//nolint goerr113
			err.Add(fmt.Errorf("config field '%s' is not validated by constraint '%s'", e.Namespace(), e.ActualTag()))
		}
	}

	if err.HasParent() {
		return err
	}

	return nil

}

// Server creates a new HTTP server instance from this configuration.
// This is a convenience method that calls the New function.
func (c *Config) Server(defLog liblog.FuncLog) (Server, error) {
	return New(*c, defLog)
}

func (o *srv) GetConfig() *Config {
	if i, l := o.c.Load(cfgConfig); !l {
		return nil
	} else if v, k := i.(Config); !k {
		return nil
	} else {
		return &v
	}
}

func (o *srv) makeOptServer(cfg Config) *optServer {
	return &optServer{
		ReadTimeout:                  cfg.ReadTimeout.Time(),
		ReadHeaderTimeout:            cfg.ReadHeaderTimeout.Time(),
		WriteTimeout:                 cfg.WriteTimeout.Time(),
		MaxHeaderBytes:               cfg.MaxHeaderBytes,
		MaxHandlers:                  cfg.MaxHandlers,
		MaxConcurrentStreams:         cfg.MaxConcurrentStreams,
		MaxReadFrameSize:             cfg.MaxReadFrameSize,
		PermitProhibitedCipherSuites: cfg.PermitProhibitedCipherSuites,
		IdleTimeout:                  cfg.IdleTimeout.Time(),
		MaxUploadBufferPerConnection: cfg.MaxUploadBufferPerConnection,
		MaxUploadBufferPerStream:     cfg.MaxUploadBufferPerStream,
		DisableKeepAlive:             cfg.DisableKeepAlive,
	}
}

func (o *srv) SetConfig(cfg Config, defLog liblog.FuncLog) error {
	if e := o.cfgSetTLS(&cfg); e != nil {
		return e
	} else if e = o.setLogger(defLog, cfg.Logger); e != nil {
		return e
	}

	if o.HandlerHas(cfg.HandlerKey) {
		o.HandlerStoreFct(cfg.HandlerKey)
	} else {
		return ErrorServerValidate.Error(fmt.Errorf("handler is missing or not existing"))
	}

	o.c.Store(cfgName, cfg.Name)
	o.c.Store(cfgListen, cfg.GetListen())
	o.c.Store(cfgExpose, cfg.GetExpose())
	o.c.Store(cfgDisabled, cfg.Disabled)
	o.c.Store(cfgServerOptions, o.makeOptServer(cfg))
	o.c.Store(cfgConfig, cfg)

	return nil
}

func (o *srv) setLogger(def liblog.FuncLog, opt logcfg.Options) error {
	if o == nil || o.l == nil {
		return ErrorServerValidate.Error(nil)
	}

	var l, e = liblog.NewFrom(o.c, &opt, o.l.Load(), def)

	l.SetFields(l.GetFields().Add("bind", o.GetBindable()))
	o.l.Store(func() liblog.Logger {
		return l
	})

	return e
}

func (o *srv) logger() liblog.Logger {
	if o == nil || o.l == nil {
		return liblog.New(o.c)
	}

	if f := o.l.Load(); f != nil {
		return f()
	}

	l := liblog.New(o.c)
	l.SetFields(l.GetFields().Add("bind", o.GetBindable()))
	return l
}

func (o *srv) cfgSetTLS(cfg *Config) error {
	o.c.Store(cfgTLSMandatory, cfg.TLSMandatory)
	if t, e := cfg.CheckTLS(); e != nil && cfg.TLSMandatory {
		return e
	} else if e != nil {
		o.c.Delete(cfgTLS)
		return nil
	} else {
		o.c.Store(cfgTLS, t)
		return nil
	}
}

func (o *srv) cfgGetTLS() libtls.TLSConfig {
	if i, l := o.c.Load(cfgTLS); !l {
		return nil
	} else if v, k := i.(libtls.TLSConfig); !k {
		return nil
	} else {
		return v
	}
}

func (o *srv) cfgTLSMandatory() bool {
	if i, l := o.c.Load(cfgTLSMandatory); !l {
		return false
	} else if v, k := i.(bool); !k {
		return false
	} else {
		return v
	}
}

func (o *srv) cfgGetServer() *optServer {
	if i, l := o.c.Load(cfgServerOptions); !l {
		return &optServer{}
	} else if v, k := i.(*optServer); !k {
		return &optServer{}
	} else {
		return v
	}
}
