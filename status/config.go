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
	// keyConfigReturnCode is the internal key for storing HTTP return codes in context config.
	keyConfigReturnCode = "cfgReturnCode"

	// keyConfigMandatory is the internal key for storing mandatory component list in context config.
	keyConfigMandatory = "cfgMandatory"
)

// Mandatory defines a group of components with a specific control mode.
// It allows grouping multiple components and defining how their health affects overall status.
//
// See github.com/nabbar/golib/status/control for Mode details.
type Mandatory struct {
	// Mode defines how this group of components affects overall status.
	// Possible values: Ignore, Should, AnyOf, Quorum.
	Mode stsctr.Mode

	// Keys is the list of component names in this group.
	// Component names must match those registered in the monitor pool.
	Keys []string
}

// ParseMandatory converts a mandatory.Mandatory interface to a Mandatory struct.
// This is a utility function for converting between the interface and struct representations.
//
// Parameters:
//   - m: the mandatory interface to convert
//
// Returns a Mandatory struct with the mode and keys from the interface.
// Returns an empty Mandatory struct if m is nil.
//
// See github.com/nabbar/golib/status/mandatory for the Mandatory interface.
func ParseMandatory(m stsmdt.Mandatory) Mandatory {
	if m == nil {
		return Mandatory{}
	}

	return Mandatory{
		Mode: m.GetMode(),
		Keys: m.KeyList(),
	}
}

// ParseList converts a slice of mandatory.Mandatory interfaces to Mandatory structs.
// This is a utility function for bulk conversion between interface and struct representations.
//
// Parameters:
//   - m: variadic list of mandatory interfaces to convert
//
// Returns a slice of Mandatory structs. Nil entries are skipped.
// This function is useful when loading configuration or marshaling/unmarshaling.
//
// See github.com/nabbar/golib/status/mandatory for the Mandatory interface.
func ParseList(m ...stsmdt.Mandatory) []Mandatory {
	r := make([]Mandatory, 0, len(m))
	for _, i := range m {
		if i != nil {
			r = append(r, ParseMandatory(i))
		}
	}
	return r
}

// Config defines the configuration for status computation and HTTP responses.
// It controls HTTP status codes and component health evaluation strategies.
type Config struct {
	// ReturnCode maps health status to HTTP status codes.
	// Keys: monsts.OK, monsts.Warn, monsts.KO
	// Default values if not set:
	//   - monsts.OK: 200 (http.StatusOK)
	//   - monsts.Warn: 207 (http.StatusMultiStatus)
	//   - monsts.KO: 500 (http.StatusInternalServerError)
	ReturnCode map[monsts.Status]int

	// MandatoryComponent defines groups of components with control modes.
	// Each group specifies how its components' health affects overall status.
	// See Mandatory type and github.com/nabbar/golib/status/control for details.
	MandatoryComponent []Mandatory
}

// Validate checks if the configuration is valid.
// It uses the validator package to validate struct fields.
//
// Returns an error if validation fails, nil otherwise.
// The error will contain details about which fields failed validation.
func (o Config) Validate() error {
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

// SetConfig applies the given configuration to the status instance.
// It sets HTTP return codes and mandatory component definitions.
//
// If ReturnCode is empty, default values are used:
//   - OK: 200, Warn: 207, KO: 500
//
// The configuration is stored in thread-safe context storage.
//
// Parameters:
//   - cfg: the configuration to apply
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

	var lst = stslmd.New()

	if len(cfg.MandatoryComponent) > 0 {
		for _, i := range cfg.MandatoryComponent {
			var m = stsmdt.New()
			m.SetMode(i.Mode)
			m.KeyAdd(i.Keys...)
			lst.Add(m)
		}
	}

	o.x.Store(keyConfigMandatory, lst)
}

// cfgGetReturnCode retrieves the HTTP status code for a given health status.
// Returns http.StatusInternalServerError (500) if not configured or on error.
//
// Parameters:
//   - s: the health status to get the HTTP code for
//
// Returns the configured HTTP status code.
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

// cfgGetMandatory retrieves the list of mandatory component configurations.
// Returns nil if not configured or on error.
//
// Returns the ListMandatory containing all component groups and their control modes.
// See github.com/nabbar/golib/status/listmandatory for ListMandatory details.
func (o *sts) cfgGetMandatory() stslmd.ListMandatory {
	if i, l := o.x.Load(keyConfigMandatory); !l {
		return nil
	} else if v, k := i.(stslmd.ListMandatory); !k {
		return nil
	} else {
		return v
	}
}

// cfgGetMode retrieves the control mode for a specific component.
// Returns stsctr.Ignore if the component is not in any mandatory group.
//
// Parameters:
//   - key: the component name to get the mode for
//
// Returns the control mode (Ignore, Should, AnyOf, or Quorum).
// See github.com/nabbar/golib/status/control for Mode details.
func (o *sts) cfgGetMode(key string) stsctr.Mode {
	if l := o.cfgGetMandatory(); l == nil {
		return stsctr.Ignore
	} else {
		return l.GetMode(key)
	}
}

// cfgGetOne retrieves all component names in the same group as the given component.
// This is used for AnyOf and Quorum modes to find all related components.
//
// Parameters:
//   - key: the component name to find the group for
//
// Returns a slice of component names in the same group, or empty slice if not found.
func (o *sts) cfgGetOne(key string) []string {
	if l := o.cfgGetMandatory(); l == nil {
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
