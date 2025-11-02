# Cobra Package

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Production-ready wrapper around [spf13/cobra](https://github.com/spf13/cobra) for building professional CLI applications with enhanced features.

---

## Table of Contents

- [Overview](#overview)
- [Key Features](#key-features)
- [Installation](#installation)
- [Architecture](#architecture)
- [Quick Start](#quick-start)
- [Built-in Commands](#built-in-commands)
- [Flag Management](#flag-management)
- [Configuration](#configuration)
- [Custom Commands](#custom-commands)
- [Advanced Usage](#advanced-usage)
- [Use Cases](#use-cases)
- [API Reference](#api-reference)
- [Testing](#testing)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)
- [Contributing](#contributing)
- [AI Transparency Notice](#ai-transparency-notice)
- [Resources](#resources)
- [License](#license)

---

## Overview

The cobra package provides a comprehensive solution for building CLI applications in Go. It wraps spf13/cobra with additional features for production use, including integrated version management, configuration handling, shell completion support, and thread-safe design.

### Design Philosophy

1. **Instance-Based Architecture** - No global state, ensuring thread-safety and testability
2. **Progressive Enhancement** - Start simple, add features incrementally
3. **Type Safety** - Comprehensive flag support with compile-time safety
4. **Developer Experience** - Intuitive API with clear method names

### Performance

- **Minimal overhead** - Thin wrapper (<1% impact vs raw cobra)
- **Zero-copy flags** - Direct binding to variables, no memory duplication
- **Fast startup** - Lazy initialization and efficient parsing
- **Thread-safe** - Multiple CLI instances can coexist safely

---

## Key Features

- ✅ **Zero-Boilerplate Setup** - Build complete CLI apps in minutes
- ✅ **Thread-Safe Design** - Instance-based, no global state
- ✅ **Type-Safe Flags** - Support for 20+ Go types
- ✅ **Shell Completion** - Auto-generate for bash, zsh, fish, PowerShell
- ✅ **Config Generation** - Auto-generate JSON, YAML, TOML configs
- ✅ **Version Management** - Integrated version display
- ✅ **Logging Integration** - Built-in structured logging support
- ✅ **Error Handling** - Clean error printing and code management
- ✅ **Persistent Flags** - Flags that propagate to subcommands
- ✅ **Custom Commands** - Easy addition of commands and subcommands

---

## Installation

```bash
go get github.com/nabbar/golib/cobra
```

**Requirements:**
- Go 1.18+
- Dependencies: `spf13/cobra`, `nabbar/golib/{logger,version,viper}`

---

## Architecture

### Package Structure

```
cobra/
├── interface.go      # Public API (Cobra interface)
├── model.go          # Core implementation
├── configure.go      # Config file generation
├── completion.go     # Shell completion support
└── printError.go     # Error code management
```

### Component Diagram

```
┌────────────────────────────────────────┐
│         Cobra Interface                │
│    CLI Application Builder             │
└──────────┬─────────────────────────────┘
           │
           ├──> Version (libver.Version)
           │    └─ Version, license, build info
           │
           ├──> Logger (liblog.Logger)
           │    └─ Structured logging
           │
           ├──> Config (libvpr.Viper)
           │    └─ Config loading/generation
           │
           ├──> Commands (spfcbr.Command)
           │    ├─ Root command
           │    ├─ Built-in commands
           │    └─ Custom commands
           │
           └──> Flags (20+ types)
                ├─ Persistent flags
                ├─ Local flags
                └─ Zero-copy binding
```

### Design Patterns

**Instance-Based Design**
```go
// Each instance is independent and thread-safe
app1 := libcbr.New()
app2 := libcbr.New()
// No interference
```

**Dependency Injection**
```go
// Inject as functions for flexibility
app.SetLogger(func() liblog.Logger { return myLogger })
app.SetViper(func() libvpr.Viper { return myConfig })
```

**Progressive Enhancement**
```go
app := libcbr.New()
app.Init()                        // Basic CLI
app.AddCommandCompletion()        // Add completion
app.AddCommandConfigure(...)      // Add config gen
```

---

## Quick Start

### Minimal CLI

```go
package main

import (
    libcbr "github.com/nabbar/golib/cobra"
    libver "github.com/nabbar/golib/version"
)

func main() {
    version := libver.NewVersion(
        libver.License_MIT, "myapp", "My App",
        "2024-01-01", "abc123", "v1.0.0",
        "Author", "myapp", struct{}{}, 0,
    )
    
    app := libcbr.New()
    app.SetVersion(version)
    app.Init()
    app.Execute()
}
```

### CLI with Features

```go
package main

import (
    "context"
    libcbr "github.com/nabbar/golib/cobra"
    liblog "github.com/nabbar/golib/logger"
    libver "github.com/nabbar/golib/version"
    libvpr "github.com/nabbar/golib/viper"
)

func main() {
    version := libver.NewVersion(/* ... */)
    
    app := libcbr.New()
    app.SetVersion(version)
    app.SetLogger(func() liblog.Logger {
        return liblog.New(context.Background)
    })
    app.SetViper(func() libvpr.Viper {
        return libvpr.New()
    })
    app.Init()
    
    // Add flags
    var configPath string
    var verbose int
    app.SetFlagConfig(true, &configPath)
    app.SetFlagVerbose(true, &verbose)
    
    // Add commands
    app.AddCommandCompletion()
    app.AddCommandConfigure("config", "c", "Generate config file")
    
    app.Execute()
}
```

### CLI with Custom Command

```go
package main

import (
    libcbr "github.com/nabbar/golib/cobra"
    spfcbr "github.com/spf13/cobra"
)

func main() {
    app := libcbr.New()
    app.SetVersion(version)
    app.Init()
    
    runCmd := &spfcbr.Command{
        Use:   "run",
        Short: "Run the application",
        RunE: func(cmd *spfcbr.Command, args []string) error {
            // Application logic
            return nil
        },
    }
    
    app.AddCommand(runCmd)
    app.Execute()
}
```

---

## Built-in Commands

### Completion Command

Generates shell completion scripts.

**Usage:**
```bash
# Bash
myapp --completion bash > /etc/bash_completion.d/myapp

# Zsh
myapp --completion zsh > /usr/local/share/zsh/site-functions/_myapp

# Fish
myapp --completion fish > ~/.config/fish/completions/myapp.fish

# PowerShell
myapp --completion powershell > myapp.ps1
```

**Add to CLI:**
```go
app.AddCommandCompletion()
```

**Installation:**
```bash
# Bash
sudo myapp --completion bash /etc/bash_completion.d/myapp
source ~/.bashrc

# Zsh
myapp --completion zsh /usr/local/share/zsh/site-functions/_myapp
source ~/.zshrc

# Fish
myapp --completion fish ~/.config/fish/completions/myapp.fish

# PowerShell
myapp --completion powershell > myapp.ps1
. ./myapp.ps1
```

### Configure Command

Generates configuration files in multiple formats.

**Usage:**
```bash
myapp --configure json config.json   # JSON format
myapp --configure yaml config.yaml   # YAML format
myapp --configure toml config.toml   # TOML format
```

**Add to CLI:**
```go
app.AddCommandConfigure("config", "c", "Generate configuration file")
```

The generated config includes all flags with current values and types.

### Print Error Code Command

Lists all error codes used in the application.

**Usage:**
```bash
myapp --print-error-code
```

**Add to CLI:**
```go
app.AddCommandPrintErrorCode("errors", "e", "Print error codes", 
    func(item, value string) {
        fmt.Printf("%s: %s\n", item, value)
    })
```

---

## Flag Management

### Standard Flags

**Config Flag:**
```go
var configPath string
app.SetFlagConfig(true, &configPath)
// Usage: --config /path/to/config.yaml
```

**Verbose Flag:**
```go
var verbose int
app.SetFlagVerbose(true, &verbose)
// Usage: -v (1), -vv (2), -vvv (3)
```

### Supported Flag Types

**String:**
```go
var str string
app.AddFlagString(true, &str, "name", "n", "default", "Description")
```

**Integer Types:**
```go
var i int
app.AddFlagInt(true, &i, "count", "c", 10, "Count value")

var i8 int8
app.AddFlagInt8(true, &i8, "level", "l", 5, "Level value")

var i16 int16
app.AddFlagInt16(true, &i16, "port", "p", 8080, "Port number")

var i32 int32
app.AddFlagInt32(true, &i32, "size", "s", 1024, "Size value")

var i64 int64
app.AddFlagInt64(true, &i64, "bytes", "b", 4096, "Byte count")
```

**Unsigned Integer Types:**
```go
var u uint
app.AddFlagUint(true, &u, "max", "m", 100, "Maximum value")

var u8 uint8
app.AddFlagUint8(true, &u8, "level", "l", 5, "Level")

var u16 uint16
app.AddFlagUint16(true, &u16, "port", "p", 8080, "Port")

var u32 uint32
app.AddFlagUint32(true, &u32, "size", "s", 1024, "Size")

var u64 uint64
app.AddFlagUint64(true, &u64, "bytes", "b", 4096, "Bytes")
```

**Floating-Point:**
```go
var f32 float32
app.AddFlagFloat32(true, &f32, "rate", "r", 0.5, "Rate value")

var f64 float64
app.AddFlagFloat64(true, &f64, "ratio", "r", 0.75, "Ratio value")
```

**Boolean:**
```go
var enabled bool
app.AddFlagBool(true, &enabled, "enable", "e", false, "Enable feature")
```

**Duration:**
```go
var timeout time.Duration
app.AddFlagDuration(true, &timeout, "timeout", "t", 
    30*time.Second, "Timeout duration")
```

**Count:**
```go
var verbosity int
app.AddFlagCount(true, &verbosity, "verbose", "v", "Increase verbosity")
```

**IP Address:**
```go
var ip net.IP
app.AddFlagIP(true, &ip, "addr", "a", 
    net.IPv4(127, 0, 0, 1), "IP address")

var mask net.IPMask
app.AddFlagIPMask(true, &mask, "mask", "m", 
    net.CIDRMask(24, 32), "Network mask")

var ipnet net.IPNet
app.AddFlagIPNet(true, &ipnet, "network", "n", 
    defaultNet, "IP network")
```

**Slices:**
```go
var strSlice []string
app.AddFlagStringSlice(true, &strSlice, "tags", "t", 
    []string{}, "Tags")

var intSlice []int
app.AddFlagIntSlice(true, &intSlice, "ids", "i", 
    []int{}, "IDs")
```

### Persistent vs Local Flags

**Persistent** (inherited by subcommands):
```go
app.AddFlagString(true, &global, "global", "g", "", "Global option")
```

**Local** (command-specific):
```go
app.AddFlagString(false, &local, "local", "l", "", "Local option")
```

---

## Configuration

### Initialization Function

```go
app.SetFuncInit(func() {
    // Custom initialization
    setupDatabase()
    loadPlugins()
    validateEnvironment()
})
```

### Logger Integration

```go
app.SetLogger(func() liblog.Logger {
    logger := liblog.New(context.Background)
    logger.SetLevel(liblog.InfoLevel)
    return logger
})
```

### Viper Integration

```go
app.SetViper(func() libvpr.Viper {
    v := libvpr.New()
    v.SetConfigType("yaml")
    v.AddConfigPath("/etc/myapp/")
    v.AddConfigPath("$HOME/.myapp")
    v.AddConfigPath(".")
    return v
})
```

### Force No Info Mode

```go
app.SetForceNoInfo(true)  // Disable package path info
```

---

## Custom Commands

### Basic Command

```go
cmd := &spfcbr.Command{
    Use:   "serve",
    Short: "Start the server",
    Long:  "Start the HTTP server on the specified port",
    RunE: func(cmd *spfcbr.Command, args []string) error {
        return startServer()
    },
}

var port int
cmd.Flags().IntVarP(&port, "port", "p", 8080, "Server port")

app.AddCommand(cmd)
```

### Subcommands

```go
// Parent command
dbCmd := &spfcbr.Command{
    Use:   "database",
    Short: "Database operations",
}

// Subcommands
migrateCmd := &spfcbr.Command{
    Use:   "migrate",
    Short: "Run migrations",
    RunE:  runMigrations,
}

seedCmd := &spfcbr.Command{
    Use:   "seed",
    Short: "Seed database",
    RunE:  seedDatabase,
}

dbCmd.AddCommand(migrateCmd, seedCmd)
app.AddCommand(dbCmd)

// Usage: myapp database migrate
```

### Pre/Post Hooks

```go
cmd := &spfcbr.Command{
    Use: "process",
    PreRunE: func(cmd *spfcbr.Command, args []string) error {
        return validateInput(args)  // Before execution
    },
    RunE: func(cmd *spfcbr.Command, args []string) error {
        return processData(args)    // Main logic
    },
    PostRunE: func(cmd *spfcbr.Command, args []string) error {
        return cleanup()            // After execution
    },
}
```

---

## Advanced Usage

### Multiple CLI Instances

```go
// Thread-safe: independent instances
adminCLI := libcbr.New()
adminCLI.SetVersion(adminVersion)
adminCLI.Init()

userCLI := libcbr.New()
userCLI.SetVersion(userVersion)
userCLI.Init()

go adminCLI.Execute()
go userCLI.Execute()
```

### Dynamic Command Registration

```go
app := libcbr.New()
app.Init()

plugins := loadPlugins()
for _, plugin := range plugins {
    app.AddCommand(plugin.GetCommand())
}
```

### Context-Aware Commands

```go
cmd := &spfcbr.Command{
    Use: "run",
    RunE: func(cmd *spfcbr.Command, args []string) error {
        ctx := cmd.Context()
        
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
            return doWork(ctx)
        }
    },
}
```

### Custom Error Handling

```go
app.GetCommand().SilenceErrors = true
app.GetCommand().SilenceUsage = true

if err := app.Execute(); err != nil {
    logger.Error("Execution failed", "error", err)
    os.Exit(1)
}
```

---

## Use Cases

### Microservice CLI

```go
app := libcbr.New()
app.SetVersion(version)
app.Init()

serverCmd := &spfcbr.Command{
    Use:   "server",
    Short: "Start microservice",
    RunE:  runServer,
}

healthCmd := &spfcbr.Command{
    Use:   "health",
    Short: "Health check",
    RunE:  checkHealth,
}

app.AddCommand(serverCmd, healthCmd)
```

### DevOps Deployment Tool

```go
app := libcbr.New()
app.Init()

deployCmd := &spfcbr.Command{
    Use:   "deploy",
    Short: "Deploy application",
}

var env string
deployCmd.Flags().StringVarP(&env, "env", "e", "dev", "Environment")
deployCmd.RunE = func(cmd *spfcbr.Command, args []string) error {
    return deploy(env)
}

app.AddCommand(deployCmd)
```

### System Administration Tool

```go
app := libcbr.New()
app.Init()

backupCmd := &spfcbr.Command{
    Use:   "backup",
    Short: "Backup data",
    RunE:  performBackup,
}

restoreCmd := &spfcbr.Command{
    Use:   "restore",
    Short: "Restore from backup",
    RunE:  performRestore,
}

app.AddCommand(backupCmd, restoreCmd)
```

### Configuration Generator

```go
app := libcbr.New()
app.Init()
app.AddCommandConfigure("config", "c", "Generate config file")

// Users can: myapp --configure yaml config.yaml
```

---

## API Reference

### Factory Function

**New() Cobra**

Creates a new Cobra instance.

```go
app := libcbr.New()
```

### Configuration Methods

| Method | Description |
|--------|-------------|
| `SetVersion(v libver.Version)` | Set version information |
| `SetLogger(fct FuncLogger)` | Inject logger provider |
| `SetViper(fct FuncViper)` | Inject config provider |
| `SetFuncInit(fct FuncInit)` | Set initialization function |
| `SetForceNoInfo(flag bool)` | Control info display |
| `Init()` | Initialize CLI (required) |
| `Execute() error` | Execute CLI |

### Command Methods

| Method | Description |
|--------|-------------|
| `AddCommand(cmd ...*spfcbr.Command)` | Add custom commands |
| `AddCommandCompletion()` | Add completion command |
| `AddCommandConfigure(use, short, long string)` | Add config command |
| `AddCommandPrintErrorCode(...)` | Add error code command |
| `GetCommand() *spfcbr.Command` | Get root command |

### Flag Methods

| Method | Description |
|--------|-------------|
| `SetFlagConfig(persistent bool, flagVar *string)` | Config flag |
| `SetFlagVerbose(persistent bool, flagVar *int)` | Verbose flag |
| `AddFlagString(...)` | String flag |
| `AddFlagInt(...)` | Int flag |
| `AddFlagBool(...)` | Bool flag |
| `AddFlagDuration(...)` | Duration flag |
| `AddFlagFloat32/64(...)` | Float flags |
| `AddFlagInt8/16/32/64(...)` | Integer flags |
| `AddFlagUint/8/16/32/64(...)` | Unsigned integer flags |
| `AddFlagIP/IPMask/IPNet(...)` | IP address flags |
| `AddFlagStringSlice/IntSlice(...)` | Slice flags |
| `AddFlagCount(...)` | Count flag |

---

## Testing

The package uses **Ginkgo v2** + **Gomega** for comprehensive testing.

```bash
# Run tests
go test ./...

# With coverage
go test -cover ./...

# With race detection
CGO_ENABLED=1 go test -race ./...

# Using Ginkgo
go install github.com/onsi/ginkgo/v2/ginkgo@latest
ginkgo -r
```

**Coverage:**

| Component | Coverage | Specs |
|-----------|----------|-------|
| Commands | ~80% | 25 |
| Flags | ~85% | 30 |
| Configuration | ~70% | 20 |
| Completion | ~65% | 15 |
| Initialization | ~75% | 10 |
| **Total** | **~75%** | **100** |

See [TESTING.md](TESTING.md) for detailed testing documentation.

---

## Best Practices

### 1. Initialization Order

```go
// ✅ Correct
app := libcbr.New()
app.SetVersion(version)
app.SetLogger(logger)
app.SetViper(viper)
app.Init()
app.AddCommand(cmd)
app.Execute()
```

### 2. Flag Organization

```go
// ✅ Group related flags
type Config struct {
    Host string
    Port int
    TLS  bool
}

var cfg Config
app.AddFlagString(true, &cfg.Host, "host", "h", "localhost", "Host")
app.AddFlagInt(true, &cfg.Port, "port", "p", 8080, "Port")
app.AddFlagBool(true, &cfg.TLS, "tls", "t", false, "Enable TLS")
```

### 3. Error Handling

```go
// ✅ Wrap errors with context
cmd.RunE = func(cmd *spfcbr.Command, args []string) error {
    if err := validate(args); err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }
    return execute()
}
```

### 4. Command Documentation

```go
// ✅ Comprehensive documentation
cmd := &spfcbr.Command{
    Use:   "deploy [env]",
    Short: "Deploy application",
    Long: `Deploy application to environment.
    
Environments: dev, staging, prod`,
    Example: "myapp deploy prod --region us-east-1",
    Args:    spfcbr.ExactArgs(1),
    RunE:    deployCommand,
}
```

### 5. Resource Cleanup

```go
// ✅ Always clean up
cmd.RunE = func(cmd *spfcbr.Command, args []string) error {
    db, err := openDB()
    if err != nil {
        return err
    }
    defer db.Close()
    
    return process(db)
}
```

### 6. Testable Design

```go
// ✅ Factory for testing
func CreateApp(logger liblog.Logger) libcbr.Cobra {
    app := libcbr.New()
    app.SetLogger(func() liblog.Logger { return logger })
    app.Init()
    return app
}

// In tests
app := CreateApp(mockLogger)
```

---

## Troubleshooting

### Flags Not Working

**Problem**: Flags not recognized.

**Solution**: Call `Init()` before adding flags.
```go
app.Init()  // ✅ First
app.SetFlagConfig(true, &config)
```

### Command Not Found

**Problem**: Custom command not found.

**Solution**: Add commands after `Init()`.
```go
app.Init()
app.AddCommand(myCommand)  // ✅ After Init
```

### Configuration Not Loading

**Problem**: Viper config not loading.

**Solution**: Set viper before `Init()` and read config.
```go
app.SetViper(func() libvpr.Viper {
    v := libvpr.New()
    v.SetConfigFile(path)
    v.ReadInConfig()  // ✅ Read config
    return v
})
app.Init()
```

### Completion Not Working

**Problem**: Shell completion not functioning.

**Solution**: Install completion correctly and check permissions.
```bash
sudo myapp --completion bash /etc/bash_completion.d/myapp
chmod +x /etc/bash_completion.d/myapp
source ~/.bashrc
```

### Debug Mode

**Enable verbose logging:**
```go
app.SetLogger(func() liblog.Logger {
    logger := liblog.New(context.Background)
    logger.SetLevel(liblog.DebugLevel)
    return logger
})
```

---

## Contributing

Contributions welcome! Follow these guidelines:

**Code:**
- Do not use AI for implementation code
- AI may assist with tests, docs, and bug fixes
- Pass `go test -race`
- Follow existing patterns
- Add tests for new features

**Documentation:**
- Update README.md for new features
- Add examples for common use cases
- Keep docs synchronized

**Pull Requests:**
- Clear description of changes
- Reference related issues
- Include test results
- Update documentation

---

## AI Transparency Notice

In accordance with Article 50.4 of the EU AI Act, AI assistance has been used for testing, documentation, and bug fixing under human supervision.

---

## Resources

**Documentation:**
- [spf13/cobra](https://cobra.dev/)
- [Cobra User Guide](https://github.com/spf13/cobra/blob/main/user_guide.md)
- [Package GoDoc](https://pkg.go.dev/github.com/nabbar/golib/cobra)

**Related Packages:**
- [logger](../logger) - Logging implementation
- [version](../version) - Version management
- [viper](../viper) - Configuration management

**External Resources:**
- [12 Factor CLI Apps](https://medium.com/@jdxcode/12-factor-cli-apps-dd3c227a0e46)
- [CLI Best Practices](https://clig.dev/)

---

## License

MIT License - See [LICENSE](../../LICENSE) file for details.
