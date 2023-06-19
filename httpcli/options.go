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
	"time"

	libptc "github.com/nabbar/golib/network/protocol"

	cmptls "github.com/nabbar/golib/config/components/tls"
	cfgcst "github.com/nabbar/golib/config/const"

	libval "github.com/go-playground/validator/v10"

	libtls "github.com/nabbar/golib/certificates"
	liberr "github.com/nabbar/golib/errors"
)

type OptionForceIP struct {
	Enable bool   `json:"enable" yaml:"enable" toml:"enable" mapstructure:"enable"`
	Net    string `json:"net,omitempty" yaml:"net,omitempty" toml:"net,omitempty" mapstructure:"net,omitempty"`
	IP     string `json:"ip,omitempty" yaml:"ip,omitempty" toml:"ip,omitempty" mapstructure:"ip,omitempty"`
}

type OptionTLS struct {
	Enable bool          `json:"enable" yaml:"enable" toml:"enable" mapstructure:"enable"`
	Config libtls.Config `json:"tls" yaml:"tls" toml:"tls" mapstructure:"tls"`
}

type Options struct {
	Timeout time.Duration `json:"timeout" yaml:"timeout" toml:"timeout" mapstructure:"timeout"`
	Http2   bool          `json:"http2" yaml:"http2" toml:"http2" mapstructure:"http2"`
	TLS     OptionTLS     `json:"tls" yaml:"tls" toml:"tls" mapstructure:"tls"`
	ForceIP OptionForceIP `json:"force_ip" yaml:"force_ip" toml:"force_ip" mapstructure:"force_ip"`
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
         "ip":"127.0.0.1:8080"
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

	if o.ForceIP.Enable {
		return GetClientTlsForceIp(libptc.Parse(o.ForceIP.Net), o.ForceIP.IP, servername, tls, o.Http2, o.Timeout)
	} else {
		return GetClientTls(servername, tls, o.Http2, o.Timeout)
	}
}

func (o Options) _GetTLS(def libtls.TLSConfig) (libtls.TLSConfig, liberr.Error) {
	if o.TLS.Enable {
		return o.TLS.Config.NewFrom(def)
	} else {
		return libtls.Default.Clone(), nil
	}
}
