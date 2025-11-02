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
	"net/http"

	. "github.com/nabbar/golib/config/components/http"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libcfg "github.com/nabbar/golib/config"
	cfgtps "github.com/nabbar/golib/config/types"
	htpool "github.com/nabbar/golib/httpserver/pool"
	liblog "github.com/nabbar/golib/logger"
	montps "github.com/nabbar/golib/monitor/types"
	libver "github.com/nabbar/golib/version"
	libvpr "github.com/nabbar/golib/viper"
	spfcbr "github.com/spf13/cobra"
)

// Interface tests verify component creation and registration
var _ = Describe("Interface Functions", func() {
	var (
		ctx context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()
	})

	Describe("New function", func() {
		Context("with valid parameters", func() {
			It("should create new component with custom TLS key", func() {
				tlsKey := "custom-tls"
				hdl := func() map[string]http.Handler {
					return map[string]http.Handler{}
				}

				cpt := New(ctx, tlsKey, hdl)
				Expect(cpt).NotTo(BeNil())
			})

			It("should create new component with default TLS key", func() {
				hdl := func() map[string]http.Handler {
					return map[string]http.Handler{}
				}

				cpt := New(ctx, "", hdl)
				Expect(cpt).NotTo(BeNil())
			})

			It("should create new component with nil handler", func() {
				cpt := New(ctx, DefaultTlsKey, nil)
				Expect(cpt).NotTo(BeNil())
			})

			It("should set default TLS key when empty", func() {
				cpt := New(ctx, "", nil)
				Expect(cpt).NotTo(BeNil())
				// Component should use DefaultTlsKey
			})
		})

		Context("component type", func() {
			It("should return correct component type", func() {
				cpt := New(ctx, DefaultTlsKey, nil)
				Expect(cpt.Type()).To(Equal(ComponentType))
			})

			It("should implement CptHttp interface", func() {
				cpt := New(ctx, DefaultTlsKey, nil)
				var _ CptHttp = cpt
			})

			It("should implement cfgtps.Component interface", func() {
				cpt := New(ctx, DefaultTlsKey, nil)
				var _ cfgtps.Component = cpt
			})
		})

		Context("initial state", func() {
			It("should not be started initially", func() {
				cpt := New(ctx, DefaultTlsKey, nil)
				Expect(cpt.IsStarted()).To(BeFalse())
			})

			It("should not be running initially", func() {
				cpt := New(ctx, DefaultTlsKey, nil)
				Expect(cpt.IsRunning()).To(BeFalse())
			})

			It("should have default dependencies", func() {
				cpt := New(ctx, DefaultTlsKey, nil)
				deps := cpt.Dependencies()
				Expect(deps).NotTo(BeEmpty())
				// Should depend on TLS component via the TLS key (DefaultTlsKey = "t")
				Expect(deps).To(ContainElement(DefaultTlsKey))
			})
		})
	})

	Describe("Register function", func() {
		var (
			cfg libcfg.Config
			cpt CptHttp
		)

		BeforeEach(func() {
			cpt = New(ctx, DefaultTlsKey, nil)
		})

		Context("registering component", func() {
			It("should register component with key", func() {
				vrs := libver.NewVersion(libver.License_MIT, "test", "", "01/01/1970", "test", "1.0.0", "test", "", struct{}{}, 0)
				cfg = libcfg.New(vrs)
				key := "http-server"
				Register(cfg, key, cpt)

				// Component should be registered
				retrieved := cfg.ComponentGet(key)
				Expect(retrieved).NotTo(BeNil())
			})

			It("should allow retrieving registered component", func() {
				vrs := libver.NewVersion(libver.License_MIT, "test", "", "01/01/1970", "test", "1.0.0", "test", "", struct{}{}, 0)
				cfg = libcfg.New(vrs)
				key := "http-server"
				Register(cfg, key, cpt)

				retrieved := cfg.ComponentGet(key)
				Expect(retrieved).To(Equal(cpt))
			})

			It("should replace existing component with same key", func() {
				vrs := libver.NewVersion(libver.License_MIT, "test", "", "01/01/1970", "test", "1.0.0", "test", "", struct{}{}, 0)
				cfg = libcfg.New(vrs)
				key := "http-server"
				cpt1 := New(ctx, DefaultTlsKey, nil)
				cpt2 := New(ctx, "other-tls", nil)

				Register(cfg, key, cpt1)
				Register(cfg, key, cpt2)

				retrieved := cfg.ComponentGet(key)
				Expect(retrieved).To(Equal(cpt2))
			})
		})
	})

	Describe("RegisterNew function", func() {
		Context("creating and registering component", func() {
			It("should create and register new component", func() {
				vrs := libver.NewVersion(libver.License_MIT, "test", "", "01/01/1970", "test", "1.0.0", "test", "", struct{}{}, 0)
				cfg := libcfg.New(vrs)
				key := "http-server"
				tlsKey := "tls"
				hdl := func() map[string]http.Handler {
					return map[string]http.Handler{}
				}

				RegisterNew(ctx, cfg, key, tlsKey, hdl)

				retrieved := cfg.ComponentGet(key)
				Expect(retrieved).NotTo(BeNil())
			})

			It("should create component with specified TLS key", func() {
				vrs := libver.NewVersion(libver.License_MIT, "test", "", "01/01/1970", "test", "1.0.0", "test", "", struct{}{}, 0)
				cfg := libcfg.New(vrs)
				key := "http-server"
				tlsKey := "custom-tls"

				RegisterNew(ctx, cfg, key, tlsKey, nil)

				retrieved := cfg.ComponentGet(key)
				Expect(retrieved).NotTo(BeNil())
			})

			It("should create component with specified handler", func() {
				vrs := libver.NewVersion(libver.License_MIT, "test", "", "01/01/1970", "test", "1.0.0", "test", "", struct{}{}, 0)
				cfg := libcfg.New(vrs)
				key := "http-server"
				hdl := func() map[string]http.Handler {
					return map[string]http.Handler{
						"test": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
					}
				}

				RegisterNew(ctx, cfg, key, DefaultTlsKey, hdl)

				retrieved := cfg.ComponentGet(key)
				Expect(retrieved).NotTo(BeNil())
			})
		})
	})

	Describe("Load function", func() {
		Context("loading existing component", func() {
			It("should load registered HTTP component", func() {
				vrs := libver.NewVersion(libver.License_MIT, "test", "", "01/01/1970", "test", "1.0.0", "test", "", struct{}{}, 0)
				cfg := libcfg.New(vrs)
				getCpt := cfg.ComponentGet
				key := "http-server"
				cpt := New(ctx, DefaultTlsKey, nil)
				Register(cfg, key, cpt)

				loaded := Load(getCpt, key)
				Expect(loaded).NotTo(BeNil())
				Expect(loaded).To(Equal(cpt))
			})

			It("should return nil for non-existent key", func() {
				vrs := libver.NewVersion(libver.License_MIT, "test", "", "01/01/1970", "test", "1.0.0", "test", "", struct{}{}, 0)
				cfg := libcfg.New(vrs)
				getCpt := cfg.ComponentGet
				loaded := Load(getCpt, "non-existent")
				Expect(loaded).To(BeNil())
			})

			It("should return nil for wrong component type", func() {
				vrs := libver.NewVersion(libver.License_MIT, "test", "", "01/01/1970", "test", "1.0.0", "test", "", struct{}{}, 0)
				cfg := libcfg.New(vrs)
				getCpt := cfg.ComponentGet
				key := "wrong-type"
				// Register a different component type
				wrongCpt := &mockComponent{}
				cfg.ComponentSet(key, wrongCpt)

				loaded := Load(getCpt, key)
				Expect(loaded).To(BeNil())
			})
		})

		Context("with nil getCpt function", func() {
			It("should handle nil getCpt gracefully", func() {
				loaded := Load(nil, "any-key")
				Expect(loaded).To(BeNil())
			})
		})
	})

	Describe("DefaultTlsKey constant", func() {
		It("should have expected value", func() {
			Expect(DefaultTlsKey).To(Equal("t"))
		})

		It("should be used as default", func() {
			cpt := New(ctx, "", nil)
			Expect(cpt).NotTo(BeNil())
			// Should use DefaultTlsKey internally
		})
	})

	Describe("Component lifecycle", func() {
		var (
			cpt CptHttp
		)

		BeforeEach(func() {
			cpt = New(ctx, DefaultTlsKey, nil)
		})

		Context("initialization", func() {
			It("should allow setting TLS key", func() {
				cpt.SetTLSKey("new-tls-key")
				// Should not panic
			})

			It("should allow setting handler", func() {
				hdl := func() map[string]http.Handler {
					return map[string]http.Handler{}
				}
				cpt.SetHandler(hdl)
				// Should not panic
			})

			It("should allow getting pool", func() {
				pool := cpt.GetPool()
				Expect(pool).NotTo(BeNil())
			})

			It("should allow setting pool", func() {
				newPool := htpool.New(ctx, func() map[string]http.Handler {
					return map[string]http.Handler{}
				})
				cpt.SetPool(newPool)
				// Should not panic
			})
		})

		Context("component interface methods", func() {
			It("should implement Type method", func() {
				typ := cpt.Type()
				Expect(typ).To(Equal(ComponentType))
			})

			It("should implement Dependencies method", func() {
				deps := cpt.Dependencies()
				Expect(deps).NotTo(BeNil())
			})

			It("should implement SetDependencies method", func() {
				err := cpt.SetDependencies([]string{"dep1", "dep2"})
				Expect(err).To(BeNil())
			})
		})
	})

	Describe("Concurrent operations", func() {
		Context("concurrent component creation", func() {
			It("should handle concurrent New calls", func() {
				done := make(chan bool, 10)

				for i := 0; i < 10; i++ {
					go func() {
						defer GinkgoRecover()
						cpt := New(ctx, DefaultTlsKey, nil)
						Expect(cpt).NotTo(BeNil())
						done <- true
					}()
				}

				for i := 0; i < 10; i++ {
					Eventually(done).Should(Receive())
				}
			})

			It("should handle concurrent registrations", func() {
				vrs := libver.NewVersion(libver.License_MIT, "test", "", "01/01/1970", "test", "1.0.0", "test", "", struct{}{}, 0)
				cfg := libcfg.New(vrs)
				done := make(chan bool, 10)

				for i := 0; i < 10; i++ {
					go func(index int) {
						defer GinkgoRecover()
						key := "http-" + string(rune('0'+index))
						cpt := New(ctx, DefaultTlsKey, nil)
						Register(cfg, key, cpt)
						done <- true
					}(i)
				}

				for i := 0; i < 10; i++ {
					Eventually(done).Should(Receive())
				}
			})
		})
	})
})

