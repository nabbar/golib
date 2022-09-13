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

package request

import (
	"fmt"
	"net/http"

	libsts "github.com/nabbar/golib/status/config"

	libval "github.com/go-playground/validator/v10"
	libtls "github.com/nabbar/golib/certificates"
	libcfg "github.com/nabbar/golib/config"
	liberr "github.com/nabbar/golib/errors"
	libhtc "github.com/nabbar/golib/httpcli"
)

type OptionsCredentials struct {
	Enable   bool   `json:"enable" yaml:"enable" toml:"enable" mapstructure:"enable"`
	Username string `json:"username" yaml:"username" toml:"username" mapstructure:"username"`
	Password string `json:"password" yaml:"password" toml:"password" mapstructure:"password"`
}

type OptionsToken struct {
	Enable bool   `json:"enable" yaml:"enable" toml:"enable" mapstructure:"enable"`
	Token  string `json:"token" yaml:"token" toml:"token" mapstructure:"token"`
}

type OptionsAuth struct {
	Basic  OptionsCredentials `json:"basic" yaml:"basic" toml:"basic" mapstructure:"basic" validate:"required,dive"`
	Bearer OptionsToken       `json:"bearer" yaml:"bearer" toml:"bearer" mapstructure:"bearer" validate:"required,dive"`
}

type OptionsHealth struct {
	Enable   bool                `json:"enable" yaml:"enable" toml:"enable" mapstructure:"enable"`
	Endpoint string              `json:"endpoint" yaml:"endpoint" toml:"endpoint" mapstructure:"endpoint" validate:"required,url"`
	Auth     OptionsAuth         `json:"auth" yaml:"auth" toml:"auth" mapstructure:"auth" validate:"required,dive"`
	Result   OptionsHealthResult `json:"result" yaml:"result" toml:"result" mapstructure:"result" validate:"required,dive"`
	Status   libsts.ConfigStatus `json:"status" yaml:"status" toml:"status" mapstructure:"status" validate:"required,dive"`
}

type OptionsHealthResult struct {
	ValidHTTPCode   []int    `json:"valid_http_code" yaml:"valid_http_code" toml:"valid_http_code" mapstructure:"valid_http_code"`
	InvalidHTTPCode []int    `json:"invalid_http_code" yaml:"invalid_http_code" toml:"invalid_http_code" mapstructure:"invalid_http_code"`
	Contain         []string `json:"contain" yaml:"contain" toml:"contain" mapstructure:"contain"`
	NotContain      []string `json:"not_contain" yaml:"not_contain" toml:"not_contain" mapstructure:"not_contain"`
}

type Options struct {
	Endpoint   string         `json:"endpoint" yaml:"endpoint" toml:"endpoint" mapstructure:"endpoint" validate:"required,url"`
	HttpClient libhtc.Options `json:"http_client" yaml:"http_client" toml:"http_client" mapstructure:"http_client" validate:"required,dive"`
	Auth       OptionsAuth    `json:"auth" yaml:"auth" toml:"auth" mapstructure:"auth" validate:"required,dive"`
	Health     OptionsHealth  `json:"health" yaml:"health" toml:"health" mapstructure:"health" validate:"required,dive"`

	def FctTLSDefault
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

func (o Options) _GetDefaultTLS() libtls.TLSConfig {
	if o.def != nil {
		return o.def()
	}

	return nil
}

func (o Options) SetDefaultTLS(fct FctTLSDefault) {
	o.def = fct
}

func (o Options) GetClientHTTP(servername string) *http.Client {
	if c, e := o.HttpClient.GetClient(o._GetDefaultTLS(), servername); e == nil {
		return c
	}

	return &http.Client{}
}

func (o Options) New(ctx libcfg.FuncContext, cli FctHttpClient, tls FctTLSDefault) (Request, error) {
	if tls != nil {
		o.def = tls
	}

	return New(ctx, cli, o)
}

func (o Options) Update(req Request, ctx libcfg.FuncContext, cli FctHttpClient, tls FctTLSDefault) (Request, error) {
	if tls != nil {
		o.def = tls
	}

	var (
		e error
		n Request
	)

	if n, e = req.Clone(); e != nil {
		return nil, e
	}

	if ctx != nil {
		n.SetContext(ctx)
	}

	if cli != nil {
		n.SetClient(cli)
	}

	if e = n.SetOption(&o); e != nil {
		return nil, e
	}

	return n, nil
}
