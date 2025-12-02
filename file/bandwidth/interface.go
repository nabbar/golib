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

// Package bandwidth provides bandwidth throttling and rate limiting for file I/O operations.
//
// This package integrates seamlessly with the file/progress package to control data transfer
// rates in bytes per second. It implements time-based throttling using atomic operations for
// thread-safe concurrent usage.
//
// Key features:
//   - Configurable bytes-per-second limits
//   - Zero-cost when set to unlimited (0 bytes/second)
//   - Thread-safe atomic operations
//   - Seamless integration with progress tracking
//
// Example usage:
//
//	import (
//	    "github.com/nabbar/golib/file/bandwidth"
//	    "github.com/nabbar/golib/file/progress"
//	    "github.com/nabbar/golib/size"
//	)
//
//	// Create bandwidth limiter (1 MB/s)
//	bw := bandwidth.New(size.SizeMiB)
//
//	// Open file with progress tracking
//	fpg, _ := progress.Open("largefile.dat")
//	defer fpg.Close()
//
//	// Register bandwidth limiting
//	bw.RegisterIncrement(fpg, nil)
//
//	// All I/O operations will be throttled to 1 MB/s
package bandwidth

import (
	"sync/atomic"

	libfpg "github.com/nabbar/golib/file/progress"
	libsiz "github.com/nabbar/golib/size"
)

// BandWidth defines the interface for bandwidth control and rate limiting.
//
// This interface provides methods to register bandwidth limiting callbacks
// with progress-enabled file operations. It integrates seamlessly with the
// progress package to enforce bytes-per-second transfer limits.
//
// All methods are safe for concurrent use across multiple goroutines.
type BandWidth interface {

	// RegisterIncrement registers a bandwidth-limited increment callback with a progress tracker.
	//
	// This method wraps the provided callback function with bandwidth throttling logic. When
	// the progress tracker detects bytes transferred, the bandwidth limiter enforces the
	// configured rate limit before invoking the user-provided callback.
	//
	// The callback function will be invoked with the total number of bytes transferred since
	// the last increment. The callback is optional; if nil, only bandwidth limiting is applied
	// without additional notification.
	//
	// Parameters:
	//   - fpg: Progress tracker to register the callback with
	//   - fi: Optional callback function with signature func(size int64)
	//
	// The callback is invoked:
	//   - After each read/write operation that transfers data
	//   - When the file reaches EOF
	//   - Even if the file is smaller than expected
	//
	// Thread safety: This method is safe to call concurrently with other BandWidth methods.
	//
	// Example:
	//   bw.RegisterIncrement(fpg, func(size int64) {
	//       fmt.Printf("Transferred %d bytes at limited rate\n", size)
	//   })
	RegisterIncrement(fpg libfpg.Progress, fi libfpg.FctIncrement)

	// RegisterReset registers a reset callback that clears bandwidth tracking state.
	//
	// This method registers a callback to be invoked when the progress tracker is reset.
	// The bandwidth limiter clears its internal timestamp state, allowing a fresh rate
	// calculation after the reset. The user-provided callback is then invoked with
	// the reset parameters.
	//
	// Parameters:
	//   - fpg: Progress tracker to register the callback with
	//   - fr: Optional callback function with signature func(size, current int64)
	//
	// The callback receives:
	//   - size: Maximum progress reached before reset
	//   - current: Current progress at the time of reset
	//
	// The callback is invoked:
	//   - When fpg.Reset() is explicitly called
	//   - When the file is repositioned (seek operations)
	//   - When io.Copy completes and progress is finalized
	//
	// Thread safety: This method is safe to call concurrently with other BandWidth methods.
	//
	// Example:
	//   bw.RegisterReset(fpg, func(size, current int64) {
	//       fmt.Printf("Reset: max=%d current=%d\n", size, current)
	//   })
	RegisterReset(fpg libfpg.Progress, fr libfpg.FctReset)
}

// New creates a new BandWidth instance with the specified rate limit.
//
// This function returns a bandwidth limiter that enforces the given bytes-per-second
// transfer rate. The limiter uses time-based throttling with atomic operations for
// thread-safe concurrent usage.
//
// Parameters:
//   - bytesBySecond: Maximum transfer rate in bytes per second
//   - Use 0 for unlimited bandwidth (no throttling overhead)
//   - Common values: size.SizeKilo (1KB/s), size.SizeMega (1MB/s), etc.
//
// Behavior:
//   - When limit is 0: No throttling applied, zero overhead
//   - When limit > 0: Enforces rate by introducing sleep delays
//   - Rate calculation: bytes / elapsed_seconds
//   - Sleep duration: capped at 1 second maximum per operation
//
// The returned instance is safe for concurrent use across multiple goroutines.
// All methods can be called concurrently without external synchronization.
//
// Thread safety:
//   - Safe for concurrent RegisterIncrement/RegisterReset calls
//   - Internal state protected by atomic operations
//   - No mutexes required for concurrent access
//
// Performance:
//   - Zero-cost when unlimited (bytesBySecond = 0)
//   - Minimal overhead when limiting enabled (<1ms per operation)
//   - Lock-free implementation using atomic.Value
//
// Example usage:
//
//	// Unlimited bandwidth
//	bw := bandwidth.New(0)
//
//	// 1 MB/s limit
//	bw := bandwidth.New(size.SizeMega)
//
//	// Custom 512 KB/s limit
//	bw := bandwidth.New(512 * size.SizeKilo)
func New(bytesBySecond libsiz.Size) BandWidth {
	return &bw{
		t: new(atomic.Value),
		l: bytesBySecond,
	}
}
