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

package http_test

import (
	"context"
	"encoding/json"
	"net/http"

	. "github.com/nabbar/golib/config/components/http"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libcfg "github.com/nabbar/golib/config"
	cfgtps "github.com/nabbar/golib/config/types"
	htpool "github.com/nabbar/golib/httpserver/pool"
	libver "github.com/nabbar/golib/version"
)

// Integration tests verify component integration with other packages
var _ = Describe("Integration Tests", func() {
	var (
		ctx context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()
	})

	Describe("Component creation and registration", func() {
		Context("with config system", func() {
			It("should integrate with config package", func() {
				vrs := libver.NewVersion(libver.License_MIT, "test", "", "01/01/1970", "test", "1.0.0", "test", "", struct{}{}, 0)
				cfg := libcfg.New(vrs)
				Expect(cfg).NotTo(BeNil())

				key := "http-server"
				cpt := New(ctx, DefaultTlsKey, nil)
				Register(cfg, key, cpt)

				retrieved := cfg.ComponentGet(key)
				Expect(retrieved).NotTo(BeNil())
			})

			It("should work with RegisterNew helper", func() {
				vrs := libver.NewVersion(libver.License_MIT, "test", "", "01/01/1970", "test", "1.0.0", "test", "", struct{}{}, 0)
				cfg := libcfg.New(vrs)

				key := "http-server"
				tlsKey := DefaultTlsKey
				hdl := func() map[string]http.Handler {
					return map[string]http.Handler{}
				}

				RegisterNew(ctx, cfg, key, tlsKey, hdl)

				retrieved := cfg.ComponentGet(key)
				Expect(retrieved).NotTo(BeNil())
			})

			It("should work with Load helper", func() {
				vrs := libver.NewVersion(libver.License_MIT, "test", "", "01/01/1970", "test", "1.0.0", "test", "", struct{}{}, 0)
				cfg := libcfg.New(vrs)
				key := "http-server"
				cpt := New(ctx, DefaultTlsKey, nil)
				Register(cfg, key, cpt)

				loaded := Load(cfg.ComponentGet, key)
				Expect(loaded).NotTo(BeNil())
				Expect(loaded).To(Equal(cpt))
			})
		})

		Context("multiple components", func() {
			It("should handle multiple HTTP components", func() {
				vrs := libver.NewVersion(libver.License_MIT, "test", "", "01/01/1970", "test", "1.0.0", "test", "", struct{}{}, 0)
				cfg := libcfg.New(vrs)

				keys := []string{"http-api", "http-status", "http-metrics"}
				for _, key := range keys {
					cpt := New(ctx, DefaultTlsKey, nil)
					Register(cfg, key, cpt)
				}

				for _, key := range keys {
					retrieved := cfg.ComponentGet(key)
					Expect(retrieved).NotTo(BeNil())
				}
			})

			It("should handle different TLS keys per component", func() {
				vrs := libver.NewVersion(libver.License_MIT, "test", "", "01/01/1970", "test", "1.0.0", "test", "", struct{}{}, 0)
				cfg := libcfg.New(vrs)

				components := map[string]string{
					"http-api":     "tls-api",
					"http-status":  "tls-status",
					"http-metrics": "tls-metrics",
				}

				for key, tlsKey := range components {
					cpt := New(ctx, tlsKey, nil)
					Register(cfg, key, cpt)
					Expect(cpt.Dependencies()).To(ContainElement(tlsKey))
				}
			})
		})
	})

	Describe("Handler integration", func() {
		Context("with HTTP handlers", func() {
			It("should accept basic HTTP handler", func() {
				handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				})

				hdl := func() map[string]http.Handler {
					return map[string]http.Handler{
						"test": handler,
					}
				}

				cpt := New(ctx, DefaultTlsKey, hdl)
				Expect(cpt).NotTo(BeNil())
			})

			It("should accept multiple handlers", func() {
				hdl := func() map[string]http.Handler {
					return map[string]http.Handler{
						"api": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							w.WriteHeader(http.StatusOK)
						}),
						"status": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							w.WriteHeader(http.StatusOK)
						}),
						"metrics": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							w.WriteHeader(http.StatusOK)
						}),
					}
				}

				cpt := New(ctx, DefaultTlsKey, hdl)
				Expect(cpt).NotTo(BeNil())
			})

			It("should allow updating handlers", func() {
				hdl1 := func() map[string]http.Handler {
					return map[string]http.Handler{
						"v1": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
					}
				}

				cpt := New(ctx, DefaultTlsKey, hdl1)

				hdl2 := func() map[string]http.Handler {
					return map[string]http.Handler{
						"v2": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
					}
				}

				cpt.SetHandler(hdl2)
				Expect(cpt).NotTo(BeNil())
			})
		})
	})

	Describe("Pool integration", func() {
		Context("with httpserver pool", func() {
			It("should create pool on component creation", func() {
				cpt := New(ctx, DefaultTlsKey, nil)
				pool := cpt.GetPool()

				Expect(pool).NotTo(BeNil())
			})

			It("should allow setting custom pool", func() {
				cpt := New(ctx, DefaultTlsKey, nil)

				customPool := htpool.New(ctx, func() map[string]http.Handler {
					return map[string]http.Handler{}
				})

				cpt.SetPool(customPool)
				pool := cpt.GetPool()
				Expect(pool).NotTo(BeNil())
			})

			It("should handle pool operations", func() {
				hdl := func() map[string]http.Handler {
					return map[string]http.Handler{
						"test": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
					}
				}

				cpt := New(ctx, DefaultTlsKey, hdl)
				pool := cpt.GetPool()

				Expect(pool).NotTo(BeNil())
				// Pool should be created and initialized
			})
		})
	})

	Describe("TLS dependency integration", func() {
		Context("with TLS component", func() {
			It("should declare TLS dependency", func() {
				cpt := New(ctx, DefaultTlsKey, nil)
				deps := cpt.Dependencies()

				// Dependencies use TLS key, not component type
				Expect(deps).To(ContainElement(DefaultTlsKey))
			})

			It("should use custom TLS key as dependency", func() {
				customTlsKey := "my-tls-config"
				cpt := New(ctx, customTlsKey, nil)
				deps := cpt.Dependencies()

				Expect(deps).To(ContainElement(customTlsKey))
			})

			It("should update dependency when TLS key changes", func() {
				cpt := New(ctx, "tls1", nil)
				Expect(cpt.Dependencies()).To(ContainElement("tls1"))

				cpt.SetTLSKey("tls2")
				Expect(cpt.Dependencies()).To(ContainElement("tls2"))
			})
		})
	})

	Describe("Complete component lifecycle", func() {
		Context("typical usage scenario", func() {
			It("should handle complete lifecycle", func() {
				// 1. Create config
				vrs := libver.NewVersion(libver.License_MIT, "test", "", "01/01/1970", "test", "1.0.0", "test", "", struct{}{}, 0)
				cfg := libcfg.New(vrs)

				// 2. Create and register component
				key := "http-server"
				hdl := func() map[string]http.Handler {
					return map[string]http.Handler{
						"api": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
					}
				}

				RegisterNew(ctx, cfg, key, DefaultTlsKey, hdl)

				// 3. Load component
				cpt := Load(cfg.ComponentGet, key)
				Expect(cpt).NotTo(BeNil())

				// 4. Check initial state
				Expect(cpt.IsStarted()).To(BeFalse())
				Expect(cpt.IsRunning()).To(BeFalse())

				// 5. Get pool
				pool := cpt.GetPool()
				Expect(pool).NotTo(BeNil())

				// 6. Stop (should not panic even if not started)
				cpt.Stop()
			})

			It("should handle reconfiguration", func() {
				vrs := libver.NewVersion(libver.License_MIT, "test", "", "01/01/1970", "test", "1.0.0", "test", "", struct{}{}, 0)
				cfg := libcfg.New(vrs)
				key := "http-server"

				// Initial setup
				cpt := New(ctx, "tls1", nil)
				Register(cfg, key, cpt)

				// Reconfigure TLS
				cpt.SetTLSKey("tls2")

				// Reconfigure handler
				newHandler := func() map[string]http.Handler {
					return map[string]http.Handler{
						"new": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
					}
				}
				cpt.SetHandler(newHandler)

				Expect(cpt).NotTo(BeNil())
			})
		})
	})

	Describe("Error handling integration", func() {
		Context("with invalid configurations", func() {
			It("should return appropriate errors on start without init", func() {
				cpt := New(ctx, DefaultTlsKey, nil)
				err := cpt.Start()

				Expect(err).To(HaveOccurred())
				// Error can be about initialization or start failure
				Expect(err.Error()).To(Or(ContainSubstring("initialized"), ContainSubstring("start")))
			})

			It("should return errors on reload without init", func() {
				cpt := New(ctx, DefaultTlsKey, nil)
				err := cpt.Reload()

				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Concurrent integration scenarios", func() {
		Context("concurrent component operations", func() {
			It("should handle concurrent component creation and registration", func() {
				vrs := libver.NewVersion(libver.License_MIT, "test", "", "01/01/1970", "test", "1.0.0", "test", "", struct{}{}, 0)
				cfg := libcfg.New(vrs)
				done := make(chan bool, 20)

				// 10 components being created
				for i := 0; i < 10; i++ {
					go func(index int) {
						defer GinkgoRecover()
						key := "http-" + string(rune('a'+index))
						cpt := New(ctx, DefaultTlsKey, nil)
						Register(cfg, key, cpt)
						done <- true
					}(i)
				}

				// 10 components being loaded
				for i := 0; i < 10; i++ {
					go func(index int) {
						defer GinkgoRecover()
						key := "http-" + string(rune('a'+index))
						Eventually(func() CptHttp {
							return Load(cfg.ComponentGet, key)
						}).ShouldNot(BeNil())
						done <- true
					}(i)
				}

				for i := 0; i < 20; i++ {
					Eventually(done).Should(Receive())
				}
			})

			It("should handle concurrent handler updates", func() {
				cpt := New(ctx, DefaultTlsKey, nil)
				done := make(chan bool, 10)

				for i := 0; i < 10; i++ {
					go func(index int) {
						defer GinkgoRecover()
						hdl := func() map[string]http.Handler {
							return map[string]http.Handler{
								"handler": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
							}
						}
						cpt.SetHandler(hdl)
						done <- true
					}(i)
				}

				for i := 0; i < 10; i++ {
					Eventually(done).Should(Receive())
				}
			})
		})
	})

	Describe("Default configuration integration", func() {
		Context("using default configuration", func() {
			It("should provide valid default configuration", func() {
				defaultCfg := DefaultConfig("")
				Expect(defaultCfg).NotTo(BeEmpty())
			})

			It("should allow customizing default configuration", func() {
				// Keep Default JSON before play with set
				old := DefaultConfig("")
				defer func() {
					SetDefaultConfig(old)
				}()

				customDefault := []byte(`[{"name":"custom","handler_key":"custom","listen":"0.0.0.0:9999","expose":"http://localhost"}]`)
				SetDefaultConfig(customDefault)

				// DefaultConfig applies indentation, so we need to compare the unmarshaled data
				cfg := DefaultConfig("")
				var cfgData, customData []map[string]interface{}
				Expect(json.Unmarshal(cfg, &cfgData)).NotTo(HaveOccurred())
				Expect(json.Unmarshal(customDefault, &customData)).NotTo(HaveOccurred())
				Expect(cfgData).To(Equal(customData))
			})
		})
	})

	Describe("Component type verification", func() {
		Context("type constants", func() {
			It("should have correct component type", func() {
				Expect(ComponentType).To(Equal("http"))
			})

			It("should match component instance type", func() {
				cpt := New(ctx, DefaultTlsKey, nil)
				Expect(cpt.Type()).To(Equal(ComponentType))
			})
		})
	})

	Describe("Callback integration", func() {
		Context("with lifecycle callbacks", func() {
			It("should register start callbacks", func() {
				cpt := New(ctx, DefaultTlsKey, nil)

				before := func(c cfgtps.Component) error {
					return nil
				}

				after := func(c cfgtps.Component) error {
					return nil
				}

				cpt.RegisterFuncStart(before, after)

				// Callbacks registered, actual execution depends on Start
			})

			It("should register reload callbacks", func() {
				cpt := New(ctx, DefaultTlsKey, nil)

				before := func(c cfgtps.Component) error {
					return nil
				}

				after := func(c cfgtps.Component) error {
					return nil
				}

				cpt.RegisterFuncReload(before, after)

				// Callbacks registered, actual execution depends on Reload
			})
		})
	})
})
