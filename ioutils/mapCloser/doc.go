/***********************************************************************************************************************
 *
 *   MIT License
 *
 *   Copyright (c) 2025 Nicolas JUHEL
 *
 *   Permission is hereby granted, free of charge, to any person obtaining a copy
 *   of this software and associated documentation files (the "Software"), to deal
 *   in the Software without restriction, including without limitation the rights
 *   to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *   copies of the Software, and to permit persons to whom the Software is
 *   furnished to do so, subject to the following conditions:
 *
 *   The above copyright notice and this permission notice shall be included in all
 *   copies or substantial portions of the Software.
 *
 *   THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *   IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *   FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *   AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *   LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *   OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *   SOFTWARE.
 *
 *
 **********************************************************************************************************************/

/*
Package mapCloser provides a thread-safe, context-aware manager for multiple io.Closer instances.

# Design Philosophy

The mapCloser package is designed to simplify resource management in concurrent Go applications
by providing a centralized, context-aware mechanism to track and close multiple io.Closer instances.
It eliminates the need for manual cleanup loops and ensures resources are properly released when
a context is cancelled or when the application explicitly requests cleanup.

Key design principles:
  - Thread-safety through atomic operations (no mutexes)
  - Context-driven lifecycle management
  - Fail-safe: continues closing resources even if some fail
  - Minimal overhead with lock-free data structures

# Architecture

The package consists of two main components:

1. Closer Interface: The public API providing Add, Get, Clean, Clone, and Close operations.

2. closer Implementation: Internal structure using:
  - atomic.Bool for closed state tracking
  - atomic.Uint64 for counter management
  - github.com/nabbar/golib/context.Config for thread-safe storage
  - context cancellation function for lifecycle control

Data Flow:

	Context Created → New(ctx) → Background Goroutine Monitors Context
	                                   ↓
	                            Add closers dynamically
	                                   ↓
	                  Context Done OR Close() → Close All Resources

# Advantages

  - Zero mutexes: Uses only atomic operations for maximum performance
  - Automatic cleanup: Resources close when context is cancelled
  - Error aggregation: Reports all close errors, not just the first
  - Concurrent safe: All operations can be called from multiple goroutines
  - Clone support: Create independent copies for hierarchical resource management
  - Nil-safe: Gracefully handles nil closers without panicking

# Disadvantages & Limitations

  - Storage overhead: Uses github.com/nabbar/golib/context.Config which maintains a map internally
  - Counter-only Len(): The Len() method returns a counter, including nil values added
  - Background goroutine: A monitoring goroutine runs until the closer is closed
  - Overflow handling: Counter overflow (> math.MaxInt) returns 0 as documented
  - Post-close operations: Add/Get/Clean become no-ops after Close() is called
  - No selective removal: Clean() removes all closers; no per-item removal exists

# Performance Characteristics

  - Add: O(1) atomic increment + O(1) map store
  - Get: O(n) where n is the number of closers
  - Len: O(1) atomic load
  - Clean: O(1) reset operations
  - Close: O(n) where n is the number of closers

Memory usage scales linearly with the number of registered closers.

# Use Cases

Simple Case: Single-resource management with timeout
  - HTTP client with connection timeout
  - File operations with deadline
  - Database connection with TTL

Medium Case: Multiple related resources
  - Web server managing multiple file handles
  - Batch processing closing multiple readers/writers
  - Service cleanup coordinating several clients

Complex Case: Hierarchical resource management
  - Cloning closers for sub-contexts
  - Parent service managing child service resources
  - Transaction-scoped resource tracking

# Thread-Safety

All public methods are safe for concurrent use. The implementation uses atomic operations
exclusively, avoiding mutexes for performance. The internal context.Config storage is also
thread-safe by design.

# Error Handling

Close errors are aggregated and returned as a single formatted error containing all failure
messages separated by commas. The closer continues attempting to close all resources even
after encountering errors, ensuring best-effort cleanup.
*/
package mapCloser
