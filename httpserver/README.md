# `httpserver` Package

## `httpserver/pool` Subpackage

The `httpserver/pool` package provides a high-level abstraction for managing a pool of [`httpserver` Package](#package-httpserver). 
<br />It allows you to configure, start, stop, and monitor multiple HTTP server instances as a unified group, with advanced filtering, merging, and handler management capabilities.

### Features

- Manage multiple HTTP server instances as a pool
- Add, remove, and retrieve servers by bind address
- Start, stop, and restart all servers in the pool with aggregated error handling
- Filter and list servers by name, bind address, or expose address (with pattern or regex)
- Merge pools and clone pool state
- Register global handler functions for all servers
- Monitor all servers and retrieve monitoring data
- Thread-safe operations

---

### Main Types & Interfaces

#### `Pool` Interface

The main interface for managing a pool of HTTP servers. Key methods include:

- `Start(ctx context.Context) error`: Start all servers in the pool.
- `Stop(ctx context.Context) error`: Stop all servers in the pool.
- `Restart(ctx context.Context) error`: Restart all servers in the pool.
- `IsRunning() bool`: Check if any server in the pool is running.
- `Uptime() time.Duration`: Get the maximum uptime among all servers.
- `Handler(fct FuncHandler)`: Register a global handler function for all servers.
- `Monitor(vrs Version) ([]Monitor, error)`: Retrieve monitoring data for all servers.
- `Clone(ctx context.Context) Pool`: Clone the pool (optionally with a new context).
- `Merge(p Pool, defLog FuncLog) error`: Merge another pool into this one.
- `Manage` and `Filter` interfaces for advanced management and filtering.

#### `Manage` Interface

- `Walk(fct FuncWalk) bool`: Iterate over all servers.
- `StoreNew(cfg Config, defLog FuncLog) error`: Add a new server from config.
- `Load(bindAddress string) Server`: Retrieve a server by bind address.
- `Delete(bindAddress string)`: Remove a server by bind address.
- `MonitorNames() []string`: List all monitor names.

#### `Filter` Interface

- `Has(bindAddress string) bool`: Check if a server exists.
- `Len() int`: Number of servers in the pool.
- `List(fieldFilter, fieldReturn FieldType, pattern, regex string) []string`: List server fields matching criteria.
- `Filter(field FieldType, pattern, regex string) Pool`: Filter servers by field and pattern/regex.

#### `Config` Type

A slice of server configuration objects. Provides helper methods to:

- Set global handler, TLS, and context functions for all configs
- Validate all configs
- Instantiate a pool from the configs

---

### Example Usage

```go
import (
    "github.com/nabbar/golib/httpserver/pool"
    "github.com/nabbar/golib/httpserver"
    "github.com/nabbar/golib/logger"
    "context"
)

cfgs := pool.Config{
    /* ... fill with httpserver.Config objects ... */
}

err := cfgs.Validate()
if err != nil {
    // handle config validation error
}

p, err := cfgs.Pool(nil, nil, logger.Default)
if err != nil {
    // handle pool creation error
}

// Start all servers
if err := p.Start(context.Background()); err != nil {
    // handle start error
}

// Stop all servers
if err := p.Stop(context.Background()); err != nil {
    // handle stop error
}
```

---

### Error Handling

All errors are wrapped with custom codes for diagnostics, such as:

- `ErrorParamEmpty`
- `ErrorPoolAdd`
- `ErrorPoolValidate`
- `ErrorPoolStart`
- `ErrorPoolStop`
- `ErrorPoolRestart`
- `ErrorPoolMonitor`

Use `err.Error()` for user-friendly messages and check error codes for diagnostics.

---

### Filtering and Listing

You can filter or list servers in the pool by name, bind address, or expose address, using exact match or regular expressions.

```go
// List all bind addresses matching a pattern
binds := p.List(FieldBind, FieldBind, "127.0.0.1:8080", "")

// Filter pool by expose address regex
filtered := p.Filter(FieldExpose, "", "^/api")
```

---

### Monitoring

Retrieve monitoring data for all servers in the pool:

```go
monitors, err := p.Monitor(version)
if err != nil {
    // handle monitoring error
}
```

---

### Notes

- The pool is thread-safe and suitable for concurrent use.
- All operations are designed for Go 1.18+.
- Integrates with the `logger`, `context`, and `monitor` packages for advanced features.

---

## Package `httpserver`

The `httpserver` package provides advanced abstractions for configuring, running, and monitoring HTTP servers in Go. It is designed for robust, concurrent, and production-grade server management, supporting TLS, custom handlers, logging, and health monitoring.

---

### Key Features

- **Configurable HTTP/HTTPS servers** with extensive options (timeouts, keep-alive, HTTP/2, TLS, etc.)
- **Handler registration** for flexible routing and API management
- **Integrated logging** and monitoring support
- **Thread-safe** and suitable for concurrent use
- **Custom error codes** for diagnostics and troubleshooting

---

### Main Types

#### `Config`

Represents the configuration for a single HTTP server instance.  
Key fields include:

- `Name`: Unique server name (required)
- `Listen`: Bind address (host:port or unix socket, required)
- `Expose`: Public/external address (URL, required)
- `HandlerKey`: Key to associate with a specific handler
- `Disabled`: Enable/disable the server without removing its config
- `Monitor`: Monitoring configuration
- `TLSMandatory`: Require valid TLS configuration to start
- `TLS`: TLS settings (can inherit defaults)
- HTTP/2 and HTTP options: timeouts, max handlers, keep-alive, etc.
- `Logger`: Logger configuration

**Helper methods:**

- `Clone()`: Deep copy of the config
- `RegisterHandlerFunc(f)`: Register a handler function
- `SetDefaultTLS(f)`: Set default TLS provider
- `SetContext(f)`: Set parent context provider
- `Validate()`: Validate config fields and constraints
- `Server(defLog)`: Instantiate a server from the config

---

#### `Server` Interface

Represents a running HTTP server instance.

- `Start(ctx) error`: Start the server
- `Stop(ctx) error`: Stop the server gracefully
- `Restart(ctx) error`: Restart the server
- `IsRunning() bool`: Check if the server is running
- `GetConfig() *Config`: Get the current config
- `SetConfig(cfg, defLog) error`: Update the config
- `Handler(f)`: Register handler function
- `Merge(srv, defLog) error`: Merge another server's config
- `Monitor(version)`: Get monitoring data
- `MonitorName() string`: Get monitor name
- `GetName() string`: Get server name
- `GetBindable() string`: Get bind address
- `GetExpose() string`: Get expose address
- `IsDisable() bool`: Check if server is disabled
- `IsTLS() bool`: Check if TLS is enabled

---

#### Handler Management

Handlers are registered via a function returning a map of handler keys to `http.Handler` instances.  
You can associate a server with a specific handler using the `HandlerKey` field.

---

#### Monitoring

Integrated with the monitoring system, each server can expose runtime, build, and health information.  
Use `Monitor(version)` to retrieve monitoring data.

---

#### Error Handling

All errors are wrapped with custom codes for diagnostics, such as:

- `ErrorParamEmpty`
- `ErrorHTTP2Configure`
- `ErrorServerValidate`
- `ErrorServerStart`
- `ErrorPortUse`

Use `err.Error()` for user-friendly messages and check error codes for diagnostics.

---

### Example Usage

```go
import (
    "github.com/nabbar/golib/httpserver"
    "github.com/nabbar/golib/logger"
    "context"
)

cfg := httpserver.Config{
    Name:   "api-server",
    Listen: "127.0.0.1:8080",
    Expose: "http://api.example.com:8080",
    // ... other fields
}
cfg.RegisterHandlerFunc(myHandlerFunc)
cfg.SetDefaultTLS(myTLSProvider)
cfg.SetContext(myContextProvider)

if err := cfg.Validate(); err != nil {
    // handle config error
}

srv, err := httpserver.New(cfg, logger.Default)
if err != nil {
    // handle server creation error
}

if err := srv.Start(context.Background()); err != nil {
    // handle start error
}

// ... later
if err := srv.Stop(context.Background()); err != nil {
    // handle stop error
}
```

---

### Notes

- The package is designed for Go 1.18+.
- All operations are thread-safe.
- Integrates with `logger`, `context`, and `monitor` packages for advanced features.
- For advanced management of multiple servers, see the `httpserver/pool` package.
