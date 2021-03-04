/*
 * MIT License
 *
 * Copyright (c) 2020 Nicolas JUHEL
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

	"github.com/go-playground/validator/v10"

	libtls "github.com/nabbar/golib/certificates"
	liberr "github.com/nabbar/golib/errors"
)

type MapUpdPoolServerConfig func(cfg ServerConfig) ServerConfig
type MapRunPoolServerConfig func(cfg ServerConfig)
type PoolServerConfig []ServerConfig

func (p PoolServerConfig) PoolServer() (PoolServer, liberr.Error) {
	var (
		r = NewPool()
		e = ErrorPoolAdd.Error(nil)
	)

	p.MapRun(func(cfg ServerConfig) {
		var err liberr.Error

		if r, err = r.Add(cfg.Server()); err != nil {
			e.AddParentError(err)
		}
	})

	if !e.HasParent() {
		e = nil
	}

	return r, e
}

func (p PoolServerConfig) UpdatePoolServer(pSrv PoolServer) (PoolServer, liberr.Error) {
	var e = ErrorPoolAdd.Error(nil)

	p.MapRun(func(cfg ServerConfig) {
		var err liberr.Error

		if pSrv, err = pSrv.Add(cfg.Server()); err != nil {
			e.AddParentError(err)
		}
	})

	if !e.HasParent() {
		e = nil
	}

	return pSrv, e
}

func (p PoolServerConfig) Validate() liberr.Error {
	var e = ErrorPoolValidate.Error(nil)

	p.MapRun(func(cfg ServerConfig) {
		var err liberr.Error

		if err = cfg.Validate(); err != nil {
			e.AddParentError(err)
		}
	})

	if !e.HasParent() {
		e = nil
	}

	return e
}

func (p PoolServerConfig) MapUpdate(f MapUpdPoolServerConfig) PoolServerConfig {
	var r = make(PoolServerConfig, 0)

	if p != nil {
		r = p
	}

	for i, c := range r {
		r[i] = f(c)
	}

	return r
}

func (p PoolServerConfig) MapRun(f MapRunPoolServerConfig) PoolServerConfig {
	var r = make(PoolServerConfig, 0)

	if p != nil {
		r = p
	}

	for _, c := range r {
		f(c)
	}

	return r
}

type ServerConfig struct {
	getTLSDefault    func() libtls.TLSConfig
	getParentContext func() context.Context

	// Enabled allow to disable a server without clean his configuration
	Disabled bool

	// Mandatory defined if the component for status is mandatory or not
	Mandatory bool

	// TimeoutCacheInfo defined the validity time of cache for info (name, version, hash)
	TimeoutCacheInfo time.Duration

	// TimeoutCacheHealth defined the validity time of cache for healthcheck of this server
	TimeoutCacheHealth time.Duration

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
	// server will read parsing the request header's keys and
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
	MaxConcurrentStreams uint32 `json:"max_concurrent_streams" json:"max_concurrent_streams" yaml:"max_concurrent_streams" toml:"max_concurrent_streams"`

	// MaxReadFrameSize optionally specifies the largest frame
	// this server is willing to read. A valid value is between
	// 16k and 16M, inclusive. If zero or otherwise invalid, a
	// default value is used.
	MaxReadFrameSize uint32 `json:"max_read_frame_size" json:"max_read_frame_size" yaml:"max_read_frame_size" toml:"max_read_frame_size"`

	// PermitProhibitedCipherSuites, if true, permits the use of
	// cipher suites prohibited by the HTTP/2 spec.
	PermitProhibitedCipherSuites bool `json:"permit_prohibited_cipher_suites" json:"permit_prohibited_cipher_suites" yaml:"permit_prohibited_cipher_suites" toml:"permit_prohibited_cipher_suites"`

	// IdleTimeout specifies how long until idle clients should be
	// closed with a GOAWAY frame. PING frames are not considered
	// activity for the purposes of IdleTimeout.
	IdleTimeout time.Duration `json:"idle_timeout" json:"idle_timeout" yaml:"idle_timeout" toml:"idle_timeout"`

	// MaxUploadBufferPerConnection is the size of the initial flow
	// control window for each connections. The HTTP/2 spec does not
	// allow this to be smaller than 65535 or larger than 2^32-1.
	// If the value is outside this range, a default value will be
	// used instead.
	MaxUploadBufferPerConnection int32 `json:"max_upload_buffer_per_connection" json:"max_upload_buffer_per_connection" yaml:"max_upload_buffer_per_connection" toml:"max_upload_buffer_per_connection"`

	// MaxUploadBufferPerStream is the size of the initial flow control
	// window for each stream. The HTTP/2 spec does not allow this to
	// be larger than 2^32-1. If the value is zero or larger than the
	// maximum, a default value will be used instead.
	MaxUploadBufferPerStream int32 `json:"max_upload_buffer_per_stream" json:"max_upload_buffer_per_stream" yaml:"max_upload_buffer_per_stream" toml:"max_upload_buffer_per_stream"`

	// Name is the name of the current server
	// the configuration allow multipke server, which each one must be identify by a name
	// If not defined, will use the listen address
	Name string `mapstructure:"name" json:"name" yaml:"name" toml:"name" validate:"required"`

	// Listen is the local address (ip, hostname, unix socket, ...) with a port
	// The server will bind with this address only and listen for the port defined
	Listen string `mapstructure:"listen" json:"listen" yaml:"listen" toml:"listen" validate:"required,hostname_port"`

	// Expose is the address use to call this server. This can be allow to use a single fqdn to multiple server"
	Expose string `mapstructure:"expose" json:"expose" yaml:"expose" toml:"expose" validate:"required,url"`

	// HandlerKeys is an options to associate current server with a specifc handler defined by the key
	// This allow to defined multiple server in only one config for different handler to start multiple api
	HandlerKeys string

	// TLSMandatory is a flag to defined that TLS must be valid to start current server.
	TLSMandatory bool

	// TLS is the tls configuration for this server.
	// To allow tls on this server, at least the TLS Config option InheritDefault must be at true and the default TLS config must be set.
	// If you don't want any tls config, just omit or set an empty struct.
	TLS libtls.Config `mapstructure:"tls" json:"tls" yaml:"tls" toml:"tls"`
}

func (c *ServerConfig) Clone() ServerConfig {
	return ServerConfig{
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
		Name:                         c.Name,
		Listen:                       c.Listen,
		Expose:                       c.Expose,
		HandlerKeys:                  strings.ToLower(c.HandlerKeys),
		TLSMandatory:                 c.TLSMandatory,
		TLS:                          c.TLS,
	}
}

func (c *ServerConfig) SetDefaultTLS(f func() libtls.TLSConfig) {
	c.getTLSDefault = f
}

func (c *ServerConfig) SetParentContext(f func() context.Context) {
	c.getParentContext = f
}

func (c ServerConfig) GetTLS() (libtls.TLSConfig, liberr.Error) {
	var def libtls.TLSConfig

	if c.getTLSDefault != nil {
		def = c.getTLSDefault()
	}

	return c.TLS.NewFrom(def)
}

func (c ServerConfig) IsTLS() bool {
	if ssl, err := c.GetTLS(); err == nil && ssl != nil && ssl.LenCertificatePair() > 0 {
		return true
	}

	return false
}

func (c ServerConfig) getContext() context.Context {
	var ctx context.Context

	if c.getParentContext != nil {
		ctx = c.getParentContext()
	}

	if ctx == nil {
		return context.Background()
	}

	return ctx
}

func (c ServerConfig) GetListen() *url.URL {
	var (
		err error
		add *url.URL
	)

	if c.Listen != "" {
		if add, err = url.Parse(c.Listen); err != nil {
			if host, prt, err := net.SplitHostPort(c.Listen); err == nil {
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

func (c ServerConfig) GetExpose() *url.URL {
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

func (c ServerConfig) GetHandlerKey() string {
	return c.HandlerKeys
}

func (c ServerConfig) Validate() liberr.Error {
	val := validator.New()
	err := val.Struct(c)

	if e, ok := err.(*validator.InvalidValidationError); ok {
		return ErrorServerValidate.ErrorParent(e)
	}

	out := ErrorServerValidate.Error(nil)

	for _, e := range err.(validator.ValidationErrors) {
		//nolint goerr113
		out.AddParent(fmt.Errorf("config field '%s' is not validated by constraint '%s'", e.Field(), e.ActualTag()))
	}

	if out.HasParent() {
		return out
	}

	return nil

}

func (c ServerConfig) Server() Server {
	return NewServer(&c)
}
