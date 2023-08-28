/*
MIT License

Copyright (c) 2019 Nicolas JUHEL

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package config

import (
	"fmt"

	libmon "github.com/nabbar/golib/monitor/types"

	libval "github.com/go-playground/validator/v10"
	libtls "github.com/nabbar/golib/certificates"
	liberr "github.com/nabbar/golib/errors"
)

type ConfigModel struct {
	DSN     string        `json:"dsn" yaml:"dsn" toml:"dsn" mapstructure:"dsn"`
	TLS     libtls.Config `json:"tls,omitempty" yaml:"tls,omitempty" toml:"tls,omitempty" mapstructure:"tls,omitempty"`
	Monitor libmon.Config `json:"monitor,omitempty" yaml:"monitor,omitempty" toml:"monitor,omitempty" mapstructure:"monitor,omitempty"`
}

func (c ConfigModel) Validate() liberr.Error {
	err := ErrorConfigValidator.Error(nil)

	if er := libval.New().Struct(c); er != nil {
		if e, ok := er.(*libval.InvalidValidationError); ok {
			err.Add(e)
		}

		for _, e := range er.(libval.ValidationErrors) {
			//nolint goerr113
			err.Add(fmt.Errorf("config field '%s' is not validated by constraint '%s'", e.Namespace(), e.ActualTag()))
		}
	}

	if c.DSN != "" {
		if _, er := New(c); er != nil {
			err.Add(er)
		}
	} else {
		err.Add(ErrorConfigInvalidDSN.Error(nil))
	}

	if !err.HasParent() {
		err = nil
	}

	return err
}

func (c ConfigModel) Config() (Config, liberr.Error) {
	return New(c)
}
