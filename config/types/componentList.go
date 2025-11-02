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

package types

import (
	"io"

	spfcbr "github.com/spf13/cobra"
)

// ComponentListWalkFunc is a callback function type for iterating over components.
// It receives the component key and component instance for each registered component.
//
// Parameters:
//   - key: The unique identifier for the component
//   - cpt: The component instance
//
// Returns:
//   - bool: Return true to continue iteration, false to stop early
//
// Example:
//
//	cfg.ComponentWalk(func(key string, cpt Component) bool {
//	    fmt.Printf("Component %s: %s\n", key, cpt.Type())
//	    return true  // Continue to next component
//	})
type ComponentListWalkFunc func(key string, cpt Component) bool

// ComponentList provides component registry management operations.
// This interface is embedded in the Config interface to provide component
// registration, retrieval, lifecycle management, and configuration generation.
//
// The registry is thread-safe and supports:
//   - Dynamic component registration and removal
//   - Dependency-ordered lifecycle operations
//   - Configuration aggregation from all components
//   - Component enumeration and inspection
type ComponentList interface {
	// ComponentHas checks if a component with the given key is registered.
	//
	// Parameters:
	//   - key: The component key to check
	//
	// Returns:
	//   - bool: true if the component exists, false otherwise
	//
	// Thread-safe: Can be called concurrently.
	ComponentHas(key string) bool

	// ComponentType returns the type identifier of the registered component.
	//
	// Parameters:
	//   - key: The component key to query
	//
	// Returns:
	//   - string: The component type (from Component.Type()), or empty string if not found
	//
	// Thread-safe: Can be called concurrently.
	ComponentType(key string) string

	// ComponentGet retrieves a registered component by its key.
	//
	// Parameters:
	//   - key: The component key to retrieve
	//
	// Returns:
	//   - Component: The component instance, or nil if not found
	//
	// The returned component can be type-asserted to specific component interfaces:
	//
	//	if db, ok := cfg.ComponentGet("database").(DatabaseComponent); ok {
	//	    db.GetConnection()
	//	}
	//
	// Thread-safe: Can be called concurrently.
	ComponentGet(key string) Component

	// ComponentDel removes a component from the registry.
	//
	// Parameters:
	//   - key: The component key to remove
	//
	// Note: Does not call Stop() on the component. Stop the component first
	// if it's running. Other components depending on this component may fail
	// if it's removed while they're running.
	//
	// Thread-safe: Can be called concurrently.
	ComponentDel(key string)

	// ComponentSet registers a component with the given key.
	//
	// Parameters:
	//   - key: Unique identifier for the component
	//   - cpt: The component instance to register
	//
	// This method:
	//   - Calls component.Init() with dependency injection parameters
	//   - Registers the monitor pool if available
	//   - Stores the component in the registry
	//   - Replaces any existing component with the same key
	//
	// Example:
	//
	//	db := &DatabaseComponent{}
	//	cfg.ComponentSet("database", db)
	//
	// Thread-safe: Can be called concurrently.
	ComponentSet(key string, cpt Component)

	// ComponentList returns all registered components as a map.
	//
	// Returns:
	//   - map[string]Component: Map of component keys to component instances
	//
	// The returned map is a snapshot; modifications don't affect the registry.
	// Components are cleaned (invalid entries removed) during iteration.
	//
	// Thread-safe: Can be called concurrently.
	ComponentList() map[string]Component

	// ComponentWalk iterates over all registered components.
	//
	// Parameters:
	//   - fct: Callback function called for each component
	//
	// The iteration can be stopped early by returning false from the callback.
	// Invalid components are automatically removed during iteration.
	//
	// Example:
	//
	//	cfg.ComponentWalk(func(key string, cpt Component) bool {
	//	    if cpt.IsRunning() {
	//	        fmt.Printf("%s is running\n", key)
	//	    }
	//	    return true
	//	})
	//
	// Thread-safe: Can be called concurrently.
	ComponentWalk(fct ComponentListWalkFunc)

	// ComponentKeys returns the keys of all registered components.
	//
	// Returns:
	//   - []string: Slice of component keys (order not guaranteed)
	//
	// Invalid components are automatically removed during enumeration.
	//
	// Thread-safe: Can be called concurrently.
	ComponentKeys() []string

	// ComponentStart starts all registered components in dependency order.
	// This is the internal method called by Config.Start().
	//
	// Behavior:
	//   - Resolves dependencies via topological sort
	//   - Starts components in dependency order
	//   - Logs start operations if logger is available
	//   - Stops on first error (remaining components not started)
	//   - Verifies each component reports started state after Start()
	//
	// Returns:
	//   - error: Aggregated error if any component fails to start
	//
	// Components are started sequentially, not in parallel.
	ComponentStart() error

	// ComponentIsStarted checks if at least one component is started.
	//
	// Returns:
	//   - bool: true if any component reports IsStarted() == true
	//
	// Used to determine if the application has been initialized.
	//
	// Thread-safe: Can be called concurrently.
	ComponentIsStarted() bool

	// ComponentReload reloads all registered components in dependency order.
	// This is the internal method called by Config.Reload().
	//
	// Behavior:
	//   - Reloads components in dependency order
	//   - Logs reload operations if logger is available
	//   - Stops on first error (remaining components not reloaded)
	//   - Verifies each component still reports started state after Reload()
	//
	// Returns:
	//   - error: Aggregated error if any component fails to reload
	//
	// Components should implement hot-reload without full restart.
	ComponentReload() error

	// ComponentStop stops all registered components in reverse dependency order.
	// This is the internal method called by Config.Stop().
	//
	// Behavior:
	//   - Stops components in reverse dependency order
	//   - Best-effort shutdown (does not propagate errors)
	//   - Each component's Stop() is called sequentially
	//
	// This method does not return errors; all components must stop cleanly.
	ComponentStop()

	// ComponentIsRunning checks the running state of components.
	//
	// Parameters:
	//   - atLeast: If true, returns true if at least one component is running.
	//              If false, returns true only if all components are running.
	//
	// Returns:
	//   - bool: Running state based on atLeast parameter
	//
	// Example:
	//
	//	// Check if any component is running
	//	if cfg.ComponentIsRunning(true) {
	//	    fmt.Println("At least one component is running")
	//	}
	//
	//	// Check if all components are running
	//	if cfg.ComponentIsRunning(false) {
	//	    fmt.Println("All components are running")
	//	}
	//
	// Thread-safe: Can be called concurrently.
	ComponentIsRunning(atLeast bool) bool

	// DefaultConfig generates a default configuration file from all components.
	//
	// Returns:
	//   - io.Reader: JSON configuration containing default values from all components
	//
	// The generated JSON structure has each component's config under its key:
	//
	//	{
	//	  "database": {
	//	    "host": "localhost",
	//	    "port": 5432
	//	  },
	//	  "cache": {
	//	    "ttl": 300
	//	  }
	//	}
	//
	// Components without default config are omitted.
	// The JSON is properly formatted and compacted.
	//
	// Thread-safe: Can be called concurrently.
	DefaultConfig() io.Reader

	// RegisterFlag registers command-line flags for all components.
	//
	// Parameters:
	//   - Command: The cobra command to register flags with
	//
	// Returns:
	//   - error: Aggregated error if any component fails flag registration
	//
	// This delegates to each component's RegisterFlag() method.
	// Typically called during CLI initialization before parsing flags.
	//
	// Thread-safe: Can be called concurrently.
	RegisterFlag(Command *spfcbr.Command) error
}
