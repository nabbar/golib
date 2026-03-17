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

	libatm "github.com/nabbar/golib/atomic"
	stsctr "github.com/nabbar/golib/status/control"
	stsmdt "github.com/nabbar/golib/status/mandatory"
)

// model is the internal implementation of the ListMandatory interface.
//
// It uses a `libatm.MapTyped[string, stsmdt.Mandatory]`, which is a type-safe
// wrapper around `sync.Map`, to provide thread-safe concurrent access to the
// collection of mandatory groups. Each group is stored and identified by a
// string key, which is the group's name.
type model struct {
	// l stores mandatory groups, mapping a group's name to its instance.
	l libatm.MapTyped[string, stsmdt.Mandatory]
}

// Len implements ListMandatory.Len.
//
// This method iterates through all entries in the underlying thread-safe map
// and counts the number of valid mandatory groups.
//
// As a self-healing mechanism, any entry that is found to be invalid (e.g., a nil
// group or a group with no keys) is automatically purged from the map during
// the iteration. This ensures the list remains in a consistent state.
//
// The operation is fully thread-safe.
func (o *model) Len() int {
	var l int

	o.l.Range(func(k string, v stsmdt.Mandatory) bool {
		if v == nil {
			o.l.Delete(k)
			return true
		}
		if len(v.KeyList()) < 1 {
			o.l.Delete(k)
			return true
		}

		l += 1
		return true
	})

	return l
}

// Walk implements ListMandatory.Walk.
//
// This method uses the `Range` function of the underlying map to safely iterate
// over all stored mandatory groups. For each valid group, it executes the
// provided callback function `fct`, passing the group's name (key) and the group
// instance itself.
//
// The iteration will stop prematurely if the callback function returns `false`.
// Any modifications made to the `stsmdt.Mandatory` instance within the callback
// are persisted as it is a reference type. The entry is explicitly resaved using
// `Store` to guarantee the update in all scenarios.
//
// The operation is thread-safe and includes self-healing for invalid entries.
func (o *model) Walk(fct func(k string, m stsmdt.Mandatory) bool) {
	o.l.Range(func(k string, v stsmdt.Mandatory) bool {
		if v == nil {
			o.l.Delete(k)
			return true
		}
		if len(v.KeyList()) < 1 {
			o.l.Delete(k)
			return true
		}

		r := fct(k, v)
		o.l.Store(k, v)
		return r
	})
}

// Add implements ListMandatory.Add.
//
// This method iterates through the provided mandatory groups. For each valid
// group (not nil and containing at least one key), it ensures the group has a
// name by calling `stsmdt.GetNameOrDefault`. This name is then used as the key
// to store the group in the map.
//
// If a group with the same name already exists, it will be overwritten.
// The operation is thread-safe.
func (o *model) Add(m ...stsmdt.Mandatory) {
	for _, v := range m {
		if v == nil {
			continue
		}
		if len(v.KeyList()) < 1 {
			continue
		}

		s := stsmdt.GetNameOrDefault(v.GetName())
		v.SetName(s)
		o.l.Store(s, v)
	}
}

// DelKey implements ListMandatory.DelKey.
//
// This method provides a direct way to remove a group by its name. It calls
// the `Delete` method on the underlying map, which is a thread-safe operation.
func (o *model) DelKey(s string) {
	o.l.Delete(s)
}

// Del implements ListMandatory.Del.
//
// This method removes one or more mandatory groups from the list by matching
// their content. It first sorts the key list of the provided group `m` for
// consistent comparison.
//
// It then iterates through all groups in the map. For each group, it sorts its
// key list and compares it with the sorted key list of `m` using `slices.Compare`.
// If the key lists are identical, the group is removed from the map.
//
// The iteration continues through all entries, so if multiple groups happen to
// have the same set of keys, all of them will be removed.
// The operation is thread-safe.
func (o *model) Del(m stsmdt.Mandatory) {
	if m == nil {
		return
	}

	c := m.KeyList()
	sort.Strings(c)

	o.l.Range(func(k string, v stsmdt.Mandatory) bool {
		if v == nil {
			o.l.Delete(k)
			return true
		}
		if len(v.KeyList()) < 1 {
			o.l.Delete(k)
			return true
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
// This method searches for a given `key` across all mandatory groups to find
// its associated validation mode. It iterates through the groups and, for each
// one, checks if it contains the key using `v.KeyHas(key)`.
//
// The search implements a "first match wins" logic. As soon as a group
// containing the key is found, its mode is captured, and the iteration is
// stopped (by returning `false`). If the entire list is searched and no group
// contains the key, the default mode `stsctr.Ignore` is returned.
//
// The operation is thread-safe and includes self-healing for invalid entries.
func (o *model) GetMode(key string) stsctr.Mode {
	var res = stsctr.Ignore
	o.l.Range(func(k string, v stsmdt.Mandatory) bool {
		if v == nil {
			o.l.Delete(k)
			return true
		}

		if len(v.KeyList()) < 1 {
			o.l.Delete(k)
			return true
		}

		if v.KeyHas(key) {
			res = v.GetMode()
			return false // Stop iteration on first match
		}

		return true
	})
	return res
}

// SetMode implements ListMandatory.SetMode.
//
// This method finds the first group containing the given `key` and updates its
// validation mode. It iterates through the groups, and upon finding the first
// match, it calls `v.SetMode(mod)` on that group.
//
// The updated group is then saved back into the map using `o.l.Store(k, v)` to
// ensure the change is persisted. The iteration is then stopped by returning `false`.
//
// The operation is thread-safe and includes self-healing for invalid entries.
func (o *model) SetMode(key string, mod stsctr.Mode) {
	o.l.Range(func(k string, v stsmdt.Mandatory) bool {
		if v == nil {
			o.l.Delete(k)
			return true
		}

		if len(v.KeyList()) < 1 {
			o.l.Delete(k)
			return true
		}

		if v.KeyHas(key) {
			v.SetMode(mod)
			o.l.Store(k, v)
			return false // Stop iteration on first match
		}

		return true
	})
}

// GetList implements ListMandatory.GetList.
//
// This method returns a snapshot of all valid mandatory groups currently in the
// list. It initializes an empty slice and iterates through all entries in the
// map. Each valid group is appended to the slice.
//
// This provides a safe way to work with all the groups at a specific point in
// time without holding a lock or worrying about concurrent modifications
//
// affecting the loop.
//
// The operation is thread-safe and includes self-healing for invalid entries.
func (o *model) GetList() []stsmdt.Mandatory {
	var res []stsmdt.Mandatory
	o.l.Range(func(k string, v stsmdt.Mandatory) bool {
		if v == nil {
			o.l.Delete(k)
			return true
		}

		if len(v.KeyList()) < 1 {
			o.l.Delete(k)
			return true
		}

		res = append(res, v)

		return true
	})
	return res
}
