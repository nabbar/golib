## `context` Package

The `context` package extends Go's standard `context.Context` with advanced configuration and management features. It provides generic, thread-safe context storage, context cloning, merging, and key-value management, making it easier to handle complex application state and configuration.

### Features

- Generic context configuration with type-safe keys
- Thread-safe key-value storage and retrieval
- Context cloning and merging
- Walk and filter stored values
- Integration with Go's `context.Context` interface

### Main Types & Functions

- **Config\[T comparable\]**: Generic interface for context configuration and key-value management.
- **MapManage\[T\]**: Interface for map operations (load, store, delete, clean).
- **FuncContext**: Function type returning a `context.Context`.
- **NewConfig**: Create a new generic context configuration.

#### Example Usage

```go
import "github.com/nabbar/golib/context"

type MyKey string

cfg := context.NewConfig[MyKey](nil)
cfg.Store("myKey", "myValue")
val, ok := cfg.Load("myKey")
```

#### Key Methods

- `SetContext(ctx FuncContext)`
- `Clone(ctx context.Context) Config[T]`
- `Merge(cfg Config[T]) bool`
- `Walk(fct FuncWalk[T]) bool`
- `LoadOrStore(key T, cfg interface{}) (val interface{}, loaded bool)`
- `Clean()`

#### IsolateParent

Creates a new context with cancelation, isolated from the parent.

```go
ctx := context.IsolateParent(parentCtx)
```

---

## `context/gin` Subpackage

The `context/gin` subpackage provides a bridge between Go's context and the [Gin](https://github.com/gin-gonic/gin) web framework. It wraps Gin's `Context` to add context management, signal handling, and metadata utilities for HTTP request handling.

### Features

- Wraps Gin's `Context` with Go's `context.Context`
- Signal-based cancellation (e.g., on OS signals)
- Key-value storage and retrieval for request-scoped data
- Type-safe getters for common types (string, int, bool, etc.)
- Integrated logging support

### Main Types & Functions

- **GinTonic**: Interface combining `context.Context` and Gin's `Context` with extra helpers.
- **New**: Create a new `GinTonic` context from a Gin `Context` and logger.

#### Example Usage

```go
import (
    "github.com/nabbar/golib/context/gin"
    "github.com/gin-gonic/gin"
    "github.com/nabbar/golib/logger"
)

func handler(c *gin.Context) {
    ctx := gin.New(c, logger.New(nil))
    ctx.Set("userID", 123)
    id := ctx.GetInt("userID")
    // ...
}
```

#### Key Methods

- `Set(key string, value interface{})`
- `Get(key string) (value interface{}, exists bool)`
- `GetString(key string) string`
- `CancelOnSignal(s ...os.Signal)`
- `SetLogger(log liblog.FuncLog)`
