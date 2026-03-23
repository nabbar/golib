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
	// keyName is the key used to store the cached name in the atomic map.
	keyName = "__keyName__"
	// fctName is the key used to store the registered name function.
	fctName = "__funcName__"
	// fctInfo is the key used to store the registered info function.
	fctInfo = "__funcInfo__"
)

// inf is the internal implementation of the Info interface.
// It uses a thread-safe atomic map to store its state, including the default name,
// registered functions, and any manually set data.
type inf struct {
	// n holds the default name provided at creation.
	n string
	// v is a thread-safe map that stores all dynamic data, including the
	// cached name, registered functions, and info key-value pairs.
	v libatm.Map[string]
}

// SetName manually sets or overrides the cached name.
// If an empty string is provided, the cached name is deleted, causing Name()
// to fall back to the default name.
func (o *inf) SetName(s string) {
	if o == nil {
		return
	} else if len(s) < 1 {
		o.v.Delete(keyName)
	} else {
		o.v.Store(keyName, s)
	}
}

// SetData replaces all existing info data with the provided map.
// Internal keys for functions and name are preserved.
// If a nil or empty map is provided, all existing info data is cleared.
func (o *inf) SetData(m map[string]interface{}) {
	if o == nil {
		return
	}

	// Collect all non-internal keys to be deleted.
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

	// Delete the collected keys.
	for _, v := range i {
		o.v.Delete(v)
	}

	if len(m) < 1 {
		return
	}

	// Store the new data.
	for k, v := range m {
		if len(k) > 0 && v != nil {
			o.v.Store(k, v)
		}
	}
}

// AddData adds a single key-value pair to the info data.
// If the key already exists, its value is updated.
// If the provided value is nil, the key is deleted.
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

// DelData removes a single key-value pair from the info data by its key.
func (o *inf) DelData(s string) {
	if o == nil {
		return
	} else if len(s) < 1 {
		return
	}

	o.v.Delete(s)
}

// RegisterName registers a function that dynamically provides the component name.
// Calling this method will cause the next call to Name() to invoke this function.
// The implementation does not cache the result; the function is called on every
// invocation of Name().
// Providing a nil function unregisters any existing function.
func (o *inf) RegisterName(fct montps.FuncInfoName) {
	if o == nil {
		return
	} else if fct == nil {
		o.v.Delete(fctName)
	} else {
		o.v.Store(fctName, fct)
	}
}

// RegisterInfo registers a function that dynamically provides the info data map.
// Calling this method will cause the next call to Info() to invoke this function.
// The implementation does not cache the result; the function is called on every
// invocation of Info().
// Providing a nil function unregisters any existing function.
func (o *inf) RegisterInfo(fct montps.FuncInfoData) {
	if o == nil {
		return
	} else if fct == nil {
		o.v.Delete(fctInfo)
	} else {
		o.v.Store(fctInfo, fct)
	}
}

// callNameFct retrieves and executes the registered name function.
// It returns an error if the instance is nil, the function is not registered,
// or the stored value is of the wrong type.
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

// callInfoFct retrieves and executes the registered info function.
// It returns an error if the instance is nil, the function is not registered,
// or the stored value is of the wrong type.
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

// getName is the internal logic for retrieving the name.
// It prioritizes the registered function, then a manually set name,
// and finally falls back to the default name.
func (o *inf) getName() string {
	if o == nil {
		return ""
	}

	// Try to get name from registered function.
	if n, e := o.callNameFct(); e == nil && len(n) > 0 {
		return n
	}

	// Fallback to manually set/cached name.
	if i, l := o.v.Load(keyName); l && i != nil {
		if v, k := i.(string); k {
			return v
		}
	}

	// Fallback to default name.
	return o.n
}

// getInfoStore retrieves all non-internal key-value pairs from the atomic map.
// It returns a standard map of the current data.
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

// getInfo is the internal logic for retrieving the info data.
// It merges data from the store with data from the registered function.
func (o *inf) getInfo() map[string]interface{} {
	if o == nil {
		return nil
	}

	// Start with any manually set data.
	var res = o.getInfoStore()
	if len(res) < 1 {
		res = make(map[string]interface{})
	}

	// Try to get additional data from registered function.
	f, e := o.callInfoFct()
	if e != nil || len(f) < 1 {
		return res
	}

	// Merge function data into the result.
	for k, v := range f {
		res[k] = v
	}

	return res
}
