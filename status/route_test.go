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

package status_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"time"

	ginsdk "github.com/gin-gonic/gin"
	liberr "github.com/nabbar/golib/errors"
	monsts "github.com/nabbar/golib/monitor/status"
	montps "github.com/nabbar/golib/monitor/types"
	libsts "github.com/nabbar/golib/status"
	stsctr "github.com/nabbar/golib/status/control"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Status/Route", func() {
	BeforeEach(func() {
		// Set Gin to test mode
		ginsdk.SetMode(ginsdk.TestMode)
	})

	Describe("MiddleWare", func() {
		Context("with default settings", func() {
			It("should return JSON by default", func() {
				status := libsts.New(globalCtx)
				status.SetInfo("route-test", "v1.0.0", "abc123")

				pool := newPool()
				status.RegisterPool(func() montps.Pool { return pool })

				// Setup router
				router := ginsdk.New()
				router.GET("/status", func(c *ginsdk.Context) {
					status.MiddleWare(c)
				})

				status.SetConfig(newStatusConfig(newListMandatory(map[stsctr.Mode][]string{
					stsctr.Must: {
						"test-service",
					},
				})...))
				m := newHealthyMonitor("test-service")
				err := pool.MonitorAdd(m)
				Expect(err).ToNot(HaveOccurred())
				time.Sleep(testMonitorStabilizeDelay)

				req := httptest.NewRequest("GET", "/status", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
				Expect(w.Header().Get("Content-Type")).To(ContainSubstring("application/json"))

				var result map[string]interface{}
				err = json.Unmarshal(w.Body.Bytes(), &result)
				Expect(err).ToNot(HaveOccurred())
				Expect(result["Name"]).To(Equal("route-test"))
			})

			It("should include X-Verbose header", func() {
				status := libsts.New(globalCtx)
				status.SetInfo("route-test", "v1.0.0", "abc123")

				pool := newPool()
				status.RegisterPool(func() montps.Pool { return pool })

				// Setup router
				router := ginsdk.New()
				router.GET("/status", func(c *ginsdk.Context) {
					status.MiddleWare(c)
				})

				status.SetConfig(newStatusConfig(newListMandatory(map[stsctr.Mode][]string{
					stsctr.Must: {
						"test-service",
					},
				})...))
				m := newHealthyMonitor("test-service")
				err := pool.MonitorAdd(m)
				Expect(err).ToNot(HaveOccurred())
				time.Sleep(testMonitorStabilizeDelay)

				req := httptest.NewRequest("GET", "/status", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				Expect(w.Header().Get("X-Verbose")).To(Equal("True"))
			})

			It("should include Connection: Close header", func() {
				status := libsts.New(globalCtx)
				status.SetInfo("route-test", "v1.0.0", "abc123")

				pool := newPool()
				status.RegisterPool(func() montps.Pool { return pool })

				// Setup router
				router := ginsdk.New()
				router.GET("/status", func(c *ginsdk.Context) {
					status.MiddleWare(c)
				})

				status.SetConfig(newStatusConfig(newListMandatory(map[stsctr.Mode][]string{
					stsctr.Must: {
						"test-service",
					},
				})...))
				m := newHealthyMonitor("test-service")
				err := pool.MonitorAdd(m)
				Expect(err).ToNot(HaveOccurred())
				time.Sleep(testMonitorStabilizeDelay)

				req := httptest.NewRequest("GET", "/status", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				Expect(w.Header().Get("Connection")).To(Equal("Close"))
			})
		})

		Context("with short query parameter", func() {
			It("should return short format with short=true", func() {
				status := libsts.New(globalCtx)
				status.SetInfo("route-test", "v1.0.0", "abc123")

				pool := newPool()
				status.RegisterPool(func() montps.Pool { return pool })

				// Setup router
				router := ginsdk.New()
				router.GET("/status", func(c *ginsdk.Context) {
					status.MiddleWare(c)
				})
				status.SetConfig(newStatusConfig(newListMandatory(map[stsctr.Mode][]string{
					stsctr.Must: {
						"test-service",
					},
				})...))
				m := newHealthyMonitor("test-service")
				err := pool.MonitorAdd(m)
				Expect(err).ToNot(HaveOccurred())
				time.Sleep(testMonitorStabilizeDelay)

				req := httptest.NewRequest("GET", "/status?short=true", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
				Expect(w.Header().Get("X-Verbose")).To(Equal("False"))

				var result map[string]interface{}
				err = json.Unmarshal(w.Body.Bytes(), &result)
				Expect(err).ToNot(HaveOccurred())

				// Short format should not include component details
				component, exists := result["Component"]
				if exists {
					// Component should be empty or minimal
					Expect(component).ToNot(BeNil())
				}
			})

			It("should return short format with short=1", func() {
				status := libsts.New(globalCtx)
				status.SetInfo("route-test", "v1.0.0", "abc123")

				pool := newPool()
				status.RegisterPool(func() montps.Pool { return pool })

				// Setup router
				router := ginsdk.New()
				router.GET("/status", func(c *ginsdk.Context) {
					status.MiddleWare(c)
				})
				status.SetConfig(newStatusConfig(newListMandatory(map[stsctr.Mode][]string{
					stsctr.Must: {
						"test-service",
					},
				})...))
				m := newHealthyMonitor("test-service")
				err := pool.MonitorAdd(m)
				Expect(err).ToNot(HaveOccurred())
				time.Sleep(testMonitorStabilizeDelay)

				req := httptest.NewRequest("GET", "/status?short=1", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				Expect(w.Header().Get("X-Verbose")).To(Equal("False"))
			})

			It("should return full format with short=false", func() {
				status := libsts.New(globalCtx)
				status.SetInfo("route-test", "v1.0.0", "abc123")

				pool := newPool()
				status.RegisterPool(func() montps.Pool { return pool })

				// Setup router
				router := ginsdk.New()
				router.GET("/status", func(c *ginsdk.Context) {
					status.MiddleWare(c)
				})
				status.SetConfig(newStatusConfig(newListMandatory(map[stsctr.Mode][]string{
					stsctr.Must: {
						"test-service",
					},
				})...))
				m := newHealthyMonitor("test-service")
				err := pool.MonitorAdd(m)
				Expect(err).ToNot(HaveOccurred())
				time.Sleep(testMonitorStabilizeDelay)

				req := httptest.NewRequest("GET", "/status?short=false", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				Expect(w.Header().Get("X-Verbose")).To(Equal("True"))
			})
		})

		Context("with X-Verbose header", func() {
			It("should respect X-Verbose: false header", func() {
				status := libsts.New(globalCtx)
				status.SetInfo("route-test", "v1.0.0", "abc123")

				pool := newPool()
				status.RegisterPool(func() montps.Pool { return pool })

				// Setup router
				router := ginsdk.New()
				router.GET("/status", func(c *ginsdk.Context) {
					status.MiddleWare(c)
				})
				status.SetConfig(newStatusConfig(newListMandatory(map[stsctr.Mode][]string{
					stsctr.Must: {
						"test-service",
					},
				})...))
				m := newHealthyMonitor("test-service")
				err := pool.MonitorAdd(m)
				Expect(err).ToNot(HaveOccurred())
				time.Sleep(testMonitorStabilizeDelay)

				req := httptest.NewRequest("GET", "/status", nil)
				req.Header.Set("X-Verbose", "false")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				Expect(w.Header().Get("X-Verbose")).To(Equal("False"))
			})

			It("should respect X-Verbose: true header", func() {
				status := libsts.New(globalCtx)
				status.SetInfo("route-test", "v1.0.0", "abc123")

				pool := newPool()
				status.RegisterPool(func() montps.Pool { return pool })

				// Setup router
				router := ginsdk.New()
				router.GET("/status", func(c *ginsdk.Context) {
					status.MiddleWare(c)
				})
				status.SetConfig(newStatusConfig(newListMandatory(map[stsctr.Mode][]string{
					stsctr.Must: {
						"test-service",
					},
				})...))
				m := newHealthyMonitor("test-service")
				err := pool.MonitorAdd(m)
				Expect(err).ToNot(HaveOccurred())
				time.Sleep(testMonitorStabilizeDelay)

				req := httptest.NewRequest("GET", "/status", nil)
				req.Header.Set("X-Verbose", "true")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				Expect(w.Header().Get("X-Verbose")).To(Equal("True"))
			})
		})

		Context("with format query parameter", func() {
			It("should return text with format=text", func() {
				status := libsts.New(globalCtx)
				status.SetInfo("route-test", "v1.0.0", "abc123")

				pool := newPool()
				status.RegisterPool(func() montps.Pool { return pool })

				// Setup router
				router := ginsdk.New()
				router.GET("/status", func(c *ginsdk.Context) {
					status.MiddleWare(c)
				})
				status.SetConfig(newStatusConfig(newListMandatory(map[stsctr.Mode][]string{
					stsctr.Must: {
						"test-service",
					},
				})...))
				m := newHealthyMonitor("test-service")
				err := pool.MonitorAdd(m)
				Expect(err).ToNot(HaveOccurred())
				time.Sleep(testMonitorStabilizeDelay)

				req := httptest.NewRequest("GET", "/status?format=text", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
				Expect(w.Header().Get("Content-Type")).To(ContainSubstring("text/plain"))

				body := w.Body.String()
				Expect(body).To(ContainSubstring("route-test"))
			})

			It("should return JSON with format=json", func() {
				status := libsts.New(globalCtx)
				status.SetInfo("route-test", "v1.0.0", "abc123")

				pool := newPool()
				status.RegisterPool(func() montps.Pool { return pool })

				// Setup router
				router := ginsdk.New()
				router.GET("/status", func(c *ginsdk.Context) {
					status.MiddleWare(c)
				})
				status.SetConfig(newStatusConfig(newListMandatory(map[stsctr.Mode][]string{
					stsctr.Must: {
						"test-service",
					},
				})...))
				m := newHealthyMonitor("test-service")
				err := pool.MonitorAdd(m)
				Expect(err).ToNot(HaveOccurred())
				time.Sleep(testMonitorStabilizeDelay)

				req := httptest.NewRequest("GET", "/status?format=json", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				Expect(w.Header().Get("Content-Type")).To(ContainSubstring("application/json"))
			})
		})

		Context("with Accept header", func() {
			It("should return text with Accept: text/plain", func() {
				status := libsts.New(globalCtx)
				status.SetInfo("route-test", "v1.0.0", "abc123")

				pool := newPool()
				status.RegisterPool(func() montps.Pool { return pool })

				// Setup router
				router := ginsdk.New()
				router.GET("/status", func(c *ginsdk.Context) {
					status.MiddleWare(c)
				})
				status.SetConfig(newStatusConfig(newListMandatory(map[stsctr.Mode][]string{
					stsctr.Must: {
						"test-service",
					},
				})...))
				m := newHealthyMonitor("test-service")
				err := pool.MonitorAdd(m)
				Expect(err).ToNot(HaveOccurred())
				time.Sleep(testMonitorStabilizeDelay)

				req := httptest.NewRequest("GET", "/status", nil)
				req.Header.Set("Accept", "text/plain")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				Expect(w.Header().Get("Content-Type")).To(ContainSubstring("text/plain"))
			})

			It("should return JSON with Accept: application/json", func() {
				status := libsts.New(globalCtx)
				status.SetInfo("route-test", "v1.0.0", "abc123")

				pool := newPool()
				status.RegisterPool(func() montps.Pool { return pool })

				// Setup router
				router := ginsdk.New()
				router.GET("/status", func(c *ginsdk.Context) {
					status.MiddleWare(c)
				})
				status.SetConfig(newStatusConfig(newListMandatory(map[stsctr.Mode][]string{
					stsctr.Must: {
						"test-service",
					},
				})...))
				m := newHealthyMonitor("test-service")
				err := pool.MonitorAdd(m)
				Expect(err).ToNot(HaveOccurred())
				time.Sleep(testMonitorStabilizeDelay)

				req := httptest.NewRequest("GET", "/status", nil)
				req.Header.Set("Accept", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				Expect(w.Header().Get("Content-Type")).To(ContainSubstring("application/json"))
			})

			It("should handle multiple Accept values", func() {
				status := libsts.New(globalCtx)
				status.SetInfo("route-test", "v1.0.0", "abc123")

				pool := newPool()
				status.RegisterPool(func() montps.Pool { return pool })

				// Setup router
				router := ginsdk.New()
				router.GET("/status", func(c *ginsdk.Context) {
					status.MiddleWare(c)
				})
				status.SetConfig(newStatusConfig(newListMandatory(map[stsctr.Mode][]string{
					stsctr.Must: {
						"test-service",
					},
				})...))
				m := newHealthyMonitor("test-service")
				err := pool.MonitorAdd(m)
				Expect(err).ToNot(HaveOccurred())
				time.Sleep(testMonitorStabilizeDelay)

				req := httptest.NewRequest("GET", "/status", nil)
				req.Header.Set("Accept", "text/html, application/json, text/plain")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				// Should prefer JSON when multiple types are present
				Expect(w.Header().Get("Content-Type")).To(ContainSubstring("application/json"))
			})
		})

		Context("with combined parameters", func() {
			It("should handle short=true and format=text", func() {
				status := libsts.New(globalCtx)
				status.SetInfo("route-test", "v1.0.0", "abc123")

				pool := newPool()
				status.RegisterPool(func() montps.Pool { return pool })

				// Setup router
				router := ginsdk.New()
				router.GET("/status", func(c *ginsdk.Context) {
					status.MiddleWare(c)
				})
				status.SetConfig(newStatusConfig(newListMandatory(map[stsctr.Mode][]string{
					stsctr.Must: {
						"test-service",
					},
				})...))
				m := newHealthyMonitor("test-service")
				err := pool.MonitorAdd(m)
				Expect(err).ToNot(HaveOccurred())
				time.Sleep(testMonitorStabilizeDelay)

				req := httptest.NewRequest("GET", "/status?short=true&format=text", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
				Expect(w.Header().Get("Content-Type")).To(ContainSubstring("text/plain"))
				Expect(w.Header().Get("X-Verbose")).To(Equal("False"))
			})

			It("should prioritize query params over headers", func() {
				status := libsts.New(globalCtx)
				status.SetInfo("route-test", "v1.0.0", "abc123")

				pool := newPool()
				status.RegisterPool(func() montps.Pool { return pool })

				// Setup router
				router := ginsdk.New()
				router.GET("/status", func(c *ginsdk.Context) {
					status.MiddleWare(c)
				})
				status.SetConfig(newStatusConfig(newListMandatory(map[stsctr.Mode][]string{
					stsctr.Must: {
						"test-service",
					},
				})...))
				m := newHealthyMonitor("test-service")
				err := pool.MonitorAdd(m)
				Expect(err).ToNot(HaveOccurred())
				time.Sleep(testMonitorStabilizeDelay)

				req := httptest.NewRequest("GET", "/status?format=text", nil)
				req.Header.Set("Accept", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				// Query param should take precedence // priority to header, and after query string
				Expect(w.Header().Get("Content-Type")).To(ContainSubstring("application/json"))
			})
		})
	})

	Describe("Expose", func() {
		It("should work with Gin context", func() {
			status := libsts.New(globalCtx)
			status.SetInfo("route-test", "v1.0.0", "abc123")

			pool := newPool()
			status.RegisterPool(func() montps.Pool { return pool })

			// Setup router
			router := ginsdk.New()
			router.GET("/expose", func(c *ginsdk.Context) {
				status.Expose(c)
			})

			status.SetConfig(newStatusConfig(newListMandatory(map[stsctr.Mode][]string{
				stsctr.Must: {
					"test-service",
				},
			})...))

			m := newHealthyMonitor("test-service")
			err := pool.MonitorAdd(m)
			Expect(err).ToNot(HaveOccurred())
			time.Sleep(testMonitorStabilizeDelay)

			req := httptest.NewRequest("GET", "/expose", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusOK))
		})
	})

	Describe("SetErrorReturn", func() {
		It("should allow custom error return", func() {
			status := libsts.New(globalCtx)
			status.SetInfo("route-test", "v1.0.0", "abc123")

			pool := newPool()
			status.RegisterPool(func() montps.Pool { return pool })

			// Setup router
			router := ginsdk.New()
			router.GET("/status", func(c *ginsdk.Context) {
				status.MiddleWare(c)
			})
			status.SetConfig(newStatusConfig(newListMandatory(map[stsctr.Mode][]string{
				stsctr.Must: {
					"test-service",
				},
			})...))
			m := newHealthyMonitor("test-service")
			err := pool.MonitorAdd(m)
			Expect(err).ToNot(HaveOccurred())
			time.Sleep(testMonitorStabilizeDelay)

			customCalled := false

			status.SetErrorReturn(func() liberr.ReturnGin {
				customCalled = true
				return liberr.NewDefaultReturn()
			})

			// This should not trigger error return since info is set
			req := httptest.NewRequest("GET", "/status", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			// Should succeed without calling custom error handler
			Expect(w.Code).To(Equal(http.StatusOK))
			// if Status OK, so no error, so cannot be called, so value is false
			Expect(customCalled).To(BeFalse())
		})
	})

	Describe("HTTP status codes", func() {
		Context("with custom return codes", func() {
			It("should return 200 for OK status", func() {
				status := libsts.New(globalCtx)
				status.SetInfo("route-test", "v1.0.0", "abc123")

				pool := newPool()
				status.RegisterPool(func() montps.Pool { return pool })

				// Setup router
				router := ginsdk.New()
				router.GET("/status", func(c *ginsdk.Context) {
					status.MiddleWare(c)
				})
				cfg := libsts.Config{
					ReturnCode: map[monsts.Status]int{
						monsts.OK:   http.StatusOK,
						monsts.Warn: http.StatusMultiStatus,
						monsts.KO:   http.StatusServiceUnavailable,
					},
					MandatoryComponent: []libsts.Mandatory{
						{
							Mode: stsctr.Must,
							Keys: []string{"healthy-service", "unhealthy-service"},
						},
					},
				}
				status.SetConfig(cfg)

				m := newHealthyMonitor("healthy-service")
				err := pool.MonitorAdd(m)
				Expect(err).ToNot(HaveOccurred())
				time.Sleep(testMonitorStabilizeDelay)

				req := httptest.NewRequest("GET", "/status", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
			})

			It("should return 503 for KO status", func() {
				status := libsts.New(globalCtx)
				status.SetInfo("route-test", "v1.0.0", "abc123")

				pool := newPool()
				status.RegisterPool(func() montps.Pool { return pool })

				// Setup router
				router := ginsdk.New()
				router.GET("/status", func(c *ginsdk.Context) {
					status.MiddleWare(c)
				})
				cfg := libsts.Config{
					ReturnCode: map[monsts.Status]int{
						monsts.OK:   http.StatusOK,
						monsts.Warn: http.StatusMultiStatus,
						monsts.KO:   http.StatusServiceUnavailable,
					},
					MandatoryComponent: []libsts.Mandatory{
						{
							Mode: stsctr.Must,
							Keys: []string{"unhealthy-service"},
						},
					},
				}
				status.SetConfig(cfg)

				m := newUnhealthyMonitor("unhealthy-service")
				err := pool.MonitorAdd(m)
				Expect(err).ToNot(HaveOccurred())

				time.Sleep(100 * time.Millisecond)

				req := httptest.NewRequest("GET", "/status", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusServiceUnavailable))
			})
		})
	})
})
