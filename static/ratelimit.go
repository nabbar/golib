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
 */

package static

import (
	"context"
	"slices"
	"strconv"
	"sync/atomic"
	"time"

	ginsdk "github.com/gin-gonic/gin"
	libatm "github.com/nabbar/golib/atomic"
	loglvl "github.com/nabbar/golib/logger/level"
)

// ipTrack stores rate limiting information for a single IP address.
// All fields use atomic operations for thread-safe access.
type ipTrack struct {
	pth libatm.MapTyped[string, time.Time] // path -> timestamp mapping
	req *atomic.Int64                      // total request counter
	fts libatm.Value[time.Time]            // first request timestamp
	lts libatm.Value[time.Time]            // last request timestamp
}

// SetRateLimit configures IP-based rate limiting.
// This method is thread-safe and uses atomic operations.
//
// If rate limiting is enabled, it starts a background goroutine
// for periodic cache cleanup.
func (s *staticHandler) SetRateLimit(cfg RateLimitConfig) {
	s.rlc.Store(&cfg)

	if cfg.Enabled {
		x, n := context.WithCancel(context.Background())

		o := s.rlx.Swap(n)
		if o != nil {
			o()
		}

		go s.cleanupRateLimitCache(x)
	}
}

// GetRateLimit returns the current rate limiting configuration.
// This method is thread-safe and uses atomic operations.
func (s *staticHandler) GetRateLimit() RateLimitConfig {
	if cfg := s.rlc.Load(); cfg != nil {
		return *cfg
	}
	return RateLimitConfig{}
}

// IsRateLimited checks if an IP address is currently rate limited.
// It counts unique file paths requested within the time window.
// Returns true if the IP has exceeded the configured MaxRequests.
func (s *staticHandler) IsRateLimited(ip string) bool {
	cfg := s.GetRateLimit()
	if !cfg.Enabled {
		return false
	}

	trk := s.getRateLimitTracker(ip)
	if trk == nil {
		return false
	}

	now := time.Now()
	windowStart := now.Add(-cfg.Window)

	// Count unique paths within the time window
	up := 0 // unique paths
	trk.pth.Range(func(_ string, ts time.Time) bool {
		if ts.After(windowStart) {
			up++
		}
		return true
	})

	return up >= cfg.MaxRequests
}

// ResetRateLimit clears all rate limiting data for a specific IP address.
// This can be used to manually unblock an IP or reset counters.
func (s *staticHandler) ResetRateLimit(ip string) {
	s.rli.Delete(ip)
}

// getRateLimitTracker retrieves or creates a rate limit tracker for an IP.
// This method is thread-safe and uses atomic map operations.
func (s *staticHandler) getRateLimitTracker(ip string) *ipTrack {

	if i, l := s.rli.Load(ip); l && i != nil {
		return i
	}

	// Create new tracker
	t := &ipTrack{
		pth: libatm.NewMapTyped[string, time.Time](),
		req: new(atomic.Int64),
		fts: libatm.NewValue[time.Time](),
		lts: libatm.NewValue[time.Time](),
	}
	s.rli.Store(ip, t)

	return t
}

// checkRateLimit checks and enforces rate limiting for a request.
// Returns true if the request is allowed, false if rate limited.
//
// This method:
//   - Extracts client IP (respecting X-Forwarded-For)
//   - Checks IP whitelist
//   - Counts unique paths in time window
//   - Returns 429 Too Many Requests if limit exceeded
//   - Adds rate limit headers (X-RateLimit-*)
//   - Notifies security backend on limit exceeded
func (s *staticHandler) checkRateLimit(c *ginsdk.Context) bool {
	cfg := s.GetRateLimit()

	if !cfg.Enabled {
		return true // Rate limiting disabled
	}

	// Extract real client IP
	ip := c.ClientIP()

	// Check if IP is whitelisted
	if slices.Contains(cfg.WhitelistIPs, ip) {
		return true
	}

	// Get or create tracker for this IP
	t := s.getRateLimitTracker(ip)
	if t == nil {
		return true // No cache, allow request
	}

	now := time.Now()
	win := now.Add(-cfg.Window)

	// Clean old entries and count requests in window
	up := 0
	t.pth.Range(func(k string, ts time.Time) bool {
		if ts.After(win) {
			up++
		} else {
			t.pth.Delete(k)
		}
		return true
	})

	// Check if limit is exceeded
	if up >= cfg.MaxRequests {
		// Calculate time until reset
		ots := time.Now()
		t.pth.Range(func(k string, ts time.Time) bool {
			if ts.After(win) && ts.Before(ots) {
				ots = ts
			}
			return true
		})

		rst := ots.Add(cfg.Window)
		rty := int(time.Until(rst).Seconds())
		if rty < 0 {
			rty = 0
		}

		// Log the event
		ent := s.getLogger().Entry(loglvl.WarnLevel, "rate limit exceeded")
		ent.FieldAdd("ip", ip)
		ent.FieldAdd("up", up)
		ent.FieldAdd("limit", cfg.MaxRequests)
		ent.FieldAdd("window", cfg.Window.String())
		ent.FieldAdd("retryAfter", rty)
		ent.Log()

		// Notify WAF/IDS/EDR
		details := map[string]string{
			"unique_paths": strconv.Itoa(up),
			"limit":        strconv.Itoa(cfg.MaxRequests),
			"retry_after":  strconv.Itoa(rty),
		}
		event := s.newSecuEvt(c, EventTypeRateLimit, "medium", true, details)
		s.notifySecuEvt(event)

		// Add rate limiting headers
		c.Header("Retry-After", strconv.Itoa(rty))
		c.Header("X-RateLimit-Limit", strconv.Itoa(cfg.MaxRequests))
		c.Header("X-RateLimit-Remaining", "0")
		c.Header("X-RateLimit-Reset", strconv.FormatInt(rst.Unix(), 10))

		c.AbortWithStatusJSON(429, ginsdk.H{
			"error":   "rate limit exceeded",
			"message": "Too many different file requests. Please retry after " + strconv.Itoa(rty) + " seconds.",
		})

		return false
	}

	// Record current request
	t.pth.Store(c.Request.URL.Path, now)
	t.lts.Store(now)
	t.req.Add(1)

	if t.fts.Load().IsZero() {
		t.fts.Store(now)
	}

	// Add informative headers
	rem := cfg.MaxRequests - up - 1
	if rem < 0 {
		rem = 0
	}

	c.Header("X-RateLimit-Limit", strconv.Itoa(cfg.MaxRequests))
	c.Header("X-RateLimit-Remaining", strconv.Itoa(rem))

	return true
}

// cleanupRateLimitCache periodically cleans up expired rate limit data.
// This background goroutine removes old entries to prevent memory leaks.
// It runs at the configured CleanupInterval and stops when context is canceled.
func (s *staticHandler) cleanupRateLimitCache(ctx context.Context) {
	cfg := s.GetRateLimit()

	if !cfg.Enabled || cfg.CleanupInterval <= 0 {
		return
	}

	ticker := time.NewTicker(cfg.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			now := time.Now()
			win := now.Add(-cfg.Window * 2) // Keep data a bit longer

			s.rli.Range(func(key string, value *ipTrack) bool {
				if value == nil {
					return true
				}

				allExp := true
				value.pth.Range(func(ph string, ts time.Time) bool {
					if ts.IsZero() {
						value.pth.Delete(ph)
						return true
					} else if ts.After(win) {
						allExp = false
						return false
					}
					return true
				})

				if allExp {
					s.rli.Delete(key)
				}

				return true
			})
		}
	}
}
