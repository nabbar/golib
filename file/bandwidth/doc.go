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

/*
Package bandwidth provides bandwidth throttling and rate limiting for file I/O operations.

# Design Philosophy

The bandwidth package implements time-based throttling using atomic operations for thread-safe
concurrent usage. It seamlessly integrates with the github.com/nabbar/golib/file/progress package
to enforce bytes-per-second transfer limits on file operations.

Key principles:
  - Zero-cost when unlimited: Setting limit to 0 disables throttling with no overhead
  - Atomic operations: Thread-safe concurrent access without mutexes
  - Callback integration: Seamless integration with progress tracking callbacks
  - Time-based limiting: Enforces rate limits by introducing sleep delays

# Architecture

The package consists of two main components:

	┌─────────────────────────────────────────────┐
	│           BandWidth Interface               │
	│  ┌───────────────────────────────────────┐  │
	│  │  RegisterIncrement(fpg, callback)     │  │
	│  │  RegisterReset(fpg, callback)         │  │
	│  └───────────────────────────────────────┘  │
	└──────────────────┬──────────────────────────┘
	                   │
	┌──────────────────▼──────────────────────────┐
	│           bw Implementation                 │
	│  ┌───────────────────────────────────────┐  │
	│  │  t: atomic.Value (timestamp)          │  │
	│  │  l: Size (bytes per second limit)     │  │
	│  └───────────────────────────────────────┘  │
	│  ┌───────────────────────────────────────┐  │
	│  │  Increment(size) - enforce limit      │  │
	│  │  Reset(size, current) - clear state   │  │
	│  └───────────────────────────────────────┘  │
	└─────────────────────────────────────────────┘
	                   │
	┌──────────────────▼──────────────────────────┐
	│      Progress Package Integration           │
	│  ┌───────────────────────────────────────┐  │
	│  │  FctIncrement callbacks               │  │
	│  │  FctReset callbacks                   │  │
	│  └───────────────────────────────────────┘  │
	└─────────────────────────────────────────────┘

The BandWidth interface provides registration methods that wrap user callbacks with
bandwidth limiting logic. The internal bw struct tracks the last operation timestamp
atomically and calculates sleep durations to enforce the configured limit.

# Rate Limiting Algorithm

The throttling algorithm works as follows:

 1. Store timestamp when bytes are transferred
 2. On next transfer, calculate elapsed time since last timestamp
 3. Calculate current rate: rate = bytes / elapsed_seconds
 4. If rate > limit, calculate required sleep: sleep = (rate / limit) * 1s
 5. Sleep to enforce limit (capped at 1 second maximum)
 6. Store new timestamp

This approach provides smooth rate limiting without strict per-operation delays,
allowing burst transfers when the average rate is below the limit.

# Performance

The package is designed for minimal overhead:

  - Zero-cost unlimited: No overhead when limit is 0
  - Atomic operations: Lock-free timestamp storage
  - No allocations: Reuses atomic.Value for timestamp storage
  - Efficient calculation: Simple floating-point math for rate calculation
  - Bounded sleep: Maximum 1 second sleep per operation prevents excessive delays

Typical overhead with limiting enabled: <1ms per operation for sleep calculation

# Use Cases

1. Network Bandwidth Control

Control upload/download speeds to avoid overwhelming network connections:

	bw := bandwidth.New(size.SizeMiB) // 1 MB/s limit
	fpg, _ := progress.Open("upload.dat")
	bw.RegisterIncrement(fpg, nil)
	io.Copy(networkConn, fpg) // Throttled to 1 MB/s

2. Disk I/O Rate Limiting

Prevent disk saturation during large file operations:

	bw := bandwidth.New(10 * size.SizeMiB) // 10 MB/s
	fpg, _ := progress.Open("large_backup.tar")
	bw.RegisterIncrement(fpg, func(sz int64) {
	    fmt.Printf("Progress: %d bytes\n", sz)
	})
	io.Copy(destination, fpg)

3. Multi-File Shared Bandwidth

Control aggregate bandwidth across multiple concurrent transfers:

	sharedBW := bandwidth.New(5 * size.SizeMiB) // Shared 5 MB/s
	for _, file := range files {
	    go func(f string) {
	        fpg, _ := progress.Open(f)
	        sharedBW.RegisterIncrement(fpg, nil)
	        io.Copy(destination, fpg)
	    }(file)
	}

4. Progress Monitoring with Rate Limiting

Combine bandwidth limiting with progress tracking:

	bw := bandwidth.New(size.SizeMiB)
	fpg, _ := progress.Open("data.bin")
	bw.RegisterIncrement(fpg, func(sz int64) {
	    pct := float64(sz) / float64(fileSize) * 100
	    fmt.Printf("Progress: %.1f%%\n", pct)
	})
	io.Copy(writer, fpg) // 1 MB/s with progress updates

# Limitations

1. Single-byte granularity: Limit specified in bytes per second
2. Time-based accuracy: Depends on system clock resolution
3. No burst control: Does not enforce strict per-operation limits
4. No traffic shaping: Simple rate limiting without advanced QoS

# Best Practices

DO:
  - Use 0 for unlimited bandwidth (zero overhead)
  - Set reasonable limits based on expected transfer rates
  - Share BandWidth instances across multiple files for aggregate limiting
  - Monitor progress with callbacks for user feedback

DON'T:
  - Use extremely small limits (<100 bytes/s) - may cause excessive sleep overhead
  - Modify limit during active transfers - create new instance instead
  - Rely on precise rate limiting - algorithm provides approximate limiting

# Thread Safety

The BandWidth instance is safe for concurrent use across multiple goroutines.
The atomic.Value provides lock-free access to the timestamp, allowing concurrent
RegisterIncrement and RegisterReset calls without contention.

However, the actual rate limiting is applied per-operation, so concurrent operations
on the same BandWidth instance will each independently enforce the limit. For true
aggregate bandwidth control, ensure operations are serialized or use separate
instances per goroutine with appropriate limits.

# Integration with Progress Package

The bandwidth package is designed to work seamlessly with github.com/nabbar/golib/file/progress:

	┌───────────┐    RegisterIncrement    ┌────────────┐
	│ BandWidth │◄───────────────────────►│  Progress  │
	└───────────┘    RegisterReset        └────────────┘
	      │                                      │
	      │ Enforce rate limit                   │ Track bytes
	      ▼                                      ▼
	  Increment(size)                     FctIncrement(size)
	  Reset(size, cur)                    FctReset(size, cur)

The BandWidth wrapper calls are inserted into the progress callback chain,
ensuring rate limiting is applied transparently during file I/O operations.

# Example Usage Patterns

Basic usage with default options:

	bw := bandwidth.New(size.SizeMiB)
	fpg, _ := progress.Open("file.dat")
	bw.RegisterIncrement(fpg, nil)
	io.Copy(destination, fpg)

With progress callback:

	bw := bandwidth.New(2 * size.SizeMiB)
	fpg, _ := progress.Open("file.dat")
	bw.RegisterIncrement(fpg, func(sz int64) {
	    fmt.Printf("Transferred: %d bytes\n", sz)
	})
	io.Copy(destination, fpg)

With reset callback:

	bw := bandwidth.New(size.SizeMiB)
	fpg, _ := progress.Open("file.dat")
	bw.RegisterReset(fpg, func(sz, cur int64) {
	    fmt.Printf("Reset: max=%d current=%d\n", sz, cur)
	})
	io.Copy(destination, fpg)

Unlimited bandwidth (no throttling):

	bw := bandwidth.New(0) // Zero overhead
	fpg, _ := progress.Open("file.dat")
	bw.RegisterIncrement(fpg, nil)
	io.Copy(destination, fpg)

# Related Packages

  - github.com/nabbar/golib/file/progress - Progress tracking for file I/O
  - github.com/nabbar/golib/size - Size constants and utilities
  - github.com/nabbar/golib/file/perm - File permissions handling

# References

  - Go Atomic Operations: https://pkg.go.dev/sync/atomic
  - Rate Limiting Patterns: https://en.wikipedia.org/wiki/Rate_limiting
  - Token Bucket Algorithm: https://en.wikipedia.org/wiki/Token_bucket
*/
package bandwidth
