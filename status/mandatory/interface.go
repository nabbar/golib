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

// Package mandatory defines the interface and implementation for managing
// groups of components with shared validation policies.
//
// See doc.go for comprehensive documentation, architecture details, and usage examples.
package mandatory

import (
	"sync/atomic"

	libatm "github.com/nabbar/golib/atomic"
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

	// SetName assigns a human-readable identifier to the mandatory group.
	// This name is primarily used for logging, debugging, and distinguishing
	// between different groups (e.g., "primary-db-cluster" vs "cache-layer").
	//
	// The input name will be sanitized by the implementation to ensure it
	// consists only of safe characters (alphanumeric, hyphens, underscores).
	//
	// Parameters:
	//   - s: The desired name for the group.
	SetName(s string)

	// GetName retrieves the current human-readable identifier for this group.
	// If a name has not been explicitly set, the implementation may return a
	// default generated name to ensure every group has a unique identifier.
	//
	// Returns:
	//   The sanitized name of the mandatory group.
	GetName() string

	// SetInfo populates the group's metadata with a map of key-value pairs.
	// This is used to store descriptive information such as a human-readable
	// description or links to relevant documentation (e.g., runbooks).
	// This method replaces any existing info with the provided map.
	//
	// Parameters:
	//   - info: A map of metadata to associate with the group.
	SetInfo(info map[string]interface{})

	// AddInfo adds or updates a specific piece of metadata for the group.
	// This allows for incremental updates to the group's information without
	// needing to replace the entire map.
	//
	// Parameters:
	//   - key: The key for the metadata entry (e.g., "description").
	//   - value: The value associated with the key.
	AddInfo(key string, value interface{})

	// GetInfo retrieves all metadata associated with the group.
	// This information is typically exposed in the status response to provide
	// context to operators.
	//
	// Returns:
	//   A map containing all metadata entries.
	GetInfo() map[string]interface{}

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
//
//	A new `Mandatory` instance ready for use.
func New() Mandatory {

	k := new(atomic.Value)
	k.Store(make([]string, 0))

	return &model{
		Mode: new(atomic.Uint32),
		Name: libatm.NewValue[string](),
		Info: libatm.NewMapTyped[string, interface{}](),
		Keys: k,
	}
}
