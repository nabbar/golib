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
	// keyConfigReturnCode is the internal key used to store the HTTP return code
	// mapping in the thread-safe context configuration.
	keyConfigReturnCode = "cfgReturnCode"

	// keyConfigMandatory is the internal key used to store the list of mandatory
	// component groups in the thread-safe context configuration.
	keyConfigMandatory = "cfgMandatory"
)

// Mandatory defines a group of components that share a specific control mode.
// It allows grouping multiple components (e.g., "all databases") and defining
// how their collective health affects the overall application status.
//
// This structure is typically used when loading configuration from a file (JSON, YAML, etc.).
//
// See github.com/nabbar/golib/status/control for details on available control modes.
type Mandatory struct {
	// Mode defines how this group of components affects the overall status.
	// Possible values include:
	//   - Ignore: The components are monitored but do not affect global status.
	//   - Should: Failure causes a warning but not a critical failure.
	//   - Must: Failure causes a critical global failure.
	//   - AnyOf: At least one component in the group must be healthy.
	//   - Quorum: A majority of components in the group must be healthy.
	Mode stsctr.Mode `mapstructure:"mode" json:"mode" yaml:"mode" toml:"mode" validate:"required"`

	// Keys is a list of static monitor names belonging to this group.
	// These names must match the names of the monitors registered in the monitor pool.
	// This field is used when the component names are known at configuration time.
	Keys []string `mapstructure:"keys" json:"keys" yaml:"keys" toml:"keys"`

	// ConfigKeys is used to specify the keys of config components. This allows for
	// dynamic resolution of monitor names. When `SetConfig` is called, the system
	// will look up the component configuration using these keys (via the function
	// registered with `RegisterGetConfigCpt`) and add the associated monitor names
	// to this mandatory group.
	ConfigKeys []string `mapstructure:"configKeys" json:"configKeys" yaml:"configKeys" toml:"configKeys"`
}

// ParseMandatory converts a `mandatory.Mandatory` interface (from the internal logic)
// to a `Mandatory` struct (for configuration/export).
//
// This is a utility function for converting between the runtime interface representation
// and the configuration struct representation.
//
// Parameters:
//   - m: The `mandatory.Mandatory` interface to convert.
//
// Returns:
//   A `Mandatory` struct populated with the mode and keys from the interface.
//   Returns an empty struct if the input is nil.
func ParseMandatory(m stsmdt.Mandatory) Mandatory {
	if m == nil {
		return Mandatory{}
	}

	return Mandatory{
		Mode:       m.GetMode(),
		Keys:       m.KeyList(),
		ConfigKeys: nil, // ConfigKeys are resolved to Keys during runtime, so we don't export them back.
	}
}

// ParseList converts a slice of `mandatory.Mandatory` interfaces to a slice of
// `Mandatory` structs.
//
// This is a utility function for bulk conversion, useful when exporting the current
// configuration state.
//
// Parameters:
//   - m: A variadic list of `mandatory.Mandatory` interfaces.
//
// Returns:
//   A slice of `Mandatory` structs. Nil entries in the input are skipped.
func ParseList(m ...stsmdt.Mandatory) []Mandatory {
	r := make([]Mandatory, 0, len(m))
	for _, i := range m {
		if i != nil {
			r = append(r, ParseMandatory(i))
		}
	}
	return r
}

// Config defines the complete configuration for the status system.
// It controls how the application's health is computed and how it is reported
// via HTTP.
type Config struct {
	// ReturnCode maps internal health status levels (OK, Warn, KO) to HTTP status codes.
	//
	// Default values if not set:
	//   - monsts.OK:   200 (http.StatusOK)
	//   - monsts.Warn: 207 (http.StatusMultiStatus)
	//   - monsts.KO:   500 (http.StatusInternalServerError)
	ReturnCode map[monsts.Status]int `mapstructure:"return-code" json:"return-code" yaml:"return-code" toml:"return-code" validate:"required"`

	// Component defines the list of mandatory component groups. Each group specifies
	// a set of components and the control mode that applies to them.
	Component []Mandatory `mapstructure:"component" json:"component" yaml:"component" toml:"component" validate:""`
}

// Validate checks if the configuration is valid using the `validator` package.
// It ensures that all required fields are present and meet the defined constraints.
//
// Returns:
//   An error if validation fails, containing details about which fields failed
//   and why. Returns nil if the configuration is valid.
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

// RegisterGetConfigCpt registers a function that retrieves a component monitor
// configuration by its key.
//
// This function is the mechanism for dynamic component resolution. It is used
// by `SetConfig` to resolve `ConfigKeys` into actual monitor names.
//
// Parameters:
//   - fct: The function to register. It should take a string key and return a
//     `cfgtps.ComponentMonitor` interface.
func (o *sts) RegisterGetConfigCpt(fct FuncGetCfgCpt) {
	o.m.Lock()
	defer o.m.Unlock()
	o.n = fct
}

