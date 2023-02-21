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

package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	cptlog "github.com/nabbar/golib/logger/config"

	cfgtps "github.com/nabbar/golib/config/const"

	libval "github.com/go-playground/validator/v10"
	liberr "github.com/nabbar/golib/errors"
	liblog "github.com/nabbar/golib/logger"
)

var _defaultConfig = []byte(`{
  "name": "",
  "check-timeout": "",
  "interval-check": "",
  "interval-fall": "",
  "interval-rise": "",
  "fall-count-ko": "",
  "fall-count-warn": "",
  "rise-count-ko": "",
  "rise-count-warn": "",
  "logger": ` + string(cptlog.DefaultConfig(cfgtps.JSONIndent+cfgtps.JSONIndent)) + `
}`)

func SetDefaultConfig(cfg []byte) {
	_defaultConfig = cfg
}

func DefaultConfig(indent string) []byte {
	var res = bytes.NewBuffer(make([]byte, 0))
	if err := json.Indent(res, _defaultConfig, indent, cfgtps.JSONIndent); err != nil {
		return _defaultConfig
	} else {
		return res.Bytes()
	}
}

type Config struct {
	// Name define the name of the monitor
	Name string `json:"name" yaml:"name" toml:"name" mapstructure:"name"`

	// CheckTimeout define the timeout use for healthcheck. Default is 5 second.
	CheckTimeout time.Duration `json:"check-timeout" yaml:"check-timeout" toml:"check-timeout" mapstructure:"check-timeout"`

	// IntervalCheck define the time waiting between 2 healthcheck. Default is 5 second.
	IntervalCheck time.Duration `json:"interval-check" yaml:"interval-check" toml:"interval-check" mapstructure:"interval-check"`

	// IntervalFall define the time waiting between 2 healthcheck when last check is KO. Default is 5 second.
	IntervalFall time.Duration `json:"interval-fall" yaml:"interval-fall" toml:"interval-fall" mapstructure:"interval-down"`

	// IntervalRise define the time waiting between 2 healthcheck when status is KO or Warn but last check is OK. Default is 5 second.
	IntervalRise time.Duration `json:"interval-rise" yaml:"interval-rise" toml:"interval-rise" mapstructure:"interval-rise"`

	// FallCountKO define the number of KO before considerate the component as down.
	FallCountKO uint8 `json:"fall-count-ko" yaml:"fall-count-ko" toml:"fall-count-ko" mapstructure:"fall-count-ko"`

	// FallCountWarn define the number of KO before considerate the component as warn.
	FallCountWarn uint8 `json:"fall-count-warn" yaml:"fall-count-warn" toml:"fall-count-warn" mapstructure:"fall-count-warn"`

	// RiseCountKO define the number of OK when status is KO before considerate the component as up.
	RiseCountKO uint8 `json:"rise-count-ko" yaml:"rise-count-ko" toml:"rise-count-ko" mapstructure:"rise-count-ko"`

	// RiseCountWarn define the number of OK when status is Warn before considerate the component as up.
	RiseCountWarn uint8 `json:"rise-count-warn" yaml:"rise-count-warn" toml:"rise-count-warn" mapstructure:"rise-count-warn"`

	// Logger define the logger options for current monitor log
	Logger liblog.Options `json:"logger" yaml:"logger" toml:"logger" mapstructure:"logger"`
}

func (o Config) Validate() liberr.Error {
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

func (o Config) Clone() Config {
	return Config{
		Name:          o.Name,
		CheckTimeout:  o.CheckTimeout,
		IntervalCheck: o.IntervalCheck,
		IntervalFall:  o.IntervalFall,
		IntervalRise:  o.IntervalRise,
		FallCountKO:   o.FallCountKO,
		FallCountWarn: o.FallCountWarn,
		RiseCountKO:   o.RiseCountKO,
		RiseCountWarn: o.RiseCountWarn,
		Logger:        o.Logger.Clone(),
	}
}
