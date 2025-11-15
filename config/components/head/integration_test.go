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

package head_test

import (
	"context"
	"net/http"
	"net/http/httptest"

	. "github.com/nabbar/golib/config/components/head"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	ginsdk "github.com/gin-gonic/gin"
	libcfg "github.com/nabbar/golib/config"
	cfgtps "github.com/nabbar/golib/config/types"
	liblog "github.com/nabbar/golib/logger"
	libver "github.com/nabbar/golib/version"
	libvpr "github.com/nabbar/golib/viper"
)

// Integration tests verify end-to-end scenarios
var _ = Describe("Integration Tests", func() {
	var (
		ctx context.Context
		cfg libcfg.Config
	)

	BeforeEach(func() {
		ctx = context.Background()
		cfg = libcfg.New(nil)
	})

	AfterEach(func() {
		if cfg != nil {
			cfg.Stop()
		}
	})

	Describe("Complete Component Lifecycle", func() {
		Context("full workflow", func() {
			It("should handle complete lifecycle: create, init, start, use, stop", func() {
				// Create
				cpt := New(ctx)
				Expect(cpt).NotTo(BeNil())

				// Register
				key := "integration-head"
				Register(cfg, key, cpt)
				Expect(cfg.ComponentHas(key)).To(BeTrue())

				// Setup configuration
				vpr := libvpr.New(ctx, func() liblog.Logger { return nil })
				vpr.Viper().Set(key, map[string]string{
					"X-Custom-Header": "custom-value",
					"X-API-Version":   "v1.0.0",
				})

				// Initialize
				cpt.Init(key, ctx,
					func(string) cfgtps.Component { return nil },
					func() libvpr.Viper { return vpr },
					libver.NewVersion(libver.License_MIT, "test", "", "01/01/1970", "abcd1234", "1.0.0", "maintainer", "", Empty{}, 0),
					func() liblog.Logger { return nil })

				// Start
				err := cpt.Start()
				Expect(err).To(BeNil())
				Expect(cpt.IsStarted()).To(BeTrue())

				// Use
				headers := cpt.GetHeaders()
				Expect(headers.Get("X-Custom-Header")).To(Equal("custom-value"))
				Expect(headers.Get("X-API-Version")).To(Equal("v1.0.0"))

				// Stop
				cpt.Stop()
				Expect(cpt.IsStarted()).To(BeFalse())
			})

			It("should handle multiple start-stop cycles", func() {
				cpt := New(ctx)
				key := "cycle-head"

				vpr := libvpr.New(ctx, func() liblog.Logger { return nil })
				vpr.Viper().Set(key, map[string]string{
					"X-Test": "value",
				})

				cpt.Init(key, ctx,
					func(string) cfgtps.Component { return nil },
					func() libvpr.Viper { return vpr },
					libver.NewVersion(libver.License_MIT, "test", "", "01/01/1970", "abcd1234", "1.0.0", "maintainer", "", Empty{}, 0),
					func() liblog.Logger { return nil })

				// Cycle 1
				Expect(cpt.Start()).To(Succeed())
				Expect(cpt.IsStarted()).To(BeTrue())
				cpt.Stop()
				Expect(cpt.IsStarted()).To(BeFalse())

				// Cycle 2
				Expect(cpt.Start()).To(Succeed())
				Expect(cpt.IsStarted()).To(BeTrue())
				cpt.Stop()
				Expect(cpt.IsStarted()).To(BeFalse())

				// Cycle 3
				Expect(cpt.Start()).To(Succeed())
				Expect(cpt.IsStarted()).To(BeTrue())
				cpt.Stop()
				Expect(cpt.IsStarted()).To(BeFalse())
			})
		})

		Context("with callbacks", func() {
			It("should execute all callbacks in correct order", func() {
				cpt := New(ctx)
				key := "callback-head"

				vpr := libvpr.New(ctx, func() liblog.Logger { return nil })
				vpr.Viper().Set(key, map[string]string{
					"X-Test": "value",
				})

				cpt.Init(key, ctx,
					func(string) cfgtps.Component { return nil },
					func() libvpr.Viper { return vpr },
					libver.NewVersion(libver.License_MIT, "test", "", "01/01/1970", "abcd1234", "1.0.0", "maintainer", "", Empty{}, 0),
					func() liblog.Logger { return nil })

				// Track callback execution
				var callOrder []string

				cpt.RegisterFuncStart(
					func(c cfgtps.Component) error {
						callOrder = append(callOrder, "start-before")
						return nil
					},
					func(c cfgtps.Component) error {
						callOrder = append(callOrder, "start-after")
						return nil
					},
				)

				cpt.RegisterFuncReload(
					func(c cfgtps.Component) error {
						callOrder = append(callOrder, "reload-before")
						return nil
					},
					func(c cfgtps.Component) error {
						callOrder = append(callOrder, "reload-after")
						return nil
					},
				)

				// Start
				err := cpt.Start()
				Expect(err).To(BeNil())
				Expect(callOrder).To(Equal([]string{"start-before", "start-after"}))

				// Reload
				callOrder = []string{}
				err = cpt.Reload()
				Expect(err).To(BeNil())
				Expect(callOrder).To(Equal([]string{"reload-before", "reload-after"}))
			})

			It("should handle callback errors gracefully", func() {
				cpt := New(ctx)
				key := "error-callback-head"

				vpr := libvpr.New(ctx, func() liblog.Logger { return nil })
				vpr.Viper().Set(key, map[string]string{
					"X-Test": "value",
				})

				cpt.Init(key, ctx,
					func(string) cfgtps.Component { return nil },
					func() libvpr.Viper { return vpr },
					libver.NewVersion(libver.License_MIT, "test", "", "01/01/1970", "abcd1234", "1.0.0", "maintainer", "", Empty{}, 0),
					func() liblog.Logger { return nil })

				// Register callback that returns error
				cpt.RegisterFuncStart(
					func(c cfgtps.Component) error {
						return ErrorParamInvalid.Error(nil)
					},
					nil,
				)

				// Start should fail
				err := cpt.Start()
				Expect(err).NotTo(BeNil())
			})
		})
	})

	Describe("Multi-Component Scenarios", func() {
		Context("with multiple head components", func() {
			It("should manage multiple independent components", func() {
				keys := []string{"head-api", "head-web", "head-admin"}

				for _, key := range keys {
					RegisterNew(ctx, cfg, key)

					vpr := libvpr.New(ctx, func() liblog.Logger { return nil })
					vpr.Viper().Set(key, map[string]string{
						"X-Component": key,
					})

					getCpt := func(k string) cfgtps.Component {
						return cfg.ComponentGet(k)
					}

					cpt := Load(getCpt, key)
					cpt.Init(key, ctx,
						getCpt,
						func() libvpr.Viper { return vpr },
						libver.NewVersion(libver.License_MIT, "test", "", "01/01/1970", "abcd1234", "1.0.0", "maintainer", "", Empty{}, 0),
						func() liblog.Logger { return nil })

					err := cpt.Start()
					Expect(err).To(BeNil())
				}

				// Verify all are independent
				for _, key := range keys {
					getCpt := func(k string) cfgtps.Component {
						return cfg.ComponentGet(k)
					}
					cpt := Load(getCpt, key)
					Expect(cpt).NotTo(BeNil())
					Expect(cpt.IsStarted()).To(BeTrue())

					headers := cpt.GetHeaders()
					Expect(headers.Get("X-Component")).To(Equal(key))
				}
			})

			It("should handle independent start/stop of multiple components", func() {
				cpt1 := New(ctx)
				cpt2 := New(ctx)

				key1 := "head-1"
				key2 := "head-2"

				vpr1 := libvpr.New(ctx, func() liblog.Logger { return nil })
				vpr1.Viper().Set(key1, map[string]string{"X-Test": "value1"})

				vpr2 := libvpr.New(ctx, func() liblog.Logger { return nil })
				vpr2.Viper().Set(key2, map[string]string{"X-Test": "value2"})

				cpt1.Init(key1, ctx, func(string) cfgtps.Component { return nil },
					func() libvpr.Viper { return vpr1 },
					libver.NewVersion(libver.License_MIT, "test", "", "01/01/1970", "abcd1234", "1.0.0", "maintainer", "", Empty{}, 0),
					func() liblog.Logger { return nil })

				cpt2.Init(key2, ctx, func(string) cfgtps.Component { return nil },
					func() libvpr.Viper { return vpr2 },
					libver.NewVersion(libver.License_MIT, "test", "", "01/01/1970", "abcd1234", "1.0.0", "maintainer", "", Empty{}, 0),
					func() liblog.Logger { return nil })

				// Start both
				Expect(cpt1.Start()).To(Succeed())
				Expect(cpt2.Start()).To(Succeed())

				Expect(cpt1.IsStarted()).To(BeTrue())
				Expect(cpt2.IsStarted()).To(BeTrue())

				// Stop only first
				cpt1.Stop()

				Expect(cpt1.IsStarted()).To(BeFalse())
				Expect(cpt2.IsStarted()).To(BeTrue())

				// Stop second
				cpt2.Stop()

				Expect(cpt1.IsStarted()).To(BeFalse())
				Expect(cpt2.IsStarted()).To(BeFalse())
			})
		})
	})

	Describe("HTTP Server Integration", func() {
		Context("with Gin framework", func() {
			It("should integrate with Gin router", func() {
				// Setup component
				cpt := New(ctx)
				key := "http-head"

				vpr := libvpr.New(ctx, func() liblog.Logger { return nil })
				vpr.Viper().Set(key, map[string]string{
					"X-Custom-Header": "custom-value",
					"X-Frame-Options": "DENY",
					"X-API-Version":   "v1.0.0",
				})

				cpt.Init(key, ctx,
					func(string) cfgtps.Component { return nil },
					func() libvpr.Viper { return vpr },
					libver.NewVersion(libver.License_MIT, "test", "", "01/01/1970", "abcd1234", "1.0.0", "maintainer", "", Empty{}, 0),
					func() liblog.Logger { return nil })

				err := cpt.Start()
				Expect(err).To(BeNil())

				// Setup Gin
				ginsdk.SetMode(ginsdk.TestMode)
				router := ginsdk.New()

				// Get headers and register middleware
				headers := cpt.GetHeaders()
				router.Use(headers.Handler)

				router.GET("/test", func(c *ginsdk.Context) {
					c.JSON(200, ginsdk.H{"message": "ok"})
				})

				// Test request
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/test", nil)
				router.ServeHTTP(w, req)

				// Verify response
				Expect(w.Code).To(Equal(200))
				Expect(w.Header().Get("X-Custom-Header")).To(Equal("custom-value"))
				Expect(w.Header().Get("X-Frame-Options")).To(Equal("DENY"))
				Expect(w.Header().Get("X-API-Version")).To(Equal("v1.0.0"))
			})

			It("should work with multiple routes", func() {
				cpt := New(ctx)
				key := "multi-route-head"

				vpr := libvpr.New(ctx, func() liblog.Logger { return nil })
				vpr.Viper().Set(key, map[string]string{
					"X-Test-Header": "test-value",
				})

				cpt.Init(key, ctx,
					func(string) cfgtps.Component { return nil },
					func() libvpr.Viper { return vpr },
					libver.NewVersion(libver.License_MIT, "test", "", "01/01/1970", "abcd1234", "1.0.0", "maintainer", "", Empty{}, 0),
					func() liblog.Logger { return nil })

				err := cpt.Start()
				Expect(err).To(BeNil())

				ginsdk.SetMode(ginsdk.TestMode)
				router := ginsdk.New()

				headers := cpt.GetHeaders()
				router.Use(headers.Handler)

				router.GET("/route1", func(c *ginsdk.Context) {
					c.JSON(200, ginsdk.H{"route": "1"})
				})
				router.POST("/route2", func(c *ginsdk.Context) {
					c.JSON(200, ginsdk.H{"route": "2"})
				})
				router.PUT("/route3", func(c *ginsdk.Context) {
					c.JSON(200, ginsdk.H{"route": "3"})
				})

				// Test all routes
				routes := []struct {
					method string
					path   string
				}{
					{"GET", "/route1"},
					{"POST", "/route2"},
					{"PUT", "/route3"},
				}

				for _, route := range routes {
					w := httptest.NewRecorder()
					req, _ := http.NewRequest(route.method, route.path, nil)
					router.ServeHTTP(w, req)

					Expect(w.Code).To(Equal(200))
					Expect(w.Header().Get("X-Test-Header")).To(Equal("test-value"))
				}
			})

			It("should apply headers before other handlers", func() {
				cpt := New(ctx)
				key := "order-head"

				vpr := libvpr.New(ctx, func() liblog.Logger { return nil })
				vpr.Viper().Set(key, map[string]string{
					"X-Before": "set-by-middleware",
				})

				cpt.Init(key, ctx,
					func(string) cfgtps.Component { return nil },
					func() libvpr.Viper { return vpr },
					libver.NewVersion(libver.License_MIT, "test", "", "01/01/1970", "abcd1234", "1.0.0", "maintainer", "", Empty{}, 0),
					func() liblog.Logger { return nil })

				err := cpt.Start()
				Expect(err).To(BeNil())

				ginsdk.SetMode(ginsdk.TestMode)
				router := ginsdk.New()

				headers := cpt.GetHeaders()
				router.Use(headers.Handler)

				router.GET("/test", func(c *ginsdk.Context) {
					// Try to override - should already be set
					c.Header("X-After", "set-by-handler")
					c.JSON(200, ginsdk.H{"message": "ok"})
				})

				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/test", nil)
				router.ServeHTTP(w, req)

				Expect(w.Header().Get("X-Before")).To(Equal("set-by-middleware"))
				Expect(w.Header().Get("X-After")).To(Equal("set-by-handler"))
			})
		})

		Context("security scenarios", func() {
			It("should set security headers for API endpoint", func() {
				cpt := New(ctx)
				key := "secure-api-head"

				vpr := libvpr.New(ctx, func() liblog.Logger { return nil })
				vpr.Viper().Set(key, map[string]string{
					"X-Content-Type-Options":    "nosniff",
					"X-Frame-Options":           "DENY",
					"Strict-Transport-Security": "max-age=31536000",
					"Content-Security-Policy":   "default-src 'self'",
					"X-XSS-Protection":          "1; mode=block",
					"Referrer-Policy":           "no-referrer",
				})

				cpt.Init(key, ctx,
					func(string) cfgtps.Component { return nil },
					func() libvpr.Viper { return vpr },
					libver.NewVersion(libver.License_MIT, "test", "", "01/01/1970", "abcd1234", "1.0.0", "maintainer", "", Empty{}, 0),
					func() liblog.Logger { return nil })

				err := cpt.Start()
				Expect(err).To(BeNil())

				ginsdk.SetMode(ginsdk.TestMode)
				router := ginsdk.New()

				headers := cpt.GetHeaders()
				router.Use(headers.Handler)

				router.GET("/api/data", func(c *ginsdk.Context) {
					c.JSON(200, ginsdk.H{"data": "secure"})
				})

				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/api/data", nil)
				router.ServeHTTP(w, req)

				// Verify all security headers are present
				Expect(w.Header().Get("X-Content-Type-Options")).To(Equal("nosniff"))
				Expect(w.Header().Get("X-Frame-Options")).To(Equal("DENY"))
				Expect(w.Header().Get("Strict-Transport-Security")).To(Equal("max-age=31536000"))
				Expect(w.Header().Get("Content-Security-Policy")).To(Equal("default-src 'self'"))
				Expect(w.Header().Get("X-XSS-Protection")).To(Equal("1; mode=block"))
				Expect(w.Header().Get("Referrer-Policy")).To(Equal("no-referrer"))
			})
		})
	})

	Describe("Real-World Scenarios", func() {
		Context("typical usage patterns", func() {
			It("should handle configuration from environment", func() {
				// Simulate loading config from env/file
				cpt := New(ctx)
				key := "env-head"

				// Simulate config that might come from YAML/JSON file
				vpr := libvpr.New(ctx, func() liblog.Logger { return nil })
				vpr.Viper().Set(key, map[string]string{
					"X-Service-Name":    "my-service",
					"X-Service-Version": "1.0.0",
					"X-Environment":     "production",
					"X-Region":          "us-west-2",
				})

				cpt.Init(key, ctx,
					func(string) cfgtps.Component { return nil },
					func() libvpr.Viper { return vpr },
					libver.NewVersion(libver.License_MIT, "my-service", "", "01/01/1970", "abcd1234", "1.0.0", "maintainer", "", Empty{}, 0),
					func() liblog.Logger { return nil })

				err := cpt.Start()
				Expect(err).To(BeNil())

				headers := cpt.GetHeaders()
				Expect(headers.Get("X-Service-Name")).To(Equal("my-service"))
				Expect(headers.Get("X-Service-Version")).To(Equal("1.0.0"))
				Expect(headers.Get("X-Environment")).To(Equal("production"))
				Expect(headers.Get("X-Region")).To(Equal("us-west-2"))
			})

			It("should support hot reload scenario", func() {
				cpt := New(ctx)
				key := "hot-reload-head"

				// Initial config
				vpr := libvpr.New(ctx, func() liblog.Logger { return nil })
				vpr.Viper().Set(key, map[string]string{
					"X-Config-Version": "1",
					"X-Feature-Flag":   "disabled",
				})

				cpt.Init(key, ctx,
					func(string) cfgtps.Component { return nil },
					func() libvpr.Viper { return vpr },
					libver.NewVersion(libver.License_MIT, "test", "", "01/01/1970", "abcd1234", "1.0.0", "maintainer", "", Empty{}, 0),
					func() liblog.Logger { return nil })

				err := cpt.Start()
				Expect(err).To(BeNil())

				// Simulate config file change
				vpr.Viper().Set(key, map[string]string{
					"X-Config-Version": "2",
					"X-Feature-Flag":   "enabled",
					"X-New-Setting":    "new-value",
				})

				// Hot reload
				err = cpt.Reload()
				Expect(err).To(BeNil())

				// Verify new config is active
				headers := cpt.GetHeaders()
				Expect(headers.Get("X-Config-Version")).To(Equal("2"))
				Expect(headers.Get("X-Feature-Flag")).To(Equal("enabled"))
				Expect(headers.Get("X-New-Setting")).To(Equal("new-value"))
			})
		})
	})
})
