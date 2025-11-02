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

// Package mandatory provides management of mandatory components with validation modes.
//
// This package allows defining groups of component keys with associated control modes
// that determine how strictly those components must be validated in a status monitoring
// system.
//
// # Key Features
//
//   - Thread-safe operations using atomic.Value
//   - Dynamic key management (add, delete, check, list)
//   - Control mode assignment per mandatory group
//   - Zero allocations for read operations
//
// # Usage Example
//
//	import (
//	    "github.com/nabbar/golib/status/control"
//	    "github.com/nabbar/golib/status/mandatory"
//	)
//
//	// Create a mandatory group
//	m := mandatory.New()
//
//	// Set validation mode
//	m.SetMode(control.Must)
//
//	// Add component keys
//	m.KeyAdd("database", "cache", "queue")
//
//	// Check if a key is mandatory
//	if m.KeyHas("database") {
//	    fmt.Println("Database is mandatory")
//	}
//
//	// List all mandatory keys
//	keys := m.KeyList()
//	fmt.Println("Mandatory components:", keys)
//
// # Thread Safety
//
// All operations are thread-safe and can be called concurrently from multiple
// goroutines without external synchronization.
//
// # Integration
//
// This package is designed to work with:
//   - github.com/nabbar/golib/status/control: For validation modes
//   - github.com/nabbar/golib/status/listmandatory: For managing multiple mandatory groups
//   - github.com/nabbar/golib/status: For overall status management
//
// See also:
//   - github.com/nabbar/golib/status/control for control modes
//   - github.com/nabbar/golib/status/listmandatory for list management
package mandatory

import (
	"sync/atomic"

	stsctr "github.com/nabbar/golib/status/control"
)

// Mandatory defines the interface for managing mandatory component groups.
//
// A Mandatory instance represents a group of component keys with an associated
// validation mode. It provides thread-safe operations for managing keys and
// the validation mode.
//
// All methods are safe for concurrent use.
type Mandatory interface {
	// SetMode sets the validation mode for this mandatory group.
	//
	// The mode determines how strictly the components in this group must be
	// validated. Common modes include Must (all must be healthy), Should
	// (warnings only), AnyOf (at least one), and Quorum (majority).
	//
	// This operation is thread-safe.
	//
	// Example:
	//
	//	m.SetMode(control.Must)
	SetMode(m stsctr.Mode)

	// GetMode returns the current validation mode for this mandatory group.
	//
	// If no mode has been set, it returns control.Ignore (the zero value).
	//
	// This operation is thread-safe and lock-free.
	//
	// Example:
	//
	//	mode := m.GetMode()
	//	if mode == control.Must {
	//	    // Enforce strict validation
	//	}
	GetMode() stsctr.Mode

	// KeyHas checks if a specific key exists in this mandatory group.
	//
	// The check is case-sensitive and performs an exact match.
	//
	// This operation is thread-safe and lock-free.
	//
	// Example:
	//
	//	if m.KeyHas("database") {
	//	    fmt.Println("Database is mandatory")
	//	}
	KeyHas(key string) bool

	// KeyAdd adds one or more keys to this mandatory group.
	//
	// Duplicate keys are automatically ignored. The order of keys is preserved
	// based on first insertion.
	//
	// This operation is thread-safe.
	//
	// Example:
	//
	//	m.KeyAdd("database", "cache")
	//	m.KeyAdd("queue") // Add more keys later
	KeyAdd(keys ...string)

	// KeyDel removes one or more keys from this mandatory group.
	//
	// Keys that don't exist are silently ignored. The operation is idempotent.
	//
	// This operation is thread-safe.
	//
	// Example:
	//
	//	m.KeyDel("cache") // Remove a single key
	//	m.KeyDel("database", "queue") // Remove multiple keys
	KeyDel(keys ...string)

	// KeyList returns a copy of all keys in this mandatory group.
	//
	// The returned slice is a copy and can be safely modified without affecting
	// the internal state. The order reflects insertion order.
	//
	// This operation is thread-safe and lock-free.
	//
	// Example:
	//
	//	keys := m.KeyList()
	//	fmt.Println("Mandatory components:", keys)
	KeyList() []string
}

// New creates a new Mandatory instance.
//
// The instance is initialized with:
//   - Mode: control.Ignore (no validation required)
//   - Keys: empty list
//
// All operations on the returned instance are thread-safe.
//
// Example:
//
//	m := mandatory.New()
//	m.SetMode(control.Must)
//	m.KeyAdd("database", "cache")
func New() Mandatory {
	m := new(atomic.Value)
	m.Store(stsctr.Ignore)

	k := new(atomic.Value)
	k.Store(make([]string, 0))

	return &model{
		Mode: m,
		Keys: k,
	}
}
