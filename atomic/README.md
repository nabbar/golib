# golib/atomic

This package provides generic, thread-safe atomic values and maps for Go, making it easier to work with concurrent data structures. It offers atomic value containers, atomic maps (with both `any` and typed values), and utility functions for safe casting and default value management.

## Features

- **Generic atomic values**: Store, load, swap, and compare-and-swap any type safely.
- **Atomic maps**: Thread-safe maps with generic or typed values.
- **Default value management**: Set default values for atomic operations.
- **Safe type casting utilities**.

## Installation

Add to your `go.mod`:

```
require github.com/nabbar/golib/atomic vX.Y.Z
```

## Usage

### Atomic Value

Create and use an atomic value for any type:

```go
import "github.com/nabbar/golib/atomic"

type MyStruct struct {
    Field1 string
    Field2 int
}

val := atomic.NewValue[MyStruct]()

v1 := MyStruct{
Field1: "Hello",
Field2: 42,
}

val.Store(v1)

v1 = val.Load()
fmt.Println(v1.Field1) // Output: Hello
fmt.Println(v1.Field2) // Output: 42

v2 := MyStruct{
    Field1: "World",
    Field2: 100,
}

swapped := val.CompareAndSwap(v1, v2) // swapped == true

old := val.Swap(MyStruct{
    Field1: "New",
    Field2: 200,
}) // old == v2

fmt.Println(old.Field1) // Output: World
fmt.Println(old.Field2) // Output: 100

```

Set default values for load/store:

```go
val.SetDefaultLoad(0)
val.SetDefaultStore(-1)
val.Store(0) // Will store -1 instead of 0
v := val.Load() // If empty, returns 0
```

### Atomic Map (any value)

```go
m := atomic.NewMapAny[string]()
m.Store("foo", 123)
v, ok := m.Load("foo") // v == 123, ok == true
m.Delete("foo")
```

### Atomic Map (typed value)

```go
mt := atomic.NewMapTyped[string, int]()
mt.Store("bar", 456)
v, ok := mt.Load("bar") // v == 456, ok == true
mt.Delete("bar")
```

### Range Over Map

```go
mt.Range(func(key string, value int) bool {
    fmt.Printf("%s: %d\n", key, value)
    return true // continue iteration
})
```

### Safe Casting

```go
v, ok := atomic.Cast[int](anyValue)
if ok {
    // v is of type int
}
empty := atomic.IsEmpty[string](anyValue)
```

## Interfaces

### Atomic Value

```go
type Value[T any] interface {
    SetDefaultLoad(def T)
    SetDefaultStore(def T)
    Load() (val T)
    Store(val T)
    Swap(new T) (old T)
    CompareAndSwap(old, new T) (swapped bool)
}
```

### Atomic Map

```go
// Map is a generic interface for atomic maps with any value type but typed key.
type Map[K comparable] interface {
    Load(key K) (value any, ok bool)
    Store(key K, value any)
    
    LoadOrStore(key K, value any) (actual any, loaded bool)
    LoadAndDelete(key K) (value any, loaded bool)
    
    Delete(key K)
    Swap(key K, value any) (previous any, loaded bool)
    
    CompareAndSwap(key K, old, new any) bool
    CompareAndDelete(key K, old any) (deleted bool)
    
    Range(f func(key K, value any) bool)
}

// MapTyped is a specialized version of Map for typed values in add of type key from Map.
type MapTyped[K comparable, V any] interface {
    Load(key K) (value V, ok bool)
    Store(key K, value V)
    
    LoadOrStore(key K, value V) (actual V, loaded bool)
    LoadAndDelete(key K) (value V, loaded bool)
    
    Delete(key K)
    Swap(key K, value V) (previous V, loaded bool)
    
    CompareAndSwap(key K, old, new V) bool
    CompareAndDelete(key K, old V) (deleted bool)
    
    Range(f func(key K, value V) bool)
}

```

See `atomic/interface.go` for full details.

## Error Handling

All operations are safe for concurrent use. Type assertions may fail if you use the wrong type; always check the returned boolean.

## License

MIT © Nicolas JUHEL

Generated With © Github Copilot.