// GetConfig retrieves the current configuration used by the status instance.
// It reconstructs the `Config` struct from the internal state stored in the
// context configuration.
//
// This is useful for inspecting the active configuration at runtime.
//
// Returns:
//   The current `Config` object. If no configuration has been set, it returns
//   a default configuration with standard HTTP codes and an empty component list.
func (o *sts) GetConfig() Config {
	var (
		cfg = Config{
			ReturnCode: make(map[monsts.Status]int, 0),
			Component:  make([]Mandatory, 0),
		}
	)

	// Set default return codes
	cfg.ReturnCode[monsts.KO] = http.StatusInternalServerError
	cfg.ReturnCode[monsts.Warn] = http.StatusMultiStatus
	cfg.ReturnCode[monsts.OK] = http.StatusOK

	// Try to load return codes from internal storage
	if i, l := o.x.Load(keyConfigReturnCode); !l || i == nil {
		// Not found or nil, keep defaults
	} else if r, k := i.(map[monsts.Status]int); !k || len(r) < 1 {
		// Invalid type or empty, keep defaults
	} else {
		cfg.ReturnCode = r
	}

	// Try to load mandatory components from internal storage
	if i, l := o.x.Load(keyConfigMandatory); !l || i == nil {
		// Not found or nil, keep empty list
	} else if r, k := i.(stslmd.ListMandatory); !k || r == nil || r.Len() < 1 {
		// Invalid type or empty, keep empty list
	} else {
		// Reconstruct the slice of Mandatory structs from the internal list
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
// It updates the HTTP return codes and the list of mandatory component groups.
//
// This method handles the logic for:
//   1. Setting default HTTP return codes if none are provided.
//   2. Creating a new internal list of mandatory groups.
//   3. Resolving dynamic `ConfigKeys` into monitor names using the registered
//      resolver function (if available).
//   4. Storing the configuration in the thread-safe context storage.
//
// Parameters:
//   - cfg: The configuration to apply.
func (o *sts) SetConfig(cfg Config) {
	// Handle default return codes
	if len(cfg.ReturnCode) < 1 {
		var def = make(map[monsts.Status]int, 0)
		def[monsts.KO] = http.StatusInternalServerError
		def[monsts.Warn] = http.StatusMultiStatus
		def[monsts.OK] = http.StatusOK

		o.x.Store(keyConfigReturnCode, def)
	} else {
		o.x.Store(keyConfigReturnCode, cfg.ReturnCode)
	}

	// Create a new list for mandatory groups
	var lst = stslmd.New()

	if len(cfg.Component) > 0 {
		for _, i := range cfg.Component {
			var m = stsmdt.New()
			m.SetMode(i.Mode)
			
			// Add static keys
			m.KeyAdd(i.Keys...)
			
			// Resolve and add dynamic keys from ConfigKeys
			if len(i.ConfigKeys) > 0 && o.n != nil {
				for _, k := range i.ConfigKeys {
					// Call the registered resolver function
					if c := o.n(k); c != nil {
						// Add the monitor names returned by the component
						m.KeyAdd(c.GetMonitorNames()...)
					}
				}
			}
			lst.Add(m)
		}
	}

	// Store the new list in the context configuration
	o.x.Store(keyConfigMandatory, lst)
}

// cfgGetReturnCode retrieves the configured HTTP status code for a specific
// health status (OK, Warn, KO).
//
// It accesses the thread-safe configuration storage. If the configuration is
// missing or invalid, it defaults to `http.StatusInternalServerError` (500)
// as a safety measure.
//
// Parameters:
//   - s: The health status to look up.
//
// Returns:
//   The corresponding HTTP status code.
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

// cfgGetMandatory retrieves the internal list of mandatory component groups.
//
// It accesses the thread-safe configuration storage.
//
// Returns:
//   The `stslmd.ListMandatory` interface containing the groups, or nil if
//   not configured.
func (o *sts) cfgGetMandatory() stslmd.ListMandatory {
	if i, l := o.x.Load(keyConfigMandatory); !l {
		return nil
	} else if v, k := i.(stslmd.ListMandatory); !k {
		return nil
	} else {
		return v
	}
}

// cfgGetMode retrieves the control mode for a specific component key.
//
// It delegates the lookup to the internal mandatory list. If the component
// is not found in any mandatory group, it returns `stsctr.Ignore`.
//
// Parameters:
//   - key: The component name to look up.
//
// Returns:
//   The `stsctr.Mode` associated with the component.
func (o *sts) cfgGetMode(key string) stsctr.Mode {
	if l := o.cfgGetMandatory(); l == nil {
		return stsctr.Ignore
	} else {
		return l.GetMode(key)
	}
}

// cfgGetOne retrieves all component names that belong to the same mandatory
// group as the specified key.
//
// This is used for evaluating group-based control modes like `AnyOf` and `Quorum`,
// where the status of one component depends on the status of its peers in the group.
//
// Parameters:
//   - key: The component name to find the group for.
//
// Returns:
//   A slice of strings containing all keys in the matching group. Returns an
//   empty slice if the key is not found or no groups are configured.
func (o *sts) cfgGetOne(key string) []string {
	if l := o.cfgGetMandatory(); l == nil {
		return make([]string, 0)
	} else {
		var r []string
		// Walk through the list to find the group containing the key
		l.Walk(func(m stsmdt.Mandatory) bool {
			if m.KeyHas(key) {
				r = m.KeyList()
				return false // Stop searching once found
			}

			return true
		})
		return r
	}
}
