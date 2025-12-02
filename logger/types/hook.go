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

package types

import (
	"context"
	"io"

	"github.com/sirupsen/logrus"
)

// Hook defines an extended logger hook interface for advanced log processing.
//
// This interface extends logrus.Hook with lifecycle management, I/O capabilities,
// and background processing support. It allows for sophisticated log handlers that
// can intercept log entries, write to multiple destinations, and run in background
// goroutines with graceful shutdown.
//
// Interface composition:
//   - logrus.Hook: Provides Fire(entry) and Levels() methods for log interception
//   - io.WriteCloser: Provides Write(p) and Close() methods for direct I/O
//
// Additional methods provide lifecycle management:
//   - RegisterHook: Self-registration with logger instances
//   - Run: Background execution with context-based cancellation
//   - IsRunning: State checking for monitoring and coordination
//
// Implementation requirements:
//
// Fire() method (from logrus.Hook):
//   - Called synchronously for every log entry matching Levels()
//   - MUST return quickly to avoid blocking all logging
//   - Should offload heavy processing to Run() goroutine via channels
//   - Returning an error logs the error but doesn't stop other hooks
//
// Levels() method (from logrus.Hook):
//   - Returns slice of log levels this hook processes
//   - Return logrus.AllLevels to process all levels
//   - Filter by level to reduce Fire() call overhead
//
// Write() method (from io.Writer):
//   - Allows direct writing to the hook bypassing logrus
//   - Useful for external log sources or raw data injection
//   - Should handle concurrent calls safely
//
// Close() method (from io.Closer):
//   - Called during cleanup to release resources
//   - Should be idempotent (safe to call multiple times)
//   - Should wait for in-flight operations to complete
//
// RegisterHook() method:
//   - Called once during initialization
//   - Should call log.AddHook(h) to register with logrus
//   - May perform additional initialization
//
// Run() method:
//   - Runs in background goroutine until context is cancelled
//   - Use for heavy processing: buffering, batching, network I/O
//   - Should respect ctx.Done() for graceful shutdown
//   - Typically receives work from Fire() via channels
//
// IsRunning() method:
//   - Returns true if Run() goroutine is active
//   - Used for status monitoring and coordination
//   - Should be safe for concurrent calls
//
// Thread safety:
//   - Fire() may be called concurrently from multiple goroutines
//   - All methods should handle concurrent access safely
//   - Use sync.Mutex or atomic operations for shared state
//
// Example implementation:
//
//	type MyHook struct {
//	    queue   chan *logrus.Entry
//	    running atomic.Bool
//	    mu      sync.Mutex
//	}
//
//	func (h *MyHook) Fire(entry *logrus.Entry) error {
//	    select {
//	    case h.queue <- entry:
//	        return nil
//	    default:
//	        return errors.New("queue full")
//	    }
//	}
//
//	func (h *MyHook) Levels() []logrus.Level {
//	    return logrus.AllLevels
//	}
//
//	func (h *MyHook) RegisterHook(log *logrus.Logger) {
//	    log.AddHook(h)
//	}
//
//	func (h *MyHook) Run(ctx context.Context) {
//	    h.running.Store(true)
//	    defer h.running.Store(false)
//	    for {
//	        select {
//	        case entry := <-h.queue:
//	            h.processEntry(entry)
//	        case <-ctx.Done():
//	            return
//	        }
//	    }
//	}
//
//	func (h *MyHook) IsRunning() bool {
//	    return h.running.Load()
//	}
//
//	func (h *MyHook) Write(p []byte) (n int, err error) {
//	    return len(p), nil
//	}
//
//	func (h *MyHook) Close() error {
//	    return nil
//	}
type Hook interface {
	logrus.Hook
	io.WriteCloser

	// RegisterHook registers this hook with the given logger instance.
	//
	// This method should call log.AddHook(h) to integrate the hook into the
	// logger's processing pipeline. It may also perform additional initialization
	// such as creating output files, establishing network connections, or
	// allocating buffers.
	//
	// This method does not return an error as the hook is responsible for
	// handling any initialization errors internally (e.g., by logging them
	// or storing them for later retrieval).
	//
	// The method should be called once during application startup before any
	// logging occurs. Calling it multiple times with different loggers will
	// register the same hook instance with multiple loggers.
	//
	// Parameters:
	//   - log: The logrus.Logger instance to register with
	//
	// Thread safety:
	//   - Safe to call concurrently with different logger instances
	//   - Not safe to call concurrently with the same logger instance
	RegisterHook(log *logrus.Logger)

	// Run executes the hook's background processing loop until the context is cancelled.
	//
	// This method should be called in a goroutine to enable background processing
	// of log entries without blocking the main application. It typically receives
	// work from Fire() via buffered channels and performs heavy operations like:
	//   - Batching log entries before writing
	//   - Writing to slow destinations (network, disk)
	//   - Formatting or transforming log data
	//   - Aggregating metrics from log entries
	//
	// The method must respect context cancellation and return promptly when
	// ctx.Done() is signalled. This allows for graceful shutdown where the hook
	// can flush pending entries and close resources.
	//
	// Implementation pattern:
	//
	//	func (h *Hook) Run(ctx context.Context) {
	//	    for {
	//	        select {
	//	        case work := <-h.workQueue:
	//	            h.process(work)
	//	        case <-ctx.Done():
	//	            h.flush() // Process remaining work
	//	            return
	//	        }
	//	    }
	//	}
	//
	// Parameters:
	//   - ctx: Context for cancellation and deadline control
	//
	// Thread safety:
	//   - Should only be called once per hook instance
	//   - Typically called with "go hook.Run(ctx)" pattern
	Run(ctx context.Context)

	// IsRunning returns whether the hook's Run() method is currently executing.
	//
	// This method provides a way to check the operational state of the hook,
	// useful for monitoring, status reporting, and coordination with other
	// components. It should return true from when Run() starts until it returns.
	//
	// Typical implementation uses an atomic.Bool or similar mechanism:
	//
	//	func (h *Hook) Run(ctx context.Context) {
	//	    h.running.Store(true)
	//	    defer h.running.Store(false)
	//	    // ... processing loop ...
	//	}
	//
	//	func (h *Hook) IsRunning() bool {
	//	    return h.running.Load()
	//	}
	//
	// Returns:
	//   - true if Run() is currently executing
	//   - false if Run() has not been started or has already returned
	//
	// Thread safety:
	//   - Must be safe for concurrent calls from multiple goroutines
	//   - Must provide consistent state without race conditions
	IsRunning() bool
}
