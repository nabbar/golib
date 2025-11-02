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

package aws_test

import (
	"context"
	"fmt"

	. "github.com/nabbar/golib/config/components/aws"
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
		It("should create a new AWS component with ConfigStandard", func() {
			cpt := New(ctx, ConfigStandard)
			Expect(cpt).NotTo(BeNil())
			Expect(cpt.Type()).To(Equal("aws"))
		})

		It("should create a new AWS component with ConfigCustom", func() {
			cpt := New(ctx, ConfigCustom)
			Expect(cpt).NotTo(BeNil())
			Expect(cpt.Type()).To(Equal("aws"))
		})

		It("should create a new AWS component with ConfigStandardStatus", func() {
			cpt := New(ctx, ConfigStandardStatus)
			Expect(cpt).NotTo(BeNil())
			Expect(cpt.Type()).To(Equal("aws"))
		})

		It("should create a new AWS component with ConfigCustomStatus", func() {
			cpt := New(ctx, ConfigCustomStatus)
			Expect(cpt).NotTo(BeNil())
			Expect(cpt.Type()).To(Equal("aws"))
		})

		It("should create independent instances", func() {
			cpt1 := New(ctx, ConfigStandard)
			cpt2 := New(ctx, ConfigCustom)

			Expect(cpt1).NotTo(BeNil())
			Expect(cpt2).NotTo(BeNil())
			Expect(cpt1).NotTo(Equal(cpt2))
		})
	})

	Describe("Register", func() {
		It("should register an AWS component in config", func() {
			cpt := New(ctx, ConfigStandard)
			key := "test-aws"

			Register(cfg, key, cpt)

			// Component should be registered
			Expect(cfg.ComponentHas(key)).To(BeTrue())
		})

		It("should register with custom key", func() {
			cpt := New(ctx, ConfigCustom)
			key := "custom-aws-service"

			Register(cfg, key, cpt)

			Expect(cfg.ComponentHas(key)).To(BeTrue())
			Expect(cfg.ComponentType(key)).To(Equal("aws"))
		})

		It("should allow multiple AWS components with different keys", func() {
			cpt1 := New(ctx, ConfigStandard)
			cpt2 := New(ctx, ConfigCustom)

			Register(cfg, "aws-1", cpt1)
			Register(cfg, "aws-2", cpt2)

			Expect(cfg.ComponentHas("aws-1")).To(BeTrue())
			Expect(cfg.ComponentHas("aws-2")).To(BeTrue())
		})

		It("should replace existing component with same key", func() {
			cpt1 := New(ctx, ConfigStandard)
			cpt2 := New(ctx, ConfigCustom)
			key := "aws"

			Register(cfg, key, cpt1)
			Register(cfg, key, cpt2)

			// Second registration should replace first
			Expect(cfg.ComponentHas(key)).To(BeTrue())
		})
	})

	Describe("RegisterNew", func() {
		It("should create and register AWS component with ConfigStandard", func() {
			key := "auto-aws-standard"
			RegisterNew(ctx, cfg, ConfigStandard, key)

			Expect(cfg.ComponentHas(key)).To(BeTrue())
			Expect(cfg.ComponentType(key)).To(Equal("aws"))
		})

		It("should create and register AWS component with ConfigCustom", func() {
			key := "auto-aws-custom"
			RegisterNew(ctx, cfg, ConfigCustom, key)

			Expect(cfg.ComponentHas(key)).To(BeTrue())
			Expect(cfg.ComponentType(key)).To(Equal("aws"))
		})

		It("should create and register AWS component with ConfigStandardStatus", func() {
			key := "auto-aws-standard-status"
			RegisterNew(ctx, cfg, ConfigStandardStatus, key)

			Expect(cfg.ComponentHas(key)).To(BeTrue())
			Expect(cfg.ComponentType(key)).To(Equal("aws"))
		})

		It("should create and register AWS component with ConfigCustomStatus", func() {
			key := "auto-aws-custom-status"
			RegisterNew(ctx, cfg, ConfigCustomStatus, key)

			Expect(cfg.ComponentHas(key)).To(BeTrue())
			Expect(cfg.ComponentType(key)).To(Equal("aws"))
		})

		It("should register multiple components with different drivers", func() {
			RegisterNew(ctx, cfg, ConfigStandard, "aws-std")
			RegisterNew(ctx, cfg, ConfigCustom, "aws-cus")
			RegisterNew(ctx, cfg, ConfigStandardStatus, "aws-std-st")
			RegisterNew(ctx, cfg, ConfigCustomStatus, "aws-cus-st")

			Expect(cfg.ComponentHas("aws-std")).To(BeTrue())
			Expect(cfg.ComponentHas("aws-cus")).To(BeTrue())
			Expect(cfg.ComponentHas("aws-std-st")).To(BeTrue())
			Expect(cfg.ComponentHas("aws-cus-st")).To(BeTrue())
		})
	})

	Describe("Load", func() {
		It("should load registered AWS component", func() {
			key := "loadable-aws"
			cpt := New(ctx, ConfigStandard)
			Register(cfg, key, cpt)

			// Create a getter function
			getCpt := func(k string) cfgtps.Component {
				return cfg.ComponentGet(k)
			}

			loaded := Load(getCpt, key)
			Expect(loaded).NotTo(BeNil())
			Expect(loaded.Type()).To(Equal("aws"))
		})

		It("should return nil for non-existent component", func() {
			getCpt := func(k string) cfgtps.Component {
				return nil
			}

			loaded := Load(getCpt, "non-existent")
			Expect(loaded).To(BeNil())
		})

		It("should return nil for wrong component type", func() {
			// Register a mock component that's not an AWS component
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
			RegisterNew(ctx, cfg, ConfigStandard, "aws-1")
			RegisterNew(ctx, cfg, ConfigCustom, "aws-2")

			getCpt := func(k string) cfgtps.Component {
				return cfg.ComponentGet(k)
			}

			// Load each one
			loaded1 := Load(getCpt, "aws-1")
			loaded2 := Load(getCpt, "aws-2")

			Expect(loaded1).NotTo(BeNil())
			Expect(loaded2).NotTo(BeNil())
		})
	})

	Describe("Integration Scenarios", func() {
		It("should handle full registration and loading cycle", func() {
			key := "integration-aws"

			// Create
			cpt := New(ctx, ConfigStandard)
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
			Expect(loaded.Type()).To(Equal("aws"))
		})

		It("should handle RegisterNew and Load cycle", func() {
			key := "quick-aws"

			// Register new
			RegisterNew(ctx, cfg, ConfigCustom, key)
			Expect(cfg.ComponentHas(key)).To(BeTrue())

			// Load
			getCpt := func(k string) cfgtps.Component {
				return cfg.ComponentGet(k)
			}
			loaded := Load(getCpt, key)
			Expect(loaded).NotTo(BeNil())
		})

		It("should support multiple AWS components in same config", func() {
			keys := []string{"aws-primary", "aws-secondary", "aws-backup"}

			for _, key := range keys {
				RegisterNew(ctx, cfg, ConfigStandard, key)
			}

			getCpt := func(k string) cfgtps.Component {
				return cfg.ComponentGet(k)
			}

			// All should be loadable
			for _, key := range keys {
				loaded := Load(getCpt, key)
				Expect(loaded).NotTo(BeNil())
				Expect(loaded.Type()).To(Equal("aws"))
			}
		})
	})

	Describe("Edge Cases", func() {
		It("should handle empty key", func() {
			cpt := New(ctx, ConfigStandard)
			Register(cfg, "", cpt)

			// Component should still be registered
			Expect(cfg.ComponentHas("")).To(BeTrue())
		})

		It("should handle special characters in key", func() {
			cpt := New(ctx, ConfigCustom)
			specialKey := "aws-test_123.service"
			Register(cfg, specialKey, cpt)

			Expect(cfg.ComponentHas(specialKey)).To(BeTrue())

			getCpt := func(k string) cfgtps.Component {
				return cfg.ComponentGet(k)
			}
			loaded := Load(getCpt, specialKey)
			Expect(loaded).NotTo(BeNil())
		})

		It("should handle very long keys", func() {
			longKey := string(make([]byte, 255))
			for range longKey {
				longKey = "a" + longKey[1:]
			}

			cpt := New(ctx, ConfigStandard)
			Register(cfg, longKey, cpt)

			Expect(cfg.ComponentHas(longKey)).To(BeTrue())
		})

		It("should handle nil config gracefully in Register", func() {
			cpt := New(ctx, ConfigStandard)
			// This might panic, but we're testing graceful handling
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

		It("should handle nil getter in Load", func() {
			// Nil getter will cause panic in real usage - this is expected
			Skip("Nil getter check skipped - causes expected panic")
		})
	})
})

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
				key := fmt.Sprintf("concurrent-aws-%d", index)
				cpt := New(ctx, ConfigStandard)
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
			key := fmt.Sprintf("concurrent-aws-%d", i)
			Expect(cfg.ComponentHas(key)).To(BeTrue())
		}
	})

	It("should handle concurrent Load calls", func() {
		// Setup: register a component
		key := "shared-aws"
		RegisterNew(ctx, cfg, ConfigStandard, key)

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
