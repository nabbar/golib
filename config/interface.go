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

package config

import (
	"context"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"

	libatm "github.com/nabbar/golib/atomic"
	cfgtps "github.com/nabbar/golib/config/types"
	libctx "github.com/nabbar/golib/context"
	liblog "github.com/nabbar/golib/logger"
	shlcmd "github.com/nabbar/golib/shell/command"
	libver "github.com/nabbar/golib/version"
	libvpr "github.com/nabbar/golib/viper"
)

// FuncEvent is a function type for lifecycle event hooks.
// It returns an error if the event handler fails.
// Used for before/after hooks in Start, Reload, and Stop operations.
type FuncEvent func() error

// Config is the main interface for application lifecycle management.
// It provides component orchestration, dependency resolution, event hooks,
// context management, and shell command integration.
//
// Lifecycle Operations:
//   - Start(): Initialize and start all registered components in dependency order
//   - Reload(): Hot-reload components without full restart
//   - Stop(): Gracefully shutdown all components in reverse dependency order
//   - Shutdown(code): Complete application termination with cleanup and exit
//
// Component Management:
//   - ComponentSet/Get/Del: Register and manage components
//   - ComponentList/Keys: Enumerate registered components
//   - ComponentStart/Stop/Reload: Lifecycle operations on components
//
// Event Hooks:
//   - RegisterFuncStartBefore/After: Hooks around start operations
//   - RegisterFuncReloadBefore/After: Hooks around reload operations
//   - RegisterFuncStopBefore/After: Hooks around stop operations
//
// Context & Cancellation:
//   - Context(): Shared application context for all components
//   - CancelAdd/CancelClean: Custom cancellation handlers
//
// Configuration:
//   - RegisterFuncViper: Viper configuration provider
//   - RegisterDefaultLogger: Logger provider for components
//   - DefaultConfig(): Generate default configuration file
//
// Shell Commands:
//   - GetShellCommand(): Interactive commands for runtime management
type Config interface {
	// Context returns the shared application context instance.
	// All components receive access to this context for shared state management
	// and coordinated cancellation.
	Context() libctx.Config[string]

	// CancelAdd registers custom functions to be called on context cancellation.
	// These functions execute before Stop() when the application receives
	// termination signals (SIGINT, SIGTERM, SIGQUIT) or when Shutdown() is called.
	// Useful for cleanup tasks that need to happen before component shutdown.
	CancelAdd(fct ...func())

	// CancelClean removes all registered cancel functions.
	// Use this to reset cancellation handlers, typically during testing
	// or when reconfiguring the application.
	CancelClean()

	// Start initiates the startup sequence for all registered components.
	// Components are started in dependency order (topological sort).
	// The sequence is:
	//   1. Execute RegisterFuncStartBefore hooks
	//   2. Start each component in dependency order
	//   3. Execute RegisterFuncStartAfter hooks
	// Returns an error if any hook or component fails to start.
	// On error, the start sequence is aborted immediately.
	Start() error

	// Reload triggers a configuration reload for all registered components.
	// Components are reloaded in dependency order.
	// The sequence is:
	//   1. Execute RegisterFuncReloadBefore hooks
	//   2. Reload each component in dependency order
	//   3. Execute RegisterFuncReloadAfter hooks
	// Returns an error if any hook or component fails to reload.
	// Components should implement hot-reload without full restart.
	Reload() error

	// Stop gracefully shuts down all registered components.
	// Components are stopped in reverse dependency order.
	// The sequence is:
	//   1. Execute RegisterFuncStopBefore hooks
	//   2. Stop each component in reverse dependency order
	//   3. Execute RegisterFuncStopAfter hooks
	// This function does not return errors; components must stop cleanly.
	Stop()

	// Shutdown performs a complete application termination.
	// This function:
	//   1. Executes all registered cancel functions (CancelAdd)
	//   2. Calls Stop() to shutdown components
	//   3. Exits the process with the specified exit code
	// This is typically called on fatal errors or during graceful shutdown.
	// Note: This function does not return as it calls os.Exit().
	Shutdown(code int)

	// RegisterFuncViper registers a Viper configuration provider function.
	// Components use this to access their configuration sections.
	// The function is called when components need to load or reload their configuration.
	// Typically registered once during application initialization.
	RegisterFuncViper(fct libvpr.FuncViper)

	// RegisterFuncStartBefore registers a hook executed before component startup.
	// This hook runs before any component's Start() method is called.
	// Use for: pre-start validation, initialization logging, resource preparation.
	// If the hook returns an error, the start sequence is aborted.
	RegisterFuncStartBefore(fct FuncEvent)

	// RegisterFuncStartAfter registers a hook executed after component startup.
	// This hook runs after all components have started successfully.
	// Use for: post-start validation, ready notification, monitoring setup.
	// If the hook returns an error, it's treated as a start failure.
	RegisterFuncStartAfter(fct FuncEvent)

	// RegisterFuncReloadBefore registers a hook executed before component reload.
	// This hook runs before any component's Reload() method is called.
	// Use for: pre-reload backup, configuration validation, logging.
	// If the hook returns an error, the reload sequence is aborted.
	RegisterFuncReloadBefore(fct FuncEvent)

	// RegisterFuncReloadAfter registers a hook executed after component reload.
	// This hook runs after all components have reloaded successfully.
	// Use for: post-reload validation, cache clearing, notification.
	// If the hook returns an error, it's treated as a reload failure.
	RegisterFuncReloadAfter(fct FuncEvent)

	// RegisterFuncStopBefore registers a hook executed before component shutdown.
	// This hook runs before any component's Stop() method is called.
	// Use for: pre-shutdown logging, resource flushing, notification.
	// Errors from this hook are logged but do not prevent shutdown.
	RegisterFuncStopBefore(fct FuncEvent)

	// RegisterFuncStopAfter registers a hook executed after component shutdown.
	// This hook runs after all components have stopped.
	// Use for: cleanup verification, final logging, monitoring notification.
	// Errors from this hook are logged but ignored (shutdown continues).
	RegisterFuncStopAfter(fct FuncEvent)

	// RegisterDefaultLogger registers a logger provider function for components.
	// Components use this logger for operational logging.
	// The function is called each time a component needs a logger instance.
	// If not registered, components may not have logging capability.
	RegisterDefaultLogger(fct liblog.FuncLog)

	// ComponentList provides component registry operations.
	// Includes: ComponentSet, ComponentGet, ComponentDel, ComponentList,
	// ComponentKeys, ComponentStart, ComponentStop, ComponentReload.
	cfgtps.ComponentList

	// ComponentMonitor provides monitoring integration.
	// Allows components to register health checks and metrics.
	cfgtps.ComponentMonitor

	// GetShellCommand returns interactive shell commands for runtime management.
	// Commands include: list (show components), start (start components),
	// stop (stop components), restart (restart components).
	// These commands can be integrated into CLI applications or interactive shells.
	GetShellCommand() []shlcmd.Command
}

