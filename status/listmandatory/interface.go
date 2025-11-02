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

// Package listmandatory provides management of multiple mandatory component groups.
//
// This package allows managing a collection of mandatory groups, where each group
// contains component keys with an associated validation mode. It provides thread-safe
// operations for adding, removing, and querying mandatory groups.
//
// # Key Features
//
//   - Thread-safe operations using sync.Map
//   - Dynamic group management (add, delete, walk)
//   - Key-based mode lookup across all groups
//   - Support for multiple validation strategies
//
// # Usage Example
//
//	import (
//	    "github.com/nabbar/golib/status/control"
//	    "github.com/nabbar/golib/status/mandatory"
//	    "github.com/nabbar/golib/status/listmandatory"
//	)
//
//	// Create mandatory groups
//	dbGroup := mandatory.New()
//	dbGroup.SetMode(control.Must)
//	dbGroup.KeyAdd("postgres", "redis")
//
//	cacheGroup := mandatory.New()
//	cacheGroup.SetMode(control.Should)
//	cacheGroup.KeyAdd("memcached")
//
//	// Create list and add groups
//	list := listmandatory.New(dbGroup, cacheGroup)
//
//	// Query mode for a specific key
//	mode := list.GetMode("postgres")
//	fmt.Println(mode) // Output: Must
//
//	// Walk through all groups
//	list.Walk(func(m mandatory.Mandatory) bool {
//	    fmt.Println("Keys:", m.KeyList())
//	    return true // continue walking
//	})
//
// # Thread Safety
//
// All operations are thread-safe and can be called concurrently from multiple
// goroutines without external synchronization.
//
// # Integration
//
// This package is designed to work with:
//   - github.com/nabbar/golib/status/mandatory: For individual mandatory groups
//   - github.com/nabbar/golib/status/control: For validation modes
//   - github.com/nabbar/golib/status: For overall status management
//
// See also:
//   - github.com/nabbar/golib/status/mandatory for mandatory groups
//   - github.com/nabbar/golib/status/control for control modes
package listmandatory

import (
	"sync/atomic"

	libatm "github.com/nabbar/golib/atomic"
	stsctr "github.com/nabbar/golib/status/control"
	stsmdt "github.com/nabbar/golib/status/mandatory"
)

// ListMandatory defines the interface for managing multiple mandatory groups.
//
// A ListMandatory instance maintains a collection of Mandatory groups and
// provides operations to manage them. Each group can have different validation
// modes and component keys.
//
// All methods are safe for concurrent use.
type ListMandatory interface {
	// Len returns the number of mandatory groups in the list.
	//
	// This operation is thread-safe. Invalid entries are automatically
	// cleaned up during the count.
	//
	// Example:
	//
	//	count := list.Len()
	//	fmt.Printf("Managing %d mandatory groups\n", count)
	Len() int

	// Walk iterates through all mandatory groups in the list.
	//
	// The provided function is called for each group. If the function returns
	// false, the iteration stops early. If it returns true, iteration continues.
	//
	// This operation is thread-safe. Invalid entries are automatically cleaned
	// up during iteration.
	//
	// Example:
	//
	//	list.Walk(func(m mandatory.Mandatory) bool {
	//	    fmt.Println("Mode:", m.GetMode())
	//	    fmt.Println("Keys:", m.KeyList())
	//	    return true // continue to next group
	//	})
	Walk(fct func(m stsmdt.Mandatory) bool)

	// Add adds one or more mandatory groups to the list.
	//
	// Each group is assigned a unique internal ID. Groups can be added
	// multiple times (they are treated as separate entries).
	//
	// This operation is thread-safe.
	//
	// Example:
	//
	//	dbGroup := mandatory.New()
	//	dbGroup.SetMode(control.Must)
	//	dbGroup.KeyAdd("database")
	//
	//	list.Add(dbGroup)
	Add(m ...stsmdt.Mandatory)

	// Del removes a mandatory group from the list.
	//
	// The group is identified by comparing its key list (sorted). If multiple
	// groups have the same keys, only the first match is removed.
	//
	// This operation is thread-safe.
	//
	// Example:
	//
	//	list.Del(dbGroup)
	Del(m stsmdt.Mandatory)

	// GetMode returns the validation mode for a specific component key.
	//
	// This method searches through all mandatory groups and returns the mode
	// of the first group that contains the specified key. If no group contains
	// the key, it returns control.Ignore.
	//
	// This operation is thread-safe.
	//
	// Example:
	//
	//	mode := list.GetMode("database")
	//	if mode == control.Must {
	//	    fmt.Println("Database is mandatory")
	//	}
	GetMode(key string) stsctr.Mode

	// SetMode updates the validation mode for all groups containing a key.
	//
	// This method searches through all mandatory groups and updates the mode
	// of the first group that contains the specified key.
	//
	// This operation is thread-safe.
	//
	// Example:
	//
	//	list.SetMode("database", control.Should)
	SetMode(key string, mod stsctr.Mode)

	// GetList returns all mandatory groups in the list.
	//
	// This method returns a slice containing all mandatory groups currently
	// stored in the list. Invalid entries are automatically cleaned up during
	// the iteration.
	//
	// The returned slice is a snapshot at the time of the call and can be
	// safely modified without affecting the internal state.
	//
	// This operation is thread-safe.
	//
	// Example:
	//
	//	groups := list.GetList()
	//	for _, group := range groups {
	//	    fmt.Println("Mode:", group.GetMode())
	//	    fmt.Println("Keys:", group.KeyList())
	//	}
	GetList() []stsmdt.Mandatory
}

// New creates a new ListMandatory instance.
//
// The instance can be initialized with zero or more mandatory groups.
// Additional groups can be added later using the Add method.
//
// All operations on the returned instance are thread-safe.
//
// Example:
//
//	// Create empty list
//	list := listmandatory.New()
//
//	// Create with initial groups
//	dbGroup := mandatory.New()
//	cacheGroup := mandatory.New()
//	list := listmandatory.New(dbGroup, cacheGroup)
func New(m ...stsmdt.Mandatory) ListMandatory {
	var o = &model{
		l: libatm.NewMapTyped[uint32, stsmdt.Mandatory](),
		k: new(atomic.Uint32),
	}

	o.k.Store(0)
	o.Add(m...)

	return o
}
