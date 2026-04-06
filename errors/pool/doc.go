/*
 * MIT License
 *
 * Copyright (c) 2026 Nicolas JUHEL
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

// Package pool provides a high-performance, thread-safe mechanism for collecting and managing multiple errors.
//
// In modern Go applications, especially those dealing with concurrency (e.g., goroutine pools,
// parallel processing, fan-out patterns), it's common for multiple errors to occur simultaneously.
// The standard Go error handling often makes it cumbersome to collect and report all these errors
// in a structured, thread-safe manner. The 'pool' package addresses this by offering a robust
// solution to accumulate errors from various sources without introducing race conditions.
//
// # Core Purpose and Benefits
//
// The primary goal of this package is to simplify the aggregation of errors that arise
// from concurrent operations. It allows developers to:
//   - Collect errors from multiple goroutines safely and efficiently.
//   - Retrieve specific errors by an assigned index.
//   - Get a consolidated view of all errors that occurred.
//   - Avoid manual mutex management for error collection, reducing complexity and potential bugs.
//
// # Key Features Explained
//
//   - Thread-Safety: All methods exposed by the Pool interface are designed to be
//     safe for concurrent access by multiple goroutines. This is achieved internally
//     through the use of atomic operations and a specialized concurrent map,
//     eliminating the need for explicit locks (like `sync.Mutex`) by the user.
//
//   - Automatic Indexing: When errors are added using the `Add()` method, the pool
//     automatically assigns a unique, sequential `uint64` index to each non-nil error.
//     This index starts from 1 and increments with each successful addition, providing
//     a simple way to refer to errors by their insertion order.
//
//   - Manual Indexing (`Set`): For scenarios where errors are associated with specific
//     identifiers (e.g., a worker ID, a request ID, or a position in a dataset),
//     the `Set()` method allows you to explicitly store an error at a chosen `uint64` index.
//     This is useful for overwriting existing errors or placing errors at non-sequential positions.
//
//   - High-Water Mark (`MaxId`): The `MaxId()` method returns the highest `uint64` index
//     that has ever been used in the pool, either through `Add()` or `Set()`. This acts
//     as a "high-water mark" and is useful for understanding the range of indices that
//     have been occupied. It does not decrease if errors are deleted.
//
//   - Aggregation (`Error` and `Slice`):
//
//   - `Error()`: This method provides a single `error` interface that encapsulates
//     all errors currently in the pool. If the pool is empty, it returns `nil`.
//     If only one error exists, it returns that error directly. If multiple errors
//     are present, they are combined into a single hierarchical error using the
//     `liberr.UnknownError` mechanism from the parent `errors` package, allowing
//     you to treat a collection of errors as a single failure point.
//
//   - `Slice()`: Returns all non-nil errors in the pool as a `[]error`. It's important
//     to note that the order of errors in this slice is not guaranteed to be sequential
//     by index or insertion order, as it reflects the iteration order of the underlying
//     concurrent map.
//
//   - Deletion (`Del`): Errors can be removed from the pool using their index. This
//     operation is thread-safe and efficient.
//
//   - Clearing (`Clear`): The entire pool can be emptied, and the `MaxId` reset,
//     using the `Clear()` method. The internal sequence counter for `Add()` is
//     intentionally not reset, ensuring that subsequent `Add()` calls will continue
//     to generate unique indices that do not conflict with previously used ones.
//
// # Usage Scenarios and Examples
//
// This package shines in scenarios where errors need to be collected from parallel operations:
//
//  1. **Parallel Workers / Goroutine Pools:**
//     Imagine a scenario where you launch several goroutines to perform tasks, and each
//     might return an error. The `Pool` can collect all these errors.
//
//     ```go
//     package main
//
//     import (
//     "fmt"
//     "sync"
//     "time"
//
//     "github.com/nabbar/golib/errors/pool"
//     )
//
//     func worker(id int, p pool.Pool, wg *sync.WaitGroup) {
//     defer wg.Done()
//     time.Sleep(time.Duration(id) * 10 * time.Millisecond) // Simulate work
//     if id%2 != 0 { // Simulate error for odd workers
//     p.Add(fmt.Errorf("worker %d failed its task", id))
//     }
//     }
//
//     func main() {
//     p := pool.New()
//     var wg sync.WaitGroup
//
//     for i := 1; i <= 5; i++ {
//     wg.Add(1)
//     go worker(i, p, &wg)
//     }
//
//     wg.Wait() // Wait for all workers to complete
//
//     if p.Len() > 0 {
//     fmt.Println("--- Errors Collected ---")
//     // Get all errors as a single aggregated error
//     if aggregatedErr := p.Error(); aggregatedErr != nil {
//     fmt.Printf("Aggregated Error: %v\n", aggregatedErr)
//     }
//
//     // Or iterate through individual errors
//     fmt.Println("\n--- Individual Errors ---")
//     for _, e := range p.Slice() {
//     fmt.Printf("- %v\n", e)
//     }
//     } else {
//     fmt.Println("No errors occurred.")
//     }
//     }
//     ```
//
//  2. **Sequential Tasks with Error Accumulation:**
//     Even in sequential processing, you might want to continue execution despite
//     non-fatal errors and report all of them at the end.
//
//     ```go
//     package main
//
//     import (
//     "fmt"
//     "github.com/nabbar/golib/errors/pool"
//     )
//
//     func processStep(step int, p pool.Pool) {
//     if step == 2 {
//     p.Add(fmt.Errorf("error at step %d: input validation failed", step))
//     }
//     if step == 4 {
//     p.Add(fmt.Errorf("error at step %d: database write failed", step))
//     }
//     fmt.Printf("Step %d completed.\n", step)
//     }
//
//     func main() {
//     p := pool.New()
//
//     for i := 1; i <= 5; i++ {
//     processStep(i, p)
//     }
//
//     if p.Len() > 0 {
//     fmt.Println("\n--- Summary of Errors ---")
//     fmt.Printf("Total errors: %d\n", p.Len())
//     if err := p.Error(); err != nil {
//     fmt.Printf("Combined error report: %v\n", err)
//     }
//     } else {
//     fmt.Println("All steps completed successfully.")
//     }
//     }
//     ```
//
//  3. **High-Performance Error Logging / Buffering:**
//     In systems generating a high volume of errors, you might buffer them in a pool
//     before flushing them to a logging system in batches, or for later analysis.
//
// # Important Considerations
//
//   - Indexing: Indices generated by `Add()` start at 1. `Set()` allows any `uint64` index.
//     Be mindful that `Del()` does not decrement `MaxId` or reuse indices from `Add()`.
//   - Order of `Slice()`: The `Slice()` method returns errors in an arbitrary order due
//     to the nature of `sync.Map` iteration. If a specific order is required, you must
//     sort the returned slice manually (e.g., by error code or message).
//   - `Clear()` Behavior: `Clear()` resets the `MaxId` but not the internal sequence
//     counter for `Add()`. This means if you `Clear()` and then `Add()` more errors,
//     their indices will continue from where they left off before the `Clear()`,
//     guaranteeing unique indices across the lifetime of the `Pool` instance.
package pool
