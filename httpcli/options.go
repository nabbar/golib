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

package httpcli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	libval "github.com/go-playground/validator/v10"
	libtls "github.com/nabbar/golib/certificates"
	cmptls "github.com/nabbar/golib/config/components/tls"
	cfgcst "github.com/nabbar/golib/config/const"
	liberr "github.com/nabbar/golib/errors"
	libptc "github.com/nabbar/golib/network/protocol"
)

type OptionForceIP struct {
	Enable bool                   `json:"enable" yaml:"enable" toml:"enable" mapstructure:"enable"`
	Net    libptc.NetworkProtocol `json:"net,omitempty" yaml:"net,omitempty" toml:"net,omitempty" mapstructure:"net,omitempty"`
	IP     string                 `json:"ip,omitempty" yaml:"ip,omitempty" toml:"ip,omitempty" mapstructure:"ip,omitempty"`
	Local  string                 `json:"local,omitempty" yaml:"local,omitempty" toml:"local,omitempty" mapstructure:"local,omitempty"`
}

type OptionTLS struct {
	Enable bool          `json:"enable" yaml:"enable" toml:"enable" mapstructure:"enable"`
	Config libtls.Config `json:"tls" yaml:"tls" toml:"tls" mapstructure:"tls"`
}

type OptionProxy struct {
	Enable   bool     `json:"enable" yaml:"enable" toml:"enable" mapstructure:"enable"`
	Endpoint *url.URL `json:"endpoint" yaml:"endpoint" toml:"endpoint" mapstructure:"endpoint"`
	Username string   `json:"username" yaml:"username" toml:"username" mapstructure:"username"`
	Password string   `json:"password" yaml:"password" toml:"password" mapstructure:"password"`
}

type Options struct {
	Timeout time.Duration `json:"timeout" yaml:"timeout" toml:"timeout" mapstructure:"timeout"`
	Http2   bool          `json:"http2" yaml:"http2" toml:"http2" mapstructure:"http2"`
	TLS     OptionTLS     `json:"tls" yaml:"tls" toml:"tls" mapstructure:"tls"`
	ForceIP OptionForceIP `json:"force_ip" yaml:"force_ip" toml:"force_ip" mapstructure:"force_ip"`
	Proxy   OptionProxy   `json:"proxy" yaml:"proxy" toml:"proxy" mapstructure:"proxy"`
}

func DefaultConfig(indent string) []byte {
	var (
		res = bytes.NewBuffer(make([]byte, 0))
		def = []byte(`{
       "timeout":"0s",
       "http2": true,
       "tls": ` + string(cmptls.DefaultConfig(cfgcst.JSONIndent)) + `,
       "force_ip": {
         "enable": false,
         "net":"tcp",
         "ip":"127.0.0.1:8080",
         "local":"127.0.0.1"
       },
       "proxy": {
         "enable": false,
         "endpoint":"http://example.com",
         "username":"example",
         "password":"example"
       }
}`)
	)
	if err := json.Indent(res, def, indent, cfgcst.JSONIndent); err != nil {
		return def
	} else {
		return res.Bytes()
	}
}

func (o Options) Validate() liberr.Error {
	var e = ErrorValidatorError.Error(nil)

	if err := libval.New().Struct(o); err != nil {
		if er, ok := err.(*libval.InvalidValidationError); ok {
			e.AddParent(er)
		}

		for _, er := range err.(libval.ValidationErrors) {
			//nolint #goerr113
			e.AddParent(fmt.Errorf("config field '%s' is not validated by constraint '%s'", er.Namespace(), er.ActualTag()))
		}
	}

	if !e.HasParent() {
		e = nil
	}

	return e
}

func (o Options) GetClient(def libtls.TLSConfig, servername string) (*http.Client, liberr.Error) {
	var tls libtls.TLSConfig

	if t, e := o._GetTLS(def); e != nil {
		return nil, e
	} else {
		tls = t
	}

	var tr *http.Transport

	tr = GetTransport(false, false, o.Http2)
	SetTransportTLS(tr, tls, "")
	SetTransportDial(tr, o.ForceIP.Enable, o.ForceIP.Net, o.ForceIP.IP, o.ForceIP.Local)

	if o.Proxy.Enable && o.Proxy.Endpoint != nil {
		var edp *url.URL

		edp = &url.URL{
			Scheme:      o.Proxy.Endpoint.Scheme,
			Opaque:      o.Proxy.Endpoint.Opaque,
			User:        nil,
			Host:        o.Proxy.Endpoint.Host,
			Path:        o.Proxy.Endpoint.Path,
			RawPath:     o.Proxy.Endpoint.RawPath,
			OmitHost:    o.Proxy.Endpoint.OmitHost,
			ForceQuery:  o.Proxy.Endpoint.ForceQuery,
			RawQuery:    o.Proxy.Endpoint.RawQuery,
			Fragment:    o.Proxy.Endpoint.Fragment,
			RawFragment: o.Proxy.Endpoint.RawFragment,
		}

		if len(o.Proxy.Password) > 0 {
			edp.User = url.UserPassword(o.Proxy.Username, o.Proxy.Password)
		} else if len(o.Proxy.Username) > 0 {
			edp.User = url.User(o.Proxy.Username)
		} else if o.Proxy.Endpoint.User != nil {
			if p, k := o.Proxy.Endpoint.User.Password(); k {
				edp.User = url.UserPassword(o.Proxy.Endpoint.User.Username(), p)
			} else {
				edp.User = url.User(o.Proxy.Endpoint.User.Username())
			}
		}

		if edp != nil && len(edp.String()) > 0 {
			SetTransportProxy(tr, edp)
		}
	}

	return GetClient(tr, o.Http2, o.Timeout)
}

func (o Options) _GetTLS(def libtls.TLSConfig) (libtls.TLSConfig, liberr.Error) {
	if o.TLS.Enable {
		return o.TLS.Config.NewFrom(def)
	} else {
		return libtls.Default.Clone(), nil
	}
}