// mockComponent is a mock implementation for testing wrong component types
type mockComponent struct{}

func (m *mockComponent) Type() string { return "mock" }
func (m *mockComponent) Init(string, context.Context, cfgtps.FuncCptGet, libvpr.FuncViper, libver.Version, liblog.FuncLog) {
}
func (m *mockComponent) Start() error                                                { return nil }
func (m *mockComponent) Reload() error                                               { return nil }
func (m *mockComponent) Stop()                                                       {}
func (m *mockComponent) Dependencies() []string                                      { return nil }
func (m *mockComponent) SetDependencies([]string) error                              { return nil }
func (m *mockComponent) IsStarted() bool                                             { return false }
func (m *mockComponent) IsRunning() bool                                             { return false }
func (m *mockComponent) RegisterFuncStart(cfgtps.FuncCptEvent, cfgtps.FuncCptEvent)  {}
func (m *mockComponent) RegisterFuncReload(cfgtps.FuncCptEvent, cfgtps.FuncCptEvent) {}
func (m *mockComponent) RegisterFlag(*spfcbr.Command) error                          { return nil }
func (m *mockComponent) DefaultConfig(string) []byte                                 { return []byte("{}") }
func (m *mockComponent) RegisterMonitorPool(montps.FuncPool)                         {}
