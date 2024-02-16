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
	"fmt"
	"net/http"
	"net/url"
	"time"

	libval "github.com/go-playground/validator/v10"
	libtls "github.com/nabbar/golib/certificates"
	liberr "github.com/nabbar/golib/errors"
	htcdns "github.com/nabbar/golib/httpcli/dns-mapper"
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
	Timeout            time.Duration `json:"timeout" yaml:"timeout" toml:"timeout" mapstructure:"timeout"`
	DisableKeepAlive   bool          `json:"disable-keep-alive" yaml:"disable-keep-alive" toml:"disable-keep-alive" mapstructure:"disable-keep-alive"`
	DisableCompression bool          `json:"disable-compression" yaml:"disable-compression" toml:"disable-compression" mapstructure:"disable-compression"`
	Http2              bool          `json:"http2" yaml:"http2" toml:"http2" mapstructure:"http2"`
	TLS                OptionTLS     `json:"tls" yaml:"tls" toml:"tls" mapstructure:"tls"`
	ForceIP            OptionForceIP `json:"force_ip" yaml:"force_ip" toml:"force_ip" mapstructure:"force_ip"`
	Proxy              OptionProxy   `json:"proxy" yaml:"proxy" toml:"proxy" mapstructure:"proxy"`
}

func DefaultConfig(indent string) []byte {
	return htcdns.DefaultConfig(indent)
}

func (o Options) Validate() liberr.Error {
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

func (o Options) GetClient(def libtls.TLSConfig, servername string) (*http.Client, liberr.Error) {
	return GetClient(), nil
}
