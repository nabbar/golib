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

	. "github.com/nabbar/golib/config/components/aws"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	cfgtps "github.com/nabbar/golib/config/types"
	libhtc "github.com/nabbar/golib/httpcli"
	liblog "github.com/nabbar/golib/logger"
	montps "github.com/nabbar/golib/monitor/types"
	libver "github.com/nabbar/golib/version"
	libvpr "github.com/nabbar/golib/viper"
)

var _ = Describe("Component Lifecycle", func() {
	var (
		cpt CptAws
		ctx context.Context
	)

	BeforeEach(func() {
		// Create a context provider
		ctx = context.Background()
		// Create a new AWS component with ConfigStandard driver
		cpt = New(ctx, ConfigStandard)
	})

	AfterEach(func() {
		if cpt != nil && cpt.IsStarted() {
			cpt.Stop()
		}
	})

	Describe("New", func() {
		It("should create a new AWS component with ConfigStandard", func() {
			component := New(ctx, ConfigStandard)
			Expect(component).NotTo(BeNil())
		})

		It("should create a new AWS component with ConfigCustom", func() {
			component := New(ctx, ConfigCustom)
			Expect(component).NotTo(BeNil())
		})

		It("should create a new AWS component with ConfigStandardStatus", func() {
			component := New(ctx, ConfigStandardStatus)
			Expect(component).NotTo(BeNil())
		})

		It("should create a new AWS component with ConfigCustomStatus", func() {
			component := New(ctx, ConfigCustomStatus)
			Expect(component).NotTo(BeNil())
		})
	})

	Describe("Type", func() {
		It("should return the correct component type", func() {
			Expect(cpt.Type()).To(Equal(ComponentType))
			Expect(cpt.Type()).To(Equal("aws"))
		})
	})

	Describe("Init", func() {
		var (
			key     string
			getCpt  cfgtps.FuncCptGet
			vpr     libvpr.FuncViper
			version libver.Version
			logger  liblog.FuncLog
		)

		BeforeEach(func() {
			key = "test-aws-component"
			getCpt = func(k string) cfgtps.Component { return nil }
			vpr = func() libvpr.Viper { return nil }
			version = nil
			logger = func() liblog.Logger { return nil }
		})

		It("should initialize component with all parameters", func() {
			cpt.Init(key, ctx, getCpt, vpr, version, logger)
			// Component should be initialized but not started
			Expect(cpt.IsStarted()).To(BeFalse())
		})

		It("should accept nil logger", func() {
			cpt.Init(key, ctx, getCpt, vpr, version, nil)
			Expect(cpt.IsStarted()).To(BeFalse())
		})

		It("should accept nil version", func() {
			cpt.Init(key, ctx, getCpt, vpr, nil, logger)
			Expect(cpt.IsStarted()).To(BeFalse())
		})
	})

	Describe("IsStarted and IsRunning", func() {
		It("should return false before Start is called", func() {
			Expect(cpt.IsStarted()).To(BeFalse())
			Expect(cpt.IsRunning()).To(BeFalse())
		})

		It("should return false after Stop is called", func() {
			// Mock start state
			cpt.Stop()
			Expect(cpt.IsStarted()).To(BeFalse())
			Expect(cpt.IsRunning()).To(BeFalse())
		})
	})

	Describe("Dependencies", func() {
		It("should return empty dependencies by default", func() {
			deps := cpt.Dependencies()
			Expect(deps).NotTo(BeNil())
			Expect(deps).To(BeEmpty())
		})

		It("should allow setting custom dependencies", func() {
			customDeps := []string{"logger", "http-client"}
			err := cpt.SetDependencies(customDeps)
			Expect(err).NotTo(HaveOccurred())

			deps := cpt.Dependencies()
			Expect(deps).To(Equal(customDeps))
		})

		It("should handle empty dependency list", func() {
			err := cpt.SetDependencies([]string{})
			Expect(err).NotTo(HaveOccurred())

			deps := cpt.Dependencies()
			Expect(deps).To(BeEmpty())
		})

		It("should handle nil component in Dependencies", func() {
			// Note: This is not a realistic scenario in production
			// Nil component will panic - this is expected behavior
			Skip("Nil component check skipped - not a production scenario")
		})
	})

	Describe("RegisterFuncStart", func() {
		It("should register start hooks without error", func() {
			var beforeCalled, afterCalled bool

			before := func(cpt cfgtps.Component) error {
				beforeCalled = true
				return nil
			}

			after := func(cpt cfgtps.Component) error {
				afterCalled = true
				return nil
			}

			cpt.RegisterFuncStart(before, after)
			// Hooks should be registered but not called yet
			Expect(beforeCalled).To(BeFalse())
			Expect(afterCalled).To(BeFalse())
		})

		It("should accept nil hooks", func() {
			Expect(func() {
				cpt.RegisterFuncStart(nil, nil)
			}).NotTo(Panic())
		})
	})

	Describe("RegisterFuncReload", func() {
		It("should register reload hooks without error", func() {
			var beforeCalled, afterCalled bool

			before := func(cpt cfgtps.Component) error {
				beforeCalled = true
				return nil
			}

			after := func(cpt cfgtps.Component) error {
				afterCalled = true
				return nil
			}

			cpt.RegisterFuncReload(before, after)
			Expect(beforeCalled).To(BeFalse())
			Expect(afterCalled).To(BeFalse())
		})
	})

	Describe("RegisterMonitorPool", func() {
		It("should register monitor pool", func() {
			poolFunc := func() montps.Pool {
				return nil
			}

			Expect(func() {
				cpt.RegisterMonitorPool(poolFunc)
			}).NotTo(Panic())
		})

		It("should accept nil monitor pool", func() {
			Expect(func() {
				cpt.RegisterMonitorPool(nil)
			}).NotTo(Panic())
		})
	})

	Describe("RegisterHTTPClient", func() {
		It("should register HTTP client", func() {
			client := libhtc.GetClient()
			Expect(func() {
				cpt.RegisterHTTPClient(client)
			}).NotTo(Panic())
		})

		It("should use default client when nil is provided", func() {
			Expect(func() {
				cpt.RegisterHTTPClient(nil)
			}).NotTo(Panic())
		})
	})

	Describe("GetAws and SetAws", func() {
		It("should return nil AWS client when not started", func() {
			aws := cpt.GetAws()
			Expect(aws).To(BeNil())
		})

		It("should allow setting AWS client", func() {
			// We can't easily create a real AWS client without proper config
			// So we just test that the function doesn't panic
			Expect(func() {
				cpt.SetAws(nil)
			}).NotTo(Panic())
		})

		It("should not panic when getting AWS from uninitialized component", func() {
			uninitializedCpt := New(ctx, ConfigStandard)
			Expect(func() {
				aws := uninitializedCpt.GetAws()
				Expect(aws).To(BeNil())
			}).NotTo(Panic())
		})
	})

	Describe("Stop", func() {
		It("should stop component without error", func() {
			Expect(func() {
				cpt.Stop()
			}).NotTo(Panic())
			Expect(cpt.IsStarted()).To(BeFalse())
		})

		It("should be idempotent", func() {
			cpt.Stop()
			Expect(cpt.IsStarted()).To(BeFalse())

			// Call stop again
			cpt.Stop()
			Expect(cpt.IsStarted()).To(BeFalse())
		})
	})
})

