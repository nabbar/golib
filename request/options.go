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
	"context"
	"fmt"

	libval "github.com/go-playground/validator/v10"

	libtls "github.com/nabbar/golib/certificates"
	libhtc "github.com/nabbar/golib/httpcli"
	liblog "github.com/nabbar/golib/logger"
	moncfg "github.com/nabbar/golib/monitor/types"
)

type OptionsCredentials struct {
	Enable   bool   `json:"enable" yaml:"enable" toml:"enable" mapstructure:"enable"`
	Username string `json:"username" yaml:"username" toml:"username" mapstructure:"username"`
	Password string `json:"password" yaml:"password" toml:"password" mapstructure:"password"` // #nosec nolint
}

type OptionsToken struct {
	Enable bool   `json:"enable" yaml:"enable" toml:"enable" mapstructure:"enable"`
	Token  string `json:"token" yaml:"token" toml:"token" mapstructure:"token"`
}

type OptionsAuth struct {
	Basic  OptionsCredentials `json:"basic" yaml:"basic" toml:"basic" mapstructure:"basic" validate:""`
	Bearer OptionsToken       `json:"bearer" yaml:"bearer" toml:"bearer" mapstructure:"bearer" validate:""`
}

type OptionsHealth struct {
	Enable   bool                `json:"enable" yaml:"enable" toml:"enable" mapstructure:"enable"`
	Endpoint string              `json:"endpoint" yaml:"endpoint" toml:"endpoint" mapstructure:"endpoint" validate:"url"`
	Auth     OptionsAuth         `json:"auth" yaml:"auth" toml:"auth" mapstructure:"auth" validate:""`
	Result   OptionsHealthResult `json:"result" yaml:"result" toml:"result" mapstructure:"result" validate:""`
	Monitor  moncfg.Config       `json:"monitor" yaml:"monitor" toml:"monitor" mapstructure:"monitor" validate:"required"`
}

type OptionsHealthResult struct {
	ValidHTTPCode   []int    `json:"valid_http_code" yaml:"valid_http_code" toml:"valid_http_code" mapstructure:"valid_http_code"`
	InvalidHTTPCode []int    `json:"invalid_http_code" yaml:"invalid_http_code" toml:"invalid_http_code" mapstructure:"invalid_http_code"`
	Contain         []string `json:"contain" yaml:"contain" toml:"contain" mapstructure:"contain"`
	NotContain      []string `json:"not_contain" yaml:"not_contain" toml:"not_contain" mapstructure:"not_contain"`
}

type Options struct {
	Endpoint string        `json:"endpoint" yaml:"endpoint" toml:"endpoint" mapstructure:"endpoint" validate:"url"`
	Auth     OptionsAuth   `json:"auth" yaml:"auth" toml:"auth" mapstructure:"auth" validate:""`
	Health   OptionsHealth `json:"health" yaml:"health" toml:"health" mapstructure:"health" validate:""`

	tls libtls.FctTLSDefault
	log liblog.FuncLog
}

func (o *Options) Validate() error {
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

func (o *OptionsHealth) Validate() error {
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

func (o *Options) SetDefaultTLS(fct libtls.FctTLSDefault) {
	o.tls = fct
}

func (o *Options) SetDefaultLog(fct liblog.FuncLog) {
	o.log = fct
}

func (o *Options) New(ctx context.Context, cli libhtc.HttpClient) (Request, error) {
	n, e := New(ctx, o, cli)

	if e != nil {
		return nil, e
	}

	var l liblog.Logger
	l, e = liblog.NewFrom(ctx, &o.Health.Monitor.Logger, o.log)

	if e != nil {
		_ = l.Close()
		return nil, e
	}

	n.RegisterDefaultLogger(func() liblog.Logger {
		return l
	})

	return n, e
}

func (o *Options) Update(ctx context.Context, req Request) (Request, error) {
	var (
		e error
		n Request
		l liblog.Logger
	)

	n, e = req.Clone()

	if e != nil {
		return nil, e
	}

	if ctx != nil {
		n.RegisterContext(ctx)
	}

	if e = n.SetOption(o); e != nil {
		return nil, e
	}

	l, e = liblog.NewFrom(ctx, &o.Health.Monitor.Logger, o.log)

	if e != nil {
		_ = l.Close()
		return nil, e
	}

	n.RegisterDefaultLogger(func() liblog.Logger {
		return l
	})

	return n, nil
}

func (r *request) options() *Options {
	if r.opt == nil {
		return nil
	} else if i := r.opt.Load(); i == nil {
		return nil
	} else if o, ok := i.(*Options); !ok {
		return nil
	} else {
		return o
	}
}

func (r *request) GetOption() *Options {
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

	r.opt.Store(opt)
	return nil
}
