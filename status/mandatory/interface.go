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

// Package mandatory provides a thread-safe mechanism for managing groups of
// components that share a common validation mode.
//
// This package is a fundamental building block for the status monitoring system.
// It allows you to define a logical set of components (e.g., "all critical databases",
// "all optional caches") and associate them with a single control mode (e.g., "Must",
// "Should"). This abstraction simplifies the configuration and evaluation of
// complex health check policies.
//
// # Key Features
//
//   - Thread-Safe: All operations are safe for concurrent use, leveraging `atomic.Value`
//     for high-performance, lock-free reads. This is critical for health check endpoints
//     that may be polled frequently.
//   - Dynamic Management: Component keys can be added or removed from a group at runtime,
//     allowing the system to adapt to changes in the application's topology.
//   - Control Mode Association: Each group is tightly coupled with a `control.Mode`,
//     ensuring that all components in the group are evaluated consistently.
//
// # Usage Example
//
//	import (
//	    "fmt"
//	    "github.com/nabbar/golib/status/control"
//	    "github.com/nabbar/golib/status/mandatory"
//	)
//
//	func main() {
//	    // Create a new mandatory group for critical database components.
//	    dbGroup := mandatory.New()
//
//	    // Set the validation mode to 'Must', meaning all components in this
//	    // group must be healthy for the group to be considered healthy.
//	    dbGroup.SetMode(control.Must)
//
//	    // Add the keys of the components that belong to this group.
//	    // These keys correspond to the names of the monitors registered in the pool.
//	    dbGroup.KeyAdd("primary-db", "read-replica-db")
//
//	    // Check if a specific component is part of this mandatory group.
//	    if dbGroup.KeyHas("primary-db") {
//	        fmt.Println("The primary database is a mandatory component.")
//	    }
//
//	    // Retrieve the list of all keys in the group.
//	    keys := dbGroup.KeyList()
//	    fmt.Printf("Mandatory database components: %v\n", keys)
//	}
package mandatory

import (
	"sync/atomic"

	stsctr "github.com/nabbar/golib/status/control"
)

// Mandatory defines the interface for managing a group of components that share
// a common validation mode. This allows for the logical grouping of related
// components (e.g., a cluster of services) to apply a single health check policy.
//
// Implementations of this interface must be thread-safe to support concurrent
// health checks and configuration updates.
type Mandatory interface {
	// SetMode assigns a validation mode to this group. The mode dictates how the
	// health of the components within this group will be evaluated. For example,
	// `control.Must` would require all components to be healthy, while `control.AnyOf`
	// would require only one.
	//
	// Parameters:
	//   - m: The `control.Mode` to apply to this group.
	SetMode(m stsctr.Mode)

	// GetMode retrieves the currently assigned validation mode for this group.
	// This allows the status calculation logic to know how to treat the components
	// in this group. If no mode has been explicitly set, it should return a default
	// (typically `control.Ignore`).
	//
	// Returns:
	//   The current `control.Mode`.
	GetMode() stsctr.Mode

	// KeyHas checks if a specific component key (by its unique name) is a member
	// of this group. The check is case-sensitive. This is useful for determining
	// if a component should be evaluated under this group's policy.
	//
	// Parameters:
	//   - key: The component name to check.
	//
	// Returns:
	//   `true` if the key exists in the group, `false` otherwise.
	KeyHas(key string) bool

	// KeyAdd adds one or more component keys to this group. If a key already
	// exists, it is ignored, ensuring uniqueness within the group. This allows
	// for idempotent additions.
	//
	// Parameters:
	//   - keys: A variadic list of component names to add.
	KeyAdd(keys ...string)

	// KeyDel removes one or more component keys from this group. If a key does
	// not exist in the group, the operation is a no-op for that key. This allows
	// for safe removal without checking for existence first.
	//
	// Parameters:
	//   - keys: A variadic list of component names to remove.
	KeyDel(keys ...string)

	// KeyList returns a slice containing all the component keys currently in this
	// group. The returned slice is a defensive copy, so modifications to it will
	// not affect the internal state of the group. This ensures thread safety for
	// the caller.
	//
	// Returns:
	//   A `[]string` containing all component keys.
	KeyList() []string
}

// New creates and returns a new, thread-safe instance of the `Mandatory` interface.
// The new instance is initialized with a default mode of `control.Ignore` and an
// empty list of keys.
//
// Returns:
//   A new `Mandatory` instance ready for use.
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
