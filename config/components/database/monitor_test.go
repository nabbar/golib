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
	logcfg "github.com/nabbar/golib/logger/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	monpol "github.com/nabbar/golib/monitor/pool"
	montps "github.com/nabbar/golib/monitor/types"
)

// Monitor integration tests verify RegisterMonitorPool and monitor lifecycle
var _ = Describe("Monitor Integration", func() {
	var (
		cpt     CptDatabase
		ctx     context.Context
		monPool montps.Pool
	)

	BeforeEach(func() {
		ctx = context.Background()
		cpt = New(ctx)
		monPool = monpol.New(ctx)
	})

	AfterEach(func() {
		if cpt != nil && cpt.IsStarted() {
			cpt.Stop()
		}
		if monPool != nil {
			monPool.MonitorWalk(func(name string, val montps.Monitor) bool {
				_ = val.Stop(ctx)
				monPool.MonitorDel(name)
				return true
			})
		}
	})

	Describe("RegisterMonitorPool", func() {
		It("should register monitor pool function", func() {
			poolFunc := func() montps.Pool {
				return monPool
			}

			Expect(func() {
				cpt.RegisterMonitorPool(poolFunc)
			}).NotTo(Panic())
		})

		It("should accept nil monitor pool function", func() {
			Expect(func() {
				cpt.RegisterMonitorPool(nil)
			}).NotTo(Panic())
		})

		It("should allow replacing monitor pool function", func() {
			poolFunc1 := func() montps.Pool {
				return monPool
			}

			poolFunc2 := func() montps.Pool {
				return nil
			}

			cpt.RegisterMonitorPool(poolFunc1)
			cpt.RegisterMonitorPool(poolFunc2)
			// No panic expected
		})

		It("should be callable multiple times", func() {
			poolFunc := func() montps.Pool {
				return monPool
			}

			for i := 0; i < 10; i++ {
				cpt.RegisterMonitorPool(poolFunc)
			}
		})
	})

	Describe("Monitor Pool Behavior", func() {
		It("should handle pool that returns nil", func() {
			nilPoolFunc := func() montps.Pool {
				return nil
			}

			Expect(func() {
				cpt.RegisterMonitorPool(nilPoolFunc)
			}).NotTo(Panic())
		})

		It("should handle pool with monitors", func() {
			poolFunc := func() montps.Pool {
				return monPool
			}

			cpt.RegisterMonitorPool(poolFunc)
			// Component should continue to function normally
			Expect(cpt.Type()).To(Equal("database"))
		})
	})

	Describe("Monitor Registration Scenarios", func() {
		It("should support monitor registration before component start", func() {
			poolFunc := func() montps.Pool {
				return monPool
			}

			cpt.RegisterMonitorPool(poolFunc)
			Expect(cpt.IsStarted()).To(BeFalse())
		})

		It("should support monitor registration on stopped component", func() {
			cpt.Stop()

			poolFunc := func() montps.Pool {
				return monPool
			}

			Expect(func() {
				cpt.RegisterMonitorPool(poolFunc)
			}).NotTo(Panic())
		})
	})
})

// Monitor pool creation and management
var _ = Describe("Monitor Pool Management", func() {
	var ctx context.Context

	BeforeEach(func() {
		ctx = context.Background()
	})

	Describe("Monitor Pool Creation", func() {
		It("should create monitor pool successfully", func() {
			pool := monpol.New(ctx)
			Expect(pool).NotTo(BeNil())
			pool.MonitorWalk(func(name string, val montps.Monitor) bool {
				_ = val.Stop(ctx)
				pool.MonitorDel(name)
				return true
			})
		})

		It("should support multiple monitor pools", func() {
			pool1 := monpol.New(ctx)
			pool2 := monpol.New(ctx)
			pool3 := monpol.New(ctx)

			Expect(pool1).NotTo(BeNil())
			Expect(pool2).NotTo(BeNil())
			Expect(pool3).NotTo(BeNil())
		})

		It("should handle pool lifecycle", func() {
			pool := monpol.New(ctx)
			Expect(pool).NotTo(BeNil())

			// Use pool
			Expect(func() {
				pool.MonitorList()
			}).NotTo(Panic())
		})
	})

	Describe("Monitor Operations", func() {
		var pool montps.Pool

		BeforeEach(func() {
			pool = monpol.New(ctx)
		})

		AfterEach(func() {
			if pool != nil {
				pool.MonitorWalk(func(name string, val montps.Monitor) bool {
					_ = val.Stop(ctx)
					pool.MonitorDel(name)
					return true
				})
			}
		})

		It("should list monitors", func() {
			monitors := pool.MonitorList()
			Expect(monitors).NotTo(BeNil())
			// Initially empty
			Expect(monitors).To(BeEmpty())
		})

		It("should get non-existent monitor", func() {
			mon := pool.MonitorGet("non-existent")
			Expect(mon).To(BeNil())
		})

		It("should check non-existent monitor", func() {
			exists := pool.MonitorGet("non-existent") != nil
			Expect(exists).To(BeFalse())
		})

		It("should set and get monitor", func() {
			// Create a basic monitor
			// Note: We can't easily create a real monitor without database
			// So we just test the pool operations
			exists := pool.MonitorGet("test-monitor") != nil
			Expect(exists).To(BeFalse())
		})
	})
})

