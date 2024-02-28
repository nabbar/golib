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

package dns_mapper

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	libval "github.com/go-playground/validator/v10"
	libtls "github.com/nabbar/golib/certificates"
	cmptls "github.com/nabbar/golib/config/components/tls"
	cfgcst "github.com/nabbar/golib/config/const"
	libdur "github.com/nabbar/golib/duration"
	liberr "github.com/nabbar/golib/errors"
)

type TransportConfig struct {
	Proxy     *url.URL       `json:"proxy,omitempty" yaml:"proxy,omitempty" toml:"proxy,omitempty" mapstructure:"proxy,omitempty"`
	TLSConfig *libtls.Config `json:"tls-config,omitempty" yaml:"tls-config,omitempty" toml:"tls-config,omitempty" mapstructure:"tls-config,omitempty"`

	DisableHTTP2       bool `json:"disable-http2" yaml:"disable-http2" toml:"disable-http2" mapstructure:"disable-http2"`
	DisableKeepAlive   bool `json:"disable-keepalive" yaml:"disable-keepalive" toml:"disable-keepalive" mapstructure:"disable-keepalive"`
	DisableCompression bool `json:"disable-compression" yaml:"disable-compression" toml:"disable-compression" mapstructure:"disable-compression"`

	MaxIdleConns        int `json:"max-idle-conns" yaml:"max-idle-conns" toml:"max-idle-conns" mapstructure:"max-idle-conns"`
	MaxIdleConnsPerHost int `json:"max-idle-conns-per-host" yaml:"max-idle-conns-per-host" toml:"max-idle-conns-per-host" mapstructure:"max-idle-conns-per-host"`
	MaxConnsPerHost     int `json:"max-conns-per-host" yaml:"max-conns-per-host" toml:"max-conns-per-host" mapstructure:"max-conns-per-host"`

	TimeoutGlobal         libdur.Duration `json:"timeout-global,omitempty" yaml:"timeout-global,omitempty" toml:"timeout-global,omitempty" mapstructure:"timeout-global,omitempty"`
	TimeoutKeepAlive      libdur.Duration `json:"timeout-keepalive,omitempty" yaml:"timeout-keepalive" toml:"timeout-keepalive" mapstructure:"timeout-keepalive,omitempty"`
	TimeoutTLSHandshake   libdur.Duration `json:"timeout-tls-handshake,omitempty" yaml:"timeout-tls-handshake,omitempty" toml:"timeout-tls-handshake,omitempty" mapstructure:"timeout-tls-handshake,omitempty"`
	TimeoutExpectContinue libdur.Duration `json:"timeout-expect-continue,omitempty" yaml:"timeout-expect-continue,omitempty" toml:"timeout-expect-continue,omitempty" mapstructure:"timeout-expect-continue,omitempty"`
	TimeoutIdleConn       libdur.Duration `json:"timeout-idle-conn,omitempty" yaml:"timeout-idle-conn,omitempty" toml:"timeout-idle-conn,omitempty" mapstructure:"timeout-idle-conn,omitempty"`
	TimeoutResponseHeader libdur.Duration `json:"timeout-response-header,omitempty" yaml:"timeout-response-headern,omitempty" toml:"timeout-response-header,omitempty" mapstructure:"timeout-response-header,omitempty"`
}

type Config struct {
	DNSMapper  map[string]string `json:"dns-mapper,omitempty" yaml:"dns-mapper,omitempty" toml:"dns-mapper,omitempty" mapstructure:"dns-mapper,omitempty"`
	TimerClean libdur.Duration   `json:"timer-clean,omitempty" yaml:"timer-clean,omitempty" toml:"timer-clean,omitempty" mapstructure:"timer-clean,omitempty"`
	Transport  TransportConfig   `json:"transport,omitempty" yaml:"transport,omitempty" toml:"transport,omitempty" mapstructure:"transport,omitempty"`
}

func DefaultConfig(indent string) []byte {
	var (
		res = bytes.NewBuffer(make([]byte, 0))
		def = []byte(`{
  "dns-mapper": {
    "localhost":"127.0.0.1"
  },
  "timer-clean": "3m",
  "transport": {
    "proxy": null,
    "tls-config": ` + string(cmptls.DefaultConfig(cfgcst.JSONIndent)) + `,
    "disable-http2": false,
    "disable-keepalive": false,
    "disable-compression": false,
    "max-idle-conns": 50,
    "max-idle-conns-per-host": 5,
    "max-conns-per-host": 25,
    "timeout-global": "30s",
    "timeout-keepalive": "15s",
    "timeout-tls-handshake": "10s",
    "timeout-expect-continue": "3s",
    "timeout-idle-conn": "30s",
    "timeout-response-header": "0s"
  }
}`)
	)
	if err := json.Indent(res, def, indent, cfgcst.JSONIndent); err != nil {
		return def
	} else {
		return res.Bytes()
	}
}

func (o Config) Validate() liberr.Error {
	var e = ErrorValidatorError.Error(nil)

	if err := libval.New().Struct(o); err != nil {
		if er, ok := err.(*libval.InvalidValidationError); ok {
			e.Add(er)
		}

		for _, er := range err.(libval.ValidationErrors) {
			//nolint #goerr113
			e.Add(fmt.Errorf("config field '%s' is not validated by constraint '%s'", er.Namespace(), er.ActualTag()))
		}
	}

	if !e.HasParent() {
		e = nil
	}

	return e
}

func (o Config) New(ctx context.Context, fct libtls.FctRootCA, msg FuncMessage) DNSMapper {
	return New(ctx, &o, fct, msg)
}
