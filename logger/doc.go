/*
 * MIT License
 *
 * Copyright (c) 2025 Nicolas JUHEL
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

/*
Package logger provides a comprehensive, production-ready structured logging solution built on top of
logrus with extensive customization options, multiple output destinations, and seamless integration
with popular Go frameworks and libraries.

# Overview

The logger package offers a unified logging interface that extends io.WriteCloser, making it compatible
with any Go code expecting a standard writer. It supports:

  - Multiple simultaneous output destinations (stdout/stderr, files, syslog)
  - Level-based filtering with six standard levels (Debug, Info, Warn, Error, Fatal, Panic)
  - Structured logging with custom fields
  - Automatic caller tracking (file, line, function name, goroutine ID)
  - Thread-safe concurrent logging
  - Integration with standard Go log package
  - Integration with spf13/jwalterweatherman (Hugo/Cobra logger)
  - Integration with Gin web framework
  - Integration with GORM ORM

# Design Philosophy

1. Flexibility: Support multiple output destinations and formats simultaneously
2. Performance: Async hook processing to minimize logging overhead
3. Safety: Thread-safe operations with proper synchronization
4. Compatibility: Seamless integration with existing logging infrastructure
5. Observability: Automatic context capture (caller info, stack traces, timestamps)
6. Standards: Follows structured logging best practices

# Key Features

Structured Logging:
  - Custom fields injection at logger or entry level
  - Automatic field propagation in logger clones
  - Support for complex data types (maps, structs, etc.)
  - Standardized field names via logger/types package

Multiple Output Destinations:
  - Standard output (stdout/stderr) with color support
  - File logging with rotation support
  - Syslog integration (local and remote)
  - Custom writer hooks for specialized outputs

Level Management:
  - Six standard levels: Debug, Info, Warn, Error, Fatal, Panic
  - Per-output level configuration
  - Dynamic level changes without restart
  - Separate level for io.Writer interface

Context Awareness:
  - Goroutine ID tracking
  - Caller information (file, line, function)
  - Stack traces for error scenarios
  - Timestamp with nanosecond precision

# Architecture

Package Structure:

	┌──────────────────────────────────────────────────────────┐
	│  github.com/nabbar/golib/logger                          │
	├──────────────────────────────────────────────────────────┤
	│                                                          │
	│  Logger Interface (interface.go)                         │
	│  ├─ Main logging methods (Debug, Info, Warn, Error...)   │
	│  ├─ Configuration (SetOptions, SetLevel, SetFields)      │
	│  ├─ Integrations (SetStdLogger, SetSPF13Level)           │
	│  └─ Advanced (Clone, Entry, CheckError)                  │
	│                                                          │
	│  Implementation (log.go, manage.go, model.go)            │
	│  ├─ Entry creation with automatic context                │
	│  ├─ Hook management and lifecycle                        │
	│  ├─ Formatter configuration                              │
	│  └─ Thread-safe operations                               │
	│                                                          │
	│  I/O Integration (iowritecloser.go, golog.go)            │
	│  ├─ io.WriteCloser implementation                        │
	│  ├─ Standard log.Logger bridge                           │
	│  └─ Message filtering                                    │
	│                                                          │
	│  Framework Integration (spf13.go)                        │
	│  └─ jwalterweatherman (Hugo/Cobra) integration           │
	│                                                          │
	└──────────────────────────────────────────────────────────┘
	                          │
	      ┌───────────────────┼───────────────────┐
	      │                   │                   │
	      ▼                   ▼                   ▼
	┌───────────┐       ┌───────────┐      ┌───────────┐
	│  config   │       │   entry   │      │  fields   │
	│  Package  │       │  Package  │      │  Package  │
	├───────────┤       ├───────────┤      ├───────────┤
	│ Options   │       │ Entry     │      │ Fields    │
	│ Files     │       │ Builder   │      │ Manager   │
	│ Syslog    │       │ Context   │      │ Clone     │
	└───────────┘       └───────────┘      └───────────┘
	      │                   │                   │
	      └───────────────────┼───────────────────┘
	                          │
	      ┌───────────────────┼───────────────────────┐
	      │                   │                       │
	      ▼                   ▼                       ▼
	┌───────────┐       ┌───────────┐         ┌────────────┐
	│   level   │       │   types   │         │ Hook Impl  │
	│  Package  │       │  Package  │         │  Packages  │
	├───────────┤       ├───────────┤         ├────────────┤
	│ Level     │       │ Hook      │         │ hookfile   │
	│ Constants │       │ Interface │         │ hookstdout │
	│ Converters│       │ Field     │         │ hookstderr │
	└───────────┘       │ Constants │         │ hooksyslog │
	                    └───────────┘         │ hookwriter │
	                                          └────────────┘

Data Flow:

	Application
	     │
	     ▼
	Logger.Debug/Info/Warn/Error...
	     │
	     ▼
	Entry Creation (newEntry)
	     │
	     ├─ Capture caller info (file, line, function)
	     ├─ Capture goroutine ID
	     ├─ Merge default fields
	     ├─ Attach custom data/errors
	     └─ Format message
	     │
	     ▼
	Entry.Log()
	     │
	     ▼
	Logrus Logger
	     │
	     ├─ Level filtering
	     └─ Formatter application
	     │
	     ▼
	Registered Hooks (parallel)
	     │
	     ├─ Hook.Fire() (synchronous, quick)
	     ├─ Queue entry for Hook.Run() (async)
	     │
	     └─ Hook.Run() (background goroutine)
	          │
	          ├─ File writing
	          ├─ Syslog transmission
	          ├─ Network streaming
	          └─ Custom processing

# Performance Characteristics

Hook Processing:
  - Fire() method is synchronous but must be fast (<1ms recommended)
  - Run() method processes entries asynchronously in background goroutines
  - Buffered channels prevent logging from blocking application code
  - Graceful shutdown ensures all queued entries are processed

Memory Usage:
  - Logger instance: ~512 bytes (excluding hooks and fields)
  - Entry overhead: ~256 bytes per log entry
  - Hook buffers: Configurable channel sizes (default: 1000 entries)
  - Field storage: Minimal overhead with copy-on-write semantics

Thread Safety:
  - All Logger methods are thread-safe
  - Concurrent logging from multiple goroutines is supported
  - Internal state protected by sync.RWMutex
  - Atomic operations for simple state management

# Use Cases

1. Application Logging

Basic structured logging for any Go application:

	import (
	    "context"
	    "github.com/nabbar/golib/logger"
	    "github.com/nabbar/golib/logger/config"
	    "github.com/nabbar/golib/logger/level"
	)

	func main() {
	    log := logger.New(context.Background())
	    log.SetLevel(level.InfoLevel)

	    err := log.SetOptions(&config.Options{
	        Stdout: &config.OptionsStd{
	            EnableTrace: true,
	        },
	    })
	    if err != nil {
	        panic(err)
	    }
	    defer log.Close()

	    log.Info("Application started", nil)
	    log.Debug("Debug information", map[string]interface{}{
	        "module": "main",
	        "pid":    os.Getpid(),
	    })
	}

2. HTTP Server Logging

Integration with web frameworks:

	func setupLogging(router *gin.Engine, log logger.Logger) {
	    router.Use(func(c *gin.Context) {
	        start := time.Now()
	        c.Next()
	        latency := time.Since(start)

	        log.Access(
	            c.ClientIP(), c.GetString("user"),
	            start, latency,
	            c.Request.Method, c.Request.URL.Path, c.Request.Proto,
	            c.Writer.Status(), int64(c.Writer.Size()),
	        ).Log()
	    })
	}

3. File and Syslog Logging

Multiple simultaneous destinations:

	log.SetOptions(&config.Options{
	    Stdout: &config.OptionsStd{
	        DisableStandard: false,
	    },
	    LogFile: []config.OptionsFile{
	        {
	            Filepath:    "/var/log/app/app.log",
	            Create:      true,
	            CreatePath:  true,
	            EnableTrace: true,
	        },
	    },
	    LogSyslog: []config.OptionsSyslog{
	        {
	            Network:  "udp",
	            Host:     "localhost",
	            Port:     514,
	            Severity: syslog.LOG_INFO,
	            Facility: syslog.LOG_USER,
	        },
	    },
	})

4. Third-Party Integration

Standard library integration:

	// Replace global standard logger
	log.SetStdLogger(level.WarnLevel, stdlog.LstdFlags)
	stdlog.Println("Now logged through custom logger")

	// Create logger for third-party code
	stdLogger := log.GetStdLogger(level.InfoLevel, stdlog.LstdFlags)
	thirdPartyFunc(stdLogger)

5. Advanced Error Tracking

Structured error logging with context:

	err := performOperation()
	if log.CheckError(level.ErrorLevel, level.InfoLevel, "Operation result", err) {
	    // Error was logged at ErrorLevel
	    return err
	}
	// Success was logged at InfoLevel

	// Or with full context:
	entry := log.Entry(level.ErrorLevel, "Database operation failed")
	entry.FieldAdd("query", query)
	entry.FieldAdd("duration", duration)
	entry.ErrorAdd(true, err)
	entry.Log()

# Best Practices

DO:
  - Always call Close() when done with a logger (use defer)
  - Set appropriate log levels for production vs development
  - Use structured fields instead of string formatting when possible
  - Clone loggers instead of creating new ones for similar configurations
  - Use Entry() for complex logging with multiple fields
  - Set reasonable buffer sizes for file/syslog hooks

DON'T:
  - Don't log sensitive data (passwords, tokens, etc.)
  - Don't use Fatal or Panic unless absolutely necessary
  - Don't create multiple loggers when one with fields would suffice
  - Don't block in custom hook implementations
  - Don't ignore errors from SetOptions()
  - Don't use Debug level in production without careful consideration

# Limitations

Hook Startup Timing:
  - Hooks start asynchronously when SetOptions() is called
  - There's a 500ms grace period to ensure hooks are running
  - Very early log entries might be delayed until hooks start

Log Ordering:
  - Multiple hooks may process entries in different orders
  - File writes are sequential per file but not across destinations
  - Use timestamp field for accurate event ordering

Fatal/Panic Behavior:
  - Fatal level calls os.Exit(1) after logging (deferred functions don't run)
  - Panic level triggers panic() (use with recover() for graceful handling)
  - Both levels may lose buffered log entries if exit is immediate

Memory Limits:
  - Default hook buffers can hold 1000 entries
  - Sustained high log volume can cause buffer overruns and dropped entries
  - Adjust buffer sizes based on expected load

# Sub-Packages

  - config: Logger configuration structures and validation
  - entry: Log entry creation, manipulation, and lifecycle
  - fields: Structured field management with clone and merge operations
  - level: Log level definitions, conversions, and comparisons
  - types: Core types, interfaces, and field constants
  - hookfile: File output hook with rotation support
  - hookstdout: Standard output hook with color support
  - hookstderr: Standard error hook with color support
  - hooksyslog: Syslog hook with local and remote support
  - hookwriter: Generic io.Writer hook for custom outputs
  - gorm: GORM ORM logger integration
  - hashicorp: Hashicorp libraries logger integration (Vault, Consul, etc.)

# External Dependencies

Core:
  - github.com/sirupsen/logrus: Underlying structured logging library
  - github.com/spf13/jwalterweatherman: Hugo/Cobra logger integration

Sub-packages may have additional dependencies for specialized functionality.

# Compatibility

  - Go: Requires Go 1.18 or later
  - Logrus: Compatible with logrus v1.8+
  - Platforms: Linux, macOS, Windows
  - Architectures: amd64, arm64, arm, 386

Thread Safety:
  - All public methods are thread-safe
  - Safe for concurrent use from multiple goroutines
  - Internal state protected by appropriate synchronization primitives

# See Also

Related packages in github.com/nabbar/golib:
  - github.com/nabbar/golib/context: Context management used by logger
  - github.com/nabbar/golib/ioutils/mapCloser: Closer management for hooks
  - github.com/nabbar/golib/semaphore: Concurrency control for logging

External references:
  - Logrus documentation: https://github.com/sirupsen/logrus
  - Structured logging: https://www.thoughtworks.com/insights/blog/practical-introduction-structured-logging
  - The Twelve-Factor App (Logs): https://12factor.net/logs
*/
package logger
