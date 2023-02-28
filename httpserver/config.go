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
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	srvtps "github.com/nabbar/golib/httpserver/types"

	moncfg "github.com/nabbar/golib/monitor/types"

	libval "github.com/go-playground/validator/v10"
	libtls "github.com/nabbar/golib/certificates"
	libctx "github.com/nabbar/golib/context"
	liberr "github.com/nabbar/golib/errors"
	liblog "github.com/nabbar/golib/logger"
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

// nolint #maligned
type Config struct {

	// Name is the name of the current srv
	// the configuration allow multipke srv, which each one must be identify by a name
	// If not defined, will use the listen address
	Name string `mapstructure:"name" json:"name" yaml:"name" toml:"name" validate:"required"`

	// Listen is the local address (ip, hostname, unix socket, ...) with a port
	// The srv will bind with this address only and listen for the port defined
	Listen string `mapstructure:"listen" json:"listen" yaml:"listen" toml:"listen" validate:"required,hostname_port"`

	// Expose is the address use to call this srv. This can be allow to use a single fqdn to multiple srv"
	Expose string `mapstructure:"expose" json:"expose" yaml:"expose" toml:"expose" validate:"required,url"`

	// HandlerKey is an options to associate current srv with a specifc handler defined by the key
	// This key allow to defined multiple srv in only one config for different handler to start multiple api
	HandlerKey string `mapstructure:"handler_key" json:"handler_key" yaml:"handler_key" toml:"handler_key"`

	//private
	getTLSDefault libtls.FctTLSDefault

	//private
	getParentContext libctx.FuncContext

	//private
	getHandlerFunc srvtps.FuncHandler

	// Enabled allow to disable a srv without clean his configuration
	Disabled bool `mapstructure:"disabled" json:"disabled" yaml:"disabled" toml:"disabled"`

	// Monitor defined the monitoring options to monitor the status & metrics about the health of this srv
	Monitor moncfg.Config `mapstructure:"monitor" json:"monitor" yaml:"monitor" toml:"monitor"`

	// TLSMandatory is a flag to defined that TLS must be valid to start current srv.
	TLSMandatory bool `mapstructure:"tls_mandatory" json:"tls_mandatory" yaml:"tls_mandatory" toml:"tls_mandatory"`

	// TLS is the tls configuration for this srv.
	// To allow tls on this srv, at least the TLS Config option InheritDefault must be at true and the default TLS config must be set.
	// If you don't want any tls config, just omit or set an empty struct.
	TLS libtls.Config `mapstructure:"tls" json:"tls" yaml:"tls" toml:"tls"`

	/*** http options ***/

	// ReadTimeout is the maximum duration for reading the entire
	// request, including the body.
	//
	// Because ReadTimeout does not let Handlers make per-request
	// decisions on each request body's acceptable deadline or
	// upload rate, most users will prefer to use
	// ReadHeaderTimeout. It is valid to use them both.
	ReadTimeout time.Duration `mapstructure:"read_timeout" json:"read_timeout" yaml:"read_timeout" toml:"read_timeout"`

	// ReadHeaderTimeout is the amount of time allowed to read
	// request headers. The connection's read deadline is reset
	// after reading the headers and the Handler can decide what
	// is considered too slow for the body. If ReadHeaderTimeout
	// is zero, the value of ReadTimeout is used. If both are
	// zero, there is no timeout.
	ReadHeaderTimeout time.Duration `mapstructure:"read_header_timeout" json:"read_header_timeout" yaml:"read_header_timeout" toml:"read_header_timeout"`

	// WriteTimeout is the maximum duration before timing out
	// writes of the response. It is reset whenever a new
	// request's header is read. Like ReadTimeout, it does not
	// let Handlers make decisions on a per-request basis.
	WriteTimeout time.Duration `mapstructure:"write_timeout" json:"write_timeout" yaml:"write_timeout" toml:"write_timeout"`

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
	IdleTimeout time.Duration `mapstructure:"idle_timeout" json:"idle_timeout" yaml:"idle_timeout" toml:"idle_timeout"`

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
	Logger liblog.Options `mapstructure:"logger" json:"logger" yaml:"logger" toml:"logger"`
}

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
			RootCAString:         c.TLS.RootCAString,
			RootCAFile:           c.TLS.RootCAFile,
			ClientCAString:       c.TLS.ClientCAString,
			ClientCAFiles:        c.TLS.ClientCAFiles,
			CertPairString:       c.TLS.CertPairString,
			CertPairFile:         c.TLS.CertPairFile,
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

