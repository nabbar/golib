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
 */

package aggregator

// SetLoggerError registers a custom error logging function for the aggregator instance.
//
// Technical Implementation:
// The logger is stored in a libatm.Value (atomic container), which ensures that updates
// to the logging function are thread-safe and immediately visible to all active
// goroutines without requiring global mutex locks. This is critical as the logger
// may be invoked frequently from the high-throughput 'run' loop.
//
// Operational Behavior:
//   - If 'f' is nil, the aggregator internally assigns a no-op function to prevent
//     nil pointer dereferences during error reporting.
//   - The provided function will be called for internal operational errors, including
//     writer failures, context timeouts, and recovered panics in callbacks.
//
// Parameters:
//   - f: The error logging function. It receives a descriptive message and an optional
//     slice of error objects for detailed diagnostic reporting.
//
// Example:
//
//	agg.SetLoggerError(func(msg string, err ...error) {
//	    log.Errorf("Aggregator Failure: %s | Details: %v", msg, err)
//	})
func (o *agg) SetLoggerError(f func(msg string, err ...error)) {
	// Optimization: Avoid redundant storage if both the current and new values are nil.
	if o.le.Load() == nil && f == nil {
		return
	}

	// Normalization: Ensure a non-nil function is always available for the hot path.
	if f == nil {
		o.le.Store(func(msg string, err ...error) {})
		return
	}

	o.le.Store(f)
}

// SetLoggerInfo registers a custom informational logging function for the aggregator.
//
// Technical Implementation:
// Similar to SetLoggerError, this method utilizes atomic storage for thread-safe
// updates at runtime. Informational logs are typically emitted during lifecycle
// transitions (Start, Stop, Restart) and major operational shifts.
//
// Operational Behavior:
//   - If 'f' is nil, the aggregator defaults to a silent (no-op) operation.
//   - It is used to track the progress of the internal data pipeline and the
//     status of the background processing goroutine.
//
// Parameters:
//   - f: The info logging function. It supports standard formatting with a message
//     string and an optional slice of arguments for variable substitution.
//
// Example:
//
//	agg.SetLoggerInfo(func(msg string, arg ...any) {
//	    log.Infof("Aggregator Status: "+msg, arg...)
//	})
func (o *agg) SetLoggerInfo(f func(msg string, arg ...any)) {
	// Optimization: Prevent unnecessary atomic writes if no change is needed.
	if o.li.Load() == nil && f == nil {
		return
	}

	// Normalization: Guard against nil calls in the processing loop.
	if f == nil {
		o.li.Store(func(msg string, arg ...any) {})
		return
	}

	o.li.Store(f)
}

// logError serves as the internal safe-wrapper for executing the error logger.
//
// Implementation Details:
//  1. It performs multi-level validation: checks if the aggregator is initialized,
//     if an actual error occurred, and if the logger container is non-nil.
//  2. It performs an atomic load of the logger function to ensure it uses the
//     most recently registered version without blocking other goroutines.
//  3. If no logger is configured, the call is discarded with zero performance overhead.
//
// Parameters:
//   - msg: A descriptive summary of where the error occurred (e.g., "batch processing").
//   - err: The error object returned by the underlying subsystem.
func (o *agg) logError(msg string, err error) {
	// Early exit if logging is not required or possible.
	if o == nil || err == nil || o.le == nil {
		return
	}

	// Atomic load ensures thread-safety.
	i := o.le.Load()

	if i == nil {
		return
	}

	// Execute the user-provided callback.
	i(msg, err)
}

// logInfo serves as the internal safe-wrapper for executing the info logger.
//
// Implementation Details:
// It follows the same thread-safe pattern as logError, ensuring that lifecycle
// events are recorded accurately without introducing race conditions during
// configuration changes.
//
// Parameters:
//   - msg: The informational message template.
//   - arg: Variable arguments for message formatting.
func (o *agg) logInfo(msg string, arg ...any) {
	// Early exit validation.
	if o == nil || len(msg) < 1 || o.li == nil {
		return
	}

	// Atomic load ensures consistency.
	i := o.li.Load()

	if i == nil {
		return
	}

	// Execute the user-provided callback.
	i(msg, arg...)
}
