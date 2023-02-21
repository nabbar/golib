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
	"sync/atomic"

	liblog "github.com/nabbar/golib/logger"

	moncfg "github.com/nabbar/golib/monitor/types"

	libval "github.com/go-playground/validator/v10"
	libtls "github.com/nabbar/golib/certificates"
	libctx "github.com/nabbar/golib/context"
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
	Monitor  moncfg.Config       `json:"monitor" yaml:"monitor" toml:"monitor" mapstructure:"monitor" validate:"required,dive"`
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

	tls libtls.FctTLSDefault
	log liblog.FuncLog
}

func (o *Options) Validate() liberr.Error {
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

func (o *Options) defaultTLS() libtls.TLSConfig {
	if o.tls != nil {
		return o.tls()
	}

	return nil
}

func (o *Options) SetDefaultTLS(fct libtls.FctTLSDefault) {
	o.tls = fct
}

func (o *Options) SetDefaultLog(fct liblog.FuncLog) {
	o.log = fct
}

func (o *Options) ClientHTTPTLS(tls libtls.TLSConfig, servername string) *http.Client {
	if c, e := o.HttpClient.GetClient(tls, servername); e == nil {
		return c
	}

	return &http.Client{}
}

func (o *Options) ClientHTTP(servername string) *http.Client {
	return o.ClientHTTPTLS(o.defaultTLS(), servername)
}

func (o *Options) New(ctx libctx.FuncContext) (Request, error) {
	if n, e := New(ctx, o); e != nil {
		return nil, e
	} else {
		n.RegisterDefaultLogger(o.log)
		n.RegisterHTTPClient(o.ClientHTTPTLS)
		return n, nil
	}
}

func (o *Options) Update(ctx libctx.FuncContext, req Request) (Request, error) {
	var (
		e error
		n Request
	)

	if n, e = req.Clone(); e != nil {
		return nil, e
	}

	if ctx != nil {
		n.RegisterContext(ctx)
	}

	if e = n.SetOption(o); e != nil {
		return nil, e
	}

	return n, nil
}

func (r *request) options() *Options {
	if r.o == nil {
		return nil
	} else if i := r.o.Load(); i == nil {
		return nil
	} else if o, ok := i.(*Options); !ok {
		return nil
	} else {
		return o
	}
}

func (r *request) GetOption() *Options {
	r.s.Lock()
	defer r.s.Unlock()
	return r.options()
}

func (r *request) SetOption(opt *Options) error {
	if e := r.SetEndpoint(opt.Endpoint); e != nil {
		return e
	}

	if opt.Auth.Basic.Enable {
		r.AuthBasic(opt.Auth.Basic.Username, opt.Auth.Basic.Password)
	} else if opt.Auth.Bearer.Enable {
		r.AuthBearer(opt.Auth.Bearer.Token)
	}

	r.s.Lock()
	defer r.s.Unlock()

	if r.o == nil {
		r.o = new(atomic.Value)
	}

	r.o.Store(opt)
	return nil
}

func (r *request) RegisterHTTPClient(fct libtls.FctHttpClient) {
	r.s.Lock()
	defer r.s.Unlock()

	r.f = fct
}

func (r *request) RegisterDefaultLogger(fct liblog.FuncLog) {
	r.s.Lock()
	defer r.s.Unlock()

	r.l = fct
}

func (r *request) _getDefaultLogger() liblog.Logger {
	r.s.Lock()
	defer r.s.Unlock()

	if r.l == nil {
		return nil
	} else {
		return r.l()
	}
}

func (r *request) defaultTLS() libtls.TLSConfig {
	if cfg := r.options(); cfg != nil {
		return cfg.defaultTLS()
	}

	return nil
}

func (r *request) RegisterContext(fct libctx.FuncContext) {
	r.s.Lock()
	defer r.s.Unlock()

	r.x = fct
}
