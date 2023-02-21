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

package status

import (
	"fmt"
	"net/http"

	libval "github.com/go-playground/validator/v10"
	monsts "github.com/nabbar/golib/monitor/status"
	"golang.org/x/exp/slices"
)

const (
	keyConfigReturnCode = "cfgReturnCode"
	keyConfigMandatory  = "cfgMandatory"
)

type Config struct {
	ReturnCode         map[monsts.Status]int
	MandatoryComponent []string
}

func (o Config) Validate() error {
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

func (o *sts) SetConfig(cfg Config) {
	if len(cfg.ReturnCode) < 1 {

		var def = make(map[monsts.Status]int, 0)
		def[monsts.KO] = http.StatusInternalServerError
		def[monsts.Warn] = http.StatusMultiStatus
		def[monsts.OK] = http.StatusOK

		o.x.Store(keyConfigReturnCode, def)
	} else {
		o.x.Store(keyConfigReturnCode, cfg.ReturnCode)
	}

	if len(cfg.MandatoryComponent) < 1 {
		o.x.Store(keyConfigMandatory, make([]string, 0))
	} else {
		o.x.Store(keyConfigMandatory, cfg.ReturnCode)
	}
}

func (o *sts) cfgGetReturnCode(s monsts.Status) int {
	if i, l := o.x.Load(keyConfigReturnCode); !l {
		return http.StatusInternalServerError
	} else if v, k := i.(map[monsts.Status]int); !k {
		return http.StatusInternalServerError
	} else if r, f := v[s]; !f {
		return http.StatusInternalServerError
	} else {
		return r
	}
}

func (o *sts) cfgIsMandatory(name string) bool {
	if i, l := o.x.Load(keyConfigMandatory); !l {
		return false
	} else if v, k := i.([]string); !k {
		return false
	} else {
		return slices.Contains(v, name)
	}
}

func (o *sts) cfgGetMandatory() []string {
	if i, l := o.x.Load(keyConfigMandatory); !l {
		return nil
	} else if v, k := i.([]string); !k {
		return nil
	} else {
		return v
	}
}
