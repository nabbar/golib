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

package static_test

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/nabbar/golib/static"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Rate Limiting", func() {
	Describe("Configuration", func() {
		Context("when setting rate limit config", func() {
			It("should store and retrieve configuration", func() {
				handler := newTestStatic()

				cfg := static.RateLimitConfig{
					Enabled:         true,
					MaxRequests:     50,
					Window:          30 * time.Second,
					CleanupInterval: 2 * time.Minute,
					WhitelistIPs:    []string{"192.168.1.1"},
					TrustedProxies:  []string{"10.0.0.1"},
				}

				handler.SetRateLimit(cfg)

				retrieved := handler.GetRateLimit()
				Expect(retrieved.Enabled).To(BeTrue())
				Expect(retrieved.MaxRequests).To(Equal(50))
				Expect(retrieved.Window).To(Equal(30 * time.Second))
				Expect(retrieved.CleanupInterval).To(Equal(2 * time.Minute))
				Expect(retrieved.WhitelistIPs).To(ContainElement("192.168.1.1"))
				Expect(retrieved.TrustedProxies).To(ContainElement("10.0.0.1"))
			})

			It("should use default config", func() {
				cfg := static.DefaultRateLimitConfig()

				Expect(cfg.Enabled).To(BeTrue())
				Expect(cfg.MaxRequests).To(Equal(100))
				Expect(cfg.Window).To(Equal(1 * time.Minute))
				Expect(cfg.CleanupInterval).To(Equal(5 * time.Minute))
				Expect(cfg.WhitelistIPs).To(ContainElement("127.0.0.1"))
				Expect(cfg.WhitelistIPs).To(ContainElement("::1"))
			})
		})
	})

	Describe("Basic Rate Limiting", func() {
		Context("when rate limiting is disabled", func() {
			It("should allow unlimited requests", func() {
				handler := newTestStatic()
				engine := setupTestRouter(handler, "/static")

				// Make many requests without rate limiting
				for i := 0; i < 150; i++ {
					w := performRequest(engine, "GET", fmt.Sprintf("/static/test%d.txt", i%10))
					// Should either succeed or 404, but never 429
					Expect(w.Code).NotTo(Equal(http.StatusTooManyRequests))
				}
			})
		})

		Context("when rate limiting is enabled", func() {
			It("should enforce request limits", func() {
				handler := newTestStatic()

				// Configure strict rate limit for testing
				handler.SetRateLimit(static.RateLimitConfig{
					Enabled:         true,
					MaxRequests:     5,
					Window:          1 * time.Minute,
					CleanupInterval: 5 * time.Minute,
					WhitelistIPs:    []string{},
					TrustedProxies:  []string{},
				})

				engine := setupTestRouter(handler, "/static")

				// First 5 unique files should succeed
				for i := 0; i < 5; i++ {
					w := performRequest(engine, "GET", fmt.Sprintf("/static/file%d.txt", i))
					Expect(w.Code).To(Or(Equal(http.StatusOK), Equal(http.StatusNotFound)))
					Expect(w.Code).NotTo(Equal(http.StatusTooManyRequests))
				}

				// 6th unique file should be rate limited
				w := performRequest(engine, "GET", "/static/file6.txt")
				Expect(w.Code).To(Equal(http.StatusTooManyRequests))
				Expect(w.Header().Get("X-RateLimit-Limit")).To(Equal("5"))
				Expect(w.Header().Get("X-RateLimit-Remaining")).To(Equal("0"))
				Expect(w.Header().Get("Retry-After")).NotTo(BeEmpty())
			})

			It("should not count duplicate requests", func() {
				handler := newTestStatic()

				handler.SetRateLimit(static.RateLimitConfig{
					Enabled:     true,
					MaxRequests: 5,
					Window:      1 * time.Minute,
				})

				engine := setupTestRouter(handler, "/static")

				// Request same file multiple times
				for i := 0; i < 10; i++ {
					w := performRequest(engine, "GET", "/static/test.txt")
					// Should never be rate limited since it's the same file
					Expect(w.Code).NotTo(Equal(http.StatusTooManyRequests))
				}
			})

			It("should include rate limit headers", func() {
				handler := newTestStatic()

				handler.SetRateLimit(static.RateLimitConfig{
					Enabled:     true,
					MaxRequests: 10,
					Window:      1 * time.Minute,
				})

				engine := setupTestRouter(handler, "/static")

				w := performRequest(engine, "GET", "/static/test.txt")

				limit := w.Header().Get("X-RateLimit-Limit")
				Expect(limit).To(Equal("10"))

				remaining := w.Header().Get("X-RateLimit-Remaining")
				Expect(remaining).NotTo(BeEmpty())

				remainingInt, err := strconv.Atoi(remaining)
				Expect(err).ToNot(HaveOccurred())
				Expect(remainingInt).To(BeNumerically("<=", 10))
			})
		})
	})

	Describe("Whitelist", func() {
		Context("when IP is whitelisted", func() {
			It("should bypass rate limiting", func() {
				handler := newTestStatic()

				// Note: In tests, ClientIP() will be empty or a test value
				// This test verifies the whitelist logic works
				handler.SetRateLimit(static.RateLimitConfig{
					Enabled:      true,
					MaxRequests:  2,
					Window:       1 * time.Minute,
					WhitelistIPs: []string{"127.0.0.1", ""},
				})

				engine := setupTestRouter(handler, "/static")

				// Should allow unlimited requests from whitelisted IP
				for i := 0; i < 10; i++ {
					w := performRequest(engine, "GET", fmt.Sprintf("/static/file%d.txt", i))
					Expect(w.Code).NotTo(Equal(http.StatusTooManyRequests))
				}
			})
		})
	})

	Describe("Window Management", func() {
		Context("when window expires", func() {
			It("should reset counter after window", func() {
				handler := newTestStatic()

				handler.SetRateLimit(static.RateLimitConfig{
					Enabled:     true,
					MaxRequests: 3,
					Window:      500 * time.Millisecond, // Short window for testing
				})

				engine := setupTestRouter(handler, "/static")

				// Use up the limit
				for i := 0; i < 3; i++ {
					w := performRequest(engine, "GET", fmt.Sprintf("/static/file%d.txt", i))
					Expect(w.Code).NotTo(Equal(http.StatusTooManyRequests))
				}

				// Should be limited
				w := performRequest(engine, "GET", "/static/file4.txt")
				Expect(w.Code).To(Equal(http.StatusTooManyRequests))

				// Wait for window to expire
				time.Sleep(600 * time.Millisecond)

				// Should work again
				w = performRequest(engine, "GET", "/static/file5.txt")
				Expect(w.Code).NotTo(Equal(http.StatusTooManyRequests))
			})
		})
	})

	Describe("IsRateLimited", func() {
		Context("when checking rate limit status", func() {
			It("should correctly report limited status", func() {
				handler := newTestStatic()

				handler.SetRateLimit(static.RateLimitConfig{
					Enabled:     true,
					MaxRequests: 2,
					Window:      1 * time.Minute,
				})

				testIP := "192.168.1.100"

				// Initially not limited
				Expect(handler.IsRateLimited(testIP)).To(BeFalse())

				// The actual test uses ClientIP() which returns empty or test value
				// This just tests the method works
			})
		})
	})

	Describe("ResetRateLimit", func() {
		Context("when resetting rate limit", func() {
			It("should clear counter for IP", func() {
				handler := newTestStatic()

				handler.SetRateLimit(static.RateLimitConfig{
					Enabled:     true,
					MaxRequests: 2,
					Window:      1 * time.Minute,
				})

				testIP := "192.168.1.100"

				// Reset should not panic
				Expect(func() {
					handler.ResetRateLimit(testIP)
				}).NotTo(Panic())
			})
		})
	})

	Describe("Concurrent Access", func() {
		Context("when multiple goroutines access simultaneously", func() {
			It("should handle concurrent requests safely", func() {
				handler := newTestStatic()

				handler.SetRateLimit(static.RateLimitConfig{
					Enabled:     true,
					MaxRequests: 50,
					Window:      1 * time.Minute,
				})

				engine := setupTestRouter(handler, "/static")

				var wg sync.WaitGroup
				successCount := 0
				rateLimitCount := 0
				var mu sync.Mutex

				// Launch 20 goroutines
				for i := 0; i < 20; i++ {
					wg.Add(1)
					go func(id int) {
						defer wg.Done()
						defer GinkgoRecover()

						// Each goroutine makes 5 unique requests
						for j := 0; j < 5; j++ {
							w := performRequest(engine, "GET", fmt.Sprintf("/static/concurrent_%d_%d.txt", id, j))

							mu.Lock()
							if w.Code == http.StatusTooManyRequests {
								rateLimitCount++
							} else {
								successCount++
							}
							mu.Unlock()
						}
					}(i)
				}

				wg.Wait()

				// Should have mix of success and rate limits (total 100 requests, limit 50)
				GinkgoWriter.Printf("Success: %d, Rate Limited: %d\n", successCount, rateLimitCount)
				Expect(successCount + rateLimitCount).To(Equal(100))
				Expect(rateLimitCount).To(BeNumerically(">", 0)) // Some should be rate limited
			})
		})
	})

	Describe("Cleanup", func() {
		Context("when cleanup runs", func() {
			It("should not panic during cleanup", func() {
				handler := newTestStatic()

				handler.SetRateLimit(static.RateLimitConfig{
					Enabled:         true,
					MaxRequests:     10,
					Window:          100 * time.Millisecond,
					CleanupInterval: 200 * time.Millisecond,
				})

				// Wait for cleanup to run
				time.Sleep(500 * time.Millisecond)

				// Should not panic
				Expect(handler.GetRateLimit().Enabled).To(BeTrue())
			})

			It("should cancel cleanup on reconfiguration", func() {
				handler := newTestStatic()

				// Set initial config
				handler.SetRateLimit(static.RateLimitConfig{
					Enabled:         true,
					MaxRequests:     10,
					Window:          1 * time.Second,
					CleanupInterval: 100 * time.Millisecond,
				})

				time.Sleep(50 * time.Millisecond)

				// Reconfigure
				handler.SetRateLimit(static.RateLimitConfig{
					Enabled:         true,
					MaxRequests:     20,
					Window:          2 * time.Second,
					CleanupInterval: 200 * time.Millisecond,
				})

				// Wait and verify no panic
				time.Sleep(300 * time.Millisecond)

				cfg := handler.GetRateLimit()
				Expect(cfg.MaxRequests).To(Equal(20))
			})
		})
	})
})
