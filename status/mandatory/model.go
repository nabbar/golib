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

package mandatory

import (
	"slices"
	"sync/atomic"

	stsctr "github.com/nabbar/golib/status/control"
)

// model is the internal implementation of the Mandatory interface.
//
// It uses atomic.Value for lock-free concurrent access to both the mode
// and the keys list. This provides excellent performance for read-heavy
// workloads while maintaining thread safety.
type model struct {
	// Mode stores the control.Mode value atomically
	Mode *atomic.Value

	// Keys stores a []string slice atomically
	Keys *atomic.Value
}

// SetMode implements Mandatory.SetMode.
//
// This method atomically updates the validation mode. The operation is
// thread-safe and wait-free.
func (o *model) SetMode(m stsctr.Mode) {
	o.Mode.Store(m)
}

// GetMode implements Mandatory.GetMode.
//
// This method atomically loads the current validation mode. If the mode
// has never been set or is nil, it returns control.Ignore as the default.
//
// The operation is thread-safe and lock-free.
func (o *model) GetMode() stsctr.Mode {
	m := o.Mode.Load()

	if m != nil {
		return m.(stsctr.Mode)
	}

	return stsctr.Ignore
}

// KeyHas implements Mandatory.KeyHas.
//
// This method atomically loads the keys list and checks if the specified
// key exists using a case-sensitive exact match.
//
// The operation is thread-safe and lock-free. It returns false if the
// keys list is nil or not properly initialized.
func (o *model) KeyHas(key string) bool {
	i := o.Keys.Load()

	if i == nil {
		return false
	} else if l, k := i.([]string); !k {
		return false
	} else {
		return slices.Contains(l, key)
	}
}

// KeyAdd implements Mandatory.KeyAdd.
//
// This method atomically loads the current keys list, adds any new keys
// that don't already exist, and stores the updated list back.
//
// Duplicate keys are automatically filtered out. The operation is thread-safe
// but may retry if there's contention (optimistic concurrency).
//
// If the keys list is nil or corrupted, it creates a new empty list.
func (o *model) KeyAdd(keys ...string) {
	var (
		i any
		k bool
		l []string
	)

	i = o.Keys.Load()

	if i == nil {
		l = make([]string, 0)
	} else if l, k = i.([]string); !k {
		l = make([]string, 0)
	}

	for _, key := range keys {
		if !slices.Contains(l, key) {
			l = append(l, key)
		}
	}

	o.Keys.Store(l)
}

// KeyDel implements Mandatory.KeyDel.
//
// This method atomically loads the current keys list, removes the specified
// keys, and stores the filtered list back.
//
// Keys that don't exist are silently ignored. The operation is thread-safe
// but may retry if there's contention (optimistic concurrency).
//
// If the keys list is nil or corrupted, it stores an empty list.
func (o *model) KeyDel(keys ...string) {
	var (
		i any
		k bool
		l []string
	)

	i = o.Keys.Load()

	if i == nil {
		o.Keys.Store(make([]string, 0))
		return
	} else if l, k = i.([]string); !k {
		o.Keys.Store(make([]string, 0))
		return
	}

	var res = make([]string, 0)

	for _, key := range l {
		if !slices.Contains(keys, key) {
			res = append(res, key)
		}
	}

	o.Keys.Store(res)
}

// KeyList implements Mandatory.KeyList.
//
// This method atomically loads the current keys list and returns a copy
// using slices.Clone. The returned slice can be safely modified without
// affecting the internal state.
//
// The operation is thread-safe and lock-free. It returns an empty slice
// if the keys list is nil or corrupted.
func (o *model) KeyList() []string {
	var (
		i any
		k bool
		l []string
	)

	i = o.Keys.Load()

	if i == nil {
		return make([]string, 0)
	} else if l, k = i.([]string); !k {
		return make([]string, 0)
	}

	return slices.Clone(l)
}
