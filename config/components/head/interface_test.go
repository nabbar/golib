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
	"fmt"

	. "github.com/nabbar/golib/config/components/head"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libcfg "github.com/nabbar/golib/config"
	cfgtps "github.com/nabbar/golib/config/types"
	liblog "github.com/nabbar/golib/logger"
	montps "github.com/nabbar/golib/monitor/types"
	libver "github.com/nabbar/golib/version"
	libvpr "github.com/nabbar/golib/viper"
	spfcbr "github.com/spf13/cobra"
)

// Interface functions tests verify New, Register, RegisterNew, and Load
// These tests ensure the basic factory and registration patterns work correctly.
var _ = Describe("Interface Functions", func() {
	var (
		ctx context.Context
		cfg libcfg.Config
	)

	BeforeEach(func() {
		// Create a basic context function for testing
		ctx = context.Background()
		// Create a new configuration instance
		cfg = libcfg.New(nil)
	})

	AfterEach(func() {
		// Clean up the configuration
		if cfg != nil {
			cfg.Stop()
		}
	})

	Describe("New", func() {
		Context("when creating a new component", func() {
			It("should create a non-nil component", func() {
				cpt := New(ctx)
				Expect(cpt).NotTo(BeNil())
			})

			It("should return correct component type", func() {
				cpt := New(ctx)
				Expect(cpt.Type()).To(Equal("head"))
			})

			It("should create component with empty headers initially", func() {
				cpt := New(ctx)
				headers := cpt.GetHeaders()
				Expect(headers).NotTo(BeNil())
				Expect(headers.Header()).To(BeEmpty())
			})
		})

		Context("when creating multiple instances", func() {
			It("should maintain separate state between instances", func() {
				cpt1 := New(ctx)
				cpt2 := New(ctx)

				// Modify one instance
				headers1 := cpt1.GetHeaders()
				headers1.Set("X-Test-Header", "value1")
				cpt1.SetHeaders(headers1)

				// Other instance should be unaffected
				headers2 := cpt2.GetHeaders()
				Expect(headers2.Get("X-Test-Header")).To(BeEmpty())
			})
		})

		Context("with different context types", func() {
			It("should work with basic context", func() {
				basicCtx := context.Background()
				cpt := New(basicCtx)
				Expect(cpt).NotTo(BeNil())
			})

			It("should work with context containing values", func() {
				customCtx := context.WithValue(ctx, "test-key", "test-value")
				cpt := New(customCtx)
				Expect(cpt).NotTo(BeNil())
			})

			It("should handle nil context", func() {
				var nilCtx context.Context
				cpt := New(nilCtx)
				Expect(cpt).NotTo(BeNil())
			})
		})
	})

	Describe("Register", func() {
		Context("when registering a component", func() {
			It("should register successfully", func() {
				cpt := New(ctx)
				key := "test-head"

				Register(cfg, key, cpt)

				Expect(cfg.ComponentHas(key)).To(BeTrue())
			})

			It("should register with correct type", func() {
				cpt := New(ctx)
				key := "head-service"

				Register(cfg, key, cpt)

				Expect(cfg.ComponentType(key)).To(Equal("head"))
			})

			It("should be retrievable after registration", func() {
				cpt := New(ctx)
				key := "retrievable-head"

				Register(cfg, key, cpt)

				retrieved := cfg.ComponentGet(key)
				Expect(retrieved).NotTo(BeNil())
			})
		})

		Context("with multiple components", func() {
			It("should allow multiple components with different keys", func() {
				cpt1 := New(ctx)
				cpt2 := New(ctx)

				Register(cfg, "head-1", cpt1)
				Register(cfg, "head-2", cpt2)

				Expect(cfg.ComponentHas("head-1")).To(BeTrue())
				Expect(cfg.ComponentHas("head-2")).To(BeTrue())
			})

			It("should replace existing component with same key", func() {
				cpt1 := New(ctx)
				cpt2 := New(ctx)
				key := "head-replace"

				Register(cfg, key, cpt1)
				Register(cfg, key, cpt2)

				// Second registration should replace first
				Expect(cfg.ComponentHas(key)).To(BeTrue())
			})
		})

		Context("with edge case keys", func() {
			It("should handle empty key", func() {
				cpt := New(ctx)
				Register(cfg, "", cpt)

				Expect(cfg.ComponentHas("")).To(BeTrue())
			})

			It("should handle special characters in key", func() {
				cpt := New(ctx)
				specialKey := "head-test_123.service"
				Register(cfg, specialKey, cpt)

				Expect(cfg.ComponentHas(specialKey)).To(BeTrue())
			})

			It("should handle very long keys", func() {
				longKey := ""
				for i := 0; i < 255; i++ {
					longKey += "a"
				}

				cpt := New(ctx)
				Register(cfg, longKey, cpt)

				Expect(cfg.ComponentHas(longKey)).To(BeTrue())
			})
		})

		Context("with nil values", func() {
			It("should panic with nil config", func() {
				cpt := New(ctx)
				Expect(func() {
					Register(nil, "test", cpt)
				}).To(Panic())
			})

			It("should handle nil component", func() {
				Expect(func() {
					Register(cfg, "nil-component", nil)
				}).NotTo(Panic())
			})
		})
	})

	Describe("RegisterNew", func() {
		Context("when creating and registering in one call", func() {
			It("should create and register component", func() {
				key := "auto-head"
				RegisterNew(ctx, cfg, key)

				Expect(cfg.ComponentHas(key)).To(BeTrue())
				Expect(cfg.ComponentType(key)).To(Equal("head"))
			})

			It("should create working component", func() {
				key := "working-head"
				RegisterNew(ctx, cfg, key)

				retrieved := cfg.ComponentGet(key)
				Expect(retrieved).NotTo(BeNil())
				Expect(retrieved.Type()).To(Equal("head"))
			})
		})

		Context("with multiple components", func() {
			It("should register multiple components", func() {
				RegisterNew(ctx, cfg, "head1")
				RegisterNew(ctx, cfg, "head2")
				RegisterNew(ctx, cfg, "head3")

				Expect(cfg.ComponentHas("head1")).To(BeTrue())
				Expect(cfg.ComponentHas("head2")).To(BeTrue())
				Expect(cfg.ComponentHas("head3")).To(BeTrue())
			})

			It("should create independent components", func() {
				RegisterNew(ctx, cfg, "independent-1")
				RegisterNew(ctx, cfg, "independent-2")

				cpt1 := cfg.ComponentGet("independent-1")
				cpt2 := cfg.ComponentGet("independent-2")

				Expect(cpt1).NotTo(Equal(cpt2))
			})
		})
	})

	Describe("Load", func() {
		Context("when loading registered component", func() {
			It("should load successfully", func() {
				key := "loadable-head"
				cpt := New(ctx)
				Register(cfg, key, cpt)

				getCpt := func(k string) cfgtps.Component {
					return cfg.ComponentGet(k)
				}

				loaded := Load(getCpt, key)
				Expect(loaded).NotTo(BeNil())
				Expect(loaded.Type()).To(Equal("head"))
			})

			It("should return same component type", func() {
				key := "typed-head"
				original := New(ctx)
				Register(cfg, key, original)

				getCpt := func(k string) cfgtps.Component {
					return cfg.ComponentGet(k)
				}

				loaded := Load(getCpt, key)
				Expect(loaded).NotTo(BeNil())
				// Verify it implements CptHead interface
				headers := loaded.GetHeaders()
				Expect(headers).NotTo(BeNil())
			})
		})

		Context("when component doesn't exist", func() {
			It("should return nil for non-existent component", func() {
				getCpt := func(k string) cfgtps.Component {
					return nil
				}

				loaded := Load(getCpt, "non-existent")
				Expect(loaded).To(BeNil())
			})
		})

		Context("when component is wrong type", func() {
			It("should return nil for wrong component type", func() {
				// Register a mock component that's not a Head component
				mockCpt := &mockComponent{}
				cfg.ComponentSet("wrong-type", mockCpt)

				getCpt := func(k string) cfgtps.Component {
					return cfg.ComponentGet(k)
				}

				loaded := Load(getCpt, "wrong-type")
				Expect(loaded).To(BeNil())
			})
		})

		Context("with multiple components", func() {
			It("should load from component list", func() {
				RegisterNew(ctx, cfg, "head-1")
				RegisterNew(ctx, cfg, "head-2")
				RegisterNew(ctx, cfg, "head-3")

				getCpt := func(k string) cfgtps.Component {
					return cfg.ComponentGet(k)
				}

				// Load each one
				loaded1 := Load(getCpt, "head-1")
				loaded2 := Load(getCpt, "head-2")
				loaded3 := Load(getCpt, "head-3")

				Expect(loaded1).NotTo(BeNil())
				Expect(loaded2).NotTo(BeNil())
				Expect(loaded3).NotTo(BeNil())
			})

			It("should load correct component by key", func() {
				RegisterNew(ctx, cfg, "specific-head")

				getCpt := func(k string) cfgtps.Component {
					return cfg.ComponentGet(k)
				}

				// Load specific component
				loaded := Load(getCpt, "specific-head")
				Expect(loaded).NotTo(BeNil())
				Expect(loaded.Type()).To(Equal("head"))
			})
		})
	})

	Describe("Integration Scenarios", func() {
		Context("full lifecycle", func() {
			It("should handle full registration and loading cycle", func() {
				key := "integration-head"

				// Create
				cpt := New(ctx)
				Expect(cpt).NotTo(BeNil())

				// Register
				Register(cfg, key, cpt)
				Expect(cfg.ComponentHas(key)).To(BeTrue())

				// Load
				getCpt := func(k string) cfgtps.Component {
					return cfg.ComponentGet(k)
				}
				loaded := Load(getCpt, key)
				Expect(loaded).NotTo(BeNil())
				Expect(loaded.Type()).To(Equal("head"))
			})

			It("should handle RegisterNew and Load cycle", func() {
				key := "quick-head"

				// Register new
				RegisterNew(ctx, cfg, key)
				Expect(cfg.ComponentHas(key)).To(BeTrue())

				// Load
				getCpt := func(k string) cfgtps.Component {
					return cfg.ComponentGet(k)
				}
				loaded := Load(getCpt, key)
				Expect(loaded).NotTo(BeNil())
			})
		})

		Context("multiple components", func() {
			It("should support multiple Head components in same config", func() {
				keys := []string{"head-api", "head-web", "head-admin"}

				for _, key := range keys {
					RegisterNew(ctx, cfg, key)
				}

				getCpt := func(k string) cfgtps.Component {
					return cfg.ComponentGet(k)
				}

				// All should be loadable
				for _, key := range keys {
					loaded := Load(getCpt, key)
					Expect(loaded).NotTo(BeNil())
					Expect(loaded.Type()).To(Equal("head"))
				}
			})
		})
	})
})