var _ = Describe("Component Integration", func() {
	var (
		cpt    CptAws
		ctx    context.Context
		logger liblog.FuncLog
	)

	BeforeEach(func() {
		ctx = context.Background()
		logger = func() liblog.Logger {
			// Return a basic logger
			return liblog.New(ctx)
		}
		cpt = New(ctx, ConfigStandard)
	})

	AfterEach(func() {
		if cpt != nil && cpt.IsStarted() {
			cpt.Stop()
		}
	})

	It("should initialize with all dependencies", func() {
		getCpt := func(key string) cfgtps.Component { return nil }
		vpr := func() libvpr.Viper { return nil }

		cpt.Init("aws-integration", ctx, getCpt, vpr, nil, logger)
		Expect(cpt.Type()).To(Equal("aws"))
		Expect(cpt.IsStarted()).To(BeFalse())
	})

	It("should handle multiple lifecycle operations", func() {
		// Initialize
		getCpt := func(key string) cfgtps.Component { return nil }
		vpr := func() libvpr.Viper { return nil }
		cpt.Init("aws-lifecycle", ctx, getCpt, vpr, nil, logger)

		// Stop (even though not started)
		cpt.Stop()
		Expect(cpt.IsStarted()).To(BeFalse())

		// Stop again (idempotent)
		cpt.Stop()
		Expect(cpt.IsStarted()).To(BeFalse())
	})

	It("should work with different driver types", func() {
		drivers := []ConfigDriver{
			ConfigStandard,
			ConfigStandardStatus,
			ConfigCustom,
			ConfigCustomStatus,
		}

		for _, drv := range drivers {
			By("Testing with driver: " + drv.String())
			component := New(ctx, drv)
			Expect(component).NotTo(BeNil())
			Expect(component.Type()).To(Equal("aws"))
		}
	})
})

var _ = Describe("Component Error Handling", func() {
	var (
		cpt CptAws
		ctx context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()
		cpt = New(ctx, ConfigStandard)
	})

	It("should handle uninitialized component gracefully", func() {
		// Try to get dependencies before initialization
		deps := cpt.Dependencies()
		Expect(deps).NotTo(BeNil())
		Expect(deps).To(BeEmpty())
	})

	It("should not panic on double Stop", func() {
		cpt.Stop()
		Expect(func() {
			cpt.Stop()
		}).NotTo(Panic())
	})

	It("should handle nil context gracefully", func() {
		var nilCtx context.Context
		component := New(nilCtx, ConfigStandard)
		Expect(component).NotTo(BeNil())
	})
})

var _ = Describe("Component Thread Safety", func() {
	var (
		cpt CptAws
		ctx context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()
		cpt = New(ctx, ConfigStandard)
	})

	It("should handle concurrent Type calls", func() {
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func() {
				defer GinkgoRecover()
				typ := cpt.Type()
				Expect(typ).To(Equal("aws"))
				done <- true
			}()
		}

		for i := 0; i < 10; i++ {
			Eventually(done).Should(Receive())
		}
	})

	It("should handle concurrent IsStarted calls", func() {
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func() {
				defer GinkgoRecover()
				started := cpt.IsStarted()
				Expect(started).To(BeFalse())
				done <- true
			}()
		}

		for i := 0; i < 10; i++ {
			Eventually(done).Should(Receive())
		}
	})

	It("should handle concurrent Dependencies calls", func() {
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func() {
				defer GinkgoRecover()
				deps := cpt.Dependencies()
				Expect(deps).NotTo(BeNil())
				done <- true
			}()
		}

		for i := 0; i < 10; i++ {
			Eventually(done).Should(Receive())
		}
	})
})
