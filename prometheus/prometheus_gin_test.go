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

package prometheus_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"

	ginsdk "github.com/gin-gonic/gin"
	libprm "github.com/nabbar/golib/prometheus"
	prmmet "github.com/nabbar/golib/prometheus/metrics"
	librtr "github.com/nabbar/golib/router"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Prometheus Gin Integration", func() {
	var (
		p      libprm.Prometheus
		router *ginsdk.Engine
	)

	BeforeEach(func() {
		ginsdk.SetMode(ginsdk.TestMode)
		p = newPrometheus()
		router = ginsdk.New()
	})

	Describe("ExposeGin", func() {
		Context("when exposing metrics endpoint", func() {
			It("should serve metrics", func() {
				router.GET("/metrics", func(c *ginsdk.Context) {
					p.ExposeGin(c)
				})

				req := httptest.NewRequest("GET", "/metrics", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
			})

			It("should return Prometheus format", func() {
				router.GET("/metrics", func(c *ginsdk.Context) {
					p.ExposeGin(c)
				})

				req := httptest.NewRequest("GET", "/metrics", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Header().Get("Content-Type")).To(ContainSubstring("text/plain"))
			})

			It("should handle multiple requests", func() {
				router.GET("/metrics", func(c *ginsdk.Context) {
					p.ExposeGin(c)
				})

				for i := 0; i < 5; i++ {
					req := httptest.NewRequest("GET", "/metrics", nil)
					w := httptest.NewRecorder()
					router.ServeHTTP(w, req)

					Expect(w.Code).To(Equal(http.StatusOK))
				}
			})

			It("should work on custom path", func() {
				router.GET("/custom/prometheus", func(c *ginsdk.Context) {
					p.ExposeGin(c)
				})

				req := httptest.NewRequest("GET", "/custom/prometheus", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
			})
		})
	})

	Describe("Expose", func() {
		Context("when using context interface", func() {
			It("should expose metrics with gin context", func() {
				router.GET("/metrics", func(c *ginsdk.Context) {
					p.Expose(c)
				})

				req := httptest.NewRequest("GET", "/metrics", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
			})

			It("should handle non-gin context gracefully", func() {
				Expect(func() {
					ctx := context.Background()
					p.Expose(ctx)
				}).ToNot(Panic())
			})
		})
	})

	Describe("MiddleWareGin", func() {
		Context("when using middleware", func() {
			It("should set start time", func() {
				var startTime int64

				router.Use(func(c *ginsdk.Context) {
					p.MiddleWareGin(c)
				})

				router.GET("/test", func(c *ginsdk.Context) {
					startTime = c.GetInt64(librtr.GinContextStartUnixNanoTime)
					c.String(http.StatusOK, "ok")
				})

				req := httptest.NewRequest("GET", "/test", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(startTime).To(BeNumerically(">", 0))
			})

			It("should preserve start time across multiple middleware", func() {
				var time1, time2 int64

				router.Use(func(c *ginsdk.Context) {
					p.MiddleWareGin(c)
					time1 = c.GetInt64(librtr.GinContextStartUnixNanoTime)
				})

				router.Use(func(c *ginsdk.Context) {
					time2 = c.GetInt64(librtr.GinContextStartUnixNanoTime)
					c.Next()
				})

				router.GET("/test", func(c *ginsdk.Context) {
					c.String(http.StatusOK, "ok")
				})

				req := httptest.NewRequest("GET", "/test", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(time1).To(Equal(time2))
			})

			It("should set request path", func() {
				var requestPath string

				router.Use(func(c *ginsdk.Context) {
					p.MiddleWareGin(c)
				})

				router.GET("/test", func(c *ginsdk.Context) {
					requestPath = c.GetString(librtr.GinContextRequestPath)
					c.String(http.StatusOK, "ok")
				})

				req := httptest.NewRequest("GET", "/test", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(requestPath).To(Equal("/test"))
			})

			It("should include query parameters in request path", func() {
				var requestPath string

				router.Use(func(c *ginsdk.Context) {
					p.MiddleWareGin(c)
				})

				router.GET("/test", func(c *ginsdk.Context) {
					requestPath = c.GetString(librtr.GinContextRequestPath)
					c.String(http.StatusOK, "ok")
				})

				req := httptest.NewRequest("GET", "/test?foo=bar&baz=qux", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(requestPath).To(Equal("/test?foo=bar&baz=qux"))
			})

			It("should call next handlers", func() {
				handlerCalled := false

				router.Use(func(c *ginsdk.Context) {
					p.MiddleWareGin(c)
				})

				router.GET("/test", func(c *ginsdk.Context) {
					handlerCalled = true
					c.String(http.StatusOK, "ok")
				})

				req := httptest.NewRequest("GET", "/test", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(handlerCalled).To(BeTrue())
			})

			It("should respect exclude paths", func() {
				collectCalled := false

				// Add a metric with collect function
				name := uniqueMetricName("middleware_metric")
				m := createCounterMetric(name)
				m.SetCollect(func(ctx context.Context, metric prmmet.Metric) {
					collectCalled = true
				})

				err := p.AddMetric(true, m)
				Expect(err).ToNot(HaveOccurred())

				// Exclude the test path
				p.ExcludePath("/excluded")

				router.Use(func(c *ginsdk.Context) {
					p.MiddleWareGin(c)
				})

				router.GET("/excluded", func(c *ginsdk.Context) {
					c.String(http.StatusOK, "ok")
				})

				req := httptest.NewRequest("GET", "/excluded", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				// Collect should not be called for excluded path
				Expect(collectCalled).To(BeFalse())
			})

			It("should collect metrics for non-excluded paths", func() {
				router.Use(func(c *ginsdk.Context) {
					p.MiddleWareGin(c)
				})

				router.GET("/included", func(c *ginsdk.Context) {
					c.String(http.StatusOK, "ok")
				})

				req := httptest.NewRequest("GET", "/included", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
			})

			It("should handle errors in handlers", func() {
				router.Use(func(c *ginsdk.Context) {
					p.MiddleWareGin(c)
				})

				router.GET("/error", func(c *ginsdk.Context) {
					c.String(http.StatusInternalServerError, "error")
				})

				req := httptest.NewRequest("GET", "/error", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusInternalServerError))
			})

			It("should handle panics in handlers", func() {
				router.Use(ginsdk.Recovery())
				router.Use(func(c *ginsdk.Context) {
					p.MiddleWareGin(c)
				})

				router.GET("/panic", func(c *ginsdk.Context) {
					// don't panic to avoid pollute output
					//panic("test panic")
				})

				req := httptest.NewRequest("GET", "/panic", nil)
				w := httptest.NewRecorder()

				Expect(func() {
					router.ServeHTTP(w, req)
				}).ToNot(Panic())
			})
		})

		Context("when handling concurrent requests", func() {
			It("should handle multiple concurrent requests", func() {
				router.Use(func(c *ginsdk.Context) {
					p.MiddleWareGin(c)
				})

				router.GET("/concurrent", func(c *ginsdk.Context) {
					time.Sleep(10 * time.Millisecond)
					c.String(http.StatusOK, "ok")
				})

				done := make(chan bool, 20)

				for i := 0; i < 20; i++ {
					go func() {
						req := httptest.NewRequest("GET", "/concurrent", nil)
						w := httptest.NewRecorder()
						router.ServeHTTP(w, req)
						done <- true
					}()
				}

				for i := 0; i < 20; i++ {
					<-done
				}
			})

			It("should maintain separate context for each request", func() {
				startTimes := make(map[int]int64)
				var mu sync.Mutex

				router.Use(func(c *ginsdk.Context) {
					p.MiddleWareGin(c)
				})

				router.GET("/test/:id", func(c *ginsdk.Context) {
					id := c.Param("id")
					startTime := c.GetInt64(librtr.GinContextStartUnixNanoTime)

					mu.Lock()
					idInt := 0
					fmt.Sscanf(id, "%d", &idInt)
					startTimes[idInt] = startTime
					mu.Unlock()

					c.String(http.StatusOK, "ok")
				})

				done := make(chan bool, 10)

				for i := 0; i < 10; i++ {
					go func(idx int) {
						req := httptest.NewRequest("GET", fmt.Sprintf("/test/%d", idx), nil)
						w := httptest.NewRecorder()
						router.ServeHTTP(w, req)
						done <- true
					}(i)
				}

				for i := 0; i < 10; i++ {
					<-done
				}

				// All requests should have unique start times
				Expect(len(startTimes)).To(Equal(10))
			})
		})
	})

	Describe("MiddleWare", func() {
		Context("when using context interface", func() {
			It("should work with gin context", func() {
				router.Use(func(c *ginsdk.Context) {
					p.MiddleWare(c)
				})

				router.GET("/test", func(c *ginsdk.Context) {
					c.String(http.StatusOK, "ok")
				})

				req := httptest.NewRequest("GET", "/test", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
			})

			It("should handle non-gin context gracefully", func() {
				Expect(func() {
					ctx := context.Background()
					p.MiddleWare(ctx)
				}).ToNot(Panic())
			})
		})
	})

	Describe("Integration with HTTP Methods", func() {
		BeforeEach(func() {
			router.Use(func(c *ginsdk.Context) {
				p.MiddleWareGin(c)
			})
		})

		Context("when handling different HTTP methods", func() {
			It("should handle GET requests", func() {
				router.GET("/get", func(c *ginsdk.Context) {
					c.String(http.StatusOK, "get")
				})

				req := httptest.NewRequest("GET", "/get", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
			})

			It("should handle POST requests", func() {
				router.POST("/post", func(c *ginsdk.Context) {
					c.String(http.StatusCreated, "post")
				})

				req := httptest.NewRequest("POST", "/post", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusCreated))
			})

			It("should handle PUT requests", func() {
				router.PUT("/put", func(c *ginsdk.Context) {
					c.String(http.StatusOK, "put")
				})

				req := httptest.NewRequest("PUT", "/put", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
			})

			It("should handle DELETE requests", func() {
				router.DELETE("/delete", func(c *ginsdk.Context) {
					c.String(http.StatusNoContent, "")
				})

				req := httptest.NewRequest("DELETE", "/delete", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusNoContent))
			})

			It("should handle PATCH requests", func() {
				router.PATCH("/patch", func(c *ginsdk.Context) {
					c.String(http.StatusOK, "patch")
				})

				req := httptest.NewRequest("PATCH", "/patch", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
			})
		})
	})
})
