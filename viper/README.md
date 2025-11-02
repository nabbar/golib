# Viper Package

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://golang.org/)
[![Tests](https://img.shields.io/badge/Tests-104%20Specs-green)]()
[![Coverage](https://img.shields.io/badge/Coverage-73.3%25-brightgreen)]()

Enhanced configuration management wrapper for spf13/viper with remote provider support, custom decode hooks, and file watching capabilities.

---

## Table of Contents

- [Overview](#overview)
- [Key Features](#key-features)
- [Installation](#installation)
- [Architecture](#architecture)
- [Quick Start](#quick-start)
- [Performance](#performance)
- [Use Cases](#use-cases)
- [Configuration Sources](#configuration-sources)
- [Advanced Features](#advanced-features)
- [Best Practices](#best-practices)
- [Testing](#testing)
- [Contributing](#contributing)
- [Future Enhancements](#future-enhancements)
- [License](#license)

---

## Overview

This package provides a production-ready wrapper around [spf13/viper](https://github.com/spf13/viper) for configuration management in Go applications. It extends viper's functionality with remote configuration support, custom type conversion hooks, and enhanced error handling.

### Design Philosophy

1. **Multi-Source**: Support file, environment, remote, and default configurations
2. **Type-Safe**: Custom decode hooks for complex type conversions
3. **Thread-Safe**: Atomic operations and proper synchronization for concurrent access
4. **Flexible**: Support for JSON, YAML, TOML, and other formats
5. **Observable**: File watching with reload callbacks for dynamic configuration

---

## Key Features

- **Multiple Configuration Sources**:
  - **File-based**: JSON, YAML, TOML, HCL, INI, ENV, properties
  - **Environment Variables**: Automatic binding with custom prefix
  - **Remote Providers**: ETCD support with automatic reloading
  - **Default Config**: Fallback configuration via `io.Reader`
  
- **Advanced Unmarshalling**:
  - Standard unmarshalling with type conversion
  - Exact unmarshalling (strict mode)
  - Key-specific unmarshalling for nested structures
  - Custom decode hooks for complex types

- **Dynamic Configuration**:
  - File system watching with change detection
  - Remote configuration polling and updates
  - Reload callbacks for application reconfiguration
  - Key unsetting for runtime modifications

- **Type System**:
  - Native support for all Go primitives
  - Time and duration parsing
  - Slices and maps
  - Custom type conversion via hooks

- **Thread Safety**:
  - Atomic operations for state management
  - Concurrent-safe hook registration
  - Thread-safe configuration access

---

## Installation

```bash
go get github.com/nabbar/golib/viper
```

**Requirements**:
- Go 1.18 or higher
- Dependencies automatically managed by Go modules

---

## Architecture

### Package Structure

```
viper/
├── interface.go         # Main Viper interface and factory
├── model.go            # Internal implementation and types
├── config.go           # Configuration initialization
├── viper.go            # Getter methods for all types
├── unmarshall.go       # Unmarshalling operations
├── hook.go             # Custom decode hook management
├── cleaner.go          # Configuration key unsetting
├── watch.go            # File and remote watching
├── remote.go           # Remote provider integration
├── errors.go           # Error codes and messages
└── doc.go              # Package documentation
```

### Component Overview

```
┌─────────────────────────────────────────────────────────┐
│                    Viper Interface                       │
│   New(), Config(), Get*(), Unmarshal(), Watch()         │
└──────────────┬──────────────┬──────────────┬────────────┘
               │              │              │
      ┌────────▼─────┐  ┌────▼─────┐  ┌────▼────────┐
      │ Config Mgmt  │  │  Hooks   │  │   Watch     │
      │              │  │          │  │             │
      │ File, Env    │  │ Custom   │  │ FS, Remote  │
      │ Remote, Def  │  │ Decode   │  │ Callbacks   │
      └──────────────┘  └──────────┘  └─────────────┘
```

| Component | Purpose | Thread-Safe |
|-----------|---------|-------------|
| **Config** | Multi-source configuration loading | ✅ |
| **Getters** | Type-safe value retrieval | ✅ |
| **Unmarshalling** | Struct binding with hooks | ✅ |
| **Hooks** | Custom type conversion | ✅ |
| **Watch** | Dynamic configuration updates | ✅ |
| **Remote** | ETCD integration | ✅ |

### Configuration Priority

Configuration values are resolved in the following order (highest to lowest):

```
1. Explicit Set() calls
2. Environment variables (with prefix)
3. Configuration file
4. Remote configuration (ETCD)
5. Default values
```

---

## Quick Start

### Basic File Configuration

Load configuration from a file:

```go
package main

import (
    "context"
    
    "github.com/nabbar/golib/logger"
    "github.com/nabbar/golib/logger/level"
    "github.com/nabbar/golib/viper"
)

func main() {
    ctx := func() context.Context { return context.Background() }
    log := func() logger.Logger { return logger.New(ctx) }
    
    v := viper.New(ctx, log)
    
    // Set configuration file
    if err := v.SetConfigFile("/etc/myapp/config.yaml"); err != nil {
        panic(err)
    }
    
    // Load configuration
    if err := v.Config(level.ErrorLevel, level.InfoLevel); err != nil {
        panic(err)
    }
    
    // Read values
    appName := v.GetString("app.name")
    port := v.GetInt("app.port")
    debug := v.GetBool("app.debug")
}
```

### Environment Variables

Use environment variables with prefix:

```go
v := viper.New(ctx, log)

// Set environment prefix
v.SetEnvVarsPrefix("MYAPP")

// Set home-based config
v.SetHomeBaseName("myapp")
v.SetConfigFile("") // Will look for ~/.myapp

// Load config (env vars override file values)
v.Config(level.ErrorLevel, level.InfoLevel)

// MYAPP_DATABASE_HOST overrides database.host
dbHost := v.GetString("database.host")
```

### Unmarshal to Struct

Bind configuration to Go structs:

```go
type AppConfig struct {
    Name    string
    Version string
    Server  struct {
        Host string
        Port int
    }
    Database struct {
        Driver   string
        Host     string
        Port     int
        User     string
        Password string
    }
}

v := viper.New(ctx, log)
v.SetConfigFile("config.yaml")
v.Config(level.ErrorLevel, level.InfoLevel)

var config AppConfig
if err := v.Unmarshal(&config); err != nil {
    panic(err)
}

// Or unmarshal specific key
var dbConfig DatabaseConfig
if err := v.UnmarshalKey("database", &dbConfig); err != nil {
    panic(err)
}
```

### Default Configuration

Provide fallback configuration:

```go
import "bytes"

v := viper.New(ctx, log)

// Set default config as fallback
v.SetDefaultConfig(func() io.Reader {
    defaultConfig := `{
        "app": {
            "name": "MyApp",
            "port": 8080,
            "debug": false
        }
    }`
    return bytes.NewReader([]byte(defaultConfig))
})

v.SetConfigFile("/etc/myapp/config.json")

// Will use default if file not found
err := v.Config(level.ErrorLevel, level.InfoLevel)
// err will indicate if default was used
```

---

## Performance

### Memory Efficiency

The viper wrapper maintains minimal memory overhead:

- **Configuration Storage**: O(n) where n is number of keys
- **Hook Management**: O(h) where h is number of hooks (max 255)
- **Watch Operations**: O(1) goroutine for file/remote watching

### Thread Safety

All operations are thread-safe through:

- **Atomic Operations**: `atomic.Uint32` for hook indexing
- **Context Storage**: Thread-safe `libctx.Config` for hook storage
- **Viper Core**: Underlying viper library is thread-safe

### Throughput

| Operation | Performance | Notes |
|-----------|-------------|-------|
| Get Value | ~100 ns/op | Direct map lookup |
| Unmarshal | ~10 µs/op | Depends on struct size |
| File Watch | ~1 ms/event | fsnotify overhead |
| Remote Poll | ~5s interval | Configurable |

*Measured on AMD64, Go 1.21*

---

## Use Cases

This library is designed for scenarios requiring flexible configuration management:

**Microservices**
- Centralized configuration via ETCD
- Environment-specific overrides
- Dynamic reconfiguration without restart
- Secret management with secure remote providers

**Cloud-Native Applications**
- 12-factor app compliance (config via environment)
- Container-friendly configuration
- Kubernetes ConfigMap integration
- Multi-environment deployments

**CLI Applications**
- User home directory configuration
- Command-line flag integration
- Configuration file discovery
- Default values with overrides

**Web Services**
- Hot-reload configuration changes
- Feature flags and A/B testing
- Multi-tenant configuration
- Environment-based settings

**Development Workflows**
- Local development overrides
- Test environment configuration
- CI/CD pipeline integration
- Configuration validation

---

## Configuration Sources

### File-Based Configuration

Supported formats: JSON, YAML, TOML, HCL, INI, ENV, properties

```go
v := viper.New(ctx, log)

// Explicit file path
v.SetConfigFile("/path/to/config.yaml")

// Or search in home directory
v.SetHomeBaseName("myapp")  // Looks for ~/.myapp
v.SetConfigFile("")

// Load configuration
v.Config(level.ErrorLevel, level.InfoLevel)
```

### Environment Variables

Automatic environment variable binding:

```go
v.SetEnvVarsPrefix("MYAPP")
v.SetHomeBaseName("myapp")
v.SetConfigFile("")

// MYAPP_SERVER_PORT overrides server.port
// MYAPP_DATABASE_HOST overrides database.host
```

### Remote Configuration (ETCD)

Dynamic configuration from ETCD:

```go
type Config struct {
    App AppSettings
    DB  DatabaseSettings
}

v := viper.New(ctx, log)

// Configure remote provider
v.SetRemoteProvider("etcd")
v.SetRemoteEndpoint("http://localhost:2379")
v.SetRemotePath("/config/myapp")
v.SetRemoteModel(&Config{})

// Optional: Secure connection
v.SetRemoteSecureKey("/path/to/encryption.key")

// Load and watch for changes
err := v.Config(level.ErrorLevel, level.InfoLevel)
// Configuration will auto-reload every 5 seconds
```

### Default Configuration

Fallback when no config file exists:

```go
v.SetDefaultConfig(func() io.Reader {
    return bytes.NewReader([]byte(`{
        "server": {"port": 8080},
        "database": {"host": "localhost"}
    }`))
})
```

---

## Advanced Features

### Custom Decode Hooks

Register custom type conversion hooks:

```go
import (
    "reflect"
    "strings"
    
    "github.com/go-viper/mapstructure/v2"
)

type CustomType string

// Register hook for custom type
hook := func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
    if t == reflect.TypeOf(CustomType("")) {
        if str, ok := data.(string); ok {
            return CustomType(strings.ToUpper(str)), nil
        }
    }
    return data, nil
}

v.HookRegister(hook)

// Now unmarshal will use the hook
var config MyConfig
v.Unmarshal(&config)
```

### File Watching

Watch for configuration file changes:

```go
v := viper.New(ctx, log)
v.SetConfigFile("config.yaml")
v.Config(level.ErrorLevel, level.InfoLevel)

// Set reload callback
v.SetRemoteReloadFunc(func() {
    log().Info("Configuration reloaded")
    // Reload application components
    reloadApp()
})

// Start watching
v.WatchFS(level.InfoLevel)
// Callback will be triggered on file changes
```

### Configuration Key Unsetting

Remove configuration keys at runtime:

```go
// Unset single key
v.Unset("app.debug")

// Unset multiple keys
v.Unset("cache.enabled", "cache.ttl", "cache.size")

// Unset nested keys
v.Unset("database.connection.pool.max")

// Other keys remain intact
port := v.GetInt("app.port") // Still available
```

### Exact Unmarshalling

Strict mode that fails on unknown fields:

```go
type StrictConfig struct {
    Name string
    Port int
}

// UnmarshalExact returns error if config has extra fields
err := v.UnmarshalExact(&config)
if err != nil {
    // Config contains fields not in StrictConfig
}
```

---

## Best Practices

**1. Always Handle Errors**

```go
// ✅ Good
func loadConfig() error {
    v := viper.New(ctx, log)
    if err := v.SetConfigFile("config.yaml"); err != nil {
        return fmt.Errorf("set config file: %w", err)
    }
    if err := v.Config(level.ErrorLevel, level.InfoLevel); err != nil {
        return fmt.Errorf("load config: %w", err)
    }
    return nil
}

// ❌ Bad: Silent failures
func loadConfigBad() {
    v := viper.New(ctx, log)
    v.SetConfigFile("config.yaml")
    v.Config(level.ErrorLevel, level.InfoLevel) // Ignoring error
}
```

**2. Use Struct Tags for Clarity**

```go
// ✅ Good: Clear mapping
type Config struct {
    AppName string `mapstructure:"app_name"`
    Port    int    `mapstructure:"server_port"`
}

// ❌ Bad: Implicit mapping
type Config struct {
    AppName string // Unclear what config key this maps to
    Port    int
}
```

**3. Provide Defaults**

```go
// ✅ Good: Fallback configuration
v.SetDefaultConfig(func() io.Reader {
    return bytes.NewReader([]byte(`{"port": 8080}`))
})

// ❌ Bad: No defaults, app may crash
v.SetConfigFile("config.yaml")
v.Config(level.ErrorLevel, level.InfoLevel)
port := v.GetInt("port") // Returns 0 if not set
```

**4. Validate Configuration**

```go
// ✅ Good: Validate after loading
type Config struct {
    Port int
    Host string
}

func loadAndValidate() (*Config, error) {
    var cfg Config
    if err := v.Unmarshal(&cfg); err != nil {
        return nil, err
    }
    
    if cfg.Port < 1 || cfg.Port > 65535 {
        return nil, fmt.Errorf("invalid port: %d", cfg.Port)
    }
    if cfg.Host == "" {
        return nil, fmt.Errorf("host is required")
    }
    
    return &cfg, nil
}
```

**5. Use Environment-Specific Configs**

```go
// ✅ Good: Environment-aware
env := os.Getenv("APP_ENV")
if env == "" {
    env = "development"
}

configFile := fmt.Sprintf("config.%s.yaml", env)
v.SetConfigFile(configFile)
```

---

## Testing

The package includes comprehensive test coverage:

```
Total Specs:      104
Coverage:         73.3%
Race Detection:   ✅ Zero data races
Execution Time:   ~0.05s (without race), ~1.1s (with race)
```

**Test Categories**:
- Creation and initialization (10 specs)
- Getter methods for all types (17 specs)
- Configuration loading (20 specs)
- Unmarshalling operations (18 specs)
- Custom decode hooks (10 specs)
- Key unsetting (15 specs)
- Error handling (11 specs)
- Concurrent access (3 specs)

**Quality Assurance**:
- ✅ Zero data races (verified with `-race`)
- ✅ Thread-safe concurrent operations
- ✅ All error codes tested
- ✅ Edge cases covered

See [TESTING.md](TESTING.md) for detailed testing documentation.

---

## Contributing

Contributions are welcome! Please follow these guidelines:

**Code Contributions**
- Do not use AI to generate package implementation code
- AI may assist with tests, documentation, and bug fixing
- All contributions must pass `go test -race`
- Maintain or improve test coverage (≥70%)
- Follow existing code style and patterns

**Documentation**
- Update README.md for new features
- Add examples for common use cases
- Keep TESTING.md synchronized with test changes
- Document all exported functions with GoDoc

**Testing**
- Write tests for all new features
- Test edge cases and error conditions
- Verify thread safety with race detector
- Add comments explaining complex scenarios

**Pull Requests**
- Provide clear description of changes
- Reference related issues
- Include test results
- Update documentation

See [CONTRIBUTING.md](../../CONTRIBUTING.md) for detailed guidelines.

---

## Future Enhancements

Potential improvements for future versions:

**Configuration Sources**
- Consul support
- AWS Parameter Store integration
- Azure Key Vault support
- HashiCorp Vault integration

**Features**
- Configuration validation framework
- Schema-based validation (JSON Schema)
- Configuration migration tools
- Encrypted configuration values
- Configuration versioning
- Audit logging for config changes

**Performance**
- Configuration caching layer
- Lazy loading for large configs
- Partial configuration updates
- Optimized remote polling

**Developer Experience**
- Configuration generator from structs
- Interactive configuration wizard
- Configuration diff tool
- Hot-reload without callbacks

Suggestions and contributions are welcome via GitHub issues.

---

## AI Transparency Notice

In accordance with Article 50.4 of the EU AI Act, AI assistance has been used for testing, documentation, and bug fixing under human supervision.

---

## License

MIT License - See [LICENSE](../../LICENSE) file for details.

---

## Resources

- **Viper Documentation**: [spf13/viper](https://github.com/spf13/viper)
- **Mapstructure**: [go-viper/mapstructure](https://github.com/go-viper/mapstructure)
- **Issues**: [GitHub Issues](https://github.com/nabbar/golib/issues)
- **GoDoc**: [pkg.go.dev](https://pkg.go.dev/github.com/nabbar/golib/viper)
- **Testing Guide**: [TESTING.md](TESTING.md)
- **Contributing**: [CONTRIBUTING.md](../../CONTRIBUTING.md)
- **Logger Package**: [golib/logger](../logger)
- **Context Package**: [golib/context](../context)
