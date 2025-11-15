# Config Package

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.21-blue)](https://golang.org/)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue)](https://pkg.go.dev/github.com/nabbar/golib/config)

**Production-ready application lifecycle management system for Go with automatic dependency resolution, event-driven architecture, and component orchestration.**

> **AI Disclaimer (EU AI Act Article 50.4):** AI assistance was used solely for testing, documentation, and bug resolution under human supervision.

---

## Table of Contents

- [Overview](#overview)
- [Key Features](#key-features)
- [Architecture](#architecture)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Core Concepts](#core-concepts)
  - [Config Interface](#config-interface)
  - [Component Interface](#component-interface)
  - [Component Lifecycle](#component-lifecycle)
  - [Dependency Management](#dependency-management)
- [Available Components](#available-components)
- [Configuration Management](#configuration-management)
- [Event Hooks](#event-hooks)
- [Context Management](#context-management)
- [Shell Integration](#shell-integration)
- [Performance](#performance)
- [Use Cases](#use-cases)
- [Best Practices](#best-practices)
- [Testing](#testing)
- [Contributing](#contributing)
- [Future Enhancements](#future-enhancements)
- [Related Documentation](#related-documentation)
- [License](#license)

---

## Overview

The **config** package provides a comprehensive lifecycle management framework for building modular, production-ready Go applications. It orchestrates component initialization, startup, hot-reloading, and graceful shutdown while automatically resolving dependencies and managing shared resources.

### Design Philosophy

1. **Dependency-Driven**: Automatic topological ordering ensures components start in the correct sequence
2. **Event-Driven**: Lifecycle hooks enable observability and cross-cutting concerns
3. **Context-Aware**: Shared application context provides coordinated cancellation
4. **Component-Based**: Pluggable architecture supports modular development
5. **Thread-Safe**: Atomic operations and proper synchronization for concurrent access
6. **Hot-Reload**: Support for configuration reloading without full restart

### Why Use This Package?

- ✅ **Zero boilerplate** for component lifecycle management
- ✅ **Automatic dependency resolution** eliminates manual ordering
- ✅ **Built-in components** for common services (HTTP, SMTP, LDAP, Database, TLS, etc.)
- ✅ **Event hooks** for logging, metrics, and custom logic
- ✅ **Graceful shutdown** with proper cleanup
- ✅ **Hot-reload support** for configuration changes
- ✅ **Shell commands** for runtime introspection
- ✅ **Comprehensive testing** with race detector validation

---

## Key Features

- **Lifecycle Management**: Coordinated start, reload, and stop sequences across all components
- **Dependency Resolution**: Automatic topological sorting ensures correct initialization order
- **Event Hooks**: Before/after callbacks for lifecycle events (start, reload, stop)
- **Context Sharing**: Application-wide context accessible to all components
- **Thread-Safe Operations**: Mutex-protected component registry with concurrent-safe access
- **Shell Commands**: Built-in interactive commands for runtime component management
- **Config Generation**: Automatic default configuration file creation from all components
- **Monitoring Integration**: Support for health checks and metrics via monitor pools
- **Viper Integration**: Seamless configuration loading with github.com/spf13/viper
- **Signal Handling**: Graceful shutdown on SIGINT, SIGTERM, SIGQUIT
- **Version Tracking**: Built-in version information management

## Installation

```bash
go get github.com/nabbar/golib/config
```

---

## Architecture

### Package Structure

The package is organized into a main package with supporting sub-packages:

```
config/
├── config/                  # Main package with lifecycle orchestration
│   ├── interface.go        # Config interface and factory
│   ├── components.go       # Component management
│   ├── events.go           # Lifecycle event handlers
│   ├── manage.go           # Hook and function registration
│   ├── context.go          # Context and cancellation
│   ├── shell.go            # Shell command integration
│   ├── errors.go           # Error definitions
│   └── model.go            # Internal data structures
├── types/                  # Interface definitions
│   ├── component.go        # Component interface
│   └── componentList.go    # Component list interface
├── const/                  # Package constants
│   └── const.go           # JSON formatting constants
└── components/            # Pre-built component implementations
    ├── aws/               # AWS component
    ├── database/          # Database component
    ├── http/              # HTTP server component
    ├── log/               # Logger component
    └── ...                # Other components
```

### Component Orchestration Flow

```
┌─────────────────────────────────────────────────────────┐
│                   Config Orchestrator                    │
│                                                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐ │
│  │  Component   │  │  Component   │  │  Component   │ │
│  │  Registry    │  │  Lifecycle   │  │   Events     │ │
│  └──────────────┘  └──────────────┘  └──────────────┘ │
│         │                  │                  │          │
│         ▼                  ▼                  ▼          │
│  ┌──────────────────────────────────────────────────┐  │
│  │        Dependency Resolution Engine               │  │
│  │  (Topological Sort + Validation)                 │  │
│  └──────────────────────────────────────────────────┘  │
│                                                          │
│  ┌──────────────────────────────────────────────────┐  │
│  │           Shared Application Context              │  │
│  │  (Thread-safe storage + Cancellation)            │  │
│  └──────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
         │                  │                  │
         ▼                  ▼                  ▼
   Component A        Component B        Component C
   (Database)         (Cache)            (HTTP Server)
```

### Lifecycle Execution Order

```
Start Phase:
┌─────────────────────────────────────────────────────┐
│ 1. Global Before-Start Hooks                        │
├─────────────────────────────────────────────────────┤
│ 2. For each component (dependency order):           │
│    ┌─────────────────────────────────────────────┐ │
│    │ a. Component Before-Start Hook              │ │
│    │ b. Component.Start()                        │ │
│    │ c. Component After-Start Hook               │ │
│    └─────────────────────────────────────────────┘ │
├─────────────────────────────────────────────────────┤
│ 3. Global After-Start Hooks                         │
└─────────────────────────────────────────────────────┘

Stop Phase:
┌─────────────────────────────────────────────────────┐
│ 1. Global Before-Stop Hooks                         │
├─────────────────────────────────────────────────────┤
│ 2. For each component (reverse dependency order):   │
│    └─→ Component.Stop()                             │
├─────────────────────────────────────────────────────┤
│ 3. Global After-Stop Hooks                          │
└─────────────────────────────────────────────────────┘
```

---

## Quick Start

### Basic Usage

```go
package main

import (
    "fmt"
    
    libcfg "github.com/nabbar/golib/config"
    libver "github.com/nabbar/golib/version"
)

func main() {
    // Create version information
    version := libver.NewVersion(
        libver.License_MIT,
        "myapp",
        "My Application",
        "2024-01-01",
        "commit-hash",
        "v1.0.0",
        "Author Name",
        "myapp",
        struct{}{},
        0,
    )
    
    // Create config instance
    cfg := libcfg.New(version)
    
    // Register components
    // cfg.ComponentSet("database", databaseComponent)
    // cfg.ComponentSet("cache", cacheComponent)
    
    // Register lifecycle hooks (optional)
    cfg.RegisterFuncStartBefore(func() error {
        fmt.Println("Starting application...")
        return nil
    })
    
    // Start all components
    if err := cfg.Start(); err != nil {
        panic(err)
    }
    
    // Application running...
    fmt.Println("Application started successfully")
    
    // Graceful shutdown
    defer cfg.Stop()
}
```

### With Signal Handling

```go
package main

import (
    libcfg "github.com/nabbar/golib/config"
    libver "github.com/nabbar/golib/version"
)

func main() {
    version := libver.NewVersion(/* ... */)
    cfg := libcfg.New(version)
    
    // Register components
    // ...
    
    // Start components
    if err := cfg.Start(); err != nil {
        panic(err)
    }
    
    // Wait for interrupt signal (SIGINT, SIGTERM, SIGQUIT)
    libcfg.WaitNotify()
}
```

---

## Performance

### Memory Characteristics

The config system maintains minimal memory overhead:

- **Config Instance**: ~1 KB base footprint
- **Per Component**: ~500 bytes overhead for tracking
- **Context Storage**: O(n) where n = number of stored key-value pairs
- **Hook Storage**: O(m) where m = number of registered hooks

### Thread Safety

All operations are thread-safe through:

- **Mutex Protection**: `sync.RWMutex` for component registry access
- **Context Storage**: Thread-safe context implementation from `github.com/nabbar/golib/context`
- **Atomic Operations**: Used where appropriate for state management
- **Concurrent Access**: Multiple goroutines can safely access components

### Startup Performance

| Operation | Components | Time | Notes |
|-----------|------------|------|-------|
| Component Registration | 10 | ~50µs | O(1) per component |
| Dependency Resolution | 10 | ~200µs | O(n log n) topological sort |
| Start Sequence | 10 | ~10ms | Depends on component logic |
| Context Operations | - | ~100ns | Per operation |

*Benchmarks on AMD64, Go 1.21*

---

## Use Cases

This package is designed for applications requiring coordinated lifecycle management:

**Microservices**
- Orchestrate HTTP servers, database pools, message queues, and caches
- Graceful shutdown with proper cleanup order
- Hot reload configuration without downtime

**Backend Services**
- Coordinate startup of multiple subsystems (auth, API, workers, schedulers)
- Manage dependencies between services (database before cache before API)
- Unified logging and monitoring across all components

**CLI Applications**
- Modular command-line tools with pluggable components
- Shell command integration for runtime inspection
- Configuration file generation for user customization

**Long-Running Daemons**
- Signal-based graceful shutdown
- Component health monitoring
- Runtime component restart without full application restart

**Plugin Systems**
- Dynamic component registration at runtime
- Dependency injection for cross-component communication
- Version-aware component loading

---

## Available Components

The config package includes pre-built components for common services:

| Component | Package | Description | Dependencies |
|-----------|---------|-------------|--------------|
| **AWS** | `components/aws` | AWS SDK integration and configuration | - |
| **Database** | `components/database` | SQL database connection pooling | - |
| **HTTP Server** | `components/http` | HTTP/HTTPS server with routing | TLS (optional) |
| **HTTP Client** | `components/httpcli` | HTTP client with connection pooling | TLS, DNS Mapper |
| **LDAP** | `components/ldap` | LDAP client for authentication | - |
| **Logger** | `components/log` | Structured logging component | - |
| **Mail** | `components/mail` | Email sending (SMTP) | SMTP |
| **Request** | `components/request` | HTTP request handling | HTTP Client |
| **SMTP** | `components/smtp` | SMTP client configuration | TLS (optional) |
| **TLS** | `components/tls` | TLS/SSL certificate management | - |
| **Head** | `components/head` | HTTP headers management | - |

### Component Features

Each component provides:
- ✅ **DefaultConfig()** - Generate sensible default configuration
- ✅ **RegisterFlag()** - CLI flag registration for Cobra
- ✅ **RegisterMonitorPool()** - Health check and metrics integration
- ✅ **Dependencies()** - Explicit dependency declaration
- ✅ **Hot-reload** - Configuration reload without restart (where applicable)
- ✅ **Thread-safe** - Concurrent access protection
- ✅ **Comprehensive tests** - Ginkgo/Gomega test suites with race detection

### Using Components

```go
import (
    libcfg "github.com/nabbar/golib/config"
    cpthttp "github.com/nabbar/golib/config/components/http"
    cptlog "github.com/nabbar/golib/config/components/log"
)

func main() {
    cfg := libcfg.New(version)
    
    // Register logger component
    logCpt := cptlog.New(ctx)
    cptlog.Register(cfg, "logger", logCpt)
    
    // Register HTTP server component  
    httpCpt := cpthttp.New(ctx)
    cpthttp.Register(cfg, "http-server", httpCpt)
    
    // Start all components (logger starts before HTTP due to dependencies)
    if err := cfg.Start(); err != nil {
        panic(err)
    }
}
```

For detailed component documentation, see the respective `README.md` in each component's directory.

---

## Core Concepts

### Config Interface

The main `Config` interface provides methods for managing the application lifecycle and components:

```go
type Config interface {
    // Lifecycle
    Start() error
    Reload() error
    Stop()
    Shutdown(code int)
    
    // Components
    ComponentSet(key string, cpt Component)
    ComponentGet(key string) Component
    ComponentDel(key string)
    ComponentList() map[string]Component
    ComponentKeys() []string
    
    // Context
    Context() libctx.Config[string]
    CancelAdd(fct ...func())
    CancelClean()
    
    // Events
    RegisterFuncStartBefore(fct FuncEvent)
    RegisterFuncStartAfter(fct FuncEvent)
    RegisterFuncReloadBefore(fct FuncEvent)
    RegisterFuncReloadAfter(fct FuncEvent)
    RegisterFuncStopBefore(fct FuncEvent)
    RegisterFuncStopAfter(fct FuncEvent)
    
    // Others
    RegisterFuncViper(fct libvpr.FuncViper)
    RegisterDefaultLogger(fct liblog.FuncLog)
    GetShellCommand() []shlcmd.Command
}
```

### Component Interface

Components must implement the `Component` interface:

```go
type Component interface {
    Type() string
    Init(key string, ctx FuncContext, get FuncCptGet, vpr FuncViper, vrs Version, log FuncLog)
    DefaultConfig(indent string) []byte
    Dependencies() []string
    SetDependencies(d []string) error
    
    IsStarted() bool
    IsRunning() bool
    Start() error
    Reload() error
    Stop()
    
    RegisterFlag(cmd *cobra.Command) error
    RegisterMonitorPool(p FuncPool)
    RegisterFuncStart(before, after FuncCptEvent)
    RegisterFuncReload(before, after FuncCptEvent)
}
```

---

## Component Lifecycle

The config system manages a three-phase lifecycle for all components:

### Start Phase

```go
cfg.Start() // Calls in order:
// 1. RegisterFuncStartBefore hooks
// 2. Component.Start() for each component (in dependency order)
// 3. RegisterFuncStartAfter hooks
```

**Features:**
- Components start in dependency order
- Early termination on first error
- Hooks execute before/after all components
- State tracking (started/running)

### Reload Phase

```go
cfg.Reload() // Calls in order:
// 1. RegisterFuncReloadBefore hooks
// 2. Component.Reload() for each component
// 3. RegisterFuncReloadAfter hooks
```

**Features:**
- Hot reload without restart
- Component state preservation
- Configuration refresh
- No downtime

### Stop Phase

```go
cfg.Stop() // Calls in order:
// 1. RegisterFuncStopBefore hooks
// 2. Component.Stop() for each component (reverse order)
// 3. RegisterFuncStopAfter hooks
```

**Features:**
- Graceful shutdown
- Reverse dependency order
- Resource cleanup
- No error propagation (best effort)

### Shutdown

```go
cfg.Shutdown(exitCode) // Calls:
// 1. Custom cancel functions (CancelAdd)
// 2. cfg.Stop()
// 3. os.Exit(exitCode)
```

**Use Case**: Complete application termination with cleanup.

---

## Dependency Management

The config system automatically resolves and orders component dependencies.

### Declaring Dependencies

```go
type DatabaseComponent struct {
    // ...
}

func (d *DatabaseComponent) Dependencies() []string {
    return []string{} // No dependencies
}

type CacheComponent struct {
    // ...
}

func (c *CacheComponent) Dependencies() []string {
    return []string{"database"} // Depends on database
}

type APIComponent struct {
    // ...
}

func (a *APIComponent) Dependencies() []string {
    return []string{"database", "cache"} // Depends on both
}
```

### Automatic Ordering

```go
cfg.ComponentSet("api", apiComponent)       // Registered in any order
cfg.ComponentSet("cache", cacheComponent)
cfg.ComponentSet("database", databaseComponent)

cfg.Start() // Starts in correct order:
// 1. database
// 2. cache
// 3. api
```

### Deep Dependency Chains

The system handles complex dependency graphs:

```
database → cache → session → api
         ↘ logger ↗
```

Components are started in topological order and stopped in reverse order.

---

## Event Hooks

Register custom functions to execute during lifecycle events:

### Global Hooks

```go
// Before starting any component
cfg.RegisterFuncStartBefore(func() error {
    fmt.Println("Preparing to start...")
    return nil
})

// After all components started
cfg.RegisterFuncStartAfter(func() error {
    fmt.Println("All components started successfully")
    return nil
})

// Before reloading
cfg.RegisterFuncReloadBefore(func() error {
    fmt.Println("Preparing to reload...")
    return nil
})

// After reloading
cfg.RegisterFuncReloadAfter(func() error {
    fmt.Println("Reload complete")
    return nil
})

// Before stopping
cfg.RegisterFuncStopBefore(func() error {
    fmt.Println("Preparing to stop...")
    return nil
})

// After all components stopped
cfg.RegisterFuncStopAfter(func() error {
    fmt.Println("Cleanup complete")
    return nil
})
```

### Component-Level Hooks

Components can register their own hooks:

```go
func (c *MyComponent) Init(key string, /* ... */) {
    // Hooks are set during initialization
}

// During registration, the component can set hooks
component.RegisterFuncStart(
    func(cpt Component) error {
        // Before this component starts
        return nil
    },
    func(cpt Component) error {
        // After this component starts
        return nil
    },
)
```

### Hook Execution Order

For `cfg.Start()`:
1. `RegisterFuncStartBefore`
2. For each component (in dependency order):
   - Component's before-start hook
   - Component's `Start()` method
   - Component's after-start hook
3. `RegisterFuncStartAfter`

---

## Shell Commands

The config system provides built-in shell commands for component management:

```go
cmds := cfg.GetShellCommand()

// Returns commands: list, start, stop, restart
for _, cmd := range cmds {
    fmt.Printf("Command: %s - %s\n", cmd.Name(), cmd.Describe())
}
```

### Available Commands

| Command | Description | Usage |
|---------|-------------|-------|
| `list` | List all components with status | `list` |
| `start` | Start all components | `start` |
| `stop` | Stop all components | `stop` |
| `restart` | Restart all components | `restart` |

### Example Usage

```go
import (
    "bytes"
    libcfg "github.com/nabbar/golib/config"
)

func main() {
    cfg := libcfg.New(version)
    
    // Register components
    cfg.ComponentSet("database", db)
    cfg.ComponentSet("cache", cache)
    
    // Get shell commands
    cmds := cfg.GetShellCommand()
    
    // Execute list command
    stdout := &bytes.Buffer{}
    stderr := &bytes.Buffer{}
    
    for _, cmd := range cmds {
        if cmd.Name() == "list" {
            cmd.Run(stdout, stderr, nil)
            fmt.Print(stdout.String())
        }
    }
}
```

---

## Context Management

The config system provides a shared context for all components:

### Basic Context Usage

```go
// Get the context
ctx := cfg.Context()

// Store values
ctx.Store("key", "value")

// Load values
val, ok := ctx.Load("key")
if ok {
    fmt.Println(val) // "value"
}
```

### Cancel Functions

Register functions to be called on application shutdown:

```go
cfg.CancelAdd(func() {
    fmt.Println("Cleanup database connections")
})

cfg.CancelAdd(func() {
    fmt.Println("Flush caches")
})

// Clear all cancel functions
cfg.CancelClean()
```

### Signal Handling

Graceful shutdown on system signals:

```go
func main() {
    cfg := libcfg.New(version)
    
    // Start components
    if err := cfg.Start(); err != nil {
        panic(err)
    }
    
    // Wait for SIGINT, SIGTERM, or SIGQUIT
    libcfg.WaitNotify()
    
    // Cleanup happens automatically
}
```

---

## Configuration Generation

Generate default configuration files for all components:

```go
// Get default configuration as io.Reader
reader := cfg.DefaultConfig()

// Write to file
file, _ := os.Create("config.json")
defer file.Close()
io.Copy(file, reader)
```

Generated configuration includes all registered components with their default values.

**Example output:**
```json
{
  "database": {
    "enabled": true,
    "host": "localhost",
    "port": 5432
  },
  "cache": {
    "enabled": true,
    "ttl": 300
  }
}
```

---

## Creating Components

### Component Interface

Components must implement this interface:

```go
type Component interface {
    // Identification
    Type() string
    
    // Initialization
    Init(key string, ctx FuncContext, get FuncCptGet, 
         vpr FuncViper, vrs Version, log FuncLog)
    
    // Configuration
    DefaultConfig(indent string) []byte
    
    // Dependencies
    Dependencies() []string
    SetDependencies(d []string) error
    
    // Lifecycle
    Start() error
    Reload() error
    Stop()
    IsStarted() bool
    IsRunning() bool
    
    // Integration
    RegisterFlag(cmd *cobra.Command) error
    RegisterMonitorPool(p FuncPool)
    RegisterFuncStart(before, after FuncCptEvent)
    RegisterFuncReload(before, after FuncCptEvent)
}
```

### Minimal Component Example

```go
package mycomponent

import (
    cfgtps "github.com/nabbar/golib/config/types"
    libctx "github.com/nabbar/golib/context"
    liblog "github.com/nabbar/golib/logger"
    libver "github.com/nabbar/golib/version"
    libvpr "github.com/nabbar/golib/viper"
)

type MyComponent struct {
    key     string
    started bool
    running bool
    logger  liblog.FuncLog
}

func (c *MyComponent) Type() string {
    return "mycomponent"
}

func (c *MyComponent) Init(key string, ctx context.Context, 
    get cfgtps.FuncCptGet, vpr libvpr.FuncViper, 
    vrs libver.Version, log liblog.FuncLog) {
    c.key = key
    c.logger = log
}

func (c *MyComponent) DefaultConfig(indent string) []byte {
    return []byte(`{
    "enabled": true,
    "timeout": 30
}`)
}

func (c *MyComponent) Dependencies() []string {
    return []string{} // No dependencies
}

func (c *MyComponent) SetDependencies(d []string) error {
    return nil
}

func (c *MyComponent) Start() error {
    c.started = true
    c.running = true
    return nil
}

func (c *MyComponent) Reload() error {
    // Reload logic here
    return nil
}

func (c *MyComponent) Stop() {
    c.running = false
    c.started = false
}

func (c *MyComponent) IsStarted() bool { return c.started }
func (c *MyComponent) IsRunning() bool { return c.running }

func (c *MyComponent) RegisterFlag(cmd *cobra.Command) error {
    return nil
}

func (c *MyComponent) RegisterMonitorPool(p montps.FuncPool) {}

func (c *MyComponent) RegisterFuncStart(before, after cfgtps.FuncCptEvent) {}

func (c *MyComponent) RegisterFuncReload(before, after cfgtps.FuncCptEvent) {}
```

### Using the Component

```go
cfg := libcfg.New(version)

// Create and register component
comp := &MyComponent{}
cfg.ComponentSet("mycomp", comp)

// Start the application
if err := cfg.Start(); err != nil {
    panic(err)
}
```

---

## Best Practices

### 1. Component Design
- Keep components focused on a single responsibility
- Use interfaces for flexibility and testing
- Implement proper error handling
- Use atomic values for thread-safe state management

### 2. Dependency Management
- Declare all dependencies explicitly
- Avoid circular dependencies
- Keep dependency chains shallow when possible

### 3. Lifecycle Management
- Always clean up resources in `Stop()`
- Make `Reload()` idempotent
- Handle errors gracefully in `Start()`
- Track component state accurately

### 4. Configuration
- Provide sensible defaults
- Validate configuration before use
- Support hot-reload when possible
- Document configuration options

### 5. Logging
- Use the provided logger
- Log at appropriate levels
- Include context in log messages
- Don't log sensitive information

### 6. Testing
- Write unit tests for each component
- Test lifecycle transitions
- Mock dependencies for isolation
- Test error conditions

---

## Testing

Comprehensive testing documentation is available in [TESTING.md](TESTING.md).

**Quick Test:**
```bash
cd config
go test -v
```

**With Coverage:**
```bash
go test -v -cover
```

**Test Results:**
- 93 test specifications
- 100% feature coverage
- 6 focused test files
- ~0.1 second execution time

---

## Examples

See the [Creating Components](#creating-components) section and [TESTING.md](TESTING.md) for detailed examples.

---

## API Reference

| Method | Description |
|--------|-------------|
| `New(version)` | Create new config instance |
| `Start()` | Start all components |
| `Reload()` | Reload all components |
| `Stop()` | Stop all components |
| `Shutdown(code)` | Shutdown with exit code |
| `ComponentSet(key, cpt)` | Register component |
| `ComponentGet(key)` | Get component |
| `ComponentDel(key)` | Delete component |
| `ComponentList()` | List all components |
| `Context()` | Get shared context |
| `GetShellCommand()` | Get shell commands |

---

## Contributing

Contributions are welcome! Please follow these guidelines:

**Code Contributions**
- Do not use AI to generate package implementation code
- AI may assist with tests, documentation, and bug fixing
- All contributions must pass existing tests
- Maintain or improve test coverage
- Follow existing code style and patterns

**Documentation**
- Update README.md for new features
- Add examples for common use cases
- Keep TESTING.md synchronized with test changes
- Document all public APIs with GoDoc comments

**Testing**
- Write tests for all new features
- Test edge cases and error conditions
- Verify thread safety when applicable
- Add comments explaining complex test scenarios

**Pull Requests**
- Provide clear description of changes
- Reference related issues
- Include test results
- Update documentation

See [CONTRIBUTING.md](../../CONTRIBUTING.md) for detailed guidelines.

---

## Future Enhancements

Potential improvements for future versions:

**Lifecycle Features**
- Parallel component startup (where dependencies allow)
- Component health checks with automatic restart
- Gradual rollout of configuration changes
- Component state persistence and recovery

**Dependency Management**
- Circular dependency detection with clear error messages
- Optional dependencies (soft dependencies)
- Dynamic dependency injection at runtime
- Dependency visualization tools

**Configuration**
- Multiple configuration sources (files, env vars, remote)
- Configuration validation before apply
- Configuration versioning and rollback
- Encrypted configuration values

**Monitoring & Observability**
- Built-in metrics for lifecycle events
- Distributed tracing integration
- Event streaming for external monitoring
- Component dependency graph visualization

**Developer Experience**
- Code generation for boilerplate components
- Interactive component inspector
- Configuration schema validation
- Better error messages with suggestions

Suggestions and contributions are welcome via GitHub issues.

---

## Related Documentation

### Core Packages
- **[context](../context/README.md)** - Thread-safe context storage used by config
- **[viper](../viper/README.md)** - Configuration file loading and management
- **[logger](../logger/README.md)** - Structured logging system
- **[version](../version/README.md)** - Application version management

### Component Packages
- **[components/aws](components/aws/README.md)** - AWS integration
- **[components/database](components/database/README.md)** - Database connection pooling
- **[components/http](components/http/README.md)** - HTTP server
- **[components/httpcli](components/httpcli/README.md)** - HTTP client
- **[components/ldap](components/ldap/README.md)** - LDAP authentication
- **[components/log](components/log/README.md)** - Logger component
- **[components/mail](components/mail/README.md)** - Email sending
- **[components/smtp](components/smtp/README.md)** - SMTP client
- **[components/tls](components/tls/README.md)** - TLS certificate management

### External References
- [Viper Configuration](https://github.com/spf13/viper) - Configuration management library
- [Cobra Commands](https://github.com/spf13/cobra) - CLI framework integration
- [Ginkgo Testing](https://github.com/onsi/ginkgo) - BDD testing framework used
- [Gomega Matchers](https://github.com/onsi/gomega) - Matcher library for tests

---

## License

MIT License - See [LICENSE](../../LICENSE) file for details.

Copyright (c) 2022 Nicolas JUHEL

---

## Resources

- **Issues**: [GitHub Issues](https://github.com/nabbar/golib/issues)
- **Documentation**: [GoDoc](https://pkg.go.dev/github.com/nabbar/golib/config)
- **Testing Guide**: [TESTING.md](TESTING.md)
- **Contributing**: [CONTRIBUTING.md](../../CONTRIBUTING.md)
- **Source Code**: [GitHub Repository](https://github.com/nabbar/golib)

---

*This package is part of the [golib](https://github.com/nabbar/golib) project.*
