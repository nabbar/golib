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

package info

import (
	libatm "github.com/nabbar/golib/atomic"
	montps "github.com/nabbar/golib/monitor/types"
)

const (
	// keyName is the internal key used to store a manually overridden component name in the atomic map.
	keyName = "__keyName__"
	// fctName is the internal key used to store the provider function for a dynamic component name.
	fctName = "__funcName__"
	// fctInfo is the internal key used to store the provider function for dynamic metadata (Info data).
	fctInfo = "__funcInfo__"
)

// inf is the private concrete implementation of the Info interface.
// It manages the metadata for a monitored component using an atomic map for thread-safe state storage.
// It supports a hierarchy of naming (Dynamic Function > Manual Override > Default) and dynamic data collection.
type inf struct {
	// n is the immutable default name provided during the initialization of the Info instance.
	n string
	// v is a thread-safe atomic map used to store all metadata, including internal system keys
	// and user-provided key-value pairs.
	v libatm.Map[string]
}

// SetName provides a way to manually override the component's name.
// If an empty string is provided, any existing override is removed, effectively resetting the name
// to either the dynamic function's result or the default name.
// This operation is thread-safe.
func (o *inf) SetName(s string) {
	if o == nil {
		return
	} else if len(s) < 1 {
		o.v.Delete(keyName)
	} else {
		o.v.Store(keyName, s)
	}
}

// SetData replaces all user-defined metadata with the contents of the provided map.
// Internal system keys (for name overrides and provider functions) are preserved during this operation.
// If the provided map is nil or empty, all current user-defined metadata is cleared.
// This operation is thread-safe.
func (o *inf) SetData(m map[string]interface{}) {
	if o == nil {
		return
	}

	// Identify all keys that are NOT internal system keys.
	var i = make([]string, 0)
	o.v.Range(func(key string, _ interface{}) bool {
		if len(key) < 1 {
			return true
		} else if key == keyName || key == fctName || key == fctInfo {
			return true
		}

		i = append(i, key)
		return true
	})

	// Remove all identified user-defined keys.
	for _, v := range i {
		o.v.Delete(v)
	}

	if len(m) < 1 {
		return
	}

	// Batch update with the new metadata.
	for k, v := range m {
		if len(k) > 0 && v != nil {
			o.v.Store(k, v)
		}
	}
}

// AddData inserts or updates a single metadata entry.
// If the key (s) is already present, its value is overwritten.
// If the value (i) is nil, the key is removed from the metadata store.
// This operation is thread-safe.
func (o *inf) AddData(s string, i interface{}) {
	if o == nil {
		return
	} else if len(s) < 1 {
		return
	} else if i == nil {
		o.v.Delete(s)
	} else {
		o.v.Store(s, i)
	}
}

// DelData removes a specific metadata entry identified by its key.
// If the key does not exist or is empty, the operation completes without error.
// This operation is thread-safe.
func (o *inf) DelData(s string) {
	if o == nil {
		return
	} else if len(s) < 1 {
		return
	}

	o.v.Delete(s)
}

// RegisterName associates a dynamic name provider function with this Info instance.
// When Name() is called, this function will be executed to determine the current name.
// Results are not cached; the function is invoked upon every request.
// Providing a nil function unregisters the current provider.
func (o *inf) RegisterName(fct montps.FuncInfoName) {
	if o == nil {
		return
	} else if fct == nil {
		o.v.Delete(fctName)
	} else {
		o.v.Store(fctName, fct)
	}
}

// RegisterData associates a dynamic data provider function with this Info instance.
// When Data() is called, this function will be executed, and its results will be merged
// with the static metadata stored in the instance.
// Providing a nil function unregisters the current provider.
func (o *inf) RegisterData(fct montps.FuncInfoData) {
	if o == nil {
		return
	} else if fct == nil {
		o.v.Delete(fctInfo)
	} else {
		o.v.Store(fctInfo, fct)
	}
}

// callNameFct is an internal helper that executes the registered name provider function.
// It handles type assertions and basic validation, returning an error if the function
// is missing or returns invalid data.
func (o *inf) callNameFct() (string, error) {
	if o == nil {
		return "", montps.ErrorInvalidInstance.Error()
	} else if i, l := o.v.Load(fctName); !l || i == nil {
		return "", montps.ErrorParamEmpty.Error()
	} else if v, k := i.(montps.FuncInfoName); !k || v == nil {
		return "", montps.ErrorParamEmpty.Error()
	} else {
		return v()
	}
}

// callInfoFct is an internal helper that executes the registered data provider function.
// It handles type assertions and basic validation, returning an error if the function
// is missing or returns invalid data.
func (o *inf) callInfoFct() (map[string]interface{}, error) {
	if o == nil {
		return nil, montps.ErrorInvalidInstance.Error()
	} else if i, l := o.v.Load(fctInfo); !l || i == nil {
		return nil, montps.ErrorParamEmpty.Error()
	} else if v, k := i.(montps.FuncInfoData); !k || v == nil {
		return nil, montps.ErrorParamEmpty.Error()
	} else {
		return v()
	}
}

// getName is the internal core logic for resolving the component's name.
// The resolution order is:
// 1. Result of the registered dynamic function (if any and non-empty).
// 2. Manually set name override (if any).
// 3. The default name provided at initialization.
func (o *inf) getName() string {
	if o == nil {
		return ""
	}

	// Priority 1: Dynamic function.
	if n, e := o.callNameFct(); e == nil && len(n) > 0 {
		return n
	}

	// Priority 2: Manual override.
	if i, l := o.v.Load(keyName); l && i != nil {
		if v, k := i.(string); k {
			return v
		}
	}

	// Priority 3: Default fallback.
	return o.n
}

// getInfoStore is an internal helper that extracts all user-defined metadata from the atomic map.
// It filters out internal system keys used for configuration and naming logic.
func (o *inf) getInfoStore() map[string]interface{} {
	if o == nil {
		return nil
	}

	var result = make(map[string]interface{})

	o.v.Range(func(key string, value any) bool {
		if len(key) < 1 {
			return true
		} else if key == keyName || key == fctName || key == fctInfo {
			return true
		}

		result[key] = value
		return true
	})

	return result
}

// getInfo is the internal core logic for resolving all metadata (Info data).
// It retrieves the static metadata stored in the instance and merges it with the result
// of the registered dynamic data provider function (if any).
// If keys collide, the dynamic function's data takes precedence.
func (o *inf) getInfo() map[string]interface{} {
	if o == nil {
		return nil
	}

	// Start with static data stored in the atomic map.
	var res = o.getInfoStore()
	if len(res) < 1 {
		res = make(map[string]interface{})
	}

	// Execute the dynamic provider function.
	f, e := o.callInfoFct()
	if e != nil || len(f) < 1 {
		return res
	}

	// Merge dynamic data into the result map.
	for k, v := range f {
		res[k] = v
	}

	return res
}
