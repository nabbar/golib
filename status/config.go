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
	stsctr "github.com/nabbar/golib/status/control"
	stslmd "github.com/nabbar/golib/status/listmandatory"
	stsmdt "github.com/nabbar/golib/status/mandatory"
)

const (
	keyConfigReturnCode = "cfgReturnCode"
	keyConfigMandatory  = "cfgMandatory"
)

type Config struct {
	ReturnCode         map[monsts.Status]int
	MandatoryComponent stslmd.ListMandatory
}

type Mandatory struct {
	Mode stsctr.Mode
	Keys []string
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
		o.x.Store(keyConfigMandatory, cfg.MandatoryComponent)
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

func (o *sts) cfgGetMandatory() stslmd.ListMandatory {
	if i, l := o.x.Load(keyConfigMandatory); !l {
		return nil
	} else if v, k := i.(stslmd.ListMandatory); !k {
		return nil
	} else {
		return v
	}
}

func (o *sts) cfgGetMode(key string) stsctr.Mode {
	if l := o.cfgGetMandatory(); len(l) < 1 {
		return stsctr.Ignore
	} else {
		return l.GetMode(key)
	}
}

func (o *sts) cfgGetOne(key string) []string {
	if l := o.cfgGetMandatory(); len(l) < 1 {
		return make([]string, 0)
	} else {
		var r []string
		l.Walk(func(m stsmdt.Mandatory) bool {
			if m.KeyHas(key) {
				r = m.KeyList()
				return false
			}

			return true
		})
		return r
	}
}
