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

package listmandatory

import (
	"slices"
	"sort"
	"sync/atomic"

	libatm "github.com/nabbar/golib/atomic"
	stsctr "github.com/nabbar/golib/status/control"
	stsmdt "github.com/nabbar/golib/status/mandatory"
)

// model is the internal implementation of the ListMandatory interface.
//
// It uses sync.Map for thread-safe concurrent access to the collection of
// mandatory groups. Each group is stored with a unique int32 ID that is
// atomically incremented.
type model struct {
	// l stores mandatory groups with int32 keys
	l libatm.MapTyped[uint32, stsmdt.Mandatory]

	// k is an atomic counter for generating unique IDs
	k *atomic.Uint32
}

// Len implements ListMandatory.Len.
//
// This method iterates through all entries in the sync.Map and counts valid
// mandatory groups. Invalid entries (wrong key or value types) are automatically
// deleted during the iteration.
//
// The operation is thread-safe.
func (o *model) Len() int {
	var l int

	o.l.Range(func(k uint32, v stsmdt.Mandatory) bool {
		if v == nil {
			o.l.Delete(k)
		} else {
			l += 1
		}
		return true
	})

	return l
}

// Walk implements ListMandatory.Walk.
//
// This method iterates through all mandatory groups in the sync.Map and calls
// the provided function for each valid group. Invalid entries (wrong key or
// value types) are automatically deleted during iteration.
//
// The iteration stops early if the function returns false.
//
// The operation is thread-safe.
func (o *model) Walk(fct func(m stsmdt.Mandatory) bool) {
	o.l.Range(func(k uint32, v stsmdt.Mandatory) bool {
		if v == nil {
			o.l.Delete(k)
			return true
		} else {
			r := fct(v)
			o.l.Store(k, v)
			return r
		}
	})
}

// Add implements ListMandatory.Add.
//
// This method adds one or more mandatory groups to the list. Each group is
// assigned a unique ID using the atomic counter.
//
// The operation is thread-safe.
func (o *model) Add(m ...stsmdt.Mandatory) {
	for _, v := range m {
		if m == nil {
			continue
		}

		o.k.Add(1)
		o.l.Store(o.k.Load(), v)
	}
}

// Del implements ListMandatory.Del.
//
// This method removes a mandatory group from the list by comparing key lists.
// The key lists are sorted before comparison to ensure order-independent matching.
// Only the first matching group is removed.
//
// Invalid entries are cleaned up during the search.
//
// The operation is thread-safe.
func (o *model) Del(m stsmdt.Mandatory) {
	c := m.KeyList()
	sort.Strings(c)

	o.l.Range(func(k uint32, v stsmdt.Mandatory) bool {
		if v == nil {
			o.l.Delete(k)
		}

		u := v.KeyList()
		sort.Strings(u)

		if slices.Compare(c, u) == 0 {
			o.l.Delete(k)
		}
		return true
	})
}

// GetMode implements ListMandatory.GetMode.
//
// This method searches through all mandatory groups and returns the mode of
// the first group that contains the specified key. If no group contains the
// key, it returns control.Ignore.
//
// Invalid entries are cleaned up during the search.
//
// The operation is thread-safe.
func (o *model) GetMode(key string) stsctr.Mode {
	var res = stsctr.Ignore
	o.l.Range(func(k uint32, v stsmdt.Mandatory) bool {
		if v == nil {
			o.l.Delete(k)
		} else if v.KeyHas(key) {
			res = v.GetMode()
			return false
		}
		return true
	})
	return res
}

// SetMode implements ListMandatory.SetMode.
//
// This method searches through all mandatory groups and updates the mode of
// the first group that contains the specified key. The search stops after
// the first match.
//
// Invalid entries are cleaned up during the search.
//
// The operation is thread-safe.
func (o *model) SetMode(key string, mod stsctr.Mode) {
	o.l.Range(func(k uint32, v stsmdt.Mandatory) bool {
		if v == nil {
			o.l.Delete(k)
		} else if v.KeyHas(key) {
			v.SetMode(mod)
			o.l.Store(k, v)
			return false
		}
		return true
	})
}

// GetList implements ListMandatory.GetList.
//
// This method iterates through all mandatory groups in the sync.Map and
// collects them into a slice. Invalid entries (nil values) are automatically
// deleted during iteration.
//
// The operation is thread-safe.
func (o *model) GetList() []stsmdt.Mandatory {
	var res []stsmdt.Mandatory
	o.l.Range(func(k uint32, v stsmdt.Mandatory) bool {
		if v == nil {
			o.l.Delete(k)
		}
		res = append(res, v)
		return true
	})
	return res
}
