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

import "sync"

const (
	keyDefName = "__keyDefaultName__"
	keyName    = "__keyName__"
)

// inf is the internal implementation of the Info interface.
// It uses sync.RWMutex for thread-safe operations and sync.Map for efficient concurrent access.
type inf struct {
	m sync.RWMutex // protects fi, ri, fn, rn fields
	v sync.Map     // stores name and info data

	fi FuncInfo // registered function to retrieve info data
	ri bool     // indicates if info function is registered and needs to be called

	fn FuncName // registered function to retrieve name
	rn bool     // indicates if name function is registered and needs to be called
}

// RegisterName registers a function that returns a dynamic name.
// This clears any cached name and sets the function to be called on next Name() access.
// The function is thread-safe and can be called concurrently.
func (o *inf) RegisterName(fct FuncName) {
	o.m.Lock()
	defer o.m.Unlock()

	o.v.Delete(keyName)
	o.fn = fct
	o.rn = true
}

// RegisterInfo registers a function that returns dynamic info data.
// This clears any cached info data (except internal keys) and sets the function to be called on next Info() access.
// The function is thread-safe and can be called concurrently.
func (o *inf) RegisterInfo(fct FuncInfo) {
	o.m.Lock()
	defer o.m.Unlock()

	o.v.Range(func(key, value any) bool {
		if s, k := key.(string); !k {
			o.v.Delete(key)
			return true
		} else if s == keyName {
			return true
		} else if s == keyDefName {
			return true
		} else {
			o.v.Delete(key)
			return true
		}
	})

	o.fi = fct
	o.ri = true
}

// isName checks if a name function is registered and needs to be called.
func (o *inf) isName() bool {
	o.m.RLock()
	defer o.m.RUnlock()

	return o.rn && o.fn != nil
}

// isInfo checks if an info function is registered and needs to be called.
func (o *inf) isInfo() bool {
	o.m.RLock()
	defer o.m.RUnlock()

	return o.ri && o.fi != nil
}

// callName invokes the registered name function and caches the result.
// If the function returns an error, the default name is returned instead.
func (o *inf) callName() (string, error) {
	o.m.RLock()
	defer o.m.RUnlock()

	if v, e := o.fn(); e != nil {
		return o.defaultName(), e
	} else {
		o.v.Store(keyName, v)
		return v, nil
	}
}

// callInfo invokes the registered info function and caches the results.
// If the function returns an error, nil is returned.
func (o *inf) callInfo() (map[string]interface{}, error) {
	o.m.RLock()
	defer o.m.RUnlock()

	if i, e := o.fi(); e != nil {
		return nil, e
	} else {
		for k, v := range i {
			o.v.Store(k, v)
		}
		return i, nil
	}
}

// getName retrieves the name by calling the registered function if available,
// or returns the default name. After a successful call, the function is marked
// as complete (rn = false) to use the cached value on subsequent calls.
func (o *inf) getName() string {
	if !o.isName() {
		return o.defaultName()
	}

	i, e := o.callName()

	if e == nil {
		o.m.Lock()
		defer o.m.Unlock()
		o.rn = false
	}

	return i
}

// getInfo retrieves the info data by calling the registered function if available.
// After a successful call, the function is marked as complete (ri = false)
// to use the cached values on subsequent calls.
func (o *inf) getInfo() map[string]interface{} {
	if !o.isInfo() {
		return nil
	}

	i, e := o.callInfo()

	if e == nil {
		o.m.Lock()
		defer o.m.Unlock()
		o.ri = false
	}

	return i
}
