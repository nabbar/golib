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

// Package types provides the core interfaces for the component-based application
// framework. It defines the contracts that components must adhere to for lifecycle
// management, configuration, and monitoring.
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
// It is used by components to access other components via dependency injection.
// It returns nil if the component with the given key does not exist.
type FuncCptGet func(key string) Component

// FuncCptEvent is a function type for component lifecycle event hooks.
// It is called before and after component lifecycle operations (e.g., Start, Reload).
// The function receives the component instance and returns an error if the hook fails.
type FuncCptEvent func(cpt Component) error

// ComponentEvent defines the lifecycle interface for components.
// All components must implement these methods to participate in the
// application lifecycle managed by the main config.
type ComponentEvent interface {
	// RegisterFuncStart registers hooks to be called before and after the Start method.
	// The 'before' hook executes before the component's Start() method is called.
	// The 'after' hook executes after the component has started successfully.
	// This is useful for tasks like initialization logging, dependency verification,
	// or post-start validation.
	RegisterFuncStart(before, after FuncCptEvent)

	// RegisterFuncReload registers hooks to be called before and after the Reload method.
	// The 'before' hook executes before the component's Reload() method is called.
	// The 'after' hook executes after the component has reloaded successfully.
	// This is useful for tasks like configuration backup, reload logging, or
	// post-reload validation.
	RegisterFuncReload(before, after FuncCptEvent)

	// IsStarted returns true if the component has been started.
	// This indicates that the Start() method has been called and completed successfully.
	// The component may or may not still be actively running.
	// This is used by the config manager to verify the component's initialization state.
	IsStarted() bool

	// IsRunning returns true if the component is actively running.
	// This differs from IsStarted() in that it indicates the current runtime state.
	// A component can be started but not currently running (e.g., stopped or crashed).
	// This is used by the config manager to check the component's health and readiness.
	IsRunning() bool

	// Start initializes and starts the component.
	// This method is called by the config manager in dependency order.
	// The component should perform tasks such as:
	//   - Loading its configuration.
	//   - Initializing resources (e.g., database connections, client pools).
	//   - Starting any background goroutines.
	//   - Setting its started and running state.
	// It returns an error if startup fails, which will cause the config manager
	// to abort the application's start sequence.
	Start() error

	// Reload refreshes the component's configuration and restarts it if necessary.
	// This method is called by the config manager in dependency order.
	// The component should:
	//   - Reload its configuration from Viper.
	//   - Apply changes without a full restart if possible.
	//   - Restart internal services if the configuration has changed significantly.
	// It returns an error if the reload fails, which will cause the config manager
	// to abort the reload sequence.
	Reload() error

	// Stop gracefully shuts down the component.
	// This method is called by the config manager in reverse dependency order.
	// The component must:
	//   - Stop all background goroutines.
	//   - Close connections and release all resources.
	//   - Set its running and started state to false.
	//   - Complete the shutdown process within a reasonable time.
	// This method should not return an error; it must perform cleanup on a best-effort basis.
	Stop()
}

// ComponentViper provides Viper configuration integration for components.
// Components implementing this interface can register command-line flags that are
// automatically bound to Viper configuration keys.
type ComponentViper interface {
	// RegisterFlag registers command-line flags for the component.
	// These flags are typically bound to Viper keys for configuration loading.
	// The 'key' parameter (from the Init method) can be used to namespace flags.
	//
	// Example:
	//   cmd.Flags().String("database.host", "localhost", "Database host")
	//   viper.BindPFlag("database.host", cmd.Flags().Lookup("database.host"))
	//
	// It returns an error if flag registration fails.
	RegisterFlag(Command *spfcbr.Command) error
}

// ComponentMonitor provides health check and metrics integration for components.
// Components implementing this interface can register monitors for observability.
type ComponentMonitor interface {
	// RegisterMonitorPool registers a monitor pool provider function.
	// The component can use this to register health checks, metrics, and status endpoints.
	// This method is called during the component's initialization (Init).
	//
	// The monitor pool typically provides:
	//   - Health check registration.
	//   - Metrics collection.
	//   - Status reporting.
	//
	// Components should store this function and call it when they are ready to register monitors.
	RegisterMonitorPool(p montps.FuncPool)

	// GetMonitorNames returns the names of the monitors registered by the component.
	GetMonitorNames() []string
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
// Components are registered with the config manager and managed through their lifecycle.
type Component interface {
	// Type returns a unique identifier for the component type.
	// This is used for logging, debugging, and component identification.
	// Examples: "database", "http-server", "cache", "logger".
	Type() string

	// Init is called by the config manager when the component is registered.
	// This provides the component with access to shared resources.
	//
	// Parameters:
	//   - key: The unique key this component is registered under.
	//   - ctx: A function to get the shared application context.
	//   - get: A function to retrieve other components by key (for dependency injection).
	//   - vpr: A function to get the Viper configuration instance.
	//   - vrs: Application version information.
	//   - log: A function to get the default logger instance.
	//
	// The component should store these for later use during its lifecycle (Start/Reload/Stop).
	Init(key string, ctx context.Context, get FuncCptGet, vpr libvpr.FuncViper, vrs libver.Version, log liblog.FuncLog)

	// DefaultConfig returns the default JSON configuration for this component.
	// This is used by the config manager to generate a complete default config file.
	//
	// Parameters:
	//   - indent: The indentation string to use for JSON formatting (e.g., "  ").
	//
	// Returns:
	//   - A JSON byte slice representing the default configuration.
	//
	// Example return value:
	//   {
	//     "enabled": true,
	//     "host": "localhost",
	//     "port": 5432
	//   }
	DefaultConfig(indent string) []byte

	// Dependencies returns the list of component keys that this component depends on.
	// The config manager uses this for topological sorting to ensure components start
	// in the correct order.
	//
	// Returns:
	//   - A slice of component keys (empty if no dependencies).
	//
	// Example:
	//   return []string{"database", "cache"}  // Depends on "database" and "cache" components.
	//
	// Dependencies are started before this component and stopped after this component.
	Dependencies() []string

	// SetDependencies allows customizing the component's dependencies at runtime.
	// This replaces the default dependencies returned by the Dependencies() method.
	//
	// Parameters:
	//   - d: A new list of dependency keys.
	//
	// Returns:
	//   - An error if the dependencies are invalid or create circular dependencies.
	//
	// Use with caution: Ensure that default dependencies are included if they are still needed.
	SetDependencies(d []string) error

	ComponentViper
	ComponentEvent
	ComponentMonitor
}
