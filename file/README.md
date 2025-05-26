# `file` Package

The `file` package provides a set of utilities and abstractions for file management, including bandwidth throttling, permission handling, and progress tracking.
<br/>It is organized into several subpackages, each focusing on a specific aspect of file operations.

## Subpackages

- **[bandwidth subpackage](#bandwidth-subpackage)**
  <br/>Bandwidth throttling and rate limiting for file operations.
  <br/>This subpackage is ideal for applications that need to control file transfer rates, such as backup tools,
  <br/>file servers, or any scenario where bandwidth usage must be limited.
  <br />It integrates seamlessly with the progress tracking system.
  <br />It is designed for applications that require precise control over data transfer rates.
  <br /><br />

- **[perm subpackage](#perm-subpackage)**
  <br/>File permission management and utilities.
  <br/>This subpackage is ideal for applications needing robust, portable, and easily configurable file permission management.
  <br />It supports parsing, formatting, and encoding/decoding of file permissions across various formats.
  <br />It is designed for applications that require consistent permission handling across different platforms and configurations.
  <br /><br />

- **[progress subpackage](#progress-subpackage)**
  <br/>Progress tracking and reporting for file operations.
  <br/>This subpackage is ideal for applications needing detailed file operation tracking, custom progress reporting, or advanced file management.
  <br />It provides a unified interface for file I/O with integrated progress tracking, buffer management, and event hooks.
  <br />It is designed for applications that require real-time feedback on file operations, such as file transfer applications, backup tools,
  <br />or any application that needs to monitor file I/O progress.
  <br /><br />

---

## `bandwidth` Subpackage

The `bandwidth` subpackage provides utilities for bandwidth throttling and rate limiting during file operations. It is designed to integrate seamlessly with the progress tracking system, allowing you to control the data transfer rate (in bytes per second) for file reads and writes.

### Overview

- Allows you to set a bandwidth limit for file operations.
- Integrates with the progress tracking system to monitor and control data flow.
- Provides hooks for increment and reset events to enforce bandwidth constraints.
- Thread-safe implementation using atomic operations.

---

### Main Types & Interfaces

#### `BandWidth` Interface

Defines the main methods for bandwidth control:

```go
type BandWidth interface {
    RegisterIncrement(fpg libfpg.Progress, fi libfpg.FctIncrement)
    RegisterReset(fpg libfpg.Progress, fr libfpg.FctReset)
}
```

- `RegisterIncrement`: Registers a callback to be called on each progress increment (e.g., bytes read/written). The bandwidth limiter is applied here.
- `RegisterReset`: Registers a callback to be called when the progress is reset.

#### Constructor

Create a new bandwidth limiter by specifying the maximum bytes per second:

```go
import (
    "github.com/nabbar/golib/file/bandwidth"
    "github.com/nabbar/golib/size"
)

bw := bandwidth.New(size.Size(1024 * 1024)) // 1 MB/s limit
```

---

### Usage Example

```go
import (
    "github.com/nabbar/golib/file/bandwidth"
    "github.com/nabbar/golib/file/progress"
    "github.com/nabbar/golib/size"
)

bw := bandwidth.New(size.Size(512 * 1024)) // 512 KB/s

f, err := progress.Open("example.txt")
if err != nil {
    // handle error
}
defer f.Close()

bw.RegisterIncrement(f, nil)
bw.RegisterReset(f, nil)

// Now, all read/write operations on f will be bandwidth-limited
```

---

### Error Handling

- The subpackage does not return errors directly; it relies on the progress and file operation layers for error reporting.
- Always check errors from file and progress operations.

---

### Notes

- The bandwidth limiter works by measuring the time since the last operation and introducing a sleep if the data rate exceeds the configured limit.
- The implementation is thread-safe and can be used concurrently.
- Designed to be used in conjunction with the `progress` subpackage for seamless integration.

---

## `perm` Subpackage

The `perm` subpackage provides utilities for handling file permissions in a portable and user-friendly way. It offers parsing, formatting, conversion, and encoding/decoding of file permissions, making it easy to work with permissions across different formats and configuration systems.

### Overview

- Defines a `Perm` type based on `os.FileMode` for representing file permissions.
- Supports parsing from strings, integers, and byte slices.
- Provides conversion methods to various integer types and string representations.
- Implements encoding and decoding for JSON, YAML, TOML, CBOR, and text.
- Integrates with Viper for configuration loading via a decoder hook.

---

### Main Types & Functions

#### `Perm` Type

A type alias for `os.FileMode`:

```go
type Perm os.FileMode
```

#### Parsing Functions

- `Parse(s string) (Perm, error)`: Parse a permission string (octal, e.g., "0644").
- `ParseInt(i int) (Perm, error)`: Parse from an integer (interpreted as octal).
- `ParseInt64(i int64) (Perm, error)`: Parse from an int64 (octal).
- `ParseByte(p []byte) (Perm, error)`: Parse from a byte slice.

#### Conversion Methods

- `FileMode() os.FileMode`: Convert to `os.FileMode`.
- `String() string`: Return octal string representation (e.g., "0644").
- `Int64() int64`, `Int32() int32`, `Int() int`: Convert to signed integers.
- `Uint64() uint64`, `Uint32() uint32`, `Uint() uint`: Convert to unsigned integers.

#### Encoding/Decoding

Implements marshaling and unmarshaling for:

- JSON (`MarshalJSON`, `UnmarshalJSON`)
- YAML (`MarshalYAML`, `UnmarshalYAML`)
- TOML (`MarshalTOML`, `UnmarshalTOML`)
- CBOR (`MarshalCBOR`, `UnmarshalCBOR`)
- Text (`MarshalText`, `UnmarshalText`)

This allows seamless integration with configuration files and serialization formats.

#### Viper Integration

- `ViperDecoderHook()`: Returns a Viper decode hook for automatic permission parsing from configuration files.

---

### Usage Example

```go
import (
    "github.com/nabbar/golib/file/perm"
    "os"
)

p, err := perm.Parse("0755")
if err != nil {
    // handle error
}
file, err := os.OpenFile("example.txt", os.O_CREATE, p.FileMode())
if err != nil {
    // handle error
}
defer file.Close()
```

#### With Viper

```go
import (
    "github.com/nabbar/golib/file/perm"
    "github.com/spf13/viper"
)

v := viper.New()
v.Set("file_perm", "0644")
type Config struct {
    FilePerm perm.Perm `mapstructure:"file_perm"`
}
var cfg Config
v.Unmarshal(&cfg, viper.DecodeHook(perm.ViperDecoderHook()))
```

---

### Error Handling

- Parsing functions return standard Go `error` values.
- Invalid or out-of-range permissions result in descriptive errors.
- Always check errors when parsing or decoding permissions.

---

### Notes

- Permission strings must be in octal format (e.g., "0644", "0755").
- Handles overflow and invalid values gracefully.
- Designed for use with Go 1.18+ and compatible with common configuration systems.

---

## `progress` Subpackage

The `progress` subpackage provides advanced file I/O utilities with integrated progress tracking, buffer management, and event hooks. It wraps standard file operations and exposes interfaces for monitoring and controlling file read/write progress, making it ideal for applications that need to report or limit file operation progress.

### Overview

- Wraps standard file I/O with progress tracking.
- Supports registering callbacks for increment, reset, and EOF events.
- Allows buffer size customization for optimized I/O.
- Provides file management utilities (open, create, temp, unique, truncate, sync, stat, etc.).
- Implements the full set of `io.Reader`, `io.Writer`, `io.Seeker`, and related interfaces.

---

### Main Types & Interfaces

#### Interfaces

- **Progress**: Main interface combining file operations and progress tracking.
- **File**: File management operations (stat, truncate, sync, etc.).
- **GenericIO**: Embeds all standard Go I/O interfaces.

#### Key Functions

- `New(name string, flags int, perm os.FileMode) (Progress, error)`: Open or create a file with progress tracking.
- `Open(name string) (Progress, error)`: Open an existing file.
- `Create(name string) (Progress, error)`: Create a new file.
- `Temp(pattern string) (Progress, error)`: Create a temporary file.
- `Unique(basePath, pattern string) (Progress, error)`: Create a unique temporary file.

#### Progress Event Registration

- `RegisterFctIncrement(fct FctIncrement)`: Register a callback for each progress increment (e.g., bytes read/written).
- `RegisterFctReset(fct FctReset)`: Register a callback for progress reset events.
- `RegisterFctEOF(fct FctEOF)`: Register a callback for EOF events.
- `SetBufferSize(size int32)`: Set the buffer size for I/O operations.
- `SetRegisterProgress(f Progress)`: Copy registered callbacks to another Progress instance.

#### File Operations

- `Path() string`: Get the file path.
- `Stat() (os.FileInfo, error)`: Get file info.
- `SizeBOF() (int64, error)`: Get current offset.
- `SizeEOF() (int64, error)`: Get size from current offset to EOF.
- `Truncate(size int64) error`: Truncate the file.
- `Sync() error`: Sync file to disk.
- `Close() error`: Close the file.
- `CloseDelete() error`: Close and delete the file.

#### I/O Operations

Implements all standard I/O methods:
- `Read`, `ReadAt`, `ReadFrom`, `Write`, `WriteAt`, `WriteTo`, `WriteString`, `ReadByte`, `WriteByte`, `Seek`, etc.

---

### Usage Example

```go
import (
    "github.com/nabbar/golib/file/progress"
    "os"
)

f, err := progress.New("example.txt", os.O_RDWR|os.O_CREATE, 0644)
if err != nil {
    // handle error
}
defer f.Close()

f.RegisterFctIncrement(func(size int64) {
    // Called on each read/write increment
})

f.RegisterFctEOF(func() {
    // Called on EOF
})

buf := make([]byte, 1024)
n, err := f.Read(buf)
// ... use n, err
```

---

### Error Handling

- All errors are wrapped with custom error codes for precise diagnostics (e.g., `ErrorNilPointer`, `ErrorIOFileStat`, etc.).
- Always check returned errors from file and I/O operations.
- Error messages are descriptive and help identify the source of the problem.

---

### Notes

- The buffer size can be tuned for performance using `SetBufferSize`.
- Progress callbacks allow integration with UI, logging, or bandwidth throttling.
- Temporary and unique file creation is supported for safe file operations.
- Implements all standard file and I/O interfaces for drop-in replacement.
