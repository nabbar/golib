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
	"testing"
	"time"

	libsiz "github.com/nabbar/golib/size"
)

// Test internal Increment behavior with nil receiver
func TestIncrementNilReceiver(t *testing.T) {
	var b *bw = nil

	// Should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Increment panicked with nil receiver: %v", r)
		}
	}()

	b.Increment(1024)
}

// Test Increment with zero limit (unlimited bandwidth)
func TestIncrementZeroLimit(t *testing.T) {
	b := &bw{
		t: new(atomic.Value),
		l: 0,
	}

	// First call - should just store timestamp
	b.Increment(1024)

	val := b.t.Load()
	if val == nil {
		t.Error("Expected timestamp to be stored after first Increment")
	}

	// Second call with zero limit - should not throttle
	start := time.Now()
	b.Increment(1024)
	elapsed := time.Since(start)

	// Should be very fast (no throttling)
	if elapsed > 100*time.Millisecond {
		t.Errorf("Increment with zero limit took too long: %v", elapsed)
	}
}

// Test Increment with very small elapsed time (< 1ms)
func TestIncrementSmallElapsedTime(t *testing.T) {
	b := &bw{
		t: new(atomic.Value),
		l: libsiz.SizeKilo,
	}

	// Store a timestamp very close to now
	b.t.Store(time.Now())

	// Immediately call Increment (< 1ms elapsed)
	start := time.Now()
	b.Increment(512)
	elapsed := time.Since(start)

	// Should skip throttling due to < 1ms elapsed
	if elapsed > 10*time.Millisecond {
		t.Errorf("Increment skipped throttling but took too long: %v", elapsed)
	}
}

// Test Increment with rate below limit
func TestIncrementRateBelowLimit(t *testing.T) {
	b := &bw{
		t: new(atomic.Value),
		l: libsiz.SizeMega, // 1 MB/s
	}

	// Store timestamp 100ms ago
	b.t.Store(time.Now().Add(-100 * time.Millisecond))

	// Transfer 1KB in 100ms = 10 KB/s (well below 1 MB/s limit)
	start := time.Now()
	b.Increment(1024)
	elapsed := time.Since(start)

	// Should not throttle (rate below limit)
	if elapsed > 10*time.Millisecond {
		t.Errorf("Increment throttled when rate was below limit: %v", elapsed)
	}
}

// Test Increment with rate above limit but reasonable
func TestIncrementRateAboveLimitReasonable(t *testing.T) {
	b := &bw{
		t: new(atomic.Value),
		l: libsiz.SizeKilo, // 1 KB/s
	}

	// Store timestamp 10ms ago
	b.t.Store(time.Now().Add(-10 * time.Millisecond))

	// Transfer 2KB in 10ms = 200 KB/s (way above 1 KB/s limit)
	// But should be capped at 1 second sleep
	start := time.Now()
	b.Increment(2048)
	elapsed := time.Since(start)

	// Should throttle (rate above limit)
	// Expected sleep should be capped at 1 second
	if elapsed > 1100*time.Millisecond {
		t.Errorf("Increment slept longer than 1 second cap: %v", elapsed)
	}
}

// Test Increment first call (no previous timestamp)
func TestIncrementFirstCall(t *testing.T) {
	b := &bw{
		t: new(atomic.Value),
		l: libsiz.SizeKilo,
	}

	// First call - no previous timestamp
	start := time.Now()
	b.Increment(1024)
	elapsed := time.Since(start)

	// Should not throttle on first call
	if elapsed > 10*time.Millisecond {
		t.Errorf("Increment throttled on first call: %v", elapsed)
	}

	// Timestamp should be stored
	val := b.t.Load()
	if val == nil {
		t.Error("Expected timestamp to be stored after first Increment")
	}

	if _, ok := val.(time.Time); !ok {
		t.Error("Stored value is not a time.Time")
	}
}

// Test Increment with nil stored value (Load returns nil)
func TestIncrementNilStoredValue(t *testing.T) {
	b := &bw{
		t: new(atomic.Value),
		l: libsiz.SizeKilo,
	}

	// Don't store anything - Load() will return nil
	// Should treat as zero time and not panic
	start := time.Now()
	b.Increment(1024)
	elapsed := time.Since(start)

	// Should not throttle (treated as first call)
	if elapsed > 10*time.Millisecond {
		t.Errorf("Increment throttled with nil stored value: %v", elapsed)
	}
}

// Test Reset functionality
func TestReset(t *testing.T) {
	b := &bw{
		t: new(atomic.Value),
		l: libsiz.SizeKilo,
	}

	// Store a timestamp
	b.t.Store(time.Now())

	// Reset
	b.Reset(1024, 512)

	// Timestamp should be zero
	val := b.t.Load()
	if val == nil {
		t.Error("Expected zero time to be stored after Reset")
		return
	}

	if ts, ok := val.(time.Time); !ok {
		t.Error("Stored value is not a time.Time after Reset")
	} else if !ts.IsZero() {
		t.Error("Expected zero time after Reset")
	}
}

// Test multiple Increment calls with proper spacing
func TestMultipleIncrementsWithSpacing(t *testing.T) {
	b := &bw{
		t: new(atomic.Value),
		l: 0, // Unlimited for fast test
	}

	// Multiple increments
	for i := 0; i < 5; i++ {
		b.Increment(512)
		time.Sleep(2 * time.Millisecond) // Small delay
	}

	// Should complete without issues
	val := b.t.Load()
	if val == nil {
		t.Error("Expected timestamp to be stored")
	}
}