// Concurrent access tests verify thread-safety of the interface functions
var _ = Describe("Concurrent Access", func() {
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

	Context("concurrent registrations", func() {
		It("should handle concurrent Register calls", func() {
			done := make(chan bool, 10)

			for i := 0; i < 10; i++ {
				go func(index int) {
					defer GinkgoRecover()
					key := fmt.Sprintf("concurrent-head-%d", index)
					cpt := New(ctx)
					Register(cfg, key, cpt)
					done <- true
				}(i)
			}

			// Wait for all goroutines
			for i := 0; i < 10; i++ {
				Eventually(done).Should(Receive())
			}

			// Verify all components are registered
			for i := 0; i < 10; i++ {
				key := fmt.Sprintf("concurrent-head-%d", i)
				Expect(cfg.ComponentHas(key)).To(BeTrue())
			}
		})

		It("should handle concurrent RegisterNew calls", func() {
			done := make(chan bool, 10)

			for i := 0; i < 10; i++ {
				go func(index int) {
					defer GinkgoRecover()
					key := fmt.Sprintf("concurrent-new-head-%d", index)
					RegisterNew(ctx, cfg, key)
					done <- true
				}(i)
			}

			// Wait for all goroutines
			for i := 0; i < 10; i++ {
				Eventually(done).Should(Receive())
			}

			// Verify all components are registered
			for i := 0; i < 10; i++ {
				key := fmt.Sprintf("concurrent-new-head-%d", i)
				Expect(cfg.ComponentHas(key)).To(BeTrue())
			}
		})
	})

	Context("concurrent loads", func() {
		It("should handle concurrent Load calls", func() {
			// Setup: register a component
			key := "shared-head"
			RegisterNew(ctx, cfg, key)

			getCpt := func(k string) cfgtps.Component {
				return cfg.ComponentGet(k)
			}

			done := make(chan bool, 10)

			// Load concurrently
			for i := 0; i < 10; i++ {
				go func() {
					defer GinkgoRecover()
					loaded := Load(getCpt, key)
					Expect(loaded).NotTo(BeNil())
					done <- true
				}()
			}

			// Wait for all goroutines
			for i := 0; i < 10; i++ {
				Eventually(done).Should(Receive())
			}
		})
	})

	Context("mixed operations", func() {
		It("should handle concurrent Register and Load", func() {
			done := make(chan bool, 20)

			// Register components
			for i := 0; i < 10; i++ {
				go func(index int) {
					defer GinkgoRecover()
					key := fmt.Sprintf("mixed-head-%d", index)
					RegisterNew(ctx, cfg, key)
					done <- true
				}(i)
			}

			getCpt := func(k string) cfgtps.Component {
				return cfg.ComponentGet(k)
			}

			// Load components (may not exist yet)
			for i := 0; i < 10; i++ {
				go func(index int) {
					defer GinkgoRecover()
					key := fmt.Sprintf("mixed-head-%d", index)
					Load(getCpt, key) // May be nil initially
					done <- true
				}(i)
			}

			// Wait for all goroutines
			for i := 0; i < 20; i++ {
				Eventually(done).Should(Receive())
			}
		})
	})
})

// mockComponent is a mock implementation for testing wrong type scenarios
type mockComponent struct{}

func (m *mockComponent) Type() string { return "mock" }
func (m *mockComponent) Init(string, context.Context, cfgtps.FuncCptGet, libvpr.FuncViper, libver.Version, liblog.FuncLog) {
}
func (m *mockComponent) RegisterFuncStart(cfgtps.FuncCptEvent, cfgtps.FuncCptEvent)  {}
func (m *mockComponent) RegisterFuncReload(cfgtps.FuncCptEvent, cfgtps.FuncCptEvent) {}
func (m *mockComponent) IsStarted() bool                                             { return false }
func (m *mockComponent) IsRunning() bool                                             { return false }
func (m *mockComponent) Start() error                                                { return nil }
func (m *mockComponent) Reload() error                                               { return nil }
func (m *mockComponent) Stop()                                                       {}
func (m *mockComponent) Dependencies() []string                                      { return nil }
func (m *mockComponent) SetDependencies([]string) error                              { return nil }
func (m *mockComponent) RegisterFlag(*spfcbr.Command) error                          { return nil }
func (m *mockComponent) RegisterMonitorPool(montps.FuncPool)                         {}
func (m *mockComponent) DefaultConfig(string) []byte                                 { return nil }
