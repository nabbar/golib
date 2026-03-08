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
	// keyConfigReturnCode is the internal key for storing HTTP return codes in the context config.
	keyConfigReturnCode = "cfgReturnCode"

	// keyConfigMandatory is the internal key for storing the mandatory component list in the context config.
	keyConfigMandatory = "cfgMandatory"
)

// Mandatory defines a group of components with a specific control mode.
// It allows grouping multiple components and defining how their collective health
// affects the overall application status.
//
// See github.com/nabbar/golib/status/control for details on control modes.
type Mandatory struct {
	// Mode defines how this group of components affects the overall status.
	// Possible values: Ignore, Should, AnyOf, Quorum.
	Mode stsctr.Mode `mapstructure:"mode" json:"mode" yaml:"mode" toml:"mode" validate:"required"`

	// Keys is a list of monitor names belonging to this group.
	// These names must match the monitors registered in the monitor pool.
	Keys []string `mapstructure:"keys" json:"keys" yaml:"keys" toml:"keys"`

	// ConfigKeys is used to specify the keys of config components. The monitor
	// names from these components will be dynamically added to this mandatory group.
	ConfigKeys []string `mapstructure:"configKeys" json:"configKeys" yaml:"configKeys" toml:"configKeys"`
}

// ParseMandatory converts a mandatory.Mandatory interface to a Mandatory struct.
// This is a utility function for converting between the interface and struct representations.
//
// Parameters:
//   - m: The mandatory.Mandatory interface to convert.
//
// Returns a Mandatory struct with the mode and keys from the interface.
// Returns an empty Mandatory struct if m is nil.
func ParseMandatory(m stsmdt.Mandatory) Mandatory {
	if m == nil {
		return Mandatory{}
	}

	return Mandatory{
		Mode:       m.GetMode(),
		Keys:       m.KeyList(),
		ConfigKeys: nil,
	}
}

// ParseList converts a slice of mandatory.Mandatory interfaces to Mandatory structs.
// This is a utility function for bulk conversion.
//
// Parameters:
//   - m: A variadic list of mandatory.Mandatory interfaces to convert.
//
// Returns a slice of Mandatory structs. Nil entries in the input are skipped.
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
	// ReturnCode maps health status (OK, Warn, KO) to HTTP status codes.
	// Default values if not set:
	//   - monsts.OK:   200 (http.StatusOK)
	//   - monsts.Warn: 207 (http.StatusMultiStatus)
	//   - monsts.KO:   500 (http.StatusInternalServerError)
	ReturnCode map[monsts.Status]int `mapstructure:"return-code" json:"return-code" yaml:"return-code" toml:"return-code" validate:"required"`

	// Component defines groups of components with specific control modes.
	// Each group specifies how its components' health affects the overall status.
	Component []Mandatory `mapstructure:"component" json:"component" yaml:"component" toml:"component" validate:""`
}

// Validate checks if the configuration is valid using struct field tags.
//
// Returns an error if validation fails, otherwise nil.
func (o Config) Validate() error {
	var e = ErrorValidatorError.Error(nil)

	if err := libval.New().Struct(o); err != nil {
		if er, ok := err.(*libval.InvalidValidationError); ok {
			e.Add(er)
		}

		for _, er := range err.(libval.ValidationErrors) {
			//nolint
			e.Add(fmt.Errorf("config field '%s' is not validated by constraint '%s'", er.Namespace(), er.ActualTag()))
		}
	}

	if !e.HasParent() {
		e = nil
	}

	return e
}

// RegisterGetConfigCpt registers a function that retrieves a component by its key.
// This function is used to resolve monitor names from component configurations.
func (o *sts) RegisterGetConfigCpt(fct FuncGetCfgCpt) {
	o.m.Lock()
	defer o.m.Unlock()
	o.n = fct
}

func (o *sts) GetConfig() Config {
	var (
		cfg = Config{
			ReturnCode: make(map[monsts.Status]int, 0),
			Component:  make([]Mandatory, 0),
		}
	)

	cfg.ReturnCode[monsts.KO] = http.StatusInternalServerError
	cfg.ReturnCode[monsts.Warn] = http.StatusMultiStatus
	cfg.ReturnCode[monsts.OK] = http.StatusOK

	if i, l := o.x.Load(keyConfigReturnCode); !l || i == nil {
		//
	} else if r, k := i.(map[monsts.Status]int); !k || len(r) < 1 {
		//
	} else {
		cfg.ReturnCode = r
	}

	if i, l := o.x.Load(keyConfigMandatory); !l || i == nil {
		//
	} else if r, k := i.(stslmd.ListMandatory); !k || r == nil || r.Len() < 1 {
		//
	} else {
		r.Walk(func(m stsmdt.Mandatory) bool {
			cfg.Component = append(cfg.Component, Mandatory{
				Mode: m.GetMode(),
				Keys: m.KeyList(),
			})
			return true
		})
	}

	return cfg
}

// SetConfig applies the given configuration to the status instance.
// It configures HTTP return codes and mandatory component groups.
// If ReturnCode is empty, default values are used.
//
// This method processes both static 'Keys' and dynamic 'ConfigKeys' to build
// the list of mandatory monitors.
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

	if len(cfg.Component) > 0 {
		for _, i := range cfg.Component {
			var m = stsmdt.New()
			m.SetMode(i.Mode)
			m.KeyAdd(i.Keys...)
			if len(i.ConfigKeys) > 0 && o.n != nil {
				for _, k := range i.ConfigKeys {
					if c := o.n(k); c != nil {
						m.KeyAdd(c.GetMonitorNames()...)
					}
				}
			}
			lst.Add(m)
		}
	}

	o.x.Store(keyConfigMandatory, lst)
}

// cfgGetReturnCode retrieves the HTTP status code for a given health status.
// It returns http.StatusInternalServerError (500) if the status is not configured.
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
// It returns nil if no mandatory components are configured.
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
// It returns stsctr.Ignore if the component is not in any mandatory group.
func (o *sts) cfgGetMode(key string) stsctr.Mode {
	if l := o.cfgGetMandatory(); l == nil {
		return stsctr.Ignore
	} else {
		return l.GetMode(key)
	}
}

// cfgGetOne retrieves all component names in the same group as the given component.
// This is used for 'AnyOf' and 'Quorum' modes to find all related components.
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