func (c *Config) RegisterHandlerFunc(hdl srvtps.FuncHandler) {
	c.getHandlerFunc = hdl
}

func (c *Config) SetDefaultTLS(f libtls.FctTLSDefault) {
	c.getTLSDefault = f
}

func (c *Config) SetContext(f libctx.FuncContext) {
	c.getParentContext = f
}

func (c *Config) GetTLS() (libtls.TLSConfig, liberr.Error) {
	var def libtls.TLSConfig

	if c.TLS.InheritDefault && c.getTLSDefault != nil {
		def = c.getTLSDefault()
	}

	return c.TLS.NewFrom(def)
}

func (c *Config) CheckTLS() (libtls.TLSConfig, liberr.Error) {
	if ssl, err := c.GetTLS(); err != nil {
		return nil, err
	} else if ssl == nil || ssl.LenCertificatePair() < 1 {
		return nil, ErrorServerValidate.ErrorParent(fmt.Errorf("not certificates defined"))
	} else {
		return ssl, nil
	}
}

func (c *Config) IsTLS() bool {
	if _, err := c.CheckTLS(); err == nil {
		return true
	}

	return false
}

func (c *Config) context() context.Context {
	var ctx context.Context

	if c.getParentContext != nil {
		ctx = c.getParentContext()
	}

	if ctx == nil {
		return context.Background()
	}

	return ctx
}

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

func (c *Config) GetHandlerKey() string {
	return c.HandlerKey
}

func (c *Config) Validate() liberr.Error {
	err := ErrorServerValidate.Error(nil)

	if er := libval.New().Struct(c); er != nil {
		if e, ok := er.(*libval.InvalidValidationError); ok {
			err.AddParent(e)
		}

		for _, e := range er.(libval.ValidationErrors) {
			//nolint goerr113
			err.AddParent(fmt.Errorf("config field '%s' is not validated by constraint '%s'", e.Namespace(), e.ActualTag()))
		}
	}

	if err.HasParent() {
		return err
	}

	return nil

}

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
		ReadTimeout:                  cfg.ReadTimeout,
		ReadHeaderTimeout:            cfg.ReadHeaderTimeout,
		WriteTimeout:                 cfg.WriteTimeout,
		MaxHeaderBytes:               cfg.MaxHeaderBytes,
		MaxHandlers:                  cfg.MaxHandlers,
		MaxConcurrentStreams:         cfg.MaxConcurrentStreams,
		MaxReadFrameSize:             cfg.MaxReadFrameSize,
		PermitProhibitedCipherSuites: cfg.PermitProhibitedCipherSuites,
		IdleTimeout:                  cfg.IdleTimeout,
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
		return ErrorServerValidate.ErrorParent(fmt.Errorf("handler is missing or not existing"))
	}

	o.c.Store(cfgName, cfg.Name)
	o.c.Store(cfgListen, cfg.GetListen())
	o.c.Store(cfgExpose, cfg.GetExpose())
	o.c.Store(cfgDisabled, cfg.Disabled)
	o.c.Store(cfgServerOptions, o.makeOptServer(cfg))
	o.c.Store(cfgConfig, cfg)

	return nil
}

func (o *srv) setLogger(def liblog.FuncLog, opt liblog.Options) error {
	o.m.Lock()
	defer o.m.Unlock()

	var l liblog.Logger

	if def != nil {
		if n := def(); n != nil {
			l = n
		}
	}

	if l == nil {
		l = liblog.GetDefault()
	}

	if e := l.SetOptions(&opt); e == nil {
		o.l = func() liblog.Logger {
			return l
		}
		return nil
	} else if o.l == nil {
		o.l = liblog.GetDefault
		return e
	} else {
		return e
	}
}

func (o *srv) logger() liblog.Logger {
	o.m.RLock()
	defer o.m.RUnlock()

	var log liblog.Logger

	if o.l != nil {
		log = o.l()
	}

	if log == nil {
		log = liblog.GetDefault()
	}

	log.SetFields(log.GetFields().Add("bind", o.GetBindable()))
	return log
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
