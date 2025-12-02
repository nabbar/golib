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

package bandwidth

import (
	"sync/atomic"
	"time"

	libfpg "github.com/nabbar/golib/file/progress"
	libsiz "github.com/nabbar/golib/size"
)

// bw implements the BandWidth interface using atomic operations for thread-safe rate limiting.
//
// The structure maintains minimal state:
//   - t: Atomic timestamp of the last bandwidth measurement
//   - l: Configured bandwidth limit in bytes per second
//
// Thread safety is achieved through atomic.Value for lock-free timestamp storage.
type bw struct {
	t *atomic.Value // Stores time.Time of last Increment call
	l libsiz.Size   // Bandwidth limit in bytes per second (0 = unlimited)
}

func (o *bw) RegisterIncrement(fpg libfpg.Progress, fi libfpg.FctIncrement) {
	fpg.RegisterFctIncrement(func(size int64) {
		o.Increment(size)
		if fi != nil {
			fi(size)
		}
	})
}

func (o *bw) RegisterReset(fpg libfpg.Progress, fr libfpg.FctReset) {
	fpg.RegisterFctReset(func(size, current int64) {
		o.Reset(size, current)
		if fr != nil {
			fr(size, current)
		}
	})
}

// Increment enforces bandwidth rate limiting based on transferred bytes.
//
// This method is called internally when data is transferred. It calculates the current
// transfer rate and introduces sleep delays if the rate exceeds the configured limit.
//
// Algorithm:
//  1. Retrieve last timestamp from atomic storage
//  2. If this is first call or after reset, store timestamp and return (no throttling)
//  3. Calculate elapsed time since last call
//  4. Calculate current rate: bytes / elapsed_seconds
//  5. If rate > limit, calculate required sleep: (rate / limit) * 1 second
//  6. Sleep to enforce limit (capped at 1 second maximum)
//  7. Store current timestamp for next calculation
//
// Parameters:
//   - size: Number of bytes transferred in this increment
//
// Behavior:
//   - Returns immediately if o is nil (defensive programming)
//   - Returns immediately on first call (no previous timestamp)
//   - Returns immediately if limit is 0 (unlimited bandwidth)
//   - Sleeps if current rate exceeds configured limit
//   - Maximum sleep duration is 1 second to prevent excessive blocking
//
// Thread safety: Safe for concurrent calls due to atomic operations on timestamp storage.
func (o *bw) Increment(size int64) {
	if o == nil {
		return
	}

	var (
		i any
		t time.Time
		k bool
	)

	// Load previous timestamp atomically
	i = o.t.Load()
	if i == nil {
		t = time.Time{}
	} else if t, k = i.(time.Time); !k {
		t = time.Time{}
	}

	// Enforce rate limit if previous timestamp exists and limit is set
	if !t.IsZero() && o.l > 0 {
		ts := time.Since(t)

		// Avoid division by zero or very small values that cause excessive sleep
		if ts < time.Millisecond {
			// If less than 1ms elapsed, skip throttling for this increment
			// to avoid unrealistic rate calculations
			o.t.Store(time.Now())
			return
		} else if ts < 100*time.Millisecond {
			return
		}

		rt := float64(size) / ts.Seconds() // Current rate in bytes/second
		if lm := o.l.Float64(); rt > lm {  // Rate exceeds limit?
			wt := time.Duration((rt / lm) * float64(time.Second)) // Required sleep duration
			if wt.Seconds() > float64(time.Second) {
				time.Sleep(time.Second) // Cap sleep at 1 second
			} else {
				time.Sleep(wt) // Sleep to enforce limit
			}
		}
	}

	// Store current timestamp for next calculation
	o.t.Store(time.Now())
}

// Reset clears the bandwidth tracking state by resetting the internal timestamp.
//
// This method is called internally when the progress tracker is reset. It stores
// a zero time.Time value in the atomic storage, effectively clearing the timestamp.
// The next Increment call will then behave as if it's the first call, storing a
// new timestamp without enforcing any rate limit.
//
// Parameters:
//   - size: Maximum progress reached before reset (unused but part of callback signature)
//   - current: Current progress at reset time (unused but part of callback signature)
//
// Behavior:
//   - Clears internal timestamp to time.Time{} (zero value)
//   - Next Increment call will not enforce rate limit (no previous timestamp)
//   - Subsequent Increment calls will resume normal rate limiting
//
// Thread safety: Safe for concurrent calls due to atomic Store operation.
func (o *bw) Reset(size, current int64) {
	o.t.Store(time.Time{})
}