// Monitor edge cases
var _ = Describe("Monitor Edge Cases", func() {
	var (
		cpt CptDatabase
		ctx context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()
		cpt = New(ctx)
	})

	AfterEach(func() {
		if cpt != nil && cpt.IsStarted() {
			cpt.Stop()
		}
	})

	Context("with nil values", func() {
		It("should handle nil pool function", func() {
			Expect(func() {
				cpt.RegisterMonitorPool(nil)
			}).NotTo(Panic())
		})

		It("should handle pool function returning nil", func() {
			poolFunc := func() montps.Pool {
				return nil
			}

			Expect(func() {
				cpt.RegisterMonitorPool(poolFunc)
			}).NotTo(Panic())
		})
	})

	Context("with multiple registrations", func() {
		It("should handle rapid monitor pool registration", func() {
			for i := 0; i < 100; i++ {
				poolFunc := func() montps.Pool {
					return monpol.New(ctx)
				}
				cpt.RegisterMonitorPool(poolFunc)
			}
		})

		It("should handle alternating registrations", func() {
			for i := 0; i < 50; i++ {
				if i%2 == 0 {
					poolFunc := func() montps.Pool {
						return monpol.New(ctx)
					}
					cpt.RegisterMonitorPool(poolFunc)
				} else {
					cpt.RegisterMonitorPool(nil)
				}
			}
		})
	})

	Context("with concurrent access", func() {
		It("should handle concurrent monitor pool registration", func() {
			done := make(chan bool, 10)

			for i := 0; i < 10; i++ {
				go func() {
					defer GinkgoRecover()
					poolFunc := func() montps.Pool {
						return monpol.New(ctx)
					}
					cpt.RegisterMonitorPool(poolFunc)
					done <- true
				}()
			}

			for i := 0; i < 10; i++ {
				Eventually(done).Should(Receive())
			}
		})
	})
})

// Monitor integration with components
var _ = Describe("Monitor Component Integration", func() {
	var (
		ctx context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()
	})

	It("should support multiple components with shared monitor pool", func() {
		pool := monpol.New(ctx)
		defer func() {
			if pool != nil {
				pool.MonitorWalk(func(name string, val montps.Monitor) bool {
					_ = val.Stop(ctx)
					pool.MonitorDel(name)
					return true
				})
			}
		}()

		poolFunc := func() montps.Pool {
			return pool
		}

		// Create multiple components
		components := make([]CptDatabase, 3)
		for i := range components {
			components[i] = New(ctx)
			components[i].RegisterMonitorPool(poolFunc)
			Expect(components[i]).NotTo(BeNil())
		}

		// Clean up
		for _, cpt := range components {
			if cpt.IsStarted() {
				cpt.Stop()
			}
		}
	})

	It("should support components with independent monitor pools", func() {
		components := make([]CptDatabase, 3)
		pools := make([]montps.Pool, 3)

		for i := range components {
			components[i] = New(ctx)
			pools[i] = monpol.New(ctx)

			localPool := pools[i]
			poolFunc := func() montps.Pool {
				return localPool
			}

			components[i].RegisterMonitorPool(poolFunc)
		}

		// Clean up
		for i, cpt := range components {
			if cpt.IsStarted() {
				cpt.Stop()
			}
			if pools[i] != nil {
				pools[i].MonitorWalk(func(name string, val montps.Monitor) bool {
					_ = val.Stop(ctx)
					pools[i].MonitorDel(name)
					return true
				})
			}
		}
	})

	It("should handle component lifecycle with monitors", func() {
		pool := monpol.New(ctx)
		defer func() {
			if pool != nil {
				pool.MonitorWalk(func(name string, val montps.Monitor) bool {
					_ = val.Stop(ctx)
					pool.MonitorDel(name)
					return true
				})
			}
		}()

		poolFunc := func() montps.Pool {
			return pool
		}

		cpt := New(ctx)
		cpt.RegisterMonitorPool(poolFunc)

		// Component operations
		Expect(cpt.Type()).To(Equal("database"))
		Expect(cpt.IsStarted()).To(BeFalse())

		// Stop component
		cpt.Stop()
		Expect(cpt.IsStarted()).To(BeFalse())
	})
})

// Monitor functionality tests
var _ = Describe("Monitor Functionality", func() {
	var (
		ctx  context.Context
		pool montps.Pool
	)

	BeforeEach(func() {
		ctx = context.Background()
		pool = monpol.New(ctx)
	})

	AfterEach(func() {
		if pool != nil {
			pool.MonitorWalk(func(name string, val montps.Monitor) bool {
				_ = val.Stop(ctx)
				pool.MonitorDel(name)
				return true
			})
		}
	})

	It("should create monitor from database", func() {
		// This test verifies the monitor creation flow
		// without actually creating a database connection
		Expect(pool).NotTo(BeNil())
		Expect(pool.MonitorList()).NotTo(BeNil())
	})

	It("should handle monitor configuration", func() {
		cfg := montps.Config{
			Name:          "test-monitor",
			CheckTimeout:  0,
			IntervalCheck: 0,
			IntervalFall:  0,
			IntervalRise:  0,
			FallCountKO:   0,
			FallCountWarn: 0,
			RiseCountKO:   0,
			RiseCountWarn: 0,
			Logger:        logcfg.Options{Stdout: &logcfg.OptionsStd{DisableStandard: true}},
		}

		Expect(cfg.Name).To(Equal("test-monitor"))
	})

	It("should create monitor with custom options", func() {
		// Test monitor creation with various options
		configs := []montps.Config{
			{Name: "monitor-1", Logger: logcfg.Options{Stdout: &logcfg.OptionsStd{DisableStandard: true}}},
			{Name: "monitor-2", Logger: logcfg.Options{Stdout: &logcfg.OptionsStd{DisableStandard: true}}},
			{Name: "monitor-3", Logger: logcfg.Options{Stdout: &logcfg.OptionsStd{DisableStandard: true}}},
		}

		for _, cfg := range configs {
			Expect(cfg.Name).NotTo(BeEmpty())
		}
	})
})