var (
	ctx context.Context
	cnl context.CancelFunc
)

func init() {
	ctx, cnl = context.WithCancel(context.Background())
}

// Shutdown cancels the main application context, triggering shutdown of all components.
// This is a package-level function that can be called from any goroutine to
// initiate a coordinated shutdown. It cancels the shared context, which is
// monitored by all Config instances.
//
// Usage:
//   - Call from signal handlers for graceful shutdown
//   - Call from error handlers for emergency shutdown
//   - Call from other goroutines to stop the application
//
// Note: This only cancels the context. Components must handle the cancellation
// through their Config.Stop() method or by monitoring the context.
func Shutdown() {
	cnl()
}

// WaitNotify blocks until an interrupt signal is received, then initiates shutdown.
// This function monitors OS signals (SIGINT, SIGTERM, SIGQUIT) and the application
// context. When either a signal is received or the context is cancelled, it calls
// Shutdown() to begin the graceful shutdown sequence.
//
// Typical usage in main():
//
//	func main() {
//	    cfg := config.New(version)
//	    cfg.Start()
//	    config.WaitNotify()  // Blocks until signal received
//	}
//
// Monitored signals:
//   - SIGINT (Ctrl+C): User interrupt from terminal
//   - SIGTERM: Termination signal (default for 'kill' command)
//   - SIGQUIT: Quit signal with core dump request
//
// The function returns immediately after calling Shutdown(), allowing the main
// function to perform final cleanup before exiting.
func WaitNotify() {
	quit := make(chan os.Signal, 1)

	defer func() {
		close(quit)
	}()

	signal.Notify(quit, syscall.SIGINT)
	signal.Notify(quit, syscall.SIGTERM)
	signal.Notify(quit, syscall.SIGQUIT)

	select {
	case <-quit:
		Shutdown()
	case <-ctx.Done():
		Shutdown()
	}
}

// New creates and initializes a new Config instance for application lifecycle management.
//
// Parameters:
//   - vrs: Version information for the application (can be nil)
//
// Returns:
//   - Config: A fully initialized configuration orchestrator
//
// The returned Config instance:
//   - Uses a shared application context for all components
//   - Automatically monitors the context for cancellation
//   - Provides thread-safe component registration and management
//   - Supports dependency resolution and ordered lifecycle operations
//
// Initialization:
//   - Creates internal registries for components, hooks, and functions
//   - Registers the provided version information
//   - Starts a goroutine to monitor context cancellation
//
// Example:
//
//	version := libver.NewVersion(...)
//	cfg := config.New(version)
//	defer cfg.Stop()
//
//	// Register components
//	cfg.ComponentSet("database", dbComponent)
//	cfg.ComponentSet("cache", cacheComponent)
//
//	// Start all components
//	if err := cfg.Start(); err != nil {
//	    log.Fatal(err)
//	}
//
// Thread Safety:
// All operations on the returned Config instance are thread-safe and can be
// called concurrently from multiple goroutines.
func New(vrs libver.Version) Config {
	c := &model{
		ctx: libctx.New[string](ctx),
		cpt: libatm.NewMapTyped[string, cfgtps.Component](),
		fct: libatm.NewMapTyped[uint8, any](),
		cnl: libatm.NewMapTyped[uint64, context.CancelFunc](),
		seq: new(atomic.Uint64),
	}

	c.RegisterVersion(vrs)

	go func() {
		<-c.ctx.Done()
		c.cancel()
	}()

	return c
}
