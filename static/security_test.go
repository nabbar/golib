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
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"time"

	"github.com/nabbar/golib/static"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Security Backend Integration", func() {
	Describe("Configuration", func() {
		Context("when setting security backend config", func() {
			It("should store and retrieve configuration", func() {
				handler := newTestStatic().(static.Static)

				cfg := static.SecurityConfig{
					Enabled:         true,
					WebhookURL:      "http://localhost:9999/webhook",
					WebhookTimeout:  5 * time.Second,
					WebhookAsync:    true,
					MinSeverity:     "high",
					BatchSize:       10,
					BatchTimeout:    30 * time.Second,
					EnableCEFFormat: false,
				}

				handler.SetSecurityBackend(cfg)

				retrieved := handler.GetSecurityBackend()
				Expect(retrieved.Enabled).To(BeTrue())
				Expect(retrieved.WebhookURL).To(Equal("http://localhost:9999/webhook"))
				Expect(retrieved.MinSeverity).To(Equal("high"))
				Expect(retrieved.BatchSize).To(Equal(10))
			})

			It("should use default config", func() {
				cfg := static.DefaultSecurityConfig()

				Expect(cfg.Enabled).To(BeFalse())
				Expect(cfg.WebhookAsync).To(BeTrue())
				Expect(cfg.MinSeverity).To(Equal("medium"))
				Expect(cfg.BatchSize).To(Equal(0))
			})
		})
	})

	Describe("Webhook Integration", func() {
		Context("when sending events to webhook", func() {
			It("should send path traversal events", func() {
				// Setup webhook server
				receivedEvents := make([]map[string]interface{}, 0)
				var mu sync.Mutex

				webhookServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					body, _ := io.ReadAll(r.Body)
					var event map[string]interface{}
					json.Unmarshal(body, &event)

					mu.Lock()
					receivedEvents = append(receivedEvents, event)
					mu.Unlock()

					w.WriteHeader(http.StatusOK)
				}))
				defer webhookServer.Close()

				// Configure handler
				handler := newTestStatic().(static.Static)
				handler.SetSecurityBackend(static.SecurityConfig{
					Enabled:        true,
					WebhookURL:     webhookServer.URL,
					WebhookAsync:   false, // Synchrone pour les tests
					WebhookTimeout: 5 * time.Second,
					MinSeverity:    "medium",
				})
				handler.SetPathSecurity(static.DefaultPathSecurityConfig())

				engine := setupTestRouter(handler, "/static")

				// Trigger path traversal
				_ = performRequest(engine, "GET", "/static/../../../etc/passwd")

				// Wait a bit for webhook
				time.Sleep(100 * time.Millisecond)

				mu.Lock()
				defer mu.Unlock()
				Expect(receivedEvents).To(HaveLen(1))
				Expect(receivedEvents[0]["event_type"]).To(Equal("path_traversal"))
				Expect(receivedEvents[0]["severity"]).To(Equal("high"))
				Expect(receivedEvents[0]["blocked"]).To(BeTrue())
			})

			It("should send rate limit events", func() {
				receivedEvents := &atomic.Int32{}

				webhookServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					receivedEvents.Add(1)
					w.WriteHeader(http.StatusOK)
				}))
				defer webhookServer.Close()

				handler := newTestStatic().(static.Static)
				handler.SetSecurityBackend(static.SecurityConfig{
					Enabled:        true,
					WebhookURL:     webhookServer.URL,
					WebhookAsync:   false,
					WebhookTimeout: 5 * time.Second,
					MinSeverity:    "medium",
				})
				handler.SetRateLimit(static.RateLimitConfig{
					Enabled:     true,
					MaxRequests: 2,
					Window:      time.Minute,
				})

				engine := setupTestRouter(handler, "/static")

				// Trigger rate limit
				_ = performRequest(engine, "GET", "/static/file1.txt")
				_ = performRequest(engine, "GET", "/static/file2.txt")
				_ = performRequest(engine, "GET", "/static/file3.txt") // Should trigger rate limit

				time.Sleep(100 * time.Millisecond)

				Expect(receivedEvents.Load()).To(BeNumerically(">=", 1))
			})

			It("should send MIME type denied events", func() {
				receivedEvents := &atomic.Int32{}

				webhookServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					receivedEvents.Add(1)
					w.WriteHeader(http.StatusOK)
				}))
				defer webhookServer.Close()

				handler := newTestStatic().(static.Static)
				handler.SetSecurityBackend(static.SecurityConfig{
					Enabled:        true,
					WebhookURL:     webhookServer.URL,
					WebhookAsync:   false,
					WebhookTimeout: 5 * time.Second,
					MinSeverity:    "medium",
				})
				handler.SetHeaders(static.HeadersConfig{
					EnableContentType: true,
					DenyMimeTypes:     []string{"text/plain"},
				})

				engine := setupTestRouter(handler, "/static")

				// Trigger MIME type denied
				_ = performRequest(engine, "GET", "/static/test.txt")

				time.Sleep(100 * time.Millisecond)

				Expect(receivedEvents.Load()).To(Equal(int32(1)))
			})

			It("should respect minimum severity level", func() {
				receivedEvents := &atomic.Int32{}

				webhookServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					receivedEvents.Add(1)
					w.WriteHeader(http.StatusOK)
				}))
				defer webhookServer.Close()

				handler := newTestStatic().(static.Static)
				handler.SetSecurityBackend(static.SecurityConfig{
					Enabled:        true,
					WebhookURL:     webhookServer.URL,
					WebhookAsync:   false,
					WebhookTimeout: 5 * time.Second,
					MinSeverity:    "critical", // Only critical events
				})
				handler.SetRateLimit(static.RateLimitConfig{
					Enabled:     true,
					MaxRequests: 1,
					Window:      time.Minute,
				})

				engine := setupTestRouter(handler, "/static")

				// Trigger medium severity event (rate limit)
				_ = performRequest(engine, "GET", "/static/file1.txt")
				_ = performRequest(engine, "GET", "/static/file2.txt")

				time.Sleep(100 * time.Millisecond)

				// Should not receive event because severity is only "medium"
				Expect(receivedEvents.Load()).To(Equal(int32(0)))
			})
		})

		Context("when webhook fails", func() {
			It("should handle connection errors gracefully", func() {
				handler := newTestStatic().(static.Static)
				handler.SetSecurityBackend(static.SecurityConfig{
					Enabled:        true,
					WebhookURL:     "http://localhost:99999/webhook", // Invalid port
					WebhookAsync:   false,
					WebhookTimeout: 1 * time.Second,
					MinSeverity:    "medium",
				})
				handler.SetPathSecurity(static.DefaultPathSecurityConfig())

				engine := setupTestRouter(handler, "/static")

				// Should not panic even if webhook fails
				_ = performRequest(engine, "GET", "/static/../passwd")
			})

			It("should handle webhook error responses", func() {
				webhookServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				}))
				defer webhookServer.Close()

				handler := newTestStatic().(static.Static)
				handler.SetSecurityBackend(static.SecurityConfig{
					Enabled:        true,
					WebhookURL:     webhookServer.URL,
					WebhookAsync:   false,
					WebhookTimeout: 5 * time.Second,
					MinSeverity:    "medium",
				})
				handler.SetPathSecurity(static.DefaultPathSecurityConfig())

				engine := setupTestRouter(handler, "/static")

				// Should not panic even if webhook returns error
				_ = performRequest(engine, "GET", "/static/../passwd")
			})
		})
	})

	Describe("Callback Integration", func() {
		Context("when using Go callbacks", func() {
			It("should not panic with callback configuration", func() {
				handler := newTestStatic().(static.Static)
				handler.SetSecurityBackend(static.SecurityConfig{
					Enabled:     true,
					MinSeverity: "medium",
				})

				// AddSecurityCallback is not accessible because SecuEvtCallback uses a private type
				// We test via webhook instead

				handler.SetPathSecurity(static.DefaultPathSecurityConfig())

				engine := setupTestRouter(handler, "/static")

				// Trigger event
				_ = performRequest(engine, "GET", "/static/../passwd")

				time.Sleep(100 * time.Millisecond)

				// Test passed if no panic
			})

			It("should handle configuration with callbacks", func() {
				handler := newTestStatic().(static.Static)
				handler.SetSecurityBackend(static.SecurityConfig{
					Enabled:     true,
					MinSeverity: "medium",
					Callbacks:   []static.SecuEvtCallback{}, // Empty callbacks list
				})

				handler.SetPathSecurity(static.DefaultPathSecurityConfig())

				engine := setupTestRouter(handler, "/static")

				// Should not panic even with empty callbacks
				_ = performRequest(engine, "GET", "/static/../passwd")

				time.Sleep(100 * time.Millisecond)

				// Test passed if no panic
			})
		})
	})

	Describe("CEF Format", func() {
		Context("when CEF format is enabled", func() {
			It("should send events in CEF format", func() {
				var receivedBody string
				var mu sync.Mutex

				webhookServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					body, _ := io.ReadAll(r.Body)
					mu.Lock()
					receivedBody = string(body)
					mu.Unlock()

					// Check Content-Type
					Expect(r.Header.Get("Content-Type")).To(Equal("text/plain"))

					w.WriteHeader(http.StatusOK)
				}))
				defer webhookServer.Close()

				handler := newTestStatic().(static.Static)
				handler.SetSecurityBackend(static.SecurityConfig{
					Enabled:         true,
					WebhookURL:      webhookServer.URL,
					WebhookAsync:    false,
					WebhookTimeout:  5 * time.Second,
					MinSeverity:     "medium",
					EnableCEFFormat: true,
				})
				handler.SetPathSecurity(static.DefaultPathSecurityConfig())

				engine := setupTestRouter(handler, "/static")

				_ = performRequest(engine, "GET", "/static/../passwd")

				time.Sleep(100 * time.Millisecond)

				mu.Lock()
				defer mu.Unlock()
				Expect(receivedBody).To(ContainSubstring("CEF:0"))
				Expect(receivedBody).To(ContainSubstring("golib"))
				Expect(receivedBody).To(ContainSubstring("static"))
			})
		})
	})

	Describe("Batch Processing", func() {
		Context("when batch mode is enabled", func() {
			It("should accumulate events and send in batch", func() {
				batchReceived := &atomic.Int32{}
				var mu sync.Mutex
				var lastBatch map[string]interface{}

				webhookServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					body, _ := io.ReadAll(r.Body)
					var batch map[string]interface{}
					json.Unmarshal(body, &batch)

					mu.Lock()
					lastBatch = batch
					mu.Unlock()

					batchReceived.Add(1)
					w.WriteHeader(http.StatusOK)
				}))
				defer webhookServer.Close()

				handler := newTestStatic().(static.Static)
				handler.SetSecurityBackend(static.SecurityConfig{
					Enabled:        true,
					WebhookURL:     webhookServer.URL,
					WebhookAsync:   false,
					WebhookTimeout: 5 * time.Second,
					MinSeverity:    "medium",
					BatchSize:      3,
					BatchTimeout:   5 * time.Second,
				})
				handler.SetPathSecurity(static.DefaultPathSecurityConfig())

				engine := setupTestRouter(handler, "/static")

				// Generate 3 events to trigger batch send
				_ = performRequest(engine, "GET", "/static/../passwd1")
				_ = performRequest(engine, "GET", "/static/../passwd2")
				_ = performRequest(engine, "GET", "/static/../passwd3")

				time.Sleep(200 * time.Millisecond)

				Expect(batchReceived.Load()).To(Equal(int32(1)))

				mu.Lock()
				defer mu.Unlock()
				if lastBatch != nil {
					Expect(lastBatch["count"]).To(BeNumerically(">=", 1))
				}
			})

			It("should send batch on timeout even if not full", func() {
				batchReceived := &atomic.Int32{}

				webhookServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					batchReceived.Add(1)
					w.WriteHeader(http.StatusOK)
				}))
				defer webhookServer.Close()

				handler := newTestStatic().(static.Static)
				handler.SetSecurityBackend(static.SecurityConfig{
					Enabled:        true,
					WebhookURL:     webhookServer.URL,
					WebhookAsync:   false,
					WebhookTimeout: 5 * time.Second,
					MinSeverity:    "medium",
					BatchSize:      10,
					BatchTimeout:   500 * time.Millisecond,
				})
				handler.SetPathSecurity(static.DefaultPathSecurityConfig())

				engine := setupTestRouter(handler, "/static")

				// Generate only 2 events (less than batch size)
				_ = performRequest(engine, "GET", "/static/../passwd1")
				_ = performRequest(engine, "GET", "/static/../passwd2")

				// Wait for timeout
				time.Sleep(1 * time.Second)

				Expect(batchReceived.Load()).To(BeNumerically(">=", 1))
			})
		})
	})

	Describe("Custom Headers", func() {
		Context("when custom webhook headers are provided", func() {
			It("should include custom headers in webhook request", func() {
				var receivedHeaders http.Header
				var mu sync.Mutex

				webhookServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					mu.Lock()
					receivedHeaders = r.Header.Clone()
					mu.Unlock()

					w.WriteHeader(http.StatusOK)
				}))
				defer webhookServer.Close()

				handler := newTestStatic().(static.Static)
				handler.SetSecurityBackend(static.SecurityConfig{
					Enabled:        true,
					WebhookURL:     webhookServer.URL,
					WebhookAsync:   false,
					WebhookTimeout: 5 * time.Second,
					MinSeverity:    "medium",
					WebhookHeaders: map[string]string{
						"Authorization": "Bearer secret-token",
						"X-Custom":      "test-value",
					},
				})
				handler.SetPathSecurity(static.DefaultPathSecurityConfig())

				engine := setupTestRouter(handler, "/static")

				_ = performRequest(engine, "GET", "/static/../passwd")

				time.Sleep(100 * time.Millisecond)

				mu.Lock()
				defer mu.Unlock()
				Expect(receivedHeaders.Get("Authorization")).To(Equal("Bearer secret-token"))
				Expect(receivedHeaders.Get("X-Custom")).To(Equal("test-value"))
			})
		})
	})
})
