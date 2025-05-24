# `logger` Package Documentation

The `logger` package provides a robust, thread-safe, and extensible logging system for Go applications.  
It is built primarily on top of the [logrus](https://github.com/sirupsen/logrus) library, but also implements wrappers and integration for other logging systems such as 
 - the standard Go `log` package, 
 - [Logrus](https://github.com/sirupsen/logrus)
 - [SPF13 / jwalterweatherman](https://github.com/spf13/jwalterweatherman), 
 - [HashiCorp loggers](https://github.com/hashicorp/go-hclog)
 - [GORM loggers](https://gorm.io/docs/logger.html).
 - ...

---

## Features

- Multiple log levels: Debug, Info, Warn, Error, Fatal, Panic, Nil
- Structured logging with custom fields and data
- Output to stdout, stderr, files, and syslog (via hooks)
- Dynamic configuration and runtime updates
- Filtering of log messages by pattern
- Integration with Go's `log.Logger`, spf13/jwalterweatherman, and others
- Cloning and context-aware loggers
- Thread-safe operations
- Access log support for HTTP servers
- Custom hooks and extensibility

---

## Main Types & Interfaces

### Logger Interface

The main interface for logging, supporting both structured and unstructured logs.

- `SetLevel(lvl Level)`, `GetLevel()`
- `SetIOWriterLevel(lvl Level)`, `GetIOWriterLevel()`
- `SetIOWriterFilter(pattern ...)`, `AddIOWriterFilter(pattern ...)`
- `SetOptions(opt *Options)`, `GetOptions()`
- `SetFields(fields Fields)`, `GetFields()`
- `Clone() Logger`
- `SetSPF13Level(lvl Level, log *jwalterweatherman.Notepad)`
- `GetStdLogger(lvl Level, logFlags int) *log.Logger`
- `SetStdLogger(lvl Level, logFlags int)`
- Logging methods: `Debug`, `Info`, `Warning`, `Error`, `Fatal`, `Panic`, `LogDetails`, `CheckError`, `Entry`, `Access`
- Implements `io.WriteCloser` for compatibility

### Logger Construction

- `New(ctx FuncContext) Logger`  
  Create a new logger instance with a context provider.

---

## Configuration

The logger is configured via the `Options` struct, which allows you to:

- Enable/disable stdout and stderr outputs, with color and trace options
- Add multiple file outputs with custom settings
- Add syslog outputs
- Set trace filters for file paths
- Control stack trace and timestamp inclusion

Configuration can be updated at runtime using `SetOptions`.

---

## Usage Example

```go
import (
    "github.com/nabbar/golib/logger"
    "github.com/nabbar/golib/logger/config"
    "github.com/nabbar/golib/logger/level"
    "context"
)

log := logger.New(func() context.Context { return context.Background() })
defer log.Close()

log.SetLevel(level.DebugLevel)
log.SetOptions(&config.Options{
    Stdout: &config.OptionsStd{EnableTrace: true},
})

log.Info("Application started", nil)
log.Debug("Debugging details: %v", map[string]interface{}{"foo": "bar"})
log.SetFields(log.GetFields().Add("service", "my-service"))
log.Error("An error occurred: %s", nil, "details")
```

---

## Error Handling

All errors are wrapped with custom codes for diagnostics.  
Use `err.Error()` for user-friendly messages and check error codes for troubleshooting.

---

## Subpackages

The `logger` package is composed of several subpackages, each providing specialized features.  
**See each section for detailed documentation:**

- [`config`](#loggerconfig-subpackage-documentation) — Logger configuration structures and helpers. See the [Config subpackage documentation](#loggerconfig-subpackage-documentation) for more details.
- [`entry`](#loggerentry-subpackage-documentation) — Log entry management and structured logging. See the [Entry subpackage documentation](#loggerentry-subpackage-documentation) for more details.
- [`fields`](#loggerfields-subpackage-documentation) — Structured fields for log entries. See the [Fields subpackage documentation](#loggerfields-subpackage-documentation) for more details.
- [`hookfile`](#loggerhookfile-subpackage-documentation) — File output hooks for logrus. See the [HookFile subpackage documentation](#loggerhookfile-subpackage-documentation) for more details.
- [`hookstderr`](#loggerhookstderr-subpackage-documentation) — Stderr output hooks for logrus. See the [HookStderr subpackage documentation](#loggerhookstderr-subpackage-documentation) for more details.
- [`hookstdout`](#loggerhookstdout-subpackage-documentation) — Stdout output hooks for logrus. See the [HookStdout subpackage documentation](#loggerhookstdout-subpackage-documentation) for more details.
- [`hooksyslog`](#loggerhooksyslog-subpackage-documentation) — Syslog output hooks for logrus. See the [HookSyslog subpackage documentation](#loggerhooksyslog-subpackage-documentation) for more details.
- [`level`](#loggerlevel-subpackage-documentation) — Log level definitions and utilities. See the [Level subpackage documentation](#loggerlevel-subpackage-documentation) for more details.
- [`types`](#loggertypes-subpackage-documentation) — Common types and interfaces for logger internals. See the [Types subpackage documentation](#loggertypes-subpackage-documentation) for more details.

In add, the `logger` package the provides adapters for external loggers, enabling seamless integration and type conversion with the main logging infrastructure:
- [GORM logger adapter](#loggergorm-subpackage-documentation): Bridges `gorm.io/gorm/logger` to centralize and standardize GORM logs. See the [GORM subpackage documentation](#loggergorm-subpackage-documentation) for more details.
- [HashiCorp hclog adapter](#loggerhashicorp-subpackage-documentation): Integrates `github.com/hashicorp/go-hclog` with the unified logger system. See the [HashiCorp subpackage documentation](#loggerhashicorp-subpackage-documentation) for more details.

---

### `logger/config` Subpackage Documentation

The `logger/config` subpackage provides configuration structures and utilities for customizing the behavior and outputs of the logger system. It enables fine-grained control over log destinations, formatting, filtering, and runtime options.

---

#### Features

- Centralized configuration for all logger outputs (stdout, files, syslog)
- Support for inheritance and merging of default options
- Validation helpers for configuration correctness
- Clone and merge utilities for dynamic configuration management

---

#### Main Types

##### `Options`

The main configuration struct for the logger.  
Key fields:

- `InheritDefault` (`bool`): If true, inherits from a registered default options function.
- `TraceFilter` (`string`): Path filter for cleaning traces in log output.
- `Stdout` (`*OptionsStd`): Options for stdout/stderr logging.
- `LogFileExtend` (`bool`): If true, appends to default file outputs; otherwise, replaces them.
- `LogFile` (`OptionsFiles`): List of file output configurations.
- `LogSyslogExtend` (`bool`): If true, appends to default syslog outputs; otherwise, replaces them.
- `LogSyslog` (`OptionsSyslogs`): List of syslog output configurations.

**Key methods:**

- `RegisterDefaultFunc(fct FuncOpt)`: Register a function to provide default options for inheritance.
- `Validate()`: Validate the configuration and return a custom error if invalid.
- `Clone() Options`: Deep copy of the options.
- `Merge(opt *Options)`: Merge another options struct into the current one.
- `Options() *Options`: Return the effective options, applying inheritance if enabled.

---

##### `OptionsStd`

Configuration for standard output (stdout/stderr):

- `DisableStandard` (`bool`): Disable writing to stdout/stderr.
- `DisableStack` (`bool`): Disable goroutine ID in messages.
- `DisableTimestamp` (`bool`): Disable timestamps in messages.
- `EnableTrace` (`bool`): Enable caller/file/line tracing.
- `DisableColor` (`bool`): Disable color formatting.
- `EnableAccessLog` (`bool`): Enable access log for API routers.

**Method:**

- `Clone() *OptionsStd`: Deep copy of the struct.

---

##### `OptionsFile` and `OptionsFiles`

Configuration for file outputs:

- `LogLevel` (`[]string`): Allowed log levels for this file.
- `Filepath` (`string`): Path to the log file.
- `Create` (`bool`): Create the file if it does not exist.
- `CreatePath` (`bool`): Create the directory path if it does not exist.
- `FileMode` (`Perm`): File permissions.
- `PathMode` (`Perm`): Directory permissions.
- `DisableStack`, `DisableTimestamp`, `EnableTrace`, `EnableAccessLog`: Same as above.
- `FileBufferSize` (`Size`): Buffer size for file writes.

**Methods:**

- `Clone() OptionsFile`: Deep copy of the struct.
- `Clone() OptionsFiles`: Deep copy of the slice.

---

##### `OptionsSyslog` and `OptionsSyslogs`

Configuration for syslog outputs:

- `LogLevel` (`[]string`): Allowed log levels for this syslog.
- `Network` (`string`): Network type (e.g., tcp, udp).
- `Host` (`string`): Syslog server address.
- `Facility` (`string`): Syslog facility.
- `Tag` (`string`): Syslog tag or logger name.
- `DisableStack`, `DisableTimestamp`, `EnableTrace`, `EnableAccessLog`: Same as above.

**Methods:**

- `Clone() OptionsSyslog`: Deep copy of the struct.
- `Clone() OptionsSyslogs`: Deep copy of the slice.

---

##### Error Handling

Custom error codes are provided for configuration validation and parameter errors.  
Use `Validate()` to check configuration correctness and handle errors accordingly.

---

#### Example Usage

```go
import (
    "github.com/nabbar/golib/logger/config"
)

opts := &config.Options{
    InheritDefault: false,
    TraceFilter: "/src/",
    Stdout: &config.OptionsStd{
        EnableTrace: true,
    },
    LogFile: config.OptionsFiles{
        {
            LogLevel: []string{"Debug", "Info"},
            Filepath: "/var/log/myapp.log",
            Create: true,
            FileMode: 0644,
        },
    },
}

if err := opts.Validate(); err != nil {
    // handle configuration error
}
```

---

#### Notes

- All configuration structs support cloning and merging for dynamic and layered setups.
- Designed for use with the main logger package and its subpackages.
- Ensures thread-safe and consistent logger configuration across your application.
---

### `logger/entry` Subpackage Documentation

The `logger/entry` subpackage provides the core types and methods for creating, managing, and logging structured log entries. It enables advanced logging scenarios with support for custom fields, error handling, data attachment, and integration with frameworks like Gin.

---

#### Features

- Creation and manipulation of structured log entries
- Support for custom fields and data
- Error collection and management within log entries
- Integration with Gin context for error propagation
- Flexible logging with level and context control
- Thread-safe design

---

#### Main Types

##### `Entry` Interface

Represents a single log entry with methods for configuration and logging:

- `SetLogger(fct func() *logrus.Logger) Entry`  
  Set the logger instance provider for this entry.
- `SetLevel(lvl Level) Entry`  
  Set the log level for the entry.
- `SetMessageOnly(flag bool) Entry`  
  Log only the message, ignoring structured fields.
- `SetEntryContext(etime, stack, caller, file, line, msg) Entry`  
  Set context information (timestamp, stack, caller, etc.).
- `SetGinContext(ctx *gin.Context) Entry`  
  Attach a Gin context for error propagation.
- `DataSet(data interface{}) Entry`  
  Attach arbitrary data to the entry.
- `Check(lvlNoErr Level) bool`  
  Log the entry and return true if errors are present.
- `Log()`  
  Log the entry using the configured logger.

##### Field Management

- `FieldAdd(key string, val interface{}) Entry`  
  Add a custom field to the entry.
- `FieldMerge(fields Fields) Entry`  
  Merge multiple fields into the entry.
- `FieldSet(fields Fields) Entry`  
  Replace all custom fields.
- `FieldClean(keys ...string) Entry`  
  Remove specific fields by key.

##### Error Management

- `ErrorClean() Entry`  
  Remove all errors from the entry.
- `ErrorSet(err []error) Entry`  
  Set the error slice for the entry.
- `ErrorAdd(cleanNil bool, err ...error) Entry`  
  Add one or more errors, optionally skipping nil values.

---

#### Example Usage

```go
import (
    "github.com/nabbar/golib/logger/entry"
    "github.com/nabbar/golib/logger/level"
)

e := entry.New(level.InfoLevel).
    FieldAdd("user", "alice").
    ErrorAdd(true, someError).
    DataSet(map[string]interface{}{"extra": 123})

e.Log()
```

---

#### Integration

- **Gin**: Use `SetGinContext` to propagate errors to the Gin context.
- **Custom Fields**: Use `FieldAdd`, `FieldMerge`, and `FieldSet` for structured logging.
- **Error Handling**: Use `ErrorAdd`, `ErrorSet`, and `ErrorClean` to manage error slices within entries.

---

#### Notes

- All entry methods are chainable for fluent usage.
- Logging is performed via Logrus and supports all configured logger outputs.
- Designed for use with the main logger package and compatible with other subpackages.

---

### `logger/fields` Subpackage Documentation

The `logger/fields` subpackage provides a flexible and thread-safe way to manage structured key-value pairs (fields) for log entries. It is designed to integrate seamlessly with the logger system, supporting advanced field manipulation, cloning, and serialization.

---

#### Features

- Thread-safe storage and manipulation of log fields
- Integration with context for field inheritance and isolation
- JSON marshaling and unmarshaling for structured logging
- Conversion to Logrus fields for compatibility
- Functional mapping and dynamic field updates
- Cloning of field sets for context propagation

---

#### Main Types

##### `Fields` Interface

Represents a set of structured fields for log entries.

- Inherits from `Config[string]` (context-aware configuration)
- Implements `json.Marshaler` and `json.Unmarshaler`
- `FieldsClone(ctx context.Context) Fields`  
  Clone the fields set, optionally with a new context.
- `Add(key string, val interface{}) Fields`  
  Add or update a key-value pair in the fields.
- `Logrus() logrus.Fields`  
  Convert the fields to a `logrus.Fields` map for Logrus integration.
- `Map(fct func(key string, val interface{}) interface{}) Fields`  
  Apply a function to each field value and update it.

##### Construction

- `New(ctx FuncContext) Fields`  
  Create a new `Fields` instance with a context provider.

---

#### Example Usage

```go
import (
    "github.com/nabbar/golib/logger/fields"
    "context"
)

f := fields.New(func() context.Context { return context.Background() })
f = f.Add("user", "alice").Add("role", "admin")

logrusFields := f.Logrus() // Use with Logrus logger

// Clone fields for a new context
f2 := f.FieldsClone(context.TODO())

// Map example: uppercase all string values
f.Map(func(key string, val interface{}) interface{} {
    if s, ok := val.(string); ok {
        return strings.ToUpper(s)
    }
    return val
})
```

---

#### Integration

- Use `Fields` to attach structured data to log entries.
- Supports context-based field inheritance for request-scoped logging.
- Compatible with Logrus and JSON-based loggers.

---

#### Notes

- All operations are safe for concurrent use.
- Fields can be serialized/deserialized as JSON for structured logging.
- Designed for use with the main logger package and its subpackages.

---

### `logger/hookfile` Subpackage Documentation

The `logger/hookfile` subpackage provides file output hooks for the logger system, enabling efficient, buffered, and concurrent logging to files. It is designed for integration with Logrus and supports advanced file management features.

---

#### Features

- Logrus-compatible file output hook
- Supports multiple log levels per file
- Buffered and batched writes for performance
- Automatic file and directory creation with configurable permissions
- Optional stack trace, timestamp, and trace information filtering
- Access log support for API routers
- Thread-safe and context-aware operation
- Graceful shutdown and buffer flushing

---

#### Main Types

##### `HookFile` Interface

Represents a file output hook for Logrus.

- Inherits from the logger `Hook` interface
- `Done() <-chan struct{}`: Returns a channel closed when the hook is stopped

##### Construction

- `New(opt OptionsFile, format logrus.Formatter) (HookFile, error)`  
  Creates a new file hook with the given configuration and formatter.  
  Returns an error if the file path is missing or cannot be created.

---

#### Configuration

The hook is configured using an `OptionsFile` struct, which includes:

- `LogLevel`: List of log levels to write to this file
- `Filepath`: Path to the log file
- `Create`: Whether to create the file if it does not exist
- `CreatePath`: Whether to create the directory path if it does not exist
- `FileMode`, `PathMode`: File and directory permissions
- `DisableStack`, `DisableTimestamp`, `EnableTrace`, `EnableAccessLog`: Output options
- `FileBufferSize`: Buffer size for batched writes

---

#### Usage Example

```go
import (
    "github.com/nabbar/golib/logger/hookfile"
    "github.com/nabbar/golib/logger/config"
    "github.com/sirupsen/logrus"
    "context"
)

opt := config.OptionsFile{
    Filepath: "/var/log/myapp.log",
    Create: true,
    FileMode: 0644,
    LogLevel: []string{"Info", "Error"},
}

hook, err := hookfile.New(opt, &logrus.TextFormatter{})
if err != nil {
    // handle error
}

log := logrus.New()
hook.RegisterHook(log)

// Start the hook's background writer
ctx, cancel := context.WithCancel(context.Background())
go hook.(*hookfile.HookFileImpl).Run(ctx)

// ... use logrus as usual

// On shutdown
cancel()
<-hook.Done()
```

---

#### Buffering and Performance

- Writes are buffered and flushed periodically or when the buffer is full.
- On shutdown, all buffered logs are flushed to disk.
- Buffer size is configurable for performance tuning.

---

#### Error Handling

- Returns errors for missing file paths, closed streams, or file system issues.
- Errors are surfaced during hook creation or log writing.

---

#### Notes

- Designed for use with the main logger package and Logrus.
- All operations are safe for concurrent use.
- Supports dynamic log level filtering and flexible file management.
- Integrates with the logger configuration system for unified setup.

---

### `logger/hookstderr` Subpackage Documentation

The `logger/hookstderr` subpackage provides a Logrus-compatible hook for logging to `stderr`, with support for color output, log level filtering, and advanced formatting options. It is designed for seamless integration with the main logger system and supports both standard and access log modes.

---

#### Features

- Logrus hook for writing logs to `stderr`
- Supports colorized output (with automatic detection)
- Configurable log levels per hook
- Optional stack trace, timestamp, and trace information filtering
- Access log mode for API routers
- Thread-safe and context-aware
- Compatible with custom formatters

---

#### Main Types

##### `HookStdErr` Interface

Represents a `stderr` output hook for Logrus.

- Inherits from the logger `Hook` interface

##### Construction

- `New(opt *OptionsStd, lvls []logrus.Level, f logrus.Formatter) (HookStdErr, error)`  
  Creates a new `stderr` hook with the given configuration, log levels, and formatter.  
  Returns `nil` if standard output is disabled.

---

#### Configuration

The hook is configured using an `OptionsStd` struct, which includes:

- `DisableStandard`: Disable writing to `stderr`
- `DisableStack`: Remove stack trace from log output
- `DisableTimestamp`: Remove timestamps from log output
- `EnableTrace`: Include caller, file, and line information
- `DisableColor`: Disable color formatting
- `EnableAccessLog`: Enable access log mode (plain message output)

---

#### Usage Example

```go
import (
    "github.com/nabbar/golib/logger/hookstderr"
    "github.com/nabbar/golib/logger/config"
    "github.com/sirupsen/logrus"
)

opt := &config.OptionsStd{
    EnableTrace: true,
    DisableColor: false,
}

hook, err := hookstderr.New(opt, []logrus.Level{logrus.InfoLevel, logrus.ErrorLevel}, &logrus.TextFormatter{})
if err != nil {
    // handle error
}

log := logrus.New()
hook.RegisterHook(log)

// Use logrus as usual; logs will be sent to stderr via the hook
log.Info("This is an info message")
```

---

#### Output Behavior

- If color is enabled, output is colorized for better readability.
- In access log mode, only the message is output, with a newline.
- Stack trace, timestamp, and trace fields can be included or filtered based on configuration.
- The hook is safe for concurrent use.

---

#### Error Handling

- Returns an error if the writer is not set up.
- All write operations are checked for errors.

---

#### Notes

- Designed for use with the main logger package and Logrus.
- Integrates with the logger configuration system for unified setup.
- All operations are thread-safe and suitable for production environments.

---

### `logger/hookstdout` Subpackage Documentation

The `logger/hookstdout` subpackage provides a Logrus-compatible hook for logging to `stdout`, supporting color output, log level filtering, and advanced formatting options. It is designed for seamless integration with the main logger system and supports both standard and access log modes.

---

#### Features

- Logrus hook for writing logs to `stdout`
- Supports colorized output (with automatic detection)
- Configurable log levels per hook
- Optional stack trace, timestamp, and trace information filtering
- Access log mode for API routers
- Thread-safe and context-aware
- Compatible with custom formatters

---

#### Main Types

##### `HookStdOut` Interface

Represents a `stdout` output hook for Logrus.

- Inherits from the logger `Hook` interface

##### Construction

- `New(opt *OptionsStd, lvls []logrus.Level, f logrus.Formatter) (HookStdOut, error)`  
  Creates a new `stdout` hook with the given configuration, log levels, and formatter.  
  Returns `nil` if standard output is disabled.

---

#### Configuration

The hook is configured using an `OptionsStd` struct, which includes:

- `DisableStandard`: Disable writing to `stdout`
- `DisableStack`: Remove stack trace from log output
- `DisableTimestamp`: Remove timestamps from log output
- `EnableTrace`: Include caller, file, and line information
- `DisableColor`: Disable color formatting
- `EnableAccessLog`: Enable access log mode (plain message output)

---

#### Usage Example

```go
import (
    "github.com/nabbar/golib/logger/hookstdout"
    "github.com/nabbar/golib/logger/config"
    "github.com/sirupsen/logrus"
)

opt := &config.OptionsStd{
    EnableTrace: true,
    DisableColor: false,
}

hook, err := hookstdout.New(opt, []logrus.Level{logrus.InfoLevel, logrus.ErrorLevel}, &logrus.TextFormatter{})
if err != nil {
    // handle error
}

log := logrus.New()
hook.RegisterHook(log)

// Use logrus as usual; logs will be sent to stdout via the hook
log.Info("This is an info message")
```

---

#### Output Behavior

- If color is enabled, output is colorized for better readability.
- In access log mode, only the message is output, with a newline.
- Stack trace, timestamp, and trace fields can be included or filtered based on configuration.
- The hook is safe for concurrent use.

---

#### Error Handling

- Returns an error if the writer is not set up.
- All write operations are checked for errors.

---

#### Notes

- Designed for use with the main logger package and Logrus.
- Integrates with the logger configuration system for unified setup.
- All operations are thread-safe and suitable for production environments.

---

### `logger/hooksyslog` Subpackage Documentation

The `logger/hooksyslog` subpackage provides a Logrus-compatible hook for sending logs to syslog servers, supporting both Unix and Windows platforms. It offers advanced configuration for syslog facilities, severities, network protocols, and formatting, making it suitable for production-grade logging in distributed systems.

---

#### Features

- Logrus hook for sending logs to syslog (local or remote)
- Supports all standard syslog facilities and severities
- Configurable network protocol (e.g., UDP, TCP, Unix socket)
- Customizable log levels per hook
- Optional stack trace, timestamp, and trace information filtering
- Access log mode for API routers
- Thread-safe and context-aware
- Graceful shutdown and error handling
- Compatible with custom formatters

---

#### Main Types

##### `HookSyslog` Interface

Represents a syslog output hook for Logrus.

- Inherits from the logger `Hook` interface
- `Done() <-chan struct{}`: Returns a channel closed when the hook is stopped
- `WriteSev(s SyslogSeverity, p []byte) (n int, err error)`: Write a message with a specific syslog severity

##### Construction

- `New(opt OptionsSyslog, format logrus.Formatter) (HookSyslog, error)`  
  Creates a new syslog hook with the given configuration and formatter.

---

#### Configuration

The hook is configured using an `OptionsSyslog` struct, which includes:

- `LogLevel`: List of log levels to send to syslog
- `Network`: Network protocol (e.g., tcp, udp, unix)
- `Host`: Syslog server address or socket path
- `Facility`: Syslog facility (e.g., LOCAL0, DAEMON)
- `Tag`: Syslog tag or logger name
- `DisableStack`: Remove stack trace from log output
- `DisableTimestamp`: Remove timestamps from log output
- `EnableTrace`: Include caller, file, and line information
- `EnableAccessLog`: Enable access log mode (plain message output)

---

#### Syslog Severity and Facility

- `SyslogSeverity`: Enum for syslog severities (EMERG, ALERT, CRIT, ERR, WARNING, NOTICE, INFO, DEBUG)
- `SyslogFacility`: Enum for syslog facilities (KERN, USER, MAIL, DAEMON, AUTH, SYSLOG, LPR, NEWS, UUCP, CRON, AUTHPRIV, FTP, LOCAL0-LOCAL7)
- Use `MakeSeverity(string)` and `MakeFacility(string)` to parse string values

---

#### Usage Example

```go
import (
    "github.com/nabbar/golib/logger/hooksyslog"
    "github.com/nabbar/golib/logger/config"
    "github.com/sirupsen/logrus"
    "context"
)

opt := config.OptionsSyslog{
    Network:  "udp",
    Host:     "127.0.0.1:514",
    Facility: "LOCAL0",
    Tag:      "myapp",
    LogLevel: []string{"info", "error"},
}

hook, err := hooksyslog.New(opt, &logrus.TextFormatter{})
if err != nil {
    panic(err)
}

log := logrus.New()
hook.RegisterHook(log)

// Start the syslog hook background process
ctx, cancel := context.WithCancel(context.Background())
go hook.(*hooksyslog.HookSyslogImpl).Run(ctx)

// Use logrus as usual; logs will be sent to syslog
log.Info("This is an info message")

// On shutdown
cancel()
<-hook.Done()
```

---

#### Output Behavior

- Maps Logrus levels to syslog severities automatically
- In access log mode, only the message is sent, with a newline
- Stack trace, timestamp, and trace fields can be included or filtered based on configuration
- Handles connection setup and reconnection transparently

---

#### Error Handling

- Returns errors for connection issues, closed streams, or syslog server errors
- All write operations are checked for errors and reported

---

#### Notes

- Designed for use with the main logger package and Logrus
- Integrates with the logger configuration system for unified setup
- All operations are thread-safe and suitable for production environments
- Supports both Unix syslog and Windows event log (with platform-specific behavior)
- Graceful shutdown ensures all logs are flushed before exit

---

### `logger/level` Subpackage Documentation

The `logger/level` subpackage defines log levels and provides utilities for parsing, converting, and integrating log levels with other logging systems such as Logrus.

---

#### Features

- Definition of standard log levels (Panic, Fatal, Error, Warn, Info, Debug, Nil)
- String and numeric conversion utilities
- Parsing from string to level
- Integration helpers for Logrus compatibility
- Listing of all available log levels

---

#### Main Types

##### `Level` Type

Represents the log level as a `uint8` type.

###### Constants

- `PanicLevel`: Critical error, triggers a panic (trace + fatal)
- `FatalLevel`: Fatal error, triggers process exit
- `ErrorLevel`: Error, process should stop and return to caller
- `WarnLevel`: Warning, process continues but an issue occurred
- `InfoLevel`: Informational message, no impact on process
- `DebugLevel`: Debug message, useful for troubleshooting
- `NilLevel`: Disables logging for this entry

---

#### Functions & Methods

##### `ListLevels() []string`

Returns a list of all available log level names as lowercase strings.

##### `Parse(l string) Level`

Parses a string and returns the corresponding `Level`. If the string does not match a known level, returns `InfoLevel`.

##### `Level.String() string`

Returns the string representation of the log level (e.g., "Debug", "Info", "Warning", "Error", "Fatal", "Critical").

##### `Level.Uint8() uint8`

Returns the numeric value of the log level.

##### `Level.Logrus() logrus.Level`

Converts the custom `Level` to the corresponding Logrus log level.

---

#### Example Usage

```go
import (
    "github.com/nabbar/golib/logger/level"
    "github.com/sirupsen/logrus"
)

lvl := level.Parse("debug")
if lvl == level.DebugLevel {
    // Enable debug logging
}

logrusLevel := lvl.Logrus()
logrus.SetLevel(logrusLevel)

for _, l := range level.ListLevels() {
    println(l)
}
```

---

#### Notes

- `NilLevel` disables logging and should not be used with `SetLogLevel`.
- String representations are case-insensitive when parsing.
- Designed for seamless integration with the main logger package and Logrus.

---

### `logger/types` Subpackage Documentation

The `logger/types` subpackage provides common types, constants, and interfaces used throughout the logger system. It defines standard field names for structured logging and the base interface for logger hooks, ensuring consistency and extensibility across all logger outputs.

---

#### Features

- Standardized field names for structured log entries
- Base `Hook` interface for implementing custom logrus hooks
- Integration with context and I/O interfaces
- Ensures compatibility and extensibility for logger outputs

---

#### Main Types

##### Field Name Constants

Defines string constants for common log entry fields:

- `FieldTime`: Timestamp of the log entry
- `FieldLevel`: Log level (e.g., info, error)
- `FieldStack`: Stack trace information
- `FieldCaller`: Caller function or method
- `FieldFile`: Source file name
- `FieldLine`: Source line number
- `FieldMessage`: Log message
- `FieldError`: Error details
- `FieldData`: Additional structured data

Use these constants to ensure consistent field naming in structured logs.

---

##### `Hook` Interface

Represents the base interface for logger hooks, designed for integration with Logrus and custom outputs.

- Inherits from `logrus.Hook` for log event handling
- Inherits from `io.WriteCloser` for I/O compatibility
- `RegisterHook(log *logrus.Logger)`: Register the hook with a Logrus logger
- `Run(ctx context.Context)`: Start the hook's background process (if needed)

This interface allows the creation of custom hooks that can be registered with the logger and manage their own lifecycle.

---

#### Example Usage

```go
import (
    "github.com/nabbar/golib/logger/types"
    "github.com/sirupsen/logrus"
    "context"
)

type MyCustomHook struct{}

func (h *MyCustomHook) Fire(entry *logrus.Entry) error { /* ... */ return nil }
func (h *MyCustomHook) Levels() []logrus.Level         { /* ... */ return nil }
func (h *MyCustomHook) Write(p []byte) (int, error)    { /* ... */ return 0, nil }
func (h *MyCustomHook) Close() error                   { /* ... */ return nil }
func (h *MyCustomHook) RegisterHook(log *logrus.Logger) { log.AddHook(h) }
func (h *MyCustomHook) Run(ctx context.Context)        { /* ... */ }

var hook types.Hook = &MyCustomHook{}
log := logrus.New()
hook.RegisterHook(log)
go hook.Run(context.Background())
```

---

#### Notes

- The field name constants should be used for all structured log entries to maintain consistency.
- The `Hook` interface is the foundation for all logger output hooks (stdout, stderr, file, syslog, etc.).
- Designed for use with the main logger package and its subpackages.
- All operations are thread-safe and suitable for concurrent environments.

---

### `logger/gorm` Subpackage Documentation

The `logger/gorm` subpackage provides an adapter to integrate the main logger system with the [GORM](https://gorm.io/) ORM logger interface. It enables centralized, structured, and configurable logging for all GORM database operations, supporting log level mapping, error handling, and slow query detection.

---

#### Features

- Implements the `gorm.io/gorm/logger.Interface` for seamless GORM integration
- Maps GORM log levels to the main logger's levels
- Structured logging with custom fields for SQL queries, rows, and elapsed time
- Configurable slow query threshold and error filtering
- Option to ignore "record not found" errors in logs
- Thread-safe and context-aware

---

#### Main Types

##### GORM Logger Adapter

- `New(fct func() Logger, ignoreRecordNotFoundError bool, slowThreshold time.Duration) gormlogger.Interface`  
  Creates a new GORM logger adapter.
    - `fct`: Function returning the main logger instance
    - `ignoreRecordNotFoundError`: If true, skips logging "record not found" errors
    - `slowThreshold`: Duration above which queries are considered slow and logged as warnings

---

#### Log Level Mapping

- `Silent`: Disables logging (`NilLevel`)
- `Info`: Logs as `InfoLevel`
- `Warn`: Logs as `WarnLevel`
- `Error`: Logs as `ErrorLevel`

---

#### Logging Methods

- `Info(ctx, msg, ...args)`: Logs informational messages
- `Warn(ctx, msg, ...args)`: Logs warnings
- `Error(ctx, msg, ...args)`: Logs errors
- `Trace(ctx, begin, fc, err)`: Logs SQL queries with execution time, rows affected, and error details
    - If the query is slow (exceeds `slowThreshold`), logs as a warning
    - If an error occurs (and is not ignored), logs as an error
    - Otherwise, logs as info

---

#### Example Usage

```go
import (
    "github.com/nabbar/golib/logger"
    "github.com/nabbar/golib/logger/gorm"
    "gorm.io/gorm"
    "time"
)

log := logger.New(/* context provider */)
gormLogger := gorm.New(
    func() logger.Logger { return log },
    true,                  // ignoreRecordNotFoundError
    200*time.Millisecond,  // slowThreshold
)

db, err := gorm.Open(/* ... */, &gorm.Config{
    Logger: gormLogger,
})
```

---

#### Output Behavior

- Each GORM operation is logged with structured fields:
    - `elapsed ms`: Query duration in milliseconds
    - `rows`: Number of rows affected (or "-" if unknown)
    - `query`: The executed SQL statement
- Errors and slow queries are highlighted according to configuration

---

#### Notes

- Designed for use with the main logger package for unified logging across your application
- Supports dynamic log level changes via the `LogMode` method
- All operations are safe for concurrent use and production environments

---

### `logger/hashicorp` Subpackage Documentation

The `logger/hashicorp` subpackage provides an adapter to integrate the main logger system with the [HashiCorp hclog](https://github.com/hashicorp/go-hclog) logging interface. This enables unified, structured, and configurable logging for libraries and tools that use hclog, with full support for log level mapping, context fields, and logger options.

---

#### Features

- Implements the `hclog.Logger` interface for seamless HashiCorp integration
- Maps hclog log levels to the main logger's levels
- Supports structured logging with custom fields and logger names
- Dynamic log level control and trace support
- Thread-safe and context-aware
- Provides standard logger and writer for compatibility

---

#### Main Types

##### HashiCorp Logger Adapter

- `New(logger FuncLog) hclog.Logger`  
  Creates a new hclog-compatible logger adapter.
    - `logger`: Function returning the main logger instance

- `SetDefault(log FuncLog)`  
  Sets the default hclog logger globally to use the adapter.

---

#### Log Level Mapping

- `NoLevel`, `Off`: Disables logging (`NilLevel`)
- `Trace`, `Debug`: Logs as `DebugLevel` (with trace support for `Trace`)
- `Info`: Logs as `InfoLevel`
- `Warn`: Logs as `WarnLevel`
- `Error`: Logs as `ErrorLevel`

---

#### Logging Methods

- `Log(level, msg, ...args)`: Generic log method for all levels
- `Trace(msg, ...args)`, `Debug(msg, ...args)`, `Info(msg, ...args)`, `Warn(msg, ...args)`, `Error(msg, ...args)`: Level-specific logging
- `IsTrace()`, `IsDebug()`, `IsInfo()`, `IsWarn()`, `IsError()`: Check if a level is enabled
- `With(args...)`: Returns a logger with additional context fields
- `Name()`, `Named(name)`, `ResetNamed(name)`: Manage logger names for context
- `SetLevel(level)`, `GetLevel()`: Set or get the current log level
- `ImpliedArgs()`: Returns the current context fields
- `StandardLogger(opts)`, `StandardWriter(opts)`: Provides standard `log.Logger` and `io.Writer` for compatibility

---

#### Example Usage

```go
import (
    "github.com/nabbar/golib/logger"
    "github.com/nabbar/golib/logger/hashicorp"
    "github.com/hashicorp/go-hclog"
    "context"
)

log := logger.New(func() context.Context { return context.Background() })
hclogger := hashicorp.New(func() logger.Logger { return log })

// Use hclogger as a drop-in replacement for hclog.Logger
hclogger.Info("Starting HashiCorp component", "component", "example")

// Set as the default hclog logger
hashicorp.SetDefault(func() logger.Logger { return log })
```

---

#### Output Behavior

- All hclog log messages are routed through the main logger, preserving structured fields and logger names.
- Log level and trace options are mapped according to the main logger configuration.
- Supports dynamic changes to log level and logger context.

---

#### Notes

- Designed for use with the main logger package for unified logging across your application and third-party libraries.
- All operations are safe for concurrent use and production environments.
- Supports full compatibility with the hclog API, including standard logger and writer methods.

---

## Notes

- Designed for Go 1.18+.
- All operations are thread-safe.
- Integrates with standard Go logging and third-party libraries.
- Suitable for high-concurrency and production environments.

For more details, refer to the GoDoc or the source code in the `logger` package and its subpackages.