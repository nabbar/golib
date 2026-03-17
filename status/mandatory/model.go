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

	libatm "github.com/nabbar/golib/atomic"
	stsctr "github.com/nabbar/golib/status/control"
)

// model is the internal, thread-safe implementation of the `Mandatory` interface.
// It uses atomic operations for all its fields to ensure high-performance,
// lock-free reads and safe concurrent writes.
type model struct {
	// Mode stores the `control.Mode` as a `uint32` to allow for atomic operations.
	Mode *atomic.Uint32

	// Name stores the human-readable identifier for the group. It is wrapped in
	// a generic atomic value (`libatm.Value`) to ensure thread-safe access.
	Name libatm.Value[string]

	// Info stores descriptive metadata for the group (e.g., description, links)
	// in a thread-safe map.
	Info libatm.MapTyped[string, interface{}]

	// Keys stores the `[]string` slice of component keys. `atomic.Value` is used
	// to ensure that the slice is always accessed and updated atomically.
	Keys *atomic.Value
}

// SetMode atomically updates the validation mode for this group.
func (o *model) SetMode(m stsctr.Mode) {
	o.Mode.Store(m.Uint32())
}

// GetMode atomically retrieves the current validation mode. It performs a lock-free
// read.
func (o *model) GetMode() stsctr.Mode {
	return stsctr.ParseUint32(o.Mode.Load())
}

// SetName atomically updates the name of the mandatory group.
// The input string is sanitized by `GetNameOrDefault` to guarantee a valid,
// non-empty identifier.
func (o *model) SetName(s string) {
	o.Name.Store(GetNameOrDefault(s))
}

// GetName atomically retrieves the current name of the mandatory group.
func (o *model) GetName() string {
	return o.Name.Load()
}

// SetInfo populates the group's metadata with a map of key-value pairs.
// This method iterates through the provided map and stores each entry in the
// thread-safe `Info` map, replacing any existing entries with the same keys.
// Nil values or empty keys are ignored.
func (o *model) SetInfo(info map[string]interface{}) {
	for key, val := range info {
		if len(key) < 1 {
			continue
		}
		if val == nil {
			continue
		}
		o.Info.Store(key, val)
	}
}

// AddInfo adds or updates a single piece of metadata for the group.
// This is a thread-safe operation that stores the key-value pair in the `Info` map.
// Nil values or empty keys are ignored.
func (o *model) AddInfo(key string, value interface{}) {
	if o == nil || o.Info == nil {
		return
	}

	if len(key) < 1 {
		return
	}

	if value == nil {
		return
	}

	o.Info.Store(key, value)
}

// GetInfo retrieves all metadata associated with the group.
// It iterates through the thread-safe `Info` map and returns a standard
// `map[string]interface{}` containing all the entries.
func (o *model) GetInfo() map[string]interface{} {
	if o == nil || o.Info == nil {
		return make(map[string]interface{})
	}

	var res = make(map[string]interface{})

	o.Info.Range(func(key string, value interface{}) bool {
		if len(key) < 1 {
			return true
		}

		if value == nil {
			return true
		}

		res[key] = value
		return true
	})

	return res
}

// KeyHas atomically checks if a given key exists in the group. This is a
// high-performance, lock-free read operation. It loads the current slice from
// `atomic.Value` and then iterates through it.
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

// KeyAdd atomically adds one or more keys to the group. This method implements a
// read-modify-write pattern on the `atomic.Value`. It first loads the current
// slice, creates a new slice with the added keys (while checking for duplicates),
// and then atomically stores the new slice back. This ensures that concurrent
// additions do not result in a corrupted state.
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
		// If the stored value is not a []string, start with a fresh slice.
		l = make([]string, 0)
	}

	// Create a new slice to avoid modifying the one currently in use by other goroutines.
	l = slices.Clone(l)

	for _, key := range keys {
		if !slices.Contains(l, key) {
			l = append(l, key)
		}
	}

	o.Keys.Store(slices.Clone(l))
}

// KeyDel atomically removes one or more keys from the group. Similar to `KeyAdd`,
// this method uses a read-modify-write pattern. It loads the current slice,
// creates a new slice containing only the elements to keep, and then atomically
// stores the new slice.
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

	// Filter out the keys to be deleted, creating a new slice.
	var res = make([]string, 0, len(l))
	for _, key := range l {
		if !slices.Contains(keys, key) {
			res = append(res, key)
		}
	}

	o.Keys.Store(res)
}

// KeyList atomically retrieves a copy of all keys in the group. It performs a
// lock-free read and then returns a defensive copy of the slice using `slices.Clone`.
// This is crucial to prevent the caller from modifying the internal slice, which
// could lead to race conditions.
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
