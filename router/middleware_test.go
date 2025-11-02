/*
 * MIT License
 *
 * Copyright (c) 2019 Nicolas JUHEL
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

package router_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"time"

	ginsdk "github.com/gin-gonic/gin"
	liblog "github.com/nabbar/golib/logger"
	logcfg "github.com/nabbar/golib/logger/config"
	librtr "github.com/nabbar/golib/router"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Router/Middleware", func() {
	var (
		engine *ginsdk.Engine
		log    liblog.Logger
		lgf    = func() liblog.Logger {
			return log
		}
	)

	BeforeEach(func() {
		ginsdk.SetMode(ginsdk.TestMode)
		engine = ginsdk.New()
		log = liblog.New(context.Background())
		Expect(log.SetOptions(&logcfg.Options{
			Stdout: &logcfg.OptionsStd{
				DisableStandard: true,
			},
		})).ToNot(HaveOccurred())
	})

	Describe("GinLatencyContext", func() {
		It("should set start time in context", func() {
			var startTime int64

			engine.Use(librtr.GinLatencyContext)
			engine.GET("/test", func(c *ginsdk.Context) {
				startTime = c.GetInt64(librtr.GinContextStartUnixNanoTime)
				c.String(http.StatusOK, "ok")
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/test", nil)
			engine.ServeHTTP(w, req)

			Expect(startTime).To(BeNumerically(">", 0))
			Expect(w.Code).To(Equal(http.StatusOK))
		})

		It("should allow latency calculation", func() {
			var latency time.Duration

			engine.Use(librtr.GinLatencyContext)
			engine.GET("/test", func(c *ginsdk.Context) {
				time.Sleep(10 * time.Millisecond)
				start := time.Unix(0, c.GetInt64(librtr.GinContextStartUnixNanoTime))
				latency = time.Since(start)
				c.String(http.StatusOK, "ok")
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/test", nil)
			engine.ServeHTTP(w, req)

			Expect(latency).To(BeNumerically(">=", 10*time.Millisecond))
		})
	})

	Describe("GinRequestContext", func() {
		It("should set request path in context", func() {
			var requestPath string

			engine.Use(librtr.GinRequestContext)
			engine.GET("/test", func(c *ginsdk.Context) {
				requestPath = c.GetString(librtr.GinContextRequestPath)
				c.String(http.StatusOK, "ok")
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/test", nil)
			engine.ServeHTTP(w, req)

			Expect(requestPath).To(Equal("/test"))
			Expect(w.Code).To(Equal(http.StatusOK))
		})

		It("should set request path with query parameters", func() {
			var requestPath string

			engine.Use(librtr.GinRequestContext)
			engine.GET("/test", func(c *ginsdk.Context) {
				requestPath = c.GetString(librtr.GinContextRequestPath)
				c.String(http.StatusOK, "ok")
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/test?param=value&foo=bar", nil)
			engine.ServeHTTP(w, req)

			Expect(requestPath).To(ContainSubstring("/test"))
			Expect(requestPath).To(ContainSubstring("param=value"))
			Expect(requestPath).To(ContainSubstring("foo=bar"))
		})

		It("should set request user when present in URL", func() {
			var requestUser string

			engine.Use(librtr.GinRequestContext)
			engine.GET("/test", func(c *ginsdk.Context) {
				requestUser = c.GetString(librtr.GinContextRequestUser)
				c.String(http.StatusOK, "ok")
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "http://testuser@localhost/test", nil)
			engine.ServeHTTP(w, req)

			Expect(requestUser).To(Equal("testuser"))
		})

		It("should handle request without user", func() {
			var requestUser string

			engine.Use(librtr.GinRequestContext)
			engine.Use(librtr.GinAccessLog(lgf))
			engine.GET("/test", func(c *ginsdk.Context) {
				requestUser = c.GetString(librtr.GinContextRequestUser)
				c.String(http.StatusOK, "ok")
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/test", nil)
			engine.ServeHTTP(w, req)

			Expect(requestUser).To(Equal(""))
		})
	})

	Describe("GinAccessLog", func() {
		It("should log access when logger is provided", func() {
			engine.Use(librtr.GinLatencyContext)
			engine.Use(librtr.GinRequestContext)
			engine.Use(librtr.GinAccessLog(lgf))
			engine.GET("/test", func(c *ginsdk.Context) {
				c.String(http.StatusOK, "ok")
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/test", nil)
			engine.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusOK))
		})

		It("should not panic when logger is nil", func() {
			engine.Use(librtr.GinLatencyContext)
			engine.Use(librtr.GinRequestContext)
			engine.Use(librtr.GinAccessLog(nil))
			engine.GET("/test", func(c *ginsdk.Context) {
				c.String(http.StatusOK, "ok")
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/test", nil)

			Expect(func() {
				engine.ServeHTTP(w, req)
			}).ToNot(Panic())
			Expect(w.Code).To(Equal(http.StatusOK))
		})

		It("should not panic when logger func returns nil", func() {
			engine.Use(librtr.GinLatencyContext)
			engine.Use(librtr.GinRequestContext)
			engine.Use(librtr.GinAccessLog(func() liblog.Logger {
				return nil
			}))
			engine.Use(librtr.GinErrorLog(lgf))
			engine.GET("/test", func(c *ginsdk.Context) {
				c.String(http.StatusOK, "ok")
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/test", nil)

			Expect(func() {
				engine.ServeHTTP(w, req)
			}).ToNot(Panic())
			Expect(w.Code).To(Equal(http.StatusOK))
		})
	})

	Describe("GinErrorLog", func() {
		It("should log errors from context", func() {
			engine.Use(librtr.GinRequestContext)
			engine.Use(librtr.GinAccessLog(lgf))
			engine.Use(librtr.GinErrorLog(lgf))
			engine.GET("/test", func(c *ginsdk.Context) {
				c.Error(errors.New("test error"))
				c.String(http.StatusOK, "ok")
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/test", nil)
			engine.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusOK))
		})

		It("should recover from panic", func() {
			engine.Use(librtr.GinRequestContext)
			engine.Use(librtr.GinAccessLog(lgf))
			engine.Use(librtr.GinErrorLog(lgf))
			engine.GET("/test", func(c *ginsdk.Context) {
				panic("test panic")
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/test", nil)

			Expect(func() {
				engine.ServeHTTP(w, req)
			}).ToNot(Panic())
			Expect(w.Code).To(Equal(http.StatusInternalServerError))
		})

		It("should not panic when logger is nil", func() {
			engine.Use(librtr.GinRequestContext)
			engine.Use(librtr.GinAccessLog(lgf))
			engine.Use(librtr.GinErrorLog(nil))
			engine.GET("/test", func(c *ginsdk.Context) {
				c.Error(errors.New("test error"))
				c.String(http.StatusOK, "ok")
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/test", nil)

			Expect(func() {
				engine.ServeHTTP(w, req)
			}).ToNot(Panic())
			Expect(w.Code).To(Equal(http.StatusOK))
		})

		It("should not panic when logger func returns nil", func() {
			engine.Use(librtr.GinRequestContext)
			engine.Use(librtr.GinAccessLog(lgf))
			engine.Use(librtr.GinErrorLog(lgf))
			engine.GET("/test", func(c *ginsdk.Context) {
				c.Error(errors.New("test error"))
				c.String(http.StatusOK, "ok")
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/test", nil)

			Expect(func() {
				engine.ServeHTTP(w, req)
			}).ToNot(Panic())
			Expect(w.Code).To(Equal(http.StatusOK))
		})

		It("should handle panic recovery with broken pipe", func() {
			engine.Use(librtr.GinRequestContext)
			engine.Use(librtr.GinAccessLog(lgf))
			engine.Use(librtr.GinErrorLog(lgf))
			engine.GET("/test", func(c *ginsdk.Context) {
				// Simulate broken pipe error
				panic("write: broken pipe")
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/test", nil)

			Expect(func() {
				engine.ServeHTTP(w, req)
			}).ToNot(Panic())
		})
	})

	Describe("Middleware Integration", func() {
		It("should work with all middlewares together", func() {
			engine.Use(librtr.GinLatencyContext)
			engine.Use(librtr.GinRequestContext)
			engine.Use(librtr.GinAccessLog(lgf))
			engine.Use(librtr.GinErrorLog(lgf))
			engine.GET("/test", func(c *ginsdk.Context) {
				c.String(http.StatusOK, "ok")
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/test", nil)
			engine.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusOK))
			Expect(w.Body.String()).To(Equal("ok"))
		})

		It("should handle errors with all middlewares", func() {
			engine.Use(librtr.GinLatencyContext)
			engine.Use(librtr.GinRequestContext)
			engine.Use(librtr.GinAccessLog(lgf))
			engine.Use(librtr.GinErrorLog(lgf))
			engine.GET("/test", func(c *ginsdk.Context) {
				c.Error(errors.New("test error"))
				c.String(http.StatusBadRequest, "error")
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/test", nil)
			engine.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusBadRequest))
		})
	})
})
