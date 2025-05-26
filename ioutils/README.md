# `ioutils` Package Documentation

The `ioutils` package provides utility functions and abstractions for I/O operations in Go.
<br />It includes helpers for file and directory management, as well as interfaces and wrappers to simplify and extend standard I/O patterns.

---

## Features

- File and directory existence checks and creation with permissions
- I/O wrapper interface for custom or dynamic I/O implementations
- Extensible design for advanced I/O utilities

---

## Main Types & Functions

### PathCheckCreate

Checks if a file or directory exists at the given path, creates it if necessary, and sets the appropriate permissions.

**Signature:**
```go
func PathCheckCreate(isFile bool, path string, permFile os.FileMode, permDir os.FileMode) error
```

- `isFile`: `true` to check/create a file, `false` for a directory
- `path`: target path
- `permFile`: permissions for files
- `permDir`: permissions for directories

---


## Subpackages

This package includes several subpackages, each providing specific utilities for I/O operations:
- `bufferReadCloser`: Buffered I/O utilities with close support. See [`bufferReadCloser` documentation](#bufferreadcloser-subpackage-documentation) for details.
- `fileDescriptor`: Utilities for working with file descriptors and low-level file operations. See [`fileDescriptor` documentation](#filedescriptor-subpackage-documentation) for details.
- `ioprogress`: Progress tracking and reporting for I/O operations. See [`ioprogress` documentation](#ioprogress-subpackage-documentation) for details
- `iowrapper`: Provides an interface for wrapping I/O operations with custom logic. See [`iowrapper` documentation](#iowrapper-subpackage-documentation) for details.
- `mapCloser`: Manages multiple `io.Closer` instances with mapping and batch close support. See [`mapCloser` documentation](#mapcloser-subpackage-documentation) for details.
- `maxstdio`: Handles system limits and management for maximum open standard I/O descriptors. See [`maxstdio` documentation](#maxstdio-subpackage-documentation) for details.
- `multiplexer`: Multiplexes I/O streams for advanced routing or splitting of data. See [`multiplexer` documentation](#multiplexer-subpackage-documentation) for details.
- `nopwritecloser`: Implements a no-operation `io.WriteCloser` for testing or stubbing. See [`nopwritecloser` documentation](#nopwritecloser-subpackage-documentation) for details.

---

### `bufferReadCloser` Subpackage Documentation

The `bufferReadCloser` subpackage provides buffered I/O utilities that combine standard Go buffer types
<br />with the `io.Closer` interface. It offers convenient wrappers for `bytes.Buffer`, `bufio.Reader`,
<br />`bufio.Writer`, and `bufio.ReadWriter`, allowing for easy resource management and integration with custom close logic.

---

#### Features

- Buffered read, write, and read/write utilities with close support
- Compatible with standard Go I/O interfaces
- Optional custom close function for resource cleanup
- Reset and flush logic integrated with close operations

---

#### Main Types & Constructors

##### Buffer

A buffered read/write/close utility based on `bytes.Buffer`.

**Implements:**
- `io.Reader`, `io.ReaderFrom`, `io.ByteReader`, `io.RuneReader`
- `io.Writer`, `io.WriterTo`, `io.ByteWriter`, `io.StringWriter`
- `io.Closer`

**Constructor:**
```go
func NewBuffer(b *bytes.Buffer, fct FuncClose) Buffer
```
- `b`: underlying buffer
- `fct`: optional close function

---

##### Reader

A buffered reader with close support, based on `bufio.Reader`.

**Implements:**
- `io.Reader`, `io.WriterTo`, `io.Closer`

**Constructor:**
```go
func NewReader(b *bufio.Reader, fct FuncClose) Reader
```

---

##### Writer

A buffered writer with close support, based on `bufio.Writer`.

**Implements:**
- `io.Writer`, `io.StringWriter`, `io.ReaderFrom`, `io.Closer`

**Constructor:**
```go
func NewWriter(b *bufio.Writer, fct FuncClose) Writer
```

---

##### ReadWriter

A buffered read/write utility with close support, based on `bufio.ReadWriter`.

**Implements:**
- All methods of `Reader` and `Writer`

**Constructor:**
```go
func NewReadWriter(b *bufio.ReadWriter, fct FuncClose) ReadWriter
```

---

#### Example Usage

```go
import (
    "bytes"
    "github.com/nabbar/golib/ioutils/bufferReadCloser"
)

buf := bytes.NewBuffer(nil)
brc := bufferReadCloser.NewBuffer(buf, nil)

_, _ = brc.Write([]byte("hello"))
_ = brc.Close() // resets buffer and calls optional close function
```

---

#### Notes

- The `Close()` method resets or flushes the buffer and then calls the optional close function if provided.
- These types are useful for managing in-memory buffers with explicit resource cleanup, especially in complex I/O pipelines.
- All types are compatible with standard Go I/O interfaces for seamless integration.

---

### `fileDescriptor` Subpackage Documentation

The `fileDescriptor` subpackage provides utilities to query and manage the system's file descriptor limits for the current process.
<br />It is useful for applications that need to handle many open files or network connections and want to ensure the process is configured with appropriate resource limits.

---

#### Features

- Query the current and maximum file descriptor limits for the process.
- Increase the current file descriptor limit up to the system's maximum (platform-dependent).
- Cross-platform support (Linux/Unix and Windows).

---

#### Main Function

##### SystemFileDescriptor

Returns the current and maximum file descriptor limits, or sets a new limit if requested.

**Signature:**
```go
func SystemFileDescriptor(newValue int) (current int, max int, err error)
```

- `newValue`:
    - If `0`, only queries the current and maximum limits.
    - If greater than the current limit, attempts to increase the limit (up to the system maximum).
- Returns:
    - `current`: the current file descriptor limit for the process.
    - `max`: the maximum file descriptor limit allowed by the system.
    - `err`: error if the operation fails.

---

#### Example Usage

```go
import "github.com/nabbar/golib/ioutils/fileDescriptor"

cur, max, err := fileDescriptor.SystemFileDescriptor(0)
if err != nil {
    // handle error
}

// Try to increase the limit to 4096
cur, max, err = fileDescriptor.SystemFileDescriptor(4096)
if err != nil {
    // handle error
}
```

---

#### Platform Notes

- **Linux/Unix:** Uses `syscall.Getrlimit` and `syscall.Setrlimit` to manage `RLIMIT_NOFILE`.
- **Windows:** Uses system calls to manage the maximum number of open files (`maxstdio`), with hard limits defined by the OS.

---

#### Use Cases

- Ensuring your application can open enough files or sockets for high concurrency.
- Dynamically adjusting resource limits at startup based on workload requirements.

---

#### Notes

- Changing file descriptor limits may require appropriate system permissions.
- Always check the returned `err` to ensure the operation succeeded.
- Designed for Go 1.18+ and cross-platform compatibility.

---

### `ioprogress` Subpackage Documentation

The `ioprogress` subpackage provides utilities for tracking and reporting progress during I/O operations. 
<br />It wraps standard `io.ReadCloser` and `io.WriteCloser` interfaces, allowing developers to monitor
<br />the amount of data read or written, and to register custom callbacks for progress, reset, and end-of-file events.

---

#### Features

- Progress tracking for reading and writing operations
- Register custom callbacks for increment, reset, and EOF events
- Thread-safe progress counters using atomic operations
- Simple integration with existing I/O streams

---

#### Main Interfaces

##### Progress

Defines methods to register callbacks and reset progress.

```go
type Progress interface {
    RegisterFctIncrement(fct FctIncrement)
    RegisterFctReset(fct FctReset)
    RegisterFctEOF(fct FctEOF)
    Reset(max int64)
}
```

##### Reader

Combines `io.ReadCloser` and `Progress` for read operations with progress tracking.

```go
type Reader interface {
    io.ReadCloser
    Progress
}
```

##### Writer

Combines `io.WriteCloser` and `Progress` for write operations with progress tracking.

```go
type Writer interface {
    io.WriteCloser
    Progress
}
```

---

#### Constructors

##### NewReadCloser

Wraps an `io.ReadCloser` to provide progress tracking.

```go
func NewReadCloser(r io.ReadCloser) Reader
```

##### NewWriteCloser

Wraps an `io.WriteCloser` to provide progress tracking.

```go
func NewWriteCloser(w io.WriteCloser) Writer
```

---

#### Example Usage

```go
import (
    "os"
    "github.com/nabbar/golib/ioutils/ioprogress"
)

file, _ := os.Open("data.txt")
reader := ioprogress.NewReadCloser(file)

reader.RegisterFctIncrement(func(size int64) {
    // Called after each read with the number of bytes read
})

reader.RegisterFctEOF(func() {
    // Called when EOF is reached
})

buf := make([]byte, 1024)
for {
    n, err := reader.Read(buf)
    if err != nil {
        break
    }
    // process buf[:n]
}
_ = reader.Close()
```

---

#### Notes

- Callbacks for increment, reset, and EOF can be registered at any time.
- The `Reset` method allows resetting the progress counter and optionally triggering a reset callback.
- Designed for Go 1.18+ and thread-safe usage.
- Useful for monitoring file transfers, network streams, or any I/O operation where progress feedback is needed.

---

### `iowrapper` Subpackage Documentation

The `iowrapper` subpackage provides a flexible interface to wrap and extend standard Go I/O operations (`io.Reader`, `io.Writer`, `io.Seeker`, `io.Closer`).
<br />It allows developers to inject custom logic for reading, writing, seeking, and closing, making it easy to adapt or mock I/O behaviors.

---

#### Features

- Wraps any I/O-compatible object with custom read, write, seek, and close functions
- Implements standard Go I/O interfaces for seamless integration
- Dynamic assignment of custom handlers at runtime
- Useful for testing, instrumentation, or adapting legacy I/O

---

#### Main Types & Functions

##### IOWrapper Interface

Defines a wrapper for I/O operations with methods to set custom logic.

```go
type IOWrapper interface {
    io.Reader
    io.Writer
    io.Seeker
    io.Closer

    SetRead(read FuncRead)
    SetWrite(write FuncWrite)
    SetSeek(seek FuncSeek)
    SetClose(close FuncClose)
}
```

##### Function Types

- `FuncRead`: `func(p []byte) []byte`
- `FuncWrite`: `func(p []byte) []byte`
- `FuncSeek`: `func(offset int64, whence int) (int64, error)`
- `FuncClose`: `func() error`

---

##### Constructor

Creates a new IOWrapper for any I/O-compatible object.

```go
func New(in any) IOWrapper
```
- `in`: any object implementing one or more standard I/O interfaces

---

#### Example Usage

```go
import (
    "os"
    "github.com/nabbar/golib/ioutils/iowrapper"
)

file, _ := os.Open("data.txt")
w := iowrapper.New(file)

// Set a custom read function (e.g., for logging or transformation)
w.SetRead(func(p []byte) []byte {
    // custom logic here
    return p
})

buf := make([]byte, 128)
_, _ = w.Read(buf)
_ = w.Close()
```

---

#### Notes

- If no custom function is set, the wrapper delegates to the underlying object's standard methods.
- Setting a function to `nil` restores the default behavior.
- Useful for testing, instrumentation, or adapting I/O flows without modifying the original implementation.

---

### `mapCloser` Subpackage Documentation

The `mapCloser` subpackage provides a utility to manage multiple `io.Closer` instances as a group, allowing for batch addition, retrieval, cloning, and closing of resources.
<br />It is designed for robust resource management in concurrent or context-driven applications.

---

##### Features

- Add and manage multiple `io.Closer` objects
- Retrieve all managed closers
- Clean and reset the internal state
- Clone the current set of closers
- Batch close all resources, collecting errors if any
- Context-aware: automatically closes resources when the context is cancelled
- Thread-safe operations

---

##### Main Interface

```go
type Closer interface {
    Add(clo ...io.Closer)
    Get() []io.Closer
    Len() int
    Clean()
    Clone() Closer
    Close() error
}
```

---

##### Constructor

Creates a new `Closer` instance bound to a context.

```go
func New(ctx context.Context) Closer
```
- `ctx`: The context to monitor for cancellation. When cancelled, all managed closers are closed automatically.

---

##### Example Usage

```go
import (
    "context"
    "os"
    "github.com/nabbar/golib/ioutils/mapCloser"
)

ctx := context.Background()
mc := mapCloser.New(ctx)

file1, _ := os.Open("file1.txt")
file2, _ := os.Open("file2.txt")

mc.Add(file1, file2)

// Retrieve all closers
closers := mc.Get()

// Close all resources
err := mc.Close()
```

---

##### Notes

- The `Close()` method attempts to close all managed resources and returns a combined error if any close operations fail.
- The `Clone()` method creates a copy of the current `Closer` with the same set of managed resources.
- The internal state is thread-safe and suitable for concurrent use.
- Automatically handles context cancellation for safe resource cleanup.

---

##### Use Cases

- Managing multiple files, network connections, or other closable resources in a batch.
- Ensuring all resources are properly closed on application shutdown or context cancellation.
- Simplifying resource management in complex workflows.

---

### `maxstdio` Subpackage Documentation

The `maxstdio` subpackage provides utilities to get and set the maximum number of standard I/O file descriptors (stdio) that a process can open on Windows systems.
<br />It uses cgo to call the underlying C runtime functions for managing this limit.

---

##### Features

- Query the current maximum stdio limit for the process.
- Set a new maximum stdio limit (up to the system hard limit).
- Direct integration with the Windows C runtime via cgo.

---

##### Main Functions

###### GetMaxStdio

Returns the current maximum number of stdio file descriptors allowed for the process.

```go
func GetMaxStdio() int
```

- Returns the current stdio limit as an integer.

---

###### SetMaxStdio

Sets a new maximum number of stdio file descriptors for the process.

```go
func SetMaxStdio(newMax int) int
```

- `newMax`: The desired new maximum value.
- Returns the updated stdio limit as an integer.

---

##### Example Usage

```go
import "github.com/nabbar/golib/ioutils/maxstdio"

cur := maxstdio.GetMaxStdio()

newLimit := 2048
updated := maxstdio.SetMaxStdio(newLimit)
```

---

##### Notes

- These functions are only available on Windows with cgo enabled.
- The actual hard limit is defined by the Windows OS and may not be exceeded.
- Useful for applications that need to handle many open files or sockets simultaneously.

---

##### Use Cases

- Increasing the stdio limit for high-concurrency servers or applications.
- Querying the current stdio limit for diagnostics or configuration validation.
- Ensuring resource limits are sufficient for the application's workload.

---

### `multiplexer` Subpackage Documentation

The `multiplexer` subpackage provides a generic and thread-safe way to multiplex and demultiplex I/O streams, 
<br />allowing you to route messages to different logical streams identified by a comparable key. 
<br />It is useful for advanced I/O routing, such as handling multiple output streams over a single connection.

---

#### Features

- Multiplexes multiple logical streams over a single I/O channel
- Supports any comparable type as a stream key (e.g., `int`, `string`)
- Thread-safe management of stream handlers
- Easy integration with standard Go `io.Reader` and `io.Writer` interfaces
- Uses CBOR encoding for efficient message framing

---

#### Main Types & Interfaces

##### MixStdOutErr

A generic interface for multiplexed I/O operations.

```go
type MixStdOutErr[T comparable] interface {
    io.Reader

    Writer(key T) io.Writer
    Add(key T, fct FuncWrite)
}
```

- `Writer(key T) io.Writer`: Returns a writer for the given stream key.
- `Add(key T, fct FuncWrite)`: Registers a write handler for a specific stream key.

##### FuncWrite

Function signature for custom write handlers.

```go
type FuncWrite func(p []byte) (n int, err error)
```

##### Message

Represents a multiplexed message with a stream key and payload.

```go
type Message[T comparable] struct {
    Stream  T
    Message []byte
}
```

---

#### Constructor

##### New

Creates a new multiplexer instance.

```go
func New[T comparable](r io.Reader, w io.Writer) MixStdOutErr[T]
```

- `r`: The underlying reader (for demultiplexing incoming messages)
- `w`: The underlying writer (for multiplexing outgoing messages)

---

#### Example Usage

```go
import (
    "os"
    "github.com/nabbar/golib/ioutils/multiplexer"
)

mux := multiplexer.New[string](os.Stdin, os.Stdout)

// Register a handler for a stream key
mux.Add("stdout", func(p []byte) (int, error) {
    // handle output for "stdout"
    return len(p), nil
})

// Write to a specific stream
writer := mux.Writer("stdout")
_, _ = writer.Write([]byte("Hello, multiplexed world!"))

// Read and dispatch messages
buf := make([]byte, 1024)
_, err := mux.Read(buf)
```

---

#### Notes

- Each message is encoded using CBOR, containing both the stream key and the message payload.
- The `Add` method allows you to register custom handlers for each logical stream.
- The `Writer` method provides a standard `io.Writer` for sending data to a specific stream.
- Reading from the multiplexer will decode messages and dispatch them to the appropriate handler based on the stream key.
- Suitable for scenarios where you need to route or split data between multiple logical channels over a single physical connection.

---

#### Use Cases

- Multiplexing stdout and stderr over a single network connection
- Routing logs or messages to different consumers based on type or channel
- Building advanced I/O pipelines with dynamic stream management

---

### `nopwritecloser` Subpackage Documentation

The `nopwritecloser` subpackage provides a simple utility that wraps any `io.Writer` to implement the `io.WriteCloser` interface, where the `Close()` method is a no-op. 
<br /> This is useful for cases where an `io.WriteCloser` is required but no actual resource needs to be closed, such as in testing or when working with in-memory buffers.

This subpackage is similar to the standard `io.NopCloser`, but specifically designed to wrap `io.Writer` types while providing a no-operation `Close()` method.

---

##### Features

- Wraps any `io.Writer` to provide a no-operation `Close()` method
- Fully compatible with the standard Go `io.WriteCloser` interface
- Useful for testing, stubbing, or adapting APIs that require a closer

---

##### Main Function

###### New

Creates a new `io.WriteCloser` from any `io.Writer`. The `Close()` method does nothing and always returns `nil`.

```go
func New(w io.Writer) io.WriteCloser
```

- `w`: The underlying writer to wrap

---

##### Example Usage

```go
import (
    "bytes"
    "github.com/nabbar/golib/ioutils/nopwritecloser"
)

buf := &bytes.Buffer{}
wc := nopwritecloser.New(buf)

_, _ = wc.Write([]byte("example"))
_ = wc.Close() // does nothing, always returns nil
```

---

##### Notes

- The wrapped writer is not closed or affected by the `Close()` call.
- This utility is ideal for adapting APIs that expect an `io.WriteCloser` but where closing is unnecessary or undesired.
- Designed for Go 1.18+ and compatible with all standard `io.Writer` implementations.

---

## Notes

- All utilities are designed for Go 1.18+.
- Thread-safe where applicable.
- Integrates with standard Go `io` interfaces for maximum compatibility.

---

For more details, refer to the GoDoc or the source code in the `ioutils` package and its subpackages.