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

package database_test

import (
	"context"
	"fmt"

	. "github.com/nabbar/golib/config/components/database"
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
		It("should create a new Database component", func() {
			cpt := New(ctx)
			Expect(cpt).NotTo(BeNil())
			Expect(cpt.Type()).To(Equal("database"))
		})

		It("should initialize with context", func() {
			customCtx := context.WithValue(ctx, "test", "value")
			cpt := New(customCtx)
			Expect(cpt).NotTo(BeNil())
		})

		It("should handle nil context", func() {
			var nilCtx context.Context
			cpt := New(nilCtx)
			Expect(cpt).NotTo(BeNil())
		})
	})

	Describe("Register", func() {
		It("should register a Database component in config", func() {
			cpt := New(ctx)
			key := "test-database"

			Register(cfg, key, cpt)

			// Component should be registered
			Expect(cfg.ComponentHas(key)).To(BeTrue())
		})

		It("should register with custom key", func() {
			cpt := New(ctx)
			key := "custom-database-service"

			Register(cfg, key, cpt)

			Expect(cfg.ComponentHas(key)).To(BeTrue())
			Expect(cfg.ComponentType(key)).To(Equal("database"))
		})

		It("should allow multiple Database components with different keys", func() {
			cpt1 := New(ctx)
			cpt2 := New(ctx)

			Register(cfg, "database-1", cpt1)
			Register(cfg, "database-2", cpt2)

			Expect(cfg.ComponentHas("database-1")).To(BeTrue())
			Expect(cfg.ComponentHas("database-2")).To(BeTrue())
		})

		It("should replace existing component with same key", func() {
			cpt1 := New(ctx)
			cpt2 := New(ctx)
			key := "database"

			Register(cfg, key, cpt1)
			Register(cfg, key, cpt2)

			// Second registration should replace first
			Expect(cfg.ComponentHas(key)).To(BeTrue())
		})
	})

	Describe("RegisterNew", func() {
		It("should create and register Database component", func() {
			key := "auto-database"
			RegisterNew(ctx, cfg, key)

			Expect(cfg.ComponentHas(key)).To(BeTrue())
			Expect(cfg.ComponentType(key)).To(Equal("database"))
		})

		It("should register multiple components", func() {
			RegisterNew(ctx, cfg, "db1")
			RegisterNew(ctx, cfg, "db2")
			RegisterNew(ctx, cfg, "db3")

			Expect(cfg.ComponentHas("db1")).To(BeTrue())
			Expect(cfg.ComponentHas("db2")).To(BeTrue())
			Expect(cfg.ComponentHas("db3")).To(BeTrue())
		})
	})

	Describe("Load", func() {
		It("should load registered Database component", func() {
			key := "loadable-database"
			cpt := New(ctx)
			Register(cfg, key, cpt)

			// Create a getter function
			getCpt := func(k string) cfgtps.Component {
				return cfg.ComponentGet(k)
			}

			loaded := Load(getCpt, key)
			Expect(loaded).NotTo(BeNil())
			Expect(loaded.Type()).To(Equal("database"))
		})

		It("should return nil for non-existent component", func() {
			getCpt := func(k string) cfgtps.Component {
				return nil
			}

			loaded := Load(getCpt, "non-existent")
			Expect(loaded).To(BeNil())
		})

		It("should return nil for wrong component type", func() {
			// Register a mock component that's not a Database component
			mockCpt := &mockComponent{}
			cfg.ComponentSet("wrong-type", mockCpt)

			getCpt := func(k string) cfgtps.Component {
				return cfg.ComponentGet(k)
			}

			loaded := Load(getCpt, "wrong-type")
			Expect(loaded).To(BeNil())
		})

		It("should load from component list", func() {
			// Register multiple components
			RegisterNew(ctx, cfg, "database-1")
			RegisterNew(ctx, cfg, "database-2")

			getCpt := func(k string) cfgtps.Component {
				return cfg.ComponentGet(k)
			}

			// Load each one
			loaded1 := Load(getCpt, "database-1")
			loaded2 := Load(getCpt, "database-2")

			Expect(loaded1).NotTo(BeNil())
			Expect(loaded2).NotTo(BeNil())
		})
	})

	Describe("Integration Scenarios", func() {
		It("should handle full registration and loading cycle", func() {
			key := "integration-database"

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
			Expect(loaded.Type()).To(Equal("database"))
		})

		It("should handle RegisterNew and Load cycle", func() {
			key := "quick-database"

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

		It("should support multiple Database components in same config", func() {
			keys := []string{"database-primary", "database-secondary", "database-backup"}

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
				Expect(loaded.Type()).To(Equal("database"))
			}
		})
	})

	Describe("Edge Cases", func() {
		It("should handle empty key", func() {
			cpt := New(ctx)
			Register(cfg, "", cpt)

			// Component should still be registered
			Expect(cfg.ComponentHas("")).To(BeTrue())
		})

		It("should handle special characters in key", func() {
			cpt := New(ctx)
			specialKey := "database-test_123.service"
			Register(cfg, specialKey, cpt)

			Expect(cfg.ComponentHas(specialKey)).To(BeTrue())

			getCpt := func(k string) cfgtps.Component {
				return cfg.ComponentGet(k)
			}
			loaded := Load(getCpt, specialKey)
			Expect(loaded).NotTo(BeNil())
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

		It("should handle nil config gracefully in Register", func() {
			cpt := New(ctx)
			// This will panic - expected behavior
			Expect(func() {
				Register(nil, "test", cpt)
			}).To(Panic())
		})

		It("should handle nil component in Register", func() {
			// Register nil component - should not crash
			Expect(func() {
				Register(cfg, "nil-component", nil)
			}).NotTo(Panic())
		})
	})
})

// Concurrent access tests
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

	It("should handle concurrent Register calls", func() {
		done := make(chan bool, 10)

		for i := 0; i < 10; i++ {
			go func(index int) {
				defer GinkgoRecover()
				key := fmt.Sprintf("concurrent-database-%d", index)
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
			key := fmt.Sprintf("concurrent-database-%d", i)
			Expect(cfg.ComponentHas(key)).To(BeTrue())
		}
	})

	It("should handle concurrent Load calls", func() {
		// Setup: register a component
		key := "shared-database"
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
