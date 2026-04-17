/*
 * MIT License
 *
 * Copyright (c) 2024 Nicolas JUHEL
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
 */

/*
Package atomic implements advanced, type-safe atomic primitives designed for high-performance concurrent applications.

The core philosophy of this package is to provide the safety and convenience of generics while maintaining the
extreme performance characteristics of the native sync/atomic package. It addresses common pain points in Go's
concurrency model, such as mandatory type assertions for atomic.Value and the risk of storing or retrieving
unexpected zero-values in shared state.

# ARCHITECTURE & INTERNALS

The package is built around three main components: Value[T], Map[K], MapTyped[K, V], and the Cast[T] utility.

1. High-Performance Value Container (Value[T])

The Value[T] implementation uses a tiered access model to minimize overhead:

Data Flow (Nominal Case - No Defaults):
User -> Store(T) -> atomic.Value.Store(any)
User <- Load() <- atomic.Value.Load() <- (Type Assertion) -> User

Data Flow (With Defaults Configured):
User -> Store(val T)
|
+-> Is val empty/zero? (using Cast[T])
|
+-- Yes --> atomic.Value.Store(DefaultStoreValue)
|
+-- No ---> atomic.Value.Store(val)

Internal Performance Optimization (Short-Circuiting):
To avoid the overhead of zero-value detection (which may require reflection for complex types),
the structure maintains internal atomic flags (bl, bs). If no default value is set, the logic
bypasses the entire Cast/IsEmpty subsystem, resulting in performance identical to native code.

2. Specialized Atomic Maps (MapAny and MapTyped)

These components wrap sync.Map to provide type safety without the need for external mutexes.
They implement a "self-healing" mechanism during iteration: if the Range function encounters
an entry whose key or value has been corrupted (e.g. by an external agent injecting an invalid type
directly into the underlying sync.Map), the entry is automatically evicted to ensure the integrity
of the typed view.

3. Zero-Value Detection Engine (Cast & IsEmpty)

The Cast function provides a highly optimized alternative to reflect.DeepEqual for detecting
uninitialized state. It uses a type switch for all Go scalar types (integers, floats, strings, bools)
to perform direct comparisons. For structs and complex types, it falls back to reflect.Value.IsZero(),
which is significantly faster than comparing two full objects.

# QUICK START

Basic type-safe atomic value:
    v := atomic.NewValue[string]()
    v.Store("hello")
    msg := v.Load() // "hello", no type assertion needed

Atomic value with default "safe" state:
    v := atomic.NewValueDefault[int](10, 20)
    fmt.Println(v.Load()) // Prints 10 (load default)
    v.Store(0)            // Triggers store default (20)
    fmt.Println(v.Load()) // Prints 20

Type-safe concurrent map:
    m := atomic.NewMapTyped[string, int]()
    m.Store("key", 42)
    val, ok := m.Load("key") // (42, true)

# USE CASES

1. Lifecycle Management:
Storing a context.Context or a CancelFunc in a structure where you want to ensure that
a call to Load() never returns nil, but instead returns a Background context or a no-op function.

2. Configuration Hot-Reloading:
Safely swapping complex configuration structs between goroutines while ensuring that any
invalid (empty) configuration received from a source is automatically replaced by a
known-good default.

3. High-Concurrency Counters & Metrics:
Using MapTyped to track per-user or per-service metrics without the performance penalty
of a global mutex or the verbosity of sync.Map type assertions.

# TECHNICAL CONSTRAINTS & CONTRACTS

- Immutability: Users should avoid modifying objects after they have been stored in an atomic container.
- Writer Contract: Functions stored as values must not retain references to provided buffers to prevent data races.
- Performance: To achieve nanosecond-level latency, prefer NewValue() over NewValueDefault()
  unless the default value logic is strictly required by the business logic.
*/
package atomic
