/*
 * MIT License
 *
 * Copyright (c) 2022 Nicolas JUHEL
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

package status

import (
	"sync/atomic"
	"time"

	monsts "github.com/nabbar/golib/monitor/status"
)

// timeCache is the default cache duration for the computed health status.
// The status is recomputed only if the cached value is older than this duration.
// This prevents excessive load from frequent health checks.
const timeCache = 3 * time.Second

// ch is the internal cache structure for storing the computed health status.
// It uses atomic operations for thread-safe access, avoiding locks for high-performance
// reads, which is crucial for endpoints that may be hit frequently.
type ch struct {
	m *atomic.Int32        // Maximum cache duration in seconds.
	t *atomic.Value        // Timestamp of the last computation (time.Time).
	c *atomic.Int64        // Cached status value, stored as an int64 representation of monsts.Status.
	f func() monsts.Status // The function to call to compute a fresh status when the cache is stale.
}

// Max returns the maximum cache duration.
// It returns a custom duration if one is set via configuration, otherwise it
// falls back to the default `timeCache` duration (3 seconds).
func (o *ch) Max() time.Duration {
	if m := o.m.Load(); m > 0 {
		return time.Duration(m) * time.Second
	} else {
		return timeCache
	}
}

// Time returns the timestamp of the last status computation.
// It returns a zero time if no computation has been performed yet.
func (o *ch) Time() time.Time {
	if t := o.t.Load(); t != nil {
		return t.(time.Time)
	} else {
		return time.Time{}
	}
}

// IsCache retrieves the health status, using a cached result if it is still valid.
// The cache is considered valid if:
//   - A previous computation has occurred (Time() is not zero).
//   - The time elapsed since the last computation is less than the Max() duration.
//
// If the cache is stale or non-existent, this method triggers a fresh computation
// by calling the registered function `f`, updates the cache with the new result
// and timestamp, and then returns the new status.
//
// This method is thread-safe and optimized for high-concurrency scenarios.
func (o *ch) IsCache() monsts.Status {
	if t := o.Time(); !t.IsZero() && time.Since(t) < o.Max() {
		r := o.c.Load()
		return monsts.NewFromInt(r)
	}

	if o.f != nil {
		c := o.f()
		o.c.Store(c.Int64())

		t := time.Now()
		o.t.Store(t)

		return c
	}

	return monsts.KO
}
