# golib/config

This package provides a modular configuration management system for Go applications, supporting dynamic components, context management, lifecycle events, logging, monitoring, and shell integration.

## Features

- **Centralized configuration management** using Viper.
- **Dynamic components**: add, remove, start, stop, reload, and manage dependencies.
- **Context management**: cancellation, custom hooks.
- **Lifecycle events**: hooks before/after start, reload, stop.
- **Configurable logging**.
- **Component monitoring**.
- **Integrated shell commands** for component control.

## Installation

Add to your `go.mod`:

```
require github.com/nabbar/golib/config vX.Y.Z
```

## Quick Start

```go
import (
    "github.com/nabbar/golib/config"
    "github.com/nabbar/golib/version"
)

func main() {
    vrs := version.New("1.0.0")
    cfg := config.New(vrs)

    // Register your components here
    // cfg.ComponentSet("myComponent", myComponentInstance)

    // Register hooks, logger, etc.
    // cfg.RegisterFuncStartBefore(func() error { ...; return nil })

    if err := cfg.Start(); err != nil {
        panic(err)
    }

    // ...
    cfg.Shutdown(0)
}
```

## Main Interfaces

- **Config**: main interface, see `config/interface.go`.
- **Component**: to implement for each component, see `config/types/component.go`.

## Lifecycle Methods

- `Start() error`: starts all registered components respecting dependencies.
- `Reload() error`: reloads configuration and components.
- `Stop()`: stops all components cleanly.
- `Shutdown(code int)`: stops everything and exits the process.

## Component Management

- `ComponentSet(key, cpt)`: register a component.
- `ComponentGet(key)`: retrieve a component.
- `ComponentDel(key)`: remove a component.
- `ComponentList()`: get all components.
- `ComponentStart()`, `ComponentStop()`, `ComponentReload()`: global actions.

## Event Hooks

Register functions to be called before/after each lifecycle step:

```go
cfg.RegisterFuncStartBefore(func() error { /* ... */ return nil })
cfg.RegisterFuncStopAfter(func() error { /* ... */ return nil })
cfg.RegisterFuncReloadBefore(func() error { /* ... */ return nil })
cfg.RegisterFuncReloadAfter(func() error { /* ... */ return nil })
```

## Context and Cancellation

- `Context()`: returns the config context instance.
- `CancelAdd(func())`: register custom functions to call on context cancel.
- `CancelClean()`: clear all registered cancel functions.

## Shell Commands

Expose commands to list, start, stop, and restart components:

```go
cmds := cfg.GetShellCommand()
// Integrate these into your CLI
```

Available commands:
- `list`: list all components
- `start`: start components (all or by name)
- `stop`: stop components (all or by name)
- `restart`: restart components (all or by name)

## Signal Handling

Handles system signals (SIGINT, SIGTERM, SIGQUIT) for graceful shutdown via `WaitNotify()`.

## Default Configuration Generation

Generate a default configuration (JSON) for all registered components:

```go
r := cfg.DefaultConfig()
// r is an io.Reader containing the default config JSON
```

## Component Interface

To create a component, implement the following interface (see `config/types/component.go`):

```go
type Component interface {
    Type() string
    Init(key string, ctx libctx.FuncContext, get FuncCptGet, vpr libvpr.FuncViper, vrs libver.Version, log liblog.FuncLog)
    DefaultConfig(indent string) []byte
    Dependencies() []string
    SetDependencies(d []string) error
    RegisterFlag(Command *spfcbr.Command) error
    RegisterMonitorPool(p montps.FuncPool)
    RegisterFuncStart(before, after FuncCptEvent)
    RegisterFuncReload(before, after FuncCptEvent)
    IsStarted() bool
    IsRunning() bool
    Start() error
    Reload() error
    Stop()
}
```

---

## Creating a Config Component

This guide explains how to implement a new config component for the `golib/config` system. Components are modular units that can be registered, started, stopped, reloaded, and monitored within the configuration framework.

### 1. Component Structure

A component should implement the `Component` interface, which defines lifecycle methods, dependency management, configuration, and monitoring hooks.

