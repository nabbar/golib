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

	. "github.com/nabbar/golib/config/components/database"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	cfgtps "github.com/nabbar/golib/config/types"
	libdur "github.com/nabbar/golib/duration"
	liblog "github.com/nabbar/golib/logger"
	montps "github.com/nabbar/golib/monitor/types"
	libver "github.com/nabbar/golib/version"
	libvpr "github.com/nabbar/golib/viper"
)

// Component lifecycle tests verify component initialization, state management,
// and lifecycle operations
var _ = Describe("Component Lifecycle", func() {
	var (
		cpt CptDatabase
		ctx context.Context
	)

	BeforeEach(func() {
		// Create a context provider
		ctx = context.Background()
		// Create a new Database component
		cpt = New(ctx)
	})

	AfterEach(func() {
		if cpt != nil && cpt.IsStarted() {
			cpt.Stop()
		}
	})

	Describe("New", func() {
		It("should create a new Database component", func() {
			component := New(ctx)
			Expect(component).NotTo(BeNil())
		})

		It("should initialize with nil context gracefully", func() {
			var nilCtx context.Context
			component := New(nilCtx)
			Expect(component).NotTo(BeNil())
		})
	})

	Describe("Type", func() {
		It("should return the correct component type", func() {
			Expect(cpt.Type()).To(Equal(ComponentType))
			Expect(cpt.Type()).To(Equal("database"))
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
			key = "test-database-component"
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

		It("should accept empty key", func() {
			cpt.Init("", ctx, getCpt, vpr, version, logger)
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

		It("should be consistent between IsStarted and IsRunning", func() {
			// IsRunning should be false if IsStarted is false
			if !cpt.IsStarted() {
				Expect(cpt.IsRunning()).To(BeFalse())
			}
		})
	})

	Describe("Dependencies", func() {
		It("should return empty dependencies by default", func() {
			deps := cpt.Dependencies()
			Expect(deps).NotTo(BeNil())
			Expect(deps).To(BeEmpty())
		})

		It("should allow setting custom dependencies", func() {
			customDeps := []string{"logger", "monitor"}
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

		It("should handle single dependency", func() {
			err := cpt.SetDependencies([]string{"logger"})
			Expect(err).NotTo(HaveOccurred())

			deps := cpt.Dependencies()
			Expect(deps).To(HaveLen(1))
			Expect(deps[0]).To(Equal("logger"))
		})

		It("should allow overwriting dependencies", func() {
			err := cpt.SetDependencies([]string{"logger"})
			Expect(err).NotTo(HaveOccurred())

			err = cpt.SetDependencies([]string{"monitor", "config"})
			Expect(err).NotTo(HaveOccurred())

			deps := cpt.Dependencies()
			Expect(deps).To(HaveLen(2))
			Expect(deps).To(ConsistOf("monitor", "config"))
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

		It("should accept only before hook", func() {
			before := func(cpt cfgtps.Component) error {
				return nil
			}
			Expect(func() {
				cpt.RegisterFuncStart(before, nil)
			}).NotTo(Panic())
		})

		It("should accept only after hook", func() {
			after := func(cpt cfgtps.Component) error {
				return nil
			}
			Expect(func() {
				cpt.RegisterFuncStart(nil, after)
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

		It("should accept nil hooks", func() {
			Expect(func() {
				cpt.RegisterFuncReload(nil, nil)
			}).NotTo(Panic())
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

	Describe("GetDatabase and SetDatabase", func() {
		It("should return nil database when not started", func() {
			db := cpt.GetDatabase()
			Expect(db).To(BeNil())
		})

		It("should allow setting database", func() {
			// We can't easily create a real database without proper config
			// So we just test that the function doesn't panic
			Expect(func() {
				cpt.SetDatabase(nil)
			}).NotTo(Panic())
		})

		It("should return nil after setting nil database", func() {
			cpt.SetDatabase(nil)
			db := cpt.GetDatabase()
			Expect(db).To(BeNil())
		})

		It("should not panic when getting database from uninitialized component", func() {
			uninitializedCpt := New(ctx)
			Expect(func() {
				db := uninitializedCpt.GetDatabase()
				Expect(db).To(BeNil())
			}).NotTo(Panic())
		})
	})

	Describe("SetLogOptions", func() {
		It("should set log options without error", func() {
			Expect(func() {
				cpt.SetLogOptions(true, 100)
			}).NotTo(Panic())
		})

		It("should accept false for ignoreRecordNotFoundError", func() {
			Expect(func() {
				cpt.SetLogOptions(false, 200)
			}).NotTo(Panic())
		})

		It("should accept zero slowThreshold", func() {
			Expect(func() {
				cpt.SetLogOptions(true, 0)
			}).NotTo(Panic())
		})

		It("should accept negative slowThreshold", func() {
			Expect(func() {
				cpt.SetLogOptions(true, -100)
			}).NotTo(Panic())
		})

		It("should be callable multiple times", func() {
			cpt.SetLogOptions(true, 100)
			cpt.SetLogOptions(false, 200)
			cpt.SetLogOptions(true, 0)
			// No assertion needed - just verify no panic
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

// Component integration tests
var _ = Describe("Component Integration", func() {
	var (
		cpt    CptDatabase
		ctx    context.Context
		logger liblog.FuncLog
	)

	BeforeEach(func() {
		ctx = context.Background()
		logger = func() liblog.Logger {
			// Return a basic logger
			return liblog.New(ctx)
		}
		cpt = New(ctx)
	})

	AfterEach(func() {
		if cpt != nil && cpt.IsStarted() {
			cpt.Stop()
		}
	})

	It("should initialize with all dependencies", func() {
		getCpt := func(key string) cfgtps.Component { return nil }
		vpr := func() libvpr.Viper { return nil }

		cpt.Init("database-integration", ctx, getCpt, vpr, nil, logger)
		Expect(cpt.Type()).To(Equal("database"))
		Expect(cpt.IsStarted()).To(BeFalse())
	})

	It("should handle multiple lifecycle operations", func() {
		// Initialize
		getCpt := func(key string) cfgtps.Component { return nil }
		vpr := func() libvpr.Viper { return nil }
		cpt.Init("database-lifecycle", ctx, getCpt, vpr, nil, logger)

		// Stop (even though not started)
		cpt.Stop()
		Expect(cpt.IsStarted()).To(BeFalse())

		// Stop again (idempotent)
		cpt.Stop()
		Expect(cpt.IsStarted()).To(BeFalse())
	})
})

// Component error handling tests
var _ = Describe("Component Error Handling", func() {
	var (
		cpt CptDatabase
		ctx context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()
		cpt = New(ctx)
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
		component := New(nilCtx)
		Expect(component).NotTo(BeNil())
	})
})

// Component thread safety tests
var _ = Describe("Component Thread Safety", func() {
	var (
		cpt CptDatabase
		ctx context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()
		cpt = New(ctx)
	})

	It("should handle concurrent Type calls", func() {
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func() {
				defer GinkgoRecover()
				typ := cpt.Type()
				Expect(typ).To(Equal("database"))
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

	It("should handle concurrent GetDatabase calls", func() {
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func() {
				defer GinkgoRecover()
				db := cpt.GetDatabase()
				Expect(db).To(BeNil())
				done <- true
			}()
		}

		for i := 0; i < 10; i++ {
			Eventually(done).Should(Receive())
		}
	})

	It("should handle concurrent SetLogOptions calls", func() {
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func(idx int) {
				defer GinkgoRecover()
				cpt.SetLogOptions(idx%2 == 0, libdur.Duration(idx*100))
				done <- true
			}(i)
		}

		for i := 0; i < 10; i++ {
			Eventually(done).Should(Receive())
		}
	})
})
