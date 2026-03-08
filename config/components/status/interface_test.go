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
	"context"
	"fmt"

	. "github.com/nabbar/golib/config/components/status"
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

// Interface functions tests verify the behavior of New, Register, RegisterNew, and Load.
// These tests ensure the basic factory and registration patterns work correctly.
var _ = Describe("Interface Functions", func() {
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

	Describe("New", func() {
		Context("when creating a new component", func() {
			It("should create a non-nil component", func() {
				cpt := New(ctx)
				Expect(cpt).NotTo(BeNil())
			})

			It("should return the correct component type", func() {
				cpt := New(ctx)
				Expect(cpt.Type()).To(Equal("status"))
			})
		})

		Context("when creating multiple instances", func() {
			It("should create separate instances with distinct states", func() {
				cpt1 := New(ctx)
				cpt2 := New(ctx)
				Expect(cpt1).NotTo(Equal(cpt2))
			})
		})

		Context("with different context types", func() {
			It("should work with a basic context", func() {
				basicCtx := context.Background()
				cpt := New(basicCtx)
				Expect(cpt).NotTo(BeNil())
			})

			It("should work with a context containing values", func() {
				type key string
				customCtx := context.WithValue(ctx, key("test-key"), "test-value")
				cpt := New(customCtx)
				Expect(cpt).NotTo(BeNil())
			})

			It("should handle a nil context gracefully", func() {
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
				key := "test-status"
				Register(cfg, key, cpt)
				Expect(cfg.ComponentHas(key)).To(BeTrue())
			})

			It("should be retrievable with the correct type after registration", func() {
				cpt := New(ctx)
				key := "retrievable-status"
				Register(cfg, key, cpt)
				retrieved := cfg.ComponentGet(key)
				Expect(retrieved).NotTo(BeNil())
				Expect(retrieved.Type()).To(Equal("status"))
			})
		})

		Context("with multiple components", func() {
			It("should allow multiple components with different keys", func() {
				cpt1 := New(ctx)
				cpt2 := New(ctx)
				Register(cfg, "status-1", cpt1)
				Register(cfg, "status-2", cpt2)
				Expect(cfg.ComponentHas("status-1")).To(BeTrue())
				Expect(cfg.ComponentHas("status-2")).To(BeTrue())
			})

			It("should replace an existing component with the same key", func() {
				cpt1 := New(ctx)
				cpt2 := New(ctx)
				key := "status-replace"
				Register(cfg, key, cpt1)
				retrieved1 := cfg.ComponentGet(key)
				Register(cfg, key, cpt2)
				retrieved2 := cfg.ComponentGet(key)
				Expect(retrieved1).NotTo(Equal(retrieved2))
			})
		})

		Context("with nil values", func() {
			It("should panic if the config manager is nil", func() {
				cpt := New(ctx)
				Expect(func() {
					Register(nil, "test", cpt)
				}).To(Panic())
			})

			It("should handle a nil component without panicking", func() {
				Expect(func() {
					Register(cfg, "nil-component", nil)
				}).NotTo(Panic())
			})
		})
	})

	Describe("RegisterNew", func() {
		Context("when creating and registering in one call", func() {
			It("should create and register the component successfully", func() {
				key := "auto-status"
				RegisterNew(ctx, cfg, key)
				Expect(cfg.ComponentHas(key)).To(BeTrue())
				Expect(cfg.ComponentType(key)).To(Equal("status"))
			})
		})
	})

	Describe("Load", func() {
		Context("when loading a registered component", func() {
			It("should load the component successfully", func() {
				key := "loadable-status"
				cpt := New(ctx)
				Register(cfg, key, cpt)
				getCpt := func(k string) cfgtps.Component { return cfg.ComponentGet(k) }
				loaded := Load(getCpt, key)
				Expect(loaded).NotTo(BeNil())
				Expect(loaded.Type()).To(Equal("status"))
			})
		})

		Context("when the component does not exist", func() {
			It("should return nil", func() {
				getCpt := func(k string) cfgtps.Component { return nil }
				loaded := Load(getCpt, "non-existent")
				Expect(loaded).To(BeNil())
			})
		})

		Context("when the component is of the wrong type", func() {
			It("should return nil", func() {
				mockCpt := &mockComponent{}
				cfg.ComponentSet("wrong-type", mockCpt)
				getCpt := func(k string) cfgtps.Component { return cfg.ComponentGet(k) }
				loaded := Load(getCpt, "wrong-type")
				Expect(loaded).To(BeNil())
			})
		})
	})

	Describe("Integration Scenarios", func() {
		Context("with a full registration and loading cycle", func() {
			It("should handle the cycle correctly", func() {
				key := "integration-status"
				cpt := New(ctx)
				Register(cfg, key, cpt)
				Expect(cfg.ComponentHas(key)).To(BeTrue())
				getCpt := func(k string) cfgtps.Component { return cfg.ComponentGet(k) }
				loaded := Load(getCpt, key)
				Expect(loaded).NotTo(BeNil())
				Expect(loaded.Type()).To(Equal("status"))
			})
		})
	})
})

// Concurrent access tests verify thread-safety of the interface functions.
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

	Context("with concurrent registrations", func() {
		It("should handle concurrent Register calls without race conditions", func() {
			done := make(chan bool, 10)
			for i := 0; i < 10; i++ {
				go func(index int) {
					defer GinkgoRecover()
					key := fmt.Sprintf("concurrent-status-%d", index)
					cpt := New(ctx)
					Register(cfg, key, cpt)
					done <- true
				}(i)
			}
			for i := 0; i < 10; i++ {
				Eventually(done).Should(Receive())
			}
			for i := 0; i < 10; i++ {
				key := fmt.Sprintf("concurrent-status-%d", i)
				Expect(cfg.ComponentHas(key)).To(BeTrue())
			}
		})

		It("should handle concurrent RegisterNew calls without race conditions", func() {
			done := make(chan bool, 10)
			for i := 0; i < 10; i++ {
				go func(index int) {
					defer GinkgoRecover()
					key := fmt.Sprintf("concurrent-new-status-%d", index)
					RegisterNew(ctx, cfg, key)
					done <- true
				}(i)
			}
			for i := 0; i < 10; i++ {
				Eventually(done).Should(Receive())
			}
			for i := 0; i < 10; i++ {
				key := fmt.Sprintf("concurrent-new-status-%d", i)
				Expect(cfg.ComponentHas(key)).To(BeTrue())
			}
		})
	})

	Context("with concurrent loads", func() {
		It("should handle concurrent Load calls without race conditions", func() {
			key := "shared-status"
			RegisterNew(ctx, cfg, key)
			getCpt := func(k string) cfgtps.Component { return cfg.ComponentGet(k) }
			done := make(chan bool, 10)
			for i := 0; i < 10; i++ {
				go func() {
					defer GinkgoRecover()
					loaded := Load(getCpt, key)
					Expect(loaded).NotTo(BeNil())
					done <- true
				}()
			}
			for i := 0; i < 10; i++ {
				Eventually(done).Should(Receive())
			}
		})
	})
})

// mockComponent is a mock implementation for testing wrong type scenarios.
type mockComponent struct{}

func (m *mockComponent) GetMonitorNames() []string {
	return nil
}

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
