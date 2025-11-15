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
	"context"

	liblog "github.com/nabbar/golib/logger"
	montps "github.com/nabbar/golib/monitor/types"
	libver "github.com/nabbar/golib/version"
	libvpr "github.com/nabbar/golib/viper"
	spfcbr "github.com/spf13/cobra"
)

// FuncCptGet is a function type that retrieves a component by its key.
// Used by components to access other components via dependency injection.
// Returns nil if the component doesn't exist.
type FuncCptGet func(key string) Component

// FuncCptEvent is a function type for component lifecycle event hooks.
// Called before and after component lifecycle operations (Start, Reload).
// Receives the component instance and returns an error if the hook fails.
type FuncCptEvent func(cpt Component) error

// ComponentEvent defines the lifecycle interface for components.
// All components must implement these methods to participate in the
// application lifecycle managed by Config.
type ComponentEvent interface {
	// RegisterFuncStart registers hooks to be called before and after Start().
	// The before hook executes before the component's Start() method.
	// The after hook executes after the component has started successfully.
	// Use for: initialization logging, dependency verification, post-start validation.
	RegisterFuncStart(before, after FuncCptEvent)

	// RegisterFuncReload registers hooks to be called before and after Reload().
	// The before hook executes before the component's Reload() method.
	// The after hook executes after the component has reloaded successfully.
	// Use for: configuration backup, reload logging, post-reload validation.
	RegisterFuncReload(before, after FuncCptEvent)

	// IsStarted returns true if the component has been started.
	// This indicates that Start() has been called and completed successfully.
	// The component may or may not still be actively running.
	// Used by Config to verify component initialization state.
	IsStarted() bool

	// IsRunning returns true if the component is actively running.
	// This differs from IsStarted() in that it indicates current runtime state.
	// A component can be started but not running (e.g., stopped, crashed).
	// Used by Config to check component health and readiness.
	IsRunning() bool

	// Start initializes and starts the component.
	// This is called by Config.Start() in dependency order.
	// The component should:
	//   - Load its configuration
	//   - Initialize resources (connections, pools, etc.)
	//   - Start background goroutines if needed
	//   - Set started and running state
	// Returns an error if startup fails. Config will abort the start sequence.
	Start() error

	// Reload refreshes the component's configuration and restarts if necessary.
	// This is called by Config.Reload() in dependency order.
	// The component should:
	//   - Reload its configuration from Viper
	//   - Apply changes without full restart if possible
	//   - Restart internal services if configuration changed significantly
	// Returns an error if reload fails. Config will abort the reload sequence.
	Reload() error

	// Stop gracefully shuts down the component.
	// This is called by Config.Stop() in reverse dependency order.
	// The component must:
	//   - Stop all background goroutines
	//   - Close connections and release resources
	//   - Set running and started state to false
	//   - Complete within a reasonable time
	// This method should not return an error; it must clean up best-effort.
	Stop()
}

// ComponentViper provides Viper configuration integration for components.
// Components implementing this interface can register CLI flags that are
// automatically bound to Viper configuration keys.
type ComponentViper interface {
	// RegisterFlag registers command-line flags for the component.
	// Flags are typically bound to Viper keys for configuration loading.
	// The key parameter (from Init) is used to namespace flags.
	//
	// Example:
	//   cmd.Flags().String("database.host", "localhost", "Database host")
	//   viper.BindPFlag("database.host", cmd.Flags().Lookup("database.host"))
	//
	// Returns an error if flag registration fails.
	RegisterFlag(Command *spfcbr.Command) error
}

// ComponentMonitor provides health check and metrics integration for components.
// Components implementing this interface can register monitors for observability.
type ComponentMonitor interface {
	// RegisterMonitorPool registers a monitor pool provider function.
	// The component can use this to register health checks, metrics, and status endpoints.
	// Called during component initialization (Init).
	//
	// The monitor pool typically provides:
	//   - Health check registration
	//   - Metrics collection
	//   - Status reporting
	//
	// Components should store this function and call it when ready to register monitors.
	RegisterMonitorPool(p montps.FuncPool)
}

// Component is the main interface that all components must implement.
// It combines lifecycle management (ComponentEvent), configuration (ComponentViper),
// and monitoring (ComponentMonitor) capabilities.
//
// A component represents a distinct subsystem of the application, such as:
//   - Database connections
//   - HTTP servers
//   - Cache systems
//   - Message queues
//   - Background workers
//
// Components are registered with Config and managed through their lifecycle.
type Component interface {
	// Type returns a unique identifier for the component type.
	// Used for logging, debugging, and component identification.
	// Examples: "database", "http-server", "cache", "logger"
	Type() string

	// Init is called by Config when the component is registered via ComponentSet().
	// This provides the component with access to shared resources:
	//
	// Parameters:
	//   - key: The unique key this component is registered under
	//   - ctx: Function to get the shared application context
	//   - get: Function to retrieve other components by key (dependency injection)
	//   - vpr: Function to get the Viper configuration instance
	//   - vrs: Application version information
	//   - log: Function to get the default logger instance
	//
	// The component should store these for later use during Start/Reload/Stop.
	Init(key string, ctx context.Context, get FuncCptGet, vpr libvpr.FuncViper, vrs libver.Version, log liblog.FuncLog)

	// DefaultConfig returns the default JSON configuration for this component.
	// This is used by Config.DefaultConfig() to generate a complete default config file.
	//
	// Parameters:
	//   - indent: The indentation string to use for JSON formatting (e.g., "  ")
	//
	// Returns:
	//   - A JSON byte slice representing default configuration
	//
	// Example return value:
	//   {
	//     "enabled": true,
	//     "host": "localhost",
	//     "port": 5432
	//   }
	DefaultConfig(indent string) []byte

	// Dependencies returns the list of component keys that this component depends on.
	// Config uses this for topological sorting to ensure components start in the correct order.
	//
	// Returns:
	//   - Slice of component keys (empty if no dependencies)
	//
	// Example:
	//   return []string{"database", "cache"}  // Depends on database and cache
	//
	// Dependencies are started before this component and stopped after this component.
	Dependencies() []string

	// SetDependencies allows customizing the component's dependencies at runtime.
	// This replaces the default dependencies returned by Dependencies().
	//
	// Parameters:
	//   - d: New list of dependency keys
	//
	// Returns:
	//   - Error if the dependencies are invalid or create circular dependencies
	//
	// Use with caution: Ensure default dependencies are included if still needed.
	SetDependencies(d []string) error

	ComponentViper
	ComponentEvent
	ComponentMonitor
}