**Example structure:**

```go
type componentMyType struct {
    x libctx.Config[uint8]      // Context and config storage
    // Add your own fields here (atomic values, state, etc.)
}
```

### 2. Implementing the Interface

Implement the following methods:

- `Type() string`: Returns the component type name.
- `Init(key, ctx, get, vpr, vrs, log)`: Initializes the component with its key, context, dependency getter, viper, version, and logger.
- `DefaultConfig(indent string) []byte`: Returns the default configuration as JSON (or other format).
- `Dependencies() []string`: Lists required component dependencies.
- `SetDependencies(d []string) error`: Sets dependencies.
- `RegisterFlag(cmd *cobra.Command) error`: Registers CLI flags.
- `RegisterMonitorPool(fct montps.FuncPool)`: Registers a monitor pool for health/metrics.
- `RegisterFuncStart(before, after FuncCptEvent)`: Registers hooks for start events.
- `RegisterFuncReload(before, after FuncCptEvent)`: Registers hooks for reload events.
- `IsStarted() bool`: Returns true if started.
- `IsRunning() bool`: Returns true if running.
- `Start() error`: Starts the component.
- `Reload() error`: Reloads the component.
- `Stop()`: Stops the component.

### 3. Configuration Handling

- Use Viper for configuration loading.
- Implement a method to unmarshal and validate the config section for your component.
- Provide a static default config as a JSON/YAML byte slice.

### 4. Lifecycle Management

- Use atomic values or context for thread-safe state.
- Implement logic for `Start`, `Reload`, and `Stop` to manage resources and state.
- Use hooks (`RegisterFuncStart`, `RegisterFuncReload`) to allow custom logic before/after lifecycle events.

### 5. Dependency Management

- Use `Dependencies()` and `SetDependencies()` to declare and manage required components (e.g., TLS, logger).
- Retrieve dependencies using the provided `FuncCptGet` function.

### 6. Monitoring Integration

- Implement `RegisterMonitorPool` to support health and metrics monitoring.
- Use the monitor pool to register and manage monitor instances for your component.

### 7. Example Skeleton

```go
type componentMyType struct {
    x libctx.Config[uint8]
    // Add fields as needed
}

func (o *componentMyType) Type() string { return "mytype" }
func (o *componentMyType) Init(key string, ctx libctx.FuncContext, get cfgtps.FuncCptGet, vpr libvpr.FuncViper, vrs libver.Version, log liblog.FuncLog) {
    o.x.Store( /* ... */ )
}
func (o *componentMyType) DefaultConfig(indent string) []byte { /* ... */ }
func (o *componentMyType) Dependencies() []string { /* ... */ }
func (o *componentMyType) SetDependencies(d []string) error { /* ... */ }
func (o *componentMyType) RegisterFlag(cmd *cobra.Command) error { /* ... */ }
func (o *componentMyType) RegisterMonitorPool(fct montps.FuncPool) { /* ... */ }
func (o *componentMyType) RegisterFuncStart(before, after cfgtps.FuncCptEvent) { /* ... */ }
func (o *componentMyType) RegisterFuncReload(before, after cfgtps.FuncCptEvent) { /* ... */ }
func (o *componentMyType) IsStarted() bool { /* ... */ }
func (o *componentMyType) IsRunning() bool { /* ... */ }
func (o *componentMyType) Start() error { /* ... */ }
func (o *componentMyType) Reload() error { /* ... */ }
func (o *componentMyType) Stop() { /* ... */ }
```

### 8. Registration

Register your component with the config system:

```go
cfg.ComponentSet("myComponentKey", myComponentInstance)
```

### 9. Error Handling

- Define error codes and messages for your component.
- Return errors using the `liberr.Error` type for consistency.

---

**Note:**
- Use atomic values for thread safety.
- Always validate configuration before starting the component.
- Integrate logging and monitoring as needed.
- Follow the modular and decoupled design to ensure maintainability and testability.
- Consider using context for cancellation and timeouts in long-running operations.
- Ensure proper cleanup in the `Stop` method to release resources gracefully.

