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

// model is the internal, thread-safe implementation of the `Mandatory` interface.
// It uses `atomic.Value` to store both the control mode and the list of keys.
// This design choice provides high-performance, lock-free reads, which is ideal
// for scenarios where health checks are frequent, while still ensuring safety
// for writes.
type model struct {
	// Mode stores the `control.Mode` value. Using `atomic.Value` allows for
	// thread-safe reads and writes without traditional mutexes.
	Mode *atomic.Value

	// Keys stores the `[]string` slice of component keys. `atomic.Value` is used
	// here as well to ensure that the slice is always accessed and updated atomically.
	Keys *atomic.Value
}

// SetMode atomically updates the validation mode for this group. The new mode is
// stored in the `atomic.Value`, replacing the old one.
func (o *model) SetMode(m stsctr.Mode) {
	o.Mode.Store(m)
}

// GetMode atomically retrieves the current validation mode. It performs a lock-free
// read from the `atomic.Value`. If the stored value is nil or not of the expected
// type, it safely returns the default `control.Ignore` mode.
func (o *model) GetMode() stsctr.Mode {
	m := o.Mode.Load()

	if m != nil {
		// Type assertion is safe here because we control what's stored.
		return m.(stsctr.Mode)
	}

	return stsctr.Ignore
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

	o.Keys.Store(l)
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
